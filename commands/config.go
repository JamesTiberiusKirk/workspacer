package commands

import (
	"fmt"

	"github.com/JamesTiberiusKirk/workspacer/cli"
	"github.com/JamesTiberiusKirk/workspacer/config"
	"github.com/JamesTiberiusKirk/workspacer/log"
	"github.com/JamesTiberiusKirk/workspacer/state"
)

// ConfigSubcommands defines the subcommands for the config command
var ConfigSubcommands = cli.ConfigMapType{
	"new": {
		Description: "Create new default configuration file",
		Runner:      runConfigNew,
	},
	"list": {
		Description: "Show actually loaded config and environment file paths",
		Runner:      runConfigFiles,
	},
}

func RunConfigCommand(ctx cli.ConfigMapCtx) {
	cli.HandleSubcommands(ctx, ConfigSubcommands, "Usage: config [subcommand]")
}

func runConfigNew(ctx cli.ConfigMapCtx) {
	err := config.WriteDefaultConfig()
	if err != nil {
		log.Error("Failed to create config: %s", err.Error())
		return
	}

	configPath, _ := config.GetDefaultConfigPath()
	log.Info("Created default config at: %s", configPath)
	log.Info("Edit this file to configure your workspaces")
}

func runConfigFiles(ctx cli.ConfigMapCtx) {
	// If globals haven't been set yet (command run without middleware), load manually
	if state.LoadedConfigPath == "" && state.LoadedEnvPath == "" {
		// Try to load config
		loadedConfig, err := config.LoadFromDefaultConfigPath()
		if err == nil && loadedConfig != nil {
			configPath, _ := config.GetDefaultConfigPath()
			state.LoadedConfigPath = configPath
		}

		// Note: We can't load env file without workspace config, so skip it
	}

	// Show actually loaded config file
	if state.LoadedConfigPath != "" {
		fmt.Println(state.LoadedConfigPath)
	}

	// Show actually loaded env file
	if state.LoadedEnvPath != "" {
		fmt.Println(state.LoadedEnvPath)
	}
}
