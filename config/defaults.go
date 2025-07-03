package config

import (
	gotmux "github.com/jubnzv/go-tmux"
)

var (
	// Example config and my personal config
	DefaultGlobalConfig = GlobalUserConfig{
		DefaultWorkspace: "ws",
		Workspaces: map[string]WorkspaceConfig{
			"notes": {
				Name:          "Notes",
				Prefix:        "notes",
				Path:          "~/Documents/notes/",
				SessionPreset: "notes",
			},
			"ws": {
				Name:          "Projects",
				Prefix:        "ws",
				Path:          "~/Projects/",
				GithubOrg:     "JamesTiberiusKirk",
				IsOrg:         false,
				SessionPreset: "default",
			},
		},
		SessionPresets: map[string]SessionConfig{
			"notes": {
				Windows: []WindowConfig{
					{
						Name:   "nvim",
						Layout: gotmux.LayoutMainVertical,
						Panes: []PanesConfig{
							{Command: "nvim-l"},
						},
					},
				},
			},
			"default": {
				Windows: []WindowConfig{
					{
						Name:   "nvim",
						Layout: gotmux.LayoutMainVertical,
						Panes: []PanesConfig{
							{Command: "nvim-l"},
							{Command: ""},
						},
					},
				},
			},
			"dots": {
				Path: "~/",
				Windows: []WindowConfig{
					{
						Name:   "~/",
						Layout: gotmux.LayoutMainVertical,
						Panes: []PanesConfig{
							{Command: "$EDITOR ~/.profile"},
						},
						Path: "~/.config/nvim-l",
					},
					{
						Name:   "tmux",
						Layout: gotmux.LayoutMainVertical,
						Panes: []PanesConfig{
							{Command: "$EDITOR ~/.tmux.conf ~/.tmux-linux.conf"},
						},
					},
					{
						Name:   "hyprland",
						Layout: gotmux.LayoutMainVertical,
						Panes: []PanesConfig{
							{Command: "$EDITOR"},
						},
						Path: "~/.config/hypr/",
					},
				},
			},
		},
	}
)
