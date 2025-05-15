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
				GithubOrg:     "JamesTiberiusKirk",
				IsOrg:         false,
				SessionPreset: "default",
			},
			"av": {
				Name:          "Aviva",
				Prefix:        "av",
				Path:          "~/Aviva/",
				IsOrg:         true,
				GithubOrg:     "aviva-verde",
				SessionPreset: "default-mac",
			},
		},
		SessionPresets: map[string]SessionConfig{
			"default-mac": {
				Windows: []WindowConfig{
					{
						Name:   "nvim",
						Layout: gotmux.LayoutMainVertical,
						Panes: []PanesConfig{
							{Command: "nvim"},
							{Command: "git diff --quiet && git diff --cached --quiet && git pull"},
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
						Name:   "nvim",
						Layout: gotmux.LayoutMainVertical,
						Panes: []PanesConfig{
							{Command: "nvim-l"},
						},
						Path: "~/.config/nvim-l",
					},
					{
						Name:   "tmux",
						Layout: gotmux.LayoutMainVertical,
						Panes: []PanesConfig{
							{Command: "nvim-l ~/.tmux.conf ~/.tmux-linux.conf"},
						},
					},
				},
			},
		},
	}
)
