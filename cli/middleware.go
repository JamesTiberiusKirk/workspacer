package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/JamesTiberiusKirk/workspacer/config"
	"github.com/JamesTiberiusKirk/workspacer/log"
	"github.com/JamesTiberiusKirk/workspacer/util"
	"github.com/jubnzv/go-tmux"
)

func MiddlewareCommon(r Runner) Runner {
	return MiddlewareConfigInjector(
		MiddlewareAssertWorkspace(
			func(ctx ConfigMapCtx) {
				r(ctx)
			},
		),
	)
}

// MiddlewareAssertWorkspace - this asserts the workspace from the flag.
// Needs config injector.
func MiddlewareAssertWorkspace(r Runner) Runner {
	return func(ctx ConfigMapCtx) {
		fs := flag.NewFlagSet("base", flag.ExitOnError)

		workspaceFlag := fs.String("workspace", "", "Specify workspace in which to work")
		wFlag := fs.String("W", "", "Shorthand for -workspace. This overwrites -workspace")
		fs.Parse(ctx.Args)

		workspace := ""

		if workspaceFlag != nil && *workspaceFlag != "" {
			workspace = *workspaceFlag
		}

		if wFlag != nil && *wFlag != "" {
			workspace = *wFlag
		}

		if workspace == "current" {
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
				workspace = strings.Split(name, "-")[0]
			}
		}

		wsConfig, ok := ctx.Config.Workspaces[workspace]
		if !ok {
			fmt.Printf("Workspace not found %s\n", workspace)
			os.Exit(1)
		}

		ctx.WorkspaceConfig = wsConfig
		ctx.Args = fs.Args()

		util.LoadEnvFile(ctx.WorkspaceConfig)

		r(ctx)
	}
}

// MiddlewareConfigInjector - gets config and injects it in the ctx
func MiddlewareConfigInjector(r Runner) Runner {
	return func(ctx ConfigMapCtx) {
		// Try to load config from file, fall back to default
		loadedConfig, err := config.LoadFromDefaultConfigPath()
		if err != nil {
			log.Error("Failed to load config from file, using default: %s", err.Error())
			ctx.Config = config.DefaultGlobalConfig
		} else if loadedConfig == nil {
			// Config file doesn't exist, use default
			ctx.Config = config.DefaultGlobalConfig
		} else {
			// Successfully loaded config from file
			ctx.Config = *loadedConfig
		}

		r(ctx)
	}
}
