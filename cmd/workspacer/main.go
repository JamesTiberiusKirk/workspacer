package main

import (
	"fmt"
	"os"

	"github.com/JamesTiberiusKirk/workspacer/config"
	"github.com/JamesTiberiusKirk/workspacer/log"
	"github.com/JamesTiberiusKirk/workspacer/workspacer"
)

type flags struct {
	workspace string
}

func main() {
	// TODO: config stuff
	// check for any global configs in users home directory
	// if debug and does not exist, use default config
	// if not debug complain

	args := os.Args
	mc, args := config.ParseArgs(args)

	log.Info("workspace: %s", mc.Workspace)

	if len(args) == 0 {
		if mc.Workspace == "" {
			log.Error("no workspace provided")
			return
		}

		workspaceConfig, ok := config.DefaultGlobalConfig.Workspaces[mc.Workspace]
		if !ok {
			log.Error("workspace %s not found", mc.Workspace)
			return
		}

		t, choise := workspacer.ChoseProjectFromWorkspace(mc.Workspace, workspaceConfig, nil)
		switch t {
		case "folder":
			args = append([]string{choise}, args...)
		case "git":
			// TODO: close repo
		case "nochoise":
			return
		}
	}

	log.Debug("args: %v len(args):%d", args, len(args))

	switch args[0] {
	case "nc", "new_config":
	case "c", "clone":
		// TODO: get list of all repos in an org and allow the user to clone one
		// check if the directory already exists and mark it as so in the list
		log.Info("CLONE, to be implemented")
	case "l", "list":
		// TODO: implement tmux session list only for the workspace
		log.Info("LIST, to be implemented")
	case "s", "search":
		// TODO: implement github arch

		searchArgs := ""
		if len(args) > 1 {
			fmt.Println("Sorry this does not work yet")
			return
			// for _, arg := range args[1:] {
			// 	searchArgs += arg + " "
			// }
		}

		workspacer.SearchGithubInUserOrOrg(mc.Workspace, searchArgs)
	case "o", "open":
		workspacer.ChooseFromOpenWorkspaceProjectsAndSwitch(mc.Workspace,
			config.DefaultGlobalConfig.Workspaces[mc.Workspace],
			config.DefaultGlobalConfig.SessionPresets,
		)

	default:
		// try and open the directory
		workspacer.StartOrSwitchToSession(
			mc.Workspace,
			config.DefaultGlobalConfig.Workspaces[mc.Workspace],
			config.DefaultGlobalConfig.SessionPresets,
			args[0],
		)
	}
}
