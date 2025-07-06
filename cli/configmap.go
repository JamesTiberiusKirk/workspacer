package cli

import (
	"context"

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
