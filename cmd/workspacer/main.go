package main

import (
	"fmt"
	"sort"

	"github.com/JamesTiberiusKirk/workspacer/cli"
	"github.com/JamesTiberiusKirk/workspacer/commands"
	"github.com/JamesTiberiusKirk/workspacer/config"
	"github.com/JamesTiberiusKirk/workspacer/log"
	"github.com/JamesTiberiusKirk/workspacer/util"
	"github.com/JamesTiberiusKirk/workspacer/workspacer"
)

var ConfigMap cli.ConfigMapType = cli.ConfigMapType{
	cli.CommandTypeNoCommand: &cli.Command{
		Description: "Run the project picker.",
		Runner: cli.MiddlewareCommon(func(ctx cli.ConfigMapCtx) {
			t, choise := workspacer.ChoseProjectFromLocalWorkspace(
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

	cli.CommandTypeDefault: &cli.Command{
		Description: "Provide local project to open",
		Runner: cli.MiddlewareCommon(func(ctx cli.ConfigMapCtx) {
			// try and open the directory
			workspacer.StartOrSwitchToSession(
				ctx.WorkspaceConfig,
				ctx.Config.SessionPresets,
				ctx.Args[0],
			)
		}),
	},
	"nc,new_config": &cli.Command{
		Description: "Creates new configuration file",
		Runner: func(ctx cli.ConfigMapCtx) {
			err := config.WriteDefaultConfig()
			if err != nil {
				log.Error("Failed to create config: %s", err.Error())
				return
			}

			configPath, _ := config.GetDefaultConfigPath()
			log.Info("Created default config at: %s", configPath)
			log.Info("Edit this file to configure your workspaces")
		},
	},

	"c,clone": &cli.Command{
		Description: "Clone a project from github org or user",
		Runner: func(ctx cli.ConfigMapCtx) {
			// TODO: get list of all repos in an org and allow the user to clone one
			// check if the directory already exists and mark it as so in the list
			log.Info("CLONE, to be implemented")
		},
	},

	"l,list": &cli.Command{
		Description: "List open sessions in a workspace",
		Runner: func(ctx cli.ConfigMapCtx) {
			// TODO: implement tmux session list only for the workspace
			log.Info("LIST, to be implemented")
		},
	},

	"s,search": &cli.Command{
		Description: "Search in github org or user",
		Runner: cli.MiddlewareCommon(func(ctx cli.ConfigMapCtx) {
			searchArgs := ""
			if len(ctx.Args) > 1 {
				for _, arg := range ctx.Args[1:] {
					searchArgs += arg + " "
				}
			}

			workspacer.SearchGithubInUserOrOrg(ctx.WorkspaceConfig.GithubOrg, searchArgs)
		}),
	},

	"a,actions": &cli.Command{
		Description: "Get github actions status for tmux",
		Runner: cli.MiddlewareCommon(func(ctx cli.ConfigMapCtx) {
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

	"o,open": &cli.Command{
		Description: "Open list chooser for the currently open workspace sessions",
		Runner: cli.MiddlewareCommon(func(ctx cli.ConfigMapCtx) {
			workspacer.ChooseFromOpenWorkspaceProjectsAndSwitch(ctx.WorkspaceConfig.Prefix,
				ctx.WorkspaceConfig,
				config.DefaultGlobalConfig.SessionPresets,
			)
		}),
	},

	"CA,close-all": &cli.Command{
		Description: "Close all sessions in workspace",
		Runner: cli.MiddlewareCommon(func(ctx cli.ConfigMapCtx) {
			workspacer.CloseAllSessionsInWorkspace(ctx.WorkspaceConfig)
		}),
	},

	// Used by my tmux config to get current session
	"get-tmux-workspace-filter": &cli.Command{
		Description: `Get tmux template for filtering sessions for the current workspace (i.e. for chose-tree). 
		If no workspace present, return without any filter.
		Example config:
		bind-key s run-shell "tmux choose-tree -Zs -f \"$(workspacer -W=current get-tmux-workspace-filter)\""`,
		Runner: cli.MiddlewareCommon(func(ctx cli.ConfigMapCtx) {
			if ctx.WorkspaceConfig.Prefix == "" {
				fmt.Print("#{session_name}")
				return
			}

			template := "#{m:%s-*,#{session_name}}"
			fmt.Printf(template, ctx.WorkspaceConfig.Prefix)
		}),
	},

	"from-preset": &cli.Command{
		Description: "Open up a preset as a project. I.E. a preset for editing dot files which might be defined in session presets section",
		Runner: func(ctx cli.ConfigMapCtx) {
			if len(ctx.Args) < 2 {
				fmt.Println("Need to provide the name of a new repo")
				return
			}

			sessionPreset := ctx.Args[1]
			preset, ok := config.DefaultGlobalConfig.SessionPresets[sessionPreset]
			if !ok {

			}

			workspacer.StartOrSwitchToTmuxPreset(sessionPreset, preset.Path, preset)
		},
	},

	"n,new": &cli.Command{
		Description: "New project in the workspace project folder.",
		Runner:      cli.MiddlewareCommon(commands.RunNewCommand),
	},

	"cache": &cli.Command{
		Description: "Cache management commands. Usage: cache [clear|status]",
		Runner: cli.MiddlewareCommon(func(ctx cli.ConfigMapCtx) {
			if len(ctx.Args) < 2 {
				log.Info("Usage: cache [clear|status]")
				log.Info("  clear  - Clear cache for current workspace")
				log.Info("  status - Show cache statistics")
				return
			}

			subcommand := ctx.Args[1]
			switch subcommand {
			case "clear":
				err := workspacer.ClearCache(ctx.WorkspaceConfig)
				if err != nil {
					log.Error("Failed to clear cache: %s", err.Error())
					return
				}
				log.Info("Cache cleared for workspace: %s", ctx.WorkspaceConfig.Name)

			case "status":
				stats, err := workspacer.GetCacheStats(ctx.WorkspaceConfig)
				if err != nil {
					log.Error("Failed to get cache stats: %s", err.Error())
					return
				}

				fmt.Printf("Cache Status for workspace: %s\n", ctx.WorkspaceConfig.Name)
				fmt.Printf("  Path: %s\n", stats["cache_path"])
				fmt.Printf("  Exists: %v\n", stats["exists"])
				if stats["exists"].(bool) {
					fmt.Printf("  Size: %d bytes\n", stats["size_bytes"])
					fmt.Printf("  Last Modified: %v\n", stats["modified"])
					fmt.Printf("  Last Updated: %v\n", stats["last_updated"])
					fmt.Printf("  Cached Projects: %d\n", stats["num_projects"])
					fmt.Printf("  GitHub Repos: %d\n", stats["num_github_repos"])
					fmt.Printf("  Recent Accesses: %d\n", stats["num_recent_accesses"])
				}

			default:
				log.Info("Unknown subcommand: %s", subcommand)
				log.Info("Usage: cache [clear|status]")
			}
		}),
	},

	"stats": &cli.Command{
		Description: "Show workspace usage statistics",
		Runner: cli.MiddlewareCommon(func(ctx cli.ConfigMapCtx) {
			cache := workspacer.LoadCache(ctx.WorkspaceConfig)

			if len(cache.Projects) == 0 {
				log.Info("No usage statistics available yet")
				return
			}

			fmt.Printf("Usage Statistics for workspace: %s\n\n", ctx.WorkspaceConfig.Name)

			// Sort projects by total access count
			type projectStat struct {
				name   string
				total  int
				recent int
			}
			var stats []projectStat
			for name, project := range cache.Projects {
				stats = append(stats, projectStat{
					name:   name,
					total:  project.AccessCountTotal,
					recent: project.AccessCountRecent,
				})
			}

			// Sort by total count desc
			sort.Slice(stats, func(i, j int) bool {
				return stats[i].total > stats[j].total
			})

			fmt.Println("All-Time Top Projects:")
			for i, stat := range stats {
				if i >= 10 {
					break
				}
				fmt.Printf("  %2d. %-30s Total: %3d  Recent: %3d\n", i+1, stat.name, stat.total, stat.recent)
			}

			// Sort by recent count desc
			sort.Slice(stats, func(i, j int) bool {
				return stats[i].recent > stats[j].recent
			})

			fmt.Println("\nRecent Activity (last 50 accesses):")
			for i, stat := range stats {
				if i >= 10 || stat.recent == 0 {
					break
				}
				fmt.Printf("  %2d. %-30s Recent: %3d\n", i+1, stat.name, stat.recent)
			}

			fmt.Printf("\nTotal Accesses Recorded: %d\n", len(cache.RecentAccesses))
		}),
	},
}

func main() {
	cli.Run(ConfigMap)
}
