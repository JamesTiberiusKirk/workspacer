package workspacer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/JamesTiberiusKirk/workspacer/config"
	"github.com/JamesTiberiusKirk/workspacer/log"
	"github.com/JamesTiberiusKirk/workspacer/ui/list"
	"github.com/JamesTiberiusKirk/workspacer/ui/spinner"
	"github.com/JamesTiberiusKirk/workspacer/util"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	branchStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("12")) // Blue
	changesStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // Yellow
	changesCleanStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // Green
)

// sisterGitInfo holds git information for a sister repository
type sisterGitInfo struct {
	label   string
	branch  string
	changes int
}

// repoGitInfo holds git information for a repository
type repoGitInfo struct {
	name         string
	branch       string
	changesCount int
	sisters      []sisterGitInfo
	hasError     bool
}

// loadGitInfoForRepo loads git information for a single repository
func loadGitInfoForRepo(wc config.WorkspaceConfig, repoName string, gitInfoChan chan<- repoGitInfo, wg *sync.WaitGroup) {
	defer wg.Done()

	info := repoGitInfo{
		name: repoName,
	}

	// Get git branch
	info.branch = util.GetGitBranch(wc, repoName)
	if info.branch == "" {
		info.hasError = true
	}

	// Get uncommitted changes count
	info.changesCount = util.GetUncommittedChangesCount(wc, repoName)

	// Check for sister repos
	sisterRepos := util.GetSisterReposForProject(wc, repoName)
	for _, sr := range sisterRepos {
		if util.DoesProjectExist(wc, sr.Name) {
			sisterInfo := sisterGitInfo{
				label:   sr.Label,
				branch:  util.GetGitBranch(wc, sr.Name),
				changes: util.GetUncommittedChangesCount(wc, sr.Name),
			}
			info.sisters = append(info.sisters, sisterInfo)
		}
	}

	gitInfoChan <- info
}

func ChooseFromOpenWorkspaceProjectsAndSwitch(workspace string, workspaceConfig config.WorkspaceConfig, sessionPresets map[string]config.SessionConfig) {
	openProjects := util.GetOpenProjectsByWorkspace(workspace)
	if len(openProjects) == 0 {
		log.Info("No open projects in workspace %s", workspace)
	}

	lists := []list.Item{}
	for _, p := range openProjects {
		display := strings.TrimPrefix(p, workspace+"-")
		lists = append(lists, list.Item{Display: display, Value: p})
	}
	item, found, err := list.NewList("Open projects in workspace: "+workspaceConfig.Name, lists, "", nil)
	if err != nil {
		panic(err)
	}
	if !found {
		// Assume that the user just existed the list
		// panic("not found")
		os.Exit(0)
	}

	// Track usage
	if workspaceConfig.EnableUsageTracking && workspaceConfig.EnableCache {
		cache := LoadCache(workspaceConfig)
		windowSize := workspaceConfig.RecentAccessWindow
		if windowSize == 0 {
			windowSize = 50 // Default
		}
		// Extract project name from display (removes workspace prefix)
		projectName := strings.TrimPrefix(item.Value, workspace+"-")
		cache.RecordAccess(projectName, windowSize)
		if err := SaveCache(workspaceConfig, cache); err != nil {
			log.Error("Failed to save usage tracking: %s", err.Error())
		}
	}

	StartOrSwitchToSession(workspaceConfig, sessionPresets, item.Value)

	return
}

func removeRepoFromArray(repos []string, name string) []string {
	res := []string{}
	for _, r := range repos {
		if r == name {
			continue
		}
		res = append(res, r)
	}

	return res
}

