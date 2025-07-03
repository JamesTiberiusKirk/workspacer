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
		ctx.Args = flag.Args()

		util.LoadEnvFile(ctx.WorkspaceConfig)

		r(ctx)
	}
}

// MiddlewareConfigInjector - gets config and injects it in the ctx
func MiddlewareConfigInjector(r Runner) Runner {
	return func(ctx ConfigMapCtx) {
		// TODO: add support for loading config from a file
		// For now this will do
		ctx.Config = config.DefaultGlobalConfig

		r(ctx)
	}
}
