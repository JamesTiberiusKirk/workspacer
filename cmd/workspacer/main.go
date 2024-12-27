package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/JamesTiberiusKirk/workspacer/config"
	"github.com/JamesTiberiusKirk/workspacer/log"
	"github.com/JamesTiberiusKirk/workspacer/util"
	"github.com/JamesTiberiusKirk/workspacer/workspacer"
	"github.com/jubnzv/go-tmux"
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

	switch args[1] {
	case "h", "help":
		fmt.Println("workspacer ...")
		return
	}

	mc, args := config.ParseArgs(args)

	switch mc.Workspace {
	case "current":
		name, err := tmux.GetAttachedSessionName()
		if err != nil {
			log.Error("could not get current tmux session")
			return
		}

		if name == "" {
			log.Error("no tmux session attached")
			return
		}

		if strings.Contains(name, "-") {
			mc.Workspace = strings.Split(name, "-")[0]
		} else {
			mc.Workspace = config.DefaultGlobalConfig.DefaultWorkspace
		}
	}

	log.Debug("workspace: %s", mc.Workspace)

	if mc.Workspace == "" {
		log.Error("no workspace provided")
		return
	}

	workspaceConfig, ok := config.DefaultGlobalConfig.Workspaces[mc.Workspace]
	if !ok {
		log.Error("workspace %s not found", mc.Workspace)
		return
	}

	util.LoadEnvFile(workspaceConfig)
	log.Debug("env loaded")

	if len(args) == 0 {
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
	case "a", "actions":
		mainBranch := util.GetGitMainBranch(workspaceConfig, args[1])

		branch := util.GetProjectCurrentBranch(workspaceConfig, args[1])
		branches := []string{mainBranch}

		staging, prod := false, false
		if util.DoesBranchExist(workspaceConfig, args[1], "staging") {
			branches = append(branches, "staging")
			staging = true
		}

		if util.DoesBranchExist(workspaceConfig, args[1], "production") {
			branches = append(branches, "production")
			prod = true
		}

		if branch != "" && branch != mainBranch && (branch != "staging" && staging) && (branch != "production" && prod) {
			branches = append([]string{branch}, branches...)
		}

		fmt.Println("Branches: ", branches)

		results := workspacer.GetWorkFlowsStatus(mc.Workspace, args[1], branches...)
		for _, r := range results {
			fmt.Println(r)
		}

	case "o", "open":
		workspacer.ChooseFromOpenWorkspaceProjectsAndSwitch(mc.Workspace,
			workspaceConfig,
			config.DefaultGlobalConfig.SessionPresets,
		)

	case "CA", "close-all":
		workspacer.CloseAllSessionsInWorkspace(workspaceConfig)

	case "get-tmux-workspace-filter":
		template := "#{m:%s-*,#{session_name}}"
		fmt.Printf(template, mc.Workspace)
	default:
		// try and open the directory
		workspacer.StartOrSwitchToSession(
			mc.Workspace,
			workspaceConfig,
			config.DefaultGlobalConfig.SessionPresets,
			args[0],
		)
	}
}