func buildWorkspaceItems(workspace string, wc config.WorkspaceConfig, extraOptions []list.Item) ([]list.Item, string, bool) {
	cache := LoadCache(wc)

	openProjects := util.GetOpenProjectsByWorkspace(workspace)
	path := util.GetWorkspacePath(wc)
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Error("Failed to read workspace directory: %s", err.Error())
		return []list.Item{}, "no cache", false
	}

	// Collect git repos that need info loading
	var gitRepos []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if util.IsSisterRepo(wc, e.Name()) {
			continue
		}
		if util.HasGitSubfolder(filepath.Join(path, e.Name())) {
			gitRepos = append(gitRepos, e.Name())
		}
	}

	// Load git info (from cache or fresh fetch)
	gitInfoMap := make(map[string]repoGitInfo)
	if wc.EnableGitInfo && len(gitRepos) > 0 {
		useCache := wc.EnableCache
		if useCache {
			for _, repoName := range gitRepos {
				if projectCache, exists := cache.GetProjectCache(repoName); exists {
					var sisters []sisterGitInfo
					for label, sc := range projectCache.SisterRepos {
						sisters = append(sisters, sisterGitInfo{
							label:   label,
							branch:  sc.Branch,
							changes: sc.Changes,
						})
					}
					info := repoGitInfo{
						name:         repoName,
						branch:       projectCache.GitBranch,
						changesCount: projectCache.GitChanges,
						sisters:      sisters,
					}
					gitInfoMap[repoName] = info
				}
			}
		}

		if !useCache || len(gitInfoMap) < len(gitRepos) {
			var wg sync.WaitGroup
			gitInfoChan := make(chan repoGitInfo, len(gitRepos))

			for _, repoName := range gitRepos {
				if _, exists := gitInfoMap[repoName]; exists && useCache {
					continue
				}
				wg.Add(1)
				go loadGitInfoForRepo(wc, repoName, gitInfoChan, &wg)
			}

			go func() {
				wg.Wait()
				close(gitInfoChan)
			}()

			for info := range gitInfoChan {
				gitInfoMap[info.name] = info
				cache.UpdateGitInfo(info.name, info)
			}
		}
	}

	// Load remote repos
	var remoteRepos []string
	remoteError := false
	if wc.EnableRemoteRepos {
		cacheValid := wc.EnableCache && len(cache.GithubRepos) > 0 && cache.GithubReposShowArchived == wc.ShowArchivedRepos
		if cacheValid {
			remoteRepos = cache.GithubRepos
		} else {
			repos, err := GetRepoNames(wc)
			if err != nil {
				log.Error("Failed to fetch remote repos: %s", err.Error())
				remoteError = true
			} else {
				remoteRepos = repos
				cache.UpdateGithubRepos(repos, wc.ShowArchivedRepos)
			}
		}
	}

	// Build list items
	folders := []list.Item{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if util.IsSisterRepo(wc, e.Name()) {
			continue
		}

		item := list.Item{
			Display:  e.Name(),
			Value:    "folder:" + e.Name(),
			Subtitle: "Folder",
		}

		if info, hasGitInfo := gitInfoMap[e.Name()]; hasGitInfo {
			subtitle := "Service: "
			if info.branch != "" {
				subtitle += branchStyle.Render(info.branch)
				if info.changesCount > 0 {
					subtitle += " " + changesStyle.Render(fmt.Sprintf("(%d)", info.changesCount))
				} else {
					subtitle += " " + changesCleanStyle.Render("✓")
				}
			} else if info.hasError {
				subtitle += "(error loading git info)"
			}

			for _, sister := range info.sisters {
				item.Display = item.Display + " +" + sister.label
				subtitle += " | " + sister.label + ": "
				if sister.branch != "" {
					subtitle += branchStyle.Render(sister.branch)
					if sister.changes > 0 {
						subtitle += " " + changesStyle.Render(fmt.Sprintf("(%d)", sister.changes))
					} else {
						subtitle += " " + changesCleanStyle.Render("✓")
					}
				}
			}

			item.Subtitle = subtitle
			remoteRepos = removeRepoFromArray(remoteRepos, e.Name())
		}

		if util.Contains(openProjects, e.Name()) {
			item.Display = item.Display + " (Active)"
			item.IsActive = true
		}

		folders = append(folders, item)
	}

	// Sort
	if wc.ActiveProjectsFirst {
		sort.Slice(folders, func(i, j int) bool {
			iActive := folders[i].IsActive
			jActive := folders[j].IsActive

			if iActive != jActive {
				return iActive
			}

			if wc.EnableUsageTracking && wc.EnableCache {
				iProject := strings.TrimPrefix(folders[i].Value, "folder:")
				jProject := strings.TrimPrefix(folders[j].Value, "folder:")

				iCache, iExists := cache.GetProjectCache(iProject)
				jCache, jExists := cache.GetProjectCache(jProject)

				if iExists && jExists {
					if iCache.AccessCountRecent != jCache.AccessCountRecent {
						return iCache.AccessCountRecent > jCache.AccessCountRecent
					}
				}
			}

			return folders[i].Display < folders[j].Display
		})
	}

	// Add remote repos section
	if wc.EnableRemoteRepos && !remoteError && len(remoteRepos) == 0 {
		folders = append(folders, list.Item{
			Display:  "No remote repositories found",
			Value:    "error:no-remote-repos",
			Subtitle: "No repositories found on GitHub for this workspace",
		})
	}

	for _, remoteRepo := range remoteRepos {
		folders = append(folders, list.Item{
			Display:  remoteRepo,
			Value:    "git:" + remoteRepo,
			Subtitle: "Clone From GitHub",
		})
	}

	if remoteError {
		folders = append(folders, list.Item{
			Display:  "⚠ GitHub repos unavailable",
			Value:    "error:github",
			Subtitle: "Check network connection or GITHUB_AUTH token",
		})
	}

	if len(extraOptions) > 0 {
		folders = append(folders, extraOptions...)
	}

	// Format cache status
	cacheStatus := "no cache"
	projCount := len(cache.Projects)
	gitCount := 0
	for _, p := range cache.Projects {
		if p.GitBranch != "" {
			gitCount++
		}
	}
	repoCount := len(cache.GithubRepos)

	if !cache.LastUpdated.IsZero() {
		ago := time.Since(cache.LastUpdated).Round(time.Minute)
		parts := []string{}
		if projCount > 0 {
			parts = append(parts, fmt.Sprintf("%d proj", projCount))
		}
		if gitCount > 0 {
			parts = append(parts, fmt.Sprintf("%d git", gitCount))
		}
		if repoCount > 0 {
			parts = append(parts, fmt.Sprintf("%d gh", repoCount))
		}
		var ageStr string
		if ago < time.Minute {
			ageStr = "now"
		} else if ago < time.Hour {
			ageStr = fmt.Sprintf("%dm", int(ago.Minutes()))
		} else {
			ageStr = fmt.Sprintf("%dh", int(ago.Hours()))
		}
		parts = append(parts, ageStr)
		cacheStatus = "cache: " + strings.Join(parts, " ")
	}

	// Save cache
	if err := SaveCache(wc, cache); err != nil {
		log.Error("Failed to save cache: %s", err.Error())
	}

	return folders, cacheStatus, remoteError
}

