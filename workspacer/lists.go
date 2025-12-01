package workspacer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/JamesTiberiusKirk/workspacer/config"
	"github.com/JamesTiberiusKirk/workspacer/log"
	"github.com/JamesTiberiusKirk/workspacer/ui/list"
	"github.com/JamesTiberiusKirk/workspacer/util"
	"github.com/charmbracelet/lipgloss"
)

var (
	branchStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("12")) // Blue
	changesStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // Yellow
	changesCleanStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // Green
)

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

	openProjects := util.GetOpenProjectsByWorkspace(workspace)

	path := util.GetWorkspacePath(wc)
	entries, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}

	remoteRepos, err := GetRepoNames(wc.GithubOrg, wc.IsOrg)
	if err != nil {
		panic(err)
	}

	folders := []list.Item{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		// Skip tenant repos - they'll be shown as suffixes on service repos
		if wc.EnableTenantRepos && !util.IsServiceRepo(wc, e.Name()) {
			continue
		}

		item := list.Item{Display: e.Name(), Value: "folder:" + e.Name(), Subtitle: "Folder"}
		if util.HasGitSubfolder(filepath.Join(path, e.Name())) {
			// Get git information for service repo
			branch := util.GetGitBranch(wc, e.Name())
			changesCount := util.GetUncommittedChangesCount(wc, e.Name())

			// Build subtitle with git info and color coding
			subtitle := "Service: "
			if branch != "" {
				subtitle += branchStyle.Render(branch)
				if changesCount > 0 {
					subtitle += " " + changesStyle.Render(fmt.Sprintf("(%d)", changesCount))
				} else {
					subtitle += " " + changesCleanStyle.Render("✓")
				}
			}

			// Check if this service has a tenant repo
			if wc.EnableTenantRepos && util.DoesTenantRepoExist(wc, e.Name()) {
				item.Display = item.Display + " + tenant"

				// Get tenant git info
				tenantName := util.GetTenantRepoName(wc, e.Name())
				tenantBranch := util.GetGitBranch(wc, tenantName)
				tenantChanges := util.GetUncommittedChangesCount(wc, tenantName)

				subtitle += " | Tenant: "
				if tenantBranch != "" {
					subtitle += branchStyle.Render(tenantBranch)
					if tenantChanges > 0 {
						subtitle += " " + changesStyle.Render(fmt.Sprintf("(%d)", tenantChanges))
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
		}

		folders = append(folders, item)
	}

	// Sort active projects to the top if configured
	if wc.ActiveProjectsFirst {
		sort.Slice(folders, func(i, j int) bool {
			iActive := strings.Contains(folders[i].Display, "(Active)")
			jActive := strings.Contains(folders[j].Display, "(Active)")

			// If one is active and the other isn't, active comes first
			if iActive != jActive {
				return iActive
			}

			// Otherwise, maintain original order (alphabetical by name)
			return folders[i].Display < folders[j].Display
		})
	}

	for _, remoteRepo := range remoteRepos {
		folders = append(folders, list.Item{
			Display:  remoteRepo,
			Value:    "git:" + remoteRepo,
			Subtitle: "Clone From GitHub",
		})
	}

	// // Sort by Display field (ascending)
	// sort.Slice(folders, func(i, j int) bool {
	// 	return folders[i].Display < folders[j].Display
	// })

	if len(extraOptions) > 0 {
		folders = append(folders, extraOptions...)
	}

	item, found, err := list.NewList("Select a project", folders)
	if err != nil {
		panic(err)
	}
	if !found {
		return "nochoise", ""
	}

	if strings.HasPrefix(item.Value, "folder:") {
		return "folder", strings.TrimPrefix(item.Value, "folder:")
	} else if strings.HasPrefix(item.Value, "git:") {
		return "git", strings.TrimPrefix(item.Value, "git:")
	}
	return "", ""
}
