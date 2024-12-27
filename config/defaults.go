package config

import (
	gotmux "github.com/jubnzv/go-tmux"
)

var (
	// Example config and my personal config
	DefaultGlobalConfig = GlobalUserConfig{
		DefaultWorkspace: "ws",
		Workspaces: map[string]WorkspaceConfig{
			"ws": {
				Name:          "Projects",
				Prefix:        "ws",
				Path:          "~/Projects/",
				SessionPreset: "default",
			},
			"av": {
				Name:          "Aviva",
				Prefix:        "av",
				Path:          "~/Aviva",
				GithubOrg:     "aviva-verde",
				SessionPreset: "default",
			},
		},
		SessionPresets: map[string]SessionConfig{
			"default": {
				Windows: []WindowConfig{
					{
						Name:   "vi",
						Layout: gotmux.LayoutEvenHorizontal,
						Panes: []PanesConfig{
							{Command: "vi"},
						},
					},
					{
						Name:  "shell",
						Panes: []PanesConfig{},
					},
				},
			},
		},
	}
)
