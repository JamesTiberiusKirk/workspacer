package config

import (
	gotmux "github.com/jubnzv/go-tmux"
)

var (
	DefaultGlobalConfig = GlobalUserConfig{
		Workspaces: map[string]WorkspaceConfig{
			"ws": {
				Name:          "Projects",
				Prefix:        "",
				Path:          "~/Projects/",
				SessionPreset: "default",
			},
			"av": {
				Name:          "Aviva",
				Prefix:        "av",
				Path:          "~/Aviva",
				OrgGithub:     "github.com/aviva-verde",
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
