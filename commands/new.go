package commands

import (
	"flag"
	"fmt"

	"github.com/JamesTiberiusKirk/workspacer/cli"
	"github.com/JamesTiberiusKirk/workspacer/config"
	"github.com/JamesTiberiusKirk/workspacer/workspacer"
)

func RunNewCommand(ctx cli.ConfigMapCtx) {
	fs := flag.NewFlagSet("new", flag.ExitOnError)
	ghFlag := fs.Bool("gh", false, "Creates git and github repo and pushes a first commit")
	privateFLag := fs.Bool("private", false, "If gh repo is initialises, sets the repo to private")

	fs.Parse(ctx.Args[1:])

	name := fs.Arg(0)
	if name == "" {
		fmt.Println("Need to provide the name of a new repo")
		return
	}

	if name == "new" {
		panic("either you cant provide name new or there's a bug where the sub command got used for the name of the new project")
	}

	if *ghFlag {
		name, err := workspacer.CreateGitHubRepo(ctx.WorkspaceConfig, name, *privateFLag)
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
}
