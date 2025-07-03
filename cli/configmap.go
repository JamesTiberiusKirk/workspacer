package cli

import (
	"context"
	"fmt"

	"github.com/JamesTiberiusKirk/workspacer/config"
	"github.com/JamesTiberiusKirk/workspacer/log"
	"github.com/JamesTiberiusKirk/workspacer/util"
	"github.com/JamesTiberiusKirk/workspacer/workspacer"
)

type ConfigMapCtx struct {
	Context         context.Context
	Args            []string
	Project         string
	WorkspaceConfig config.WorkspaceConfig
	Config          config.GlobalUserConfig
}

type Runner func(ctx ConfigMapCtx)
type Command struct {
	Description string
	Help        string
	Runner      Runner
}

// commands: runners
type ConfigMapType map[string]*Command

const (
	// When no Command has been passed
	CommandTypeNoCommand = "no-command"

	// When command is not in list
	CommandTypeDefault = "default"
)

var ConfigMap ConfigMapType = ConfigMapType{
	CommandTypeNoCommand: &Command{
		Description: "Run the project picker.",
		Runner: MiddlewareCommon(func(ctx ConfigMapCtx) {
			t, choise := workspacer.ChoseProjectFromWorkspace(
				ctx.WorkspaceConfig.Prefix,
				ctx.WorkspaceConfig,
				nil,
			)
			switch t {
			case "folder":
				// args = append([]string{choise}, args...)
			case "git":
				err := workspacer.CloneRepo(ctx.WorkspaceConfig, choise)
				if err != nil {
					return
				}
				// args = append([]string{choise}, args...)
			case "nochoise":
				return

			}

			// try and open the directory
			workspacer.StartOrSwitchToSession(
				ctx.WorkspaceConfig,
				ctx.Config.SessionPresets,
				choise,
			)
			return

		}),
	},

	CommandTypeDefault: &Command{
		Description: "Provide local project to open",
		Runner: MiddlewareCommon(func(ctx ConfigMapCtx) {
			// try and open the directory
			workspacer.StartOrSwitchToSession(
				ctx.WorkspaceConfig,
				ctx.Config.SessionPresets,
				ctx.Args[0],
			)
		}),
	},
	"nc,new_config": &Command{
		Description: "Creates new configuration file",
		Runner: func(ctx ConfigMapCtx) {
			log.Info("new config to be implemented")
		},
	},

	"c,clone": &Command{
		Description: "Clone a project from github org or user",
		Runner: func(ctx ConfigMapCtx) {
			// TODO: get list of all repos in an org and allow the user to clone one
			// check if the directory already exists and mark it as so in the list
			log.Info("CLONE, to be implemented")
		},
	},

	"l,list": &Command{
		Description: "List open sessions in a workspace",
		Runner: func(ctx ConfigMapCtx) {
			// TODO: implement tmux session list only for the workspace
			log.Info("LIST, to be implemented")
		},
	},

	"s,search": &Command{
		Description: "Search in github org or user",
		Runner: MiddlewareCommon(func(ctx ConfigMapCtx) {
			searchArgs := ""
			if len(ctx.Args) > 1 {
				for _, arg := range ctx.Args[1:] {
					searchArgs += arg + " "
				}
			}

			workspacer.SearchGithubInUserOrOrg(ctx.WorkspaceConfig.GithubOrg, searchArgs)
		}),
	},

	"a,actions": &Command{
		Description: "Get github actions status for tmux",
		Runner: MiddlewareCommon(func(ctx ConfigMapCtx) {
			mainBranch := util.GetGitMainBranch(ctx.WorkspaceConfig, ctx.Project)

			branch := util.GetProjectCurrentBranch(ctx.WorkspaceConfig, ctx.Project)
			branches := []string{mainBranch}

			staging, prod := false, false
			if util.DoesBranchExist(ctx.WorkspaceConfig, ctx.Project, "staging") {
				branches = append(branches, "staging")
				staging = true
			}

			if util.DoesBranchExist(ctx.WorkspaceConfig, ctx.Project, "production") {
				branches = append(branches, "production")
				prod = true
			}

			if branch != "" && branch != mainBranch && (branch != "staging" && staging) && (branch != "production" && prod) {
				branches = append([]string{branch}, branches...)
			}

			fmt.Println("Branches: ", branches)
			results := workspacer.GetWorkFlowsStatus(ctx.WorkspaceConfig.Prefix, ctx.Project, branches...)
			for _, r := range results {
				fmt.Println(r)
			}
		}),
	},

	"o,open": &Command{
		Description: "Open list chooser for the currently open workspace sessions",
		Runner: MiddlewareCommon(func(ctx ConfigMapCtx) {
			workspacer.ChooseFromOpenWorkspaceProjectsAndSwitch(ctx.WorkspaceConfig.Prefix,
				ctx.WorkspaceConfig,
				config.DefaultGlobalConfig.SessionPresets,
			)
		}),
	},

	"CA,close-all": &Command{
		Description: "Close all sessions in workspace",
		Runner: MiddlewareCommon(func(ctx ConfigMapCtx) {
			workspacer.CloseAllSessionsInWorkspace(ctx.WorkspaceConfig)
		}),
	},

	// Used by my tmux config to get current session
	"get-tmux-workspace-filter": &Command{
		Description: `Get tmux template for filtering sessions for the current workspace (i.e. for chose-tree). 
		If no workspace present, return without any filter.
		Example config:
		bind-key s run-shell "tmux choose-tree -Zs -f \"$(workspacer -W=current get-tmux-workspace-filter)\""`,
		Runner: MiddlewareCommon(func(ctx ConfigMapCtx) {
			if ctx.WorkspaceConfig.Prefix == "" {
				fmt.Print("#{session_name}")
				return
			}

			template := "#{m:%s-*,#{session_name}}"
			fmt.Printf(template, ctx.WorkspaceConfig.Prefix)
		}),
	},

	"from-presets": &Command{
		Description: "Open up a preset as a project. I.E. a preset for editing dot files which might be defined in session presets section",
		Runner: MiddlewareCommon(func(ctx ConfigMapCtx) {
			if len(ctx.Args) < 2 {
				fmt.Println("Need to provide the name of a new repo")
				return
			}

			sessionPreset := ctx.Args[1]
			workspacer.StartOrSwitchToTmuxPreset("dots", "", config.DefaultGlobalConfig.SessionPresets[sessionPreset])
		}),
	},

	"n,new": &Command{
		Description: "New project in the workspace project folder.",
		Help: `
		-gh			Initiate Github Repo and push initial commit
		-private	Initiate repo as private if -gh is on
		-h			Print this message
		`,
		Runner: MiddlewareCommon(func(ctx ConfigMapCtx) {
			if len(ctx.Args) < 2 {
				fmt.Println("Need to provide the name of a new repo")
				return
			}

			private := false
			createRepoAndPush := false
			name := ""

			// TODO: re-writte this with go flag subcommands
			// https://gobyexample.com/command-line-subcommands
			// skip first as that will be "new" command
			args := ctx.Args[1:]
			for _, arg := range args {
				switch arg {
				case "-private":
					private = true
				case "-gh":
					createRepoAndPush = true
				case "-h", "-help":
					fmt.Println(`
					-private	if repo is created, make it private
					-gh			create gh repo and push
					`)
					return
				default:
					name = arg
				}
			}

			if createRepoAndPush {
				name, err := workspacer.CreateGitHubRepo(ctx.WorkspaceConfig, name, private)
				if err != nil {
					fmt.Println("Error creating repo:", err)
					return
				}
				fmt.Println("Created Repo:", name)

				err = workspacer.NewProjectAndPush(ctx.WorkspaceConfig, name)
				if err != nil {
					fmt.Println("Error cloning:", err)
					return
				}
			} else {
				err := workspacer.NewProject(ctx.WorkspaceConfig, name)
				if err != nil {
					fmt.Println("Error cloning:", err)
					return
				}
			}

			workspacer.StartOrSwitchToSession(
				ctx.WorkspaceConfig,
				config.DefaultGlobalConfig.SessionPresets,
				name,
			)
		}),
	},
}