func ChoseProjectFromLocalWorkspace(workspace string, wc config.WorkspaceConfig, extraOptions []list.Item) (string, string) {
	if _, err := os.Stat(util.GetWorkspacePath(wc)); os.IsNotExist(err) {
		log.Error("workspace %s does not exist", workspace)
		return "nochoise", ""
	}

	// Show loading spinner
	spinnerModel := spinner.New("Loading workspace projects...")
	p := tea.NewProgram(spinnerModel)

	// Load data in background
	type loadResult struct {
		folders          []list.Item
		cacheLastUpdated string
	}
	resultChan := make(chan loadResult, 1)

	go func() {
		defer func() {
			if p != nil {
				p.Quit()
			}
		}()

		folders, cacheStatus, _ := buildWorkspaceItems(workspace, wc, extraOptions)
		resultChan <- loadResult{folders: folders, cacheLastUpdated: cacheStatus}
	}()

	// Run spinner while loading
	if _, err := p.Run(); err != nil {
		log.Error("Error running spinner: %s", err.Error())
	}

	// Get loaded data
	result := <-resultChan

	// Refresh callback: clear cache and reload
	refreshCache := func() ([]list.Item, string) {
		_ = ClearCache(wc)
		folders, cacheStatus, _ := buildWorkspaceItems(workspace, wc, extraOptions)
		return folders, cacheStatus
	}

	// Show list
	item, found, err := list.NewList("Select a project", result.folders, result.cacheLastUpdated, refreshCache)
	if err != nil {
		panic(err)
	}
	if !found {
		return "nochoise", ""
	}

	// Ignore error items
	if strings.HasPrefix(item.Value, "error:") {
		return "nochoise", ""
	}

	var projectType, projectName string
	if strings.HasPrefix(item.Value, "folder:") {
		projectType = "folder"
		projectName = strings.TrimPrefix(item.Value, "folder:")
	} else if strings.HasPrefix(item.Value, "git:") {
		projectType = "git"
		projectName = strings.TrimPrefix(item.Value, "git:")
	} else if strings.HasPrefix(item.Value, "root:") {
		projectType = "root"
		projectName = ""
	} else {
		return "", ""
	}

	// Track usage
	if wc.EnableUsageTracking && wc.EnableCache {
		cache := LoadCache(wc)
		windowSize := wc.RecentAccessWindow
		if windowSize == 0 {
			windowSize = 50 // Default
		}
		cache.RecordAccess(projectName, windowSize)
		if err := SaveCache(wc, cache); err != nil {
			log.Error("Failed to save usage tracking: %s", err.Error())
		}
	}

	return projectType, projectName
}
