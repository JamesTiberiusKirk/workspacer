package workspacer

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-github/v66/github"
)

// githubToken resolves a GitHub token, first from GITHUB_AUTH, then from the
// gh CLI. It never falls back to an anonymous client, that only turns into a
// confusing 401 further down.
func githubToken() (string, error) {
	if t := os.Getenv("GITHUB_AUTH"); t != "" {
		return t, nil
	}

	if out, err := exec.Command("gh", "auth", "token").Output(); err == nil {
		if t := strings.TrimSpace(string(out)); t != "" {
			return t, nil
		}
	}

	return "", fmt.Errorf("no GitHub auth: set GITHUB_AUTH or run `gh auth login`")
}

var ghClient *github.Client

func newGitHubClient() (*github.Client, error) {
	if ghClient != nil {
		return ghClient, nil
	}

	token, err := githubToken()
	if err != nil {
		return nil, err
	}

	ghClient = github.NewClient(nil).WithAuthToken(token)
	return ghClient, nil
}
