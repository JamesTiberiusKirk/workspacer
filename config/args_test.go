package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// EXAMPLE
// /tmp/go-build1891856884/b001/exe/main -W=ws help

func TestArgs(t *testing.T) {
	tests := []struct {
		Name         string
		Input        []string
		ExpectedOut  Options
		ExpectedErrs []string
	}{
		{
			Name:  "Basic_workspace_detection",
			Input: []string{"workspacer", "-W=ws"},
			ExpectedOut: Options{
				Workspace: "ws",
			},
		},
		{
			Name:  "Basic_workspace_detection",
			Input: []string{"workspacer", "-W", "ws"},
			ExpectedOut: Options{
				Workspace: "ws",
			},
		},
		{
			Name:  "Basic_workspace_detection",
			Input: []string{"workspacer", "-workspace=ws"},
			ExpectedOut: Options{
				Workspace: "ws",
			},
		},
		{
			Name:  "Basic_workspace_detection",
			Input: []string{"workspacer", "-workspace", "ws"},
			ExpectedOut: Options{
				Workspace: "ws",
			},
		},
		{
			Name:  "Help_falg_detetion",
			Input: []string{"workspacer", "-H"},
			ExpectedOut: Options{
				Help: true,
			},
		},
		{
			Name:  "Help_falg_detetion",
			Input: []string{"workspacer", "-help"},
			ExpectedOut: Options{
				Help: true,
			},
		},
		{
			Name:  "Debug_falg_detetion",
			Input: []string{"workspacer", "-D"},
			ExpectedOut: Options{
				Debug: true,
			},
		},
		{
			Name:  "Debug_falg_detetion",
			Input: []string{"workspacer", "-debug"},
			ExpectedOut: Options{
				Debug: true,
			},
		},
		{
			Name:  "Detecting_all_flags",
			Input: []string{"workspacer", "-D", "-W", "ws", "-H"},
			ExpectedOut: Options{
				Workspace: "ws",
				Debug:     true,
				Help:      true,
			},
		},
		{
			Name:  "Detecting_all_flags",
			Input: []string{"workspacer", "-W", "ws", "-H", "-D"},
			ExpectedOut: Options{
				Workspace: "ws",
				Debug:     true,
				Help:      true,
			},
		},
		{
			Name:  "Detecting_all_flags",
			Input: []string{"workspacer", "-H", "-D", "-W", "ws"},
			ExpectedOut: Options{
				Workspace: "ws",
				Debug:     true,
				Help:      true,
			},
		},
		{
			Name:        "No_flags",
			Input:       []string{"workspacer"},
			ExpectedOut: Options{},
		},
		// Error tests
		{
			Name:         "Error_with_no_workspace_provided",
			Input:        []string{"workspacer", "-W"},
			ExpectedOut:  Options{},
			ExpectedErrs: []string{"Workspace flag needs the name of a workspace"},
		},
		{
			Name:         "Error_with_no_workspace_provided",
			Input:        []string{"workspacer", "-W="},
			ExpectedOut:  Options{},
			ExpectedErrs: []string{"Workspace flag needs the name of a workspace"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			opts, errs := ParseArgs(tt.Input)
			assert.Equal(t, tt.ExpectedOut.Workspace, opts.Workspace, "Unexpected Workspace")
			assert.Equal(t, tt.ExpectedOut.Debug, opts.Debug, "Unexpected Debug")
			assert.Equal(t, tt.ExpectedOut.Help, opts.Help, "Unexpected Help")

			assert.Equal(t, len(tt.ExpectedErrs), len(errs), "Unexpected length of errors slice")
			for i, e := range errs {
				assert.EqualError(t, e, tt.ExpectedErrs[i], fmt.Sprintf("Unexpected error at index %d", i))
			}
		})
	}

}
