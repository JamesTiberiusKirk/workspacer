package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/JamesTiberiusKirk/workspacer/config"
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

// HandleSubcommands routes to subcommands or prints help if no valid subcommand found.
// This is the main function to use for handling subcommands - it automatically:
// - Routes to the appropriate subcommand if found
// - Prints help if no args provided or invalid subcommand
func HandleSubcommands(ctx ConfigMapCtx, subcommandMap ConfigMapType, usageMsg string) {
	// Find the subcommand by skipping flags AND the parent command name
	subcommand := ""
	subcommandIdx := -1
	foundParentCommand := false

	for i := 0; i < len(ctx.Args); i++ {
		arg := ctx.Args[i]

		// Skip flags and their values
		if strings.HasPrefix(arg, "-") {
			// If flag has =, it's self-contained, otherwise skip next arg as value
			if !strings.Contains(arg, "=") && i+1 < len(ctx.Args) && !strings.HasPrefix(ctx.Args[i+1], "-") {
				i++ // skip the value
			}
			continue
		}

		// First non-flag arg is the parent command name, skip it
		if !foundParentCommand {
			foundParentCommand = true
			continue
		}

		// Found the actual subcommand
		subcommand = arg
		subcommandIdx = i
		break
	}

	// No subcommand found
	if subcommand == "" {
		printSubcommandHelp(subcommandMap, usageMsg)
		return
	}

	cmd, ok := subcommandMap[subcommand]
	if !ok {
		fmt.Printf("Unknown subcommand: %s\n", subcommand)
		printSubcommandHelp(subcommandMap, usageMsg)
		return
	}

	// Create new context with args shifted to start at the subcommand
	subCtx := ctx
	subCtx.Args = ctx.Args[subcommandIdx:]
	cmd.Runner(subCtx)
}

// printSubcommandHelp prints help text for available subcommands
func printSubcommandHelp(subcommandMap ConfigMapType, usageMsg string) {
	if usageMsg != "" {
		fmt.Println(usageMsg)
	}
	fmt.Println("\nAvailable subcommands:")
	for name, cmd := range subcommandMap {
		fmt.Printf("  %-10s - %s\n", name, cmd.Description)
	}
}
