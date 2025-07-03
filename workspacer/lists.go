package workspacer

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/JamesTiberiusKirk/workspacer/config"
	"github.com/JamesTiberiusKirk/workspacer/log"
	"github.com/JamesTiberiusKirk/workspacer/ui/list"
	"github.com/JamesTiberiusKirk/workspacer/util"
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

	StartOrSwitchToSession(workspace, workspaceConfig, sessionPresets, item.Value)

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

func ChoseProjectFromWorkspace(workspace string, wc config.WorkspaceConfig, extraOptions []list.Item) (string, string) {
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

		item := list.Item{Display: e.Name(), Value: "folder:" + e.Name(), Subtitle: "Folder"}
		if util.HasGitSubfolder(filepath.Join(path, e.Name())) {
			item.Subtitle = "Local Git Repo"
			remoteRepos = removeRepoFromArray(remoteRepos, e.Name())
		}

		if util.Contains(openProjects, e.Name()) {
			item.Display = item.Display + " (Active)"
		}

		folders = append(folders, item)
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
