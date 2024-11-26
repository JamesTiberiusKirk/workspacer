package main

import (
	"fmt"
	"os"

	"github.com/JamesTiberiusKirk/workspacer/config"
	"github.com/JamesTiberiusKirk/workspacer/ui/codesearch"
	"github.com/JamesTiberiusKirk/workspacer/workspacer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/go-github/v66/github"
	"github.com/joho/godotenv"
)

func newGitHubClient() *github.Client {
	godotenv.Load()
	return github.NewClient(nil).WithAuthToken(os.Getenv("GITHUB_AUTH"))
}

func main() {
	gh := newGitHubClient()

	p := tea.NewProgram(
		codesearch.New(
			config.DefaultGlobalConfig.Workspaces["av"],
			config.DefaultGlobalConfig.SessionPresets,
			gh.Search.Code,
			workspacer.StartOrSwitchToSession,
			"",
		),
	)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
