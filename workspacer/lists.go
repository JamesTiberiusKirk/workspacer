package workspacer

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/JamesTiberiusKirk/workspacer/config"
	"github.com/JamesTiberiusKirk/workspacer/log"
	"github.com/JamesTiberiusKirk/workspacer/ui/list"
)

func ChooseFromOpenWorkspaceProjectsAndSwitch(workspace string, workspaceConfig config.WorkspaceConfig, sessionPresets map[string]config.SessionConfig) {
	openProjects := getOpenProjectsByWorkspace(workspace)
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
		panic("not found")
	}

	StartOrSwitchToSession(workspace, workspaceConfig, sessionPresets, item.Value)

	return
}

func ChoseProjectFromWorkspace(workspace string, wc config.WorkspaceConfig, extraOptions []list.Item) (string, string) {
	if _, err := os.Stat(getWorkspacePath(wc)); os.IsNotExist(err) {
		log.Error("workspace %s does not exist", workspace)
		return "nochoise", ""
	}

	openProjects := getOpenProjectsByWorkspace(workspace)

	path := getWorkspacePath(wc)
	entries, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}

	folders := []list.Item{}
	for _, e := range entries {
		if e.IsDir() {
			item := list.Item{Display: e.Name(), Value: "folder:" + e.Name(), Subtitle: "Folder"}
			if hasGitSubfolder(filepath.Join(path, e.Name())) {
				item.Subtitle = "Git Repo"
			}

			if contains(openProjects, e.Name()) {
				item.Display = item.Display + " (Active)"
			}

			folders = append(folders, item)
		}
	}

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
		return "folder", item.Value[7:]
	} else if strings.HasPrefix(item.Value, "git:") {
		return "git", item.Value
	}
	return "", ""
}
