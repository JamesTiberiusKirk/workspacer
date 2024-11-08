package main

import (
	"github.com/JamesTiberiusKirk/workspacer/config"
	"github.com/JamesTiberiusKirk/workspacer/workspacer"
)

func main() {
	// workspacer.Execute()

	workspacer.StartOrSwitchToSession(
		"ws",
		config.DefaultGlobalConfig.Workspaces["ws"],
		config.DefaultGlobalConfig.SessionPresets,
		"test",
	)
}
