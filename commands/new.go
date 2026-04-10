package commands

import (
	"flag"
	"fmt"

	"github.com/JamesTiberiusKirk/workspacer/cli"
	"github.com/JamesTiberiusKirk/workspacer/ui/newwizard"
	"github.com/JamesTiberiusKirk/workspacer/workspacer"
)

func RunNewCommand(ctx cli.ConfigMapCtx) {
	var (
		name      string
		gh        bool
		isPrivate bool
	)

	// Zero user args (just "new" itself) -> open the TUI wizard.
	if len(ctx.Args) == 1 {
		result, err := newwizard.Run(ctx.WorkspaceConfig)
		if err != nil {
			fmt.Println("Error running wizard:", err)
			return
		}
		if result.Cancelled {
			return
		}
		name = result.Name
		gh = result.GitHub
		isPrivate = result.Private
	} else {
		fs := flag.NewFlagSet("new", flag.ExitOnError)
		ghFlag := fs.Bool("gh", false, "Creates git and github repo and pushes a first commit")
		privateFLag := fs.Bool("private", false, "If gh repo is initialises, sets the repo to private")

		fs.Parse(ctx.Args[1:])

		name = fs.Arg(0)
		if name == "" {
			fmt.Println("Need to provide the name of a new repo")
			return
		}

		if name == "new" {
			panic("either you cant provide name new or there's a bug where the sub command got used for the name of the new project")
		}

		gh = *ghFlag
		isPrivate = *privateFLag
	}

	if gh {
		createdName, err := workspacer.CreateGitHubRepo(ctx.WorkspaceConfig, name, isPrivate)
		if err != nil {
			fmt.Println("Error creating repo:", err)
			return
		}

		fmt.Println("Created Repo:", createdName)

		err = workspacer.NewProjectAndPush(ctx.WorkspaceConfig, createdName)
		if err != nil {
			fmt.Println("Error cloning:", err)
			return
		}
		name = createdName
	} else {
		err := workspacer.NewProject(ctx.WorkspaceConfig, name)
		if err != nil {
			fmt.Println("Error cloning:", err)
			return
		}
	}

	workspacer.StartOrSwitchToSession(
		ctx.WorkspaceConfig,
		ctx.Config.SessionPresets,
		name,
	)
}
