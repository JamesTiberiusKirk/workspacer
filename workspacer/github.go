package workspacer

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/JamesTiberiusKirk/workspacer/config"
	"github.com/JamesTiberiusKirk/workspacer/ui/codesearch"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/go-github/v66/github"
	"github.com/joho/godotenv"
)

func newGitHubClient() *github.Client {
	// add workspace to get the right token
	// temp
	_ = godotenv.Load()

	return github.NewClient(nil).WithAuthToken(os.Getenv("GITHUB_AUTH"))
}

func generateBlobURL(result *github.CodeResult, lineBegin, lineEnd int) string {
	return fmt.Sprintf("https://github.com/%s/blob/%s/%s#L%d-L%d",
		*result.Repository.FullName,
		*result.Repository.DefaultBranch,
		*result.Path,
		lineBegin,
		lineEnd,
	)
}

func getRepoDefaultBranch(owner, repoName string) string {
	client := github.NewClient(nil)
	repo, _, err := client.Repositories.Get(context.Background(), owner, repoName)
	if err != nil {
		// handle error
	}
	return repo.GetDefaultBranch()
}

// BUG: so this does not work whats so ever lol
func estimateLineNumbers(fragment *string, match *github.Match) (int, int) {
	lines := strings.Split(*fragment, "\n")
	startLine, endLine := 1, 1
	currentLine := 1
	fragmentStart := match.Indices[0]
	fragmentEnd := match.Indices[1]

	currentPos := 0
	for i, line := range lines {
		if currentPos <= fragmentStart && fragmentStart < currentPos+len(line) {
			startLine = i + 1
		}
		if currentPos <= fragmentEnd && fragmentEnd <= currentPos+len(line) {
			endLine = i + 1
			break
		}
		currentPos += len(line) + 1 // +1 for the newline character
		currentLine++
	}

	return startLine, endLine
}

func SearchGithubInUserOrOrg(userOrOrg, search string) {
	client := newGitHubClient()

	p := tea.NewProgram(
		codesearch.New(
			config.DefaultGlobalConfig.Workspaces["av"],
			config.DefaultGlobalConfig.SessionPresets,
			client.Search.Code,
			StartOrSwitchToSession,
			search,
		),
	)
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
