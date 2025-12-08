package workspacer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/JamesTiberiusKirk/workspacer/config"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// GitHubProvider is an interface for fetching GitHub repository information
type GitHubProvider interface {
	GetRepoNames(login string, isOrg bool) ([]string, error)
}

// APIProvider uses the GitHub GraphQL API
type APIProvider struct{}

// NewAPIProvider creates a new API-based GitHub provider
func NewAPIProvider() *APIProvider {
	return &APIProvider{}
}

// GetRepoNames fetches repository names using the GitHub GraphQL API
func (p *APIProvider) GetRepoNames(login string, isOrg bool) ([]string, error) {
	token := os.Getenv("GITHUB_AUTH")
	if token == "" {
		return []string{}, nil
	}

	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)

	var allRepoNames []string
	var cursor *githubv4.String

	for {
		if isOrg {
			var query struct {
				Organization struct {
					Repositories struct {
						Nodes []struct {
							Name string
						}
						PageInfo struct {
							HasNextPage bool
							EndCursor   githubv4.String
						}
					} `graphql:"repositories(first: 100, after: $cursor)"`
				} `graphql:"organization(login: $login)"`
			}

			vars := map[string]any{
				"login":  githubv4.String(login),
				"cursor": cursor,
			}

			err := client.Query(context.Background(), &query, vars)
			if err != nil {
				return nil, fmt.Errorf("GitHub GraphQL org query failed: %w", err)
			}

			for _, node := range query.Organization.Repositories.Nodes {
				allRepoNames = append(allRepoNames, node.Name)
			}

			if !query.Organization.Repositories.PageInfo.HasNextPage {
				break
			}
			cursor = &query.Organization.Repositories.PageInfo.EndCursor

		} else {
			var query struct {
				User struct {
					Repositories struct {
						Nodes []struct {
							Name string
						}
						PageInfo struct {
							HasNextPage bool
							EndCursor   githubv4.String
						}
					} `graphql:"repositories(first: 100, after: $cursor)"`
				} `graphql:"user(login: $login)"`
			}

			vars := map[string]any{
				"login":  githubv4.String(login),
				"cursor": cursor,
			}

			err := client.Query(context.Background(), &query, vars)
			if err != nil {
				if isNetworkError(err) {
					return []string{}, nil
				}

				return nil, fmt.Errorf("GitHub GraphQL user query failed: %w", err)
			}

			for _, node := range query.User.Repositories.Nodes {
				allRepoNames = append(allRepoNames, node.Name)
			}

			if !query.User.Repositories.PageInfo.HasNextPage {
				break
			}
			cursor = &query.User.Repositories.PageInfo.EndCursor
		}
	}

	return allRepoNames, nil
}

// CLIProvider uses the GitHub CLI (gh)
type CLIProvider struct{}

// NewCLIProvider creates a new CLI-based GitHub provider
func NewCLIProvider() *CLIProvider {
	return &CLIProvider{}
}

// GetRepoNames fetches repository names using the GitHub CLI
func (p *CLIProvider) GetRepoNames(login string, isOrg bool) ([]string, error) {
	// Check if gh CLI is installed
	if _, err := exec.LookPath("gh"); err != nil {
		return nil, fmt.Errorf("gh CLI not found: %w", err)
	}

	var cmd *exec.Cmd
	if isOrg {
		cmd = exec.Command("gh", "repo", "list", login, "--json", "name", "--limit", "1000")
	} else {
		cmd = exec.Command("gh", "repo", "list", login, "--json", "name", "--limit", "1000")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("gh CLI command failed: %w\nOutput: %s", err, string(output))
	}

	// Parse JSON output
	var repos []struct {
		Name string `json:"name"`
	}

	if err := json.Unmarshal(output, &repos); err != nil {
		return nil, fmt.Errorf("failed to parse gh CLI output: %w", err)
	}

	repoNames := make([]string, len(repos))
	for i, repo := range repos {
		// gh CLI returns full names like "owner/repo", we only want "repo"
		parts := strings.Split(repo.Name, "/")
		if len(parts) > 1 {
			repoNames[i] = parts[1]
		} else {
			repoNames[i] = repo.Name
		}
	}

	return repoNames, nil
}

// GetProvider returns the appropriate GitHub provider based on the workspace config
func GetProvider(wc config.WorkspaceConfig) GitHubProvider {
	switch wc.GithubBackend {
	case config.GithubBackendCLI:
		return NewCLIProvider()
	case config.GithubBackendAPI:
		fallthrough
	default:
		return NewAPIProvider()
	}
}
