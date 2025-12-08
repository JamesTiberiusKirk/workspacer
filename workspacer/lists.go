package workspacer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

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

// repoGitInfo holds git information for a repository
type repoGitInfo struct {
	name          string
	branch        string
	changesCount  int
	hasTenant     bool
	tenantBranch  string
	tenantChanges int
	hasError      bool
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

	// Check for tenant repo
	if wc.EnableTenantRepos && util.DoesTenantRepoExist(wc, repoName) {
		info.hasTenant = true
		tenantName := util.GetTenantRepoName(wc, repoName)
		info.tenantBranch = util.GetGitBranch(wc, tenantName)
		info.tenantChanges = util.GetUncommittedChangesCount(wc, tenantName)
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
	item, found, err := list.NewList("Open projects in workspace: "+workspaceConfig.Name, lists)
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
		folders      []list.Item
		remoteError  bool
	}
	resultChan := make(chan loadResult, 1)

	go func() {
		defer func() {
			// Quit the spinner when done loading
			if p != nil {
				p.Quit()
			}
		}()

		// Load cache
		cache := LoadCache(wc)

		openProjects := util.GetOpenProjectsByWorkspace(workspace)
		path := util.GetWorkspacePath(wc)
		entries, err := os.ReadDir(path)
		if err != nil {
			log.Error("Failed to read workspace directory: %s", err.Error())
			resultChan <- loadResult{folders: []list.Item{}}
			return
		}

		// Collect git repos that need info loading
		var gitRepos []string
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			// Skip tenant repos - they'll be shown as suffixes on service repos
			if wc.EnableTenantRepos && !util.IsServiceRepo(wc, e.Name()) {
				continue
			}
			if util.HasGitSubfolder(filepath.Join(path, e.Name())) {
				gitRepos = append(gitRepos, e.Name())
			}
		}

		// Load git info (from cache or fresh fetch)
		gitInfoMap := make(map[string]repoGitInfo)
		if wc.EnableGitInfo && len(gitRepos) > 0 {
			// Try loading from cache first
			useCache := wc.EnableCache
			if useCache {
				// Use cached data
				for _, repoName := range gitRepos {
					if projectCache, exists := cache.GetProjectCache(repoName); exists {
						info := repoGitInfo{
							name:          repoName,
							branch:        projectCache.GitBranch,
							changesCount:  projectCache.GitChanges,
							tenantBranch:  projectCache.TenantBranch,
							tenantChanges: projectCache.TenantChanges,
							hasTenant:     projectCache.TenantBranch != "",
						}
						gitInfoMap[repoName] = info
					}
				}
			}

			// Fetch fresh data if cache disabled or missing
			if !useCache || len(gitInfoMap) < len(gitRepos) {
				var wg sync.WaitGroup
				gitInfoChan := make(chan repoGitInfo, len(gitRepos))

				for _, repoName := range gitRepos {
					// Skip if already in cache
					if _, exists := gitInfoMap[repoName]; exists && useCache {
						continue
					}
					wg.Add(1)
					go loadGitInfoForRepo(wc, repoName, gitInfoChan, &wg)
				}

				// Close channel when all goroutines complete
				go func() {
					wg.Wait()
					close(gitInfoChan)
				}()

				// Collect results and update cache
				for info := range gitInfoChan {
					gitInfoMap[info.name] = info
					cache.UpdateGitInfo(info.name, info)
				}
			}
		}

		// Load remote repos (from cache or fresh fetch)
		var remoteRepos []string
		remoteError := false
		if wc.EnableRemoteRepos {
			// Try cache first
			if wc.EnableCache && len(cache.GithubRepos) > 0 {
				remoteRepos = cache.GithubRepos
			} else {
				// Fetch fresh data
				repos, err := GetRepoNames(wc)
				if err != nil {
					log.Error("Failed to fetch remote repos: %s", err.Error())
					remoteError = true
				} else {
					remoteRepos = repos
					cache.UpdateGithubRepos(repos)
				}
			}
		}

		// Build list items
		folders := []list.Item{}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}

			// Skip tenant repos
			if wc.EnableTenantRepos && !util.IsServiceRepo(wc, e.Name()) {
				continue
			}

			item := list.Item{
				Display: e.Name(),
				Value:   "folder:" + e.Name(),
				Subtitle: "Folder",
			}

			// Add git info if available
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

				// Add tenant info
				if info.hasTenant {
					item.Display = item.Display + " + tenant"
					subtitle += " | Tenant: "
					if info.tenantBranch != "" {
						subtitle += branchStyle.Render(info.tenantBranch)
						if info.tenantChanges > 0 {
							subtitle += " " + changesStyle.Render(fmt.Sprintf("(%d)", info.tenantChanges))
						} else {
							subtitle += " " + changesCleanStyle.Render("✓")
						}
					}
				}

				item.Subtitle = subtitle
				remoteRepos = removeRepoFromArray(remoteRepos, e.Name())
			}

			// Mark active projects
			if util.Contains(openProjects, e.Name()) {
				item.Display = item.Display + " (Active)"
				item.IsActive = true
			}

			folders = append(folders, item)
		}

		// Sort: Active first, then by recent usage, then alphabetical
		if wc.ActiveProjectsFirst {
			sort.Slice(folders, func(i, j int) bool {
				iActive := folders[i].IsActive
				jActive := folders[j].IsActive

				// Active projects first
				if iActive != jActive {
					return iActive
				}

				// Then by recent usage if tracking enabled
				if wc.EnableUsageTracking && wc.EnableCache {
					// Extract project name from value (e.g., "folder:workspacer" -> "workspacer")
					iProject := strings.TrimPrefix(folders[i].Value, "folder:")
					jProject := strings.TrimPrefix(folders[j].Value, "folder:")

					iCache, iExists := cache.GetProjectCache(iProject)
					jCache, jExists := cache.GetProjectCache(jProject)

					if iExists && jExists {
						// Sort by recent usage count (higher first)
						if iCache.AccessCountRecent != jCache.AccessCountRecent {
							return iCache.AccessCountRecent > jCache.AccessCountRecent
						}
					}
				}

				// Finally alphabetical
				return folders[i].Display < folders[j].Display
			})
		}

		// Add remote repos
		for _, remoteRepo := range remoteRepos {
			folders = append(folders, list.Item{
				Display:  remoteRepo,
				Value:    "git:" + remoteRepo,
				Subtitle: "Clone From GitHub",
			})
		}

		// Add error indicator if GitHub fetch failed
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

		// Save cache
		if err := SaveCache(wc, cache); err != nil {
			log.Error("Failed to save cache: %s", err.Error())
		}

		resultChan <- loadResult{folders: folders, remoteError: remoteError}
	}()

	// Run spinner while loading
	if _, err := p.Run(); err != nil {
		log.Error("Error running spinner: %s", err.Error())
	}

	// Get loaded data
	result := <-resultChan

	// Show list
	item, found, err := list.NewList("Select a project", result.folders)
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
