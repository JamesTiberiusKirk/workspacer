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
	ParsedFlags     *FlagValues // Parsed flag values from command's flag definitions
}

// GetBoolFlag returns a boolean flag value
// Panics if the flag doesn't exist in the command's flag definitions
func (ctx ConfigMapCtx) GetBoolFlag(name string) bool {
	if ctx.ParsedFlags == nil {
		panic("no flags parsed for this command")
	}
	return ctx.ParsedFlags.Bool(name)
}

// GetStringFlag returns a string flag value
// Panics if the flag doesn't exist in the command's flag definitions
func (ctx ConfigMapCtx) GetStringFlag(name string) string {
	if ctx.ParsedFlags == nil {
		panic("no flags parsed for this command")
	}
	return ctx.ParsedFlags.String(name)
}

// GetIntFlag returns an int flag value
// Panics if the flag doesn't exist in the command's flag definitions
func (ctx ConfigMapCtx) GetIntFlag(name string) int {
	if ctx.ParsedFlags == nil {
		panic("no flags parsed for this command")
	}
	return ctx.ParsedFlags.Int(name)
}

// GetFlagArgs returns the remaining non-flag arguments
func (ctx ConfigMapCtx) GetFlagArgs() []string {
	if ctx.ParsedFlags == nil {
		return ctx.Args
	}
	return ctx.ParsedFlags.Args()
}

type Runner func(ctx ConfigMapCtx)

type Flag struct {
	Name        string // Long name (e.g., "workspace")
	Short       string // Short name (e.g., "W"), optional
	Description string
	Type        string // "string", "bool", "int"
	Default     string // Default value as string
}

type Command struct {
	Description string
	Help        string
	Flags       []Flag        // Declarative flag definitions
	Subcommands ConfigMapType // Optional: for commands with subcommands (enables completion)
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

// GlobalFlags defines flags that are available for all commands.
// Applications using this CLI library should set this before calling Run().
// Example:
//   cli.GlobalFlags = []cli.Flag{
//       {Name: "verbose", Short: "v", Description: "Verbose output", Type: "bool"},
//   }
var GlobalFlags = []Flag{}

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
