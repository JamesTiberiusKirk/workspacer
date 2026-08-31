package workspacer

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/JamesTiberiusKirk/workspacer/config"
	"github.com/JamesTiberiusKirk/workspacer/ui/codelist"
	"github.com/google/go-github/v66/github"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// GitHubProvider is the full set of GitHub operations workspacer needs. Both
// backends implement all of it, so `github_backend` switches every call, not
// just repo listing.
type GitHubProvider interface {
	GetRepoNames() ([]string, error)
	CreateRepo(name string, private bool) (string, error)
	SearchCode(query string) ([]codelist.SearchResult, error)
	GetWorkflowStatus(repo string, branches []string) []string
}

// GetProvider returns the appropriate GitHub provider based on the workspace config
func GetProvider(wc config.WorkspaceConfig) GitHubProvider {
	switch wc.GithubBackend {
	case config.GithubBackendCLI:
		return &CLIProvider{wc: wc}
	case config.GithubBackendAPI:
		fallthrough
	default:
		return &APIProvider{wc: wc}
	}
}

// owner is the org or user repos live under.
func (p *APIProvider) owner() string { return p.wc.GithubOrg }
func (p *CLIProvider) owner() string { return p.wc.GithubOrg }

// APIProvider talks to the GitHub API directly
type APIProvider struct{ wc config.WorkspaceConfig }

func NewAPIProvider(wc config.WorkspaceConfig) *APIProvider { return &APIProvider{wc: wc} }

// GetRepoNames fetches repository names using the GitHub GraphQL API
func (p *APIProvider) GetRepoNames() ([]string, error) {
	token, err := githubToken()
	if err != nil {
		return nil, err
	}

	login := p.owner()
	showArchived := p.wc.ShowArchivedRepos

	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)

	var allRepoNames []string
	var cursor *githubv4.String

	for {
		if p.wc.IsOrg {
			var query struct {
				Organization struct {
					Repositories struct {
						Nodes []struct {
							Name       string
							IsArchived bool
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
				if !showArchived && node.IsArchived {
					continue
				}
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
							Name       string
							IsArchived bool
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
				return nil, fmt.Errorf("GitHub GraphQL user query failed: %w", err)
			}

			for _, node := range query.User.Repositories.Nodes {
				if !showArchived && node.IsArchived {
					continue
				}
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

func (p *APIProvider) CreateRepo(name string, private bool) (string, error) {
	client, err := newGitHubClient()
	if err != nil {
		return "", err
	}

	org := ""
	if p.wc.IsOrg {
		org = p.owner()
	}

	created, _, err := client.Repositories.Create(context.Background(), org, &github.Repository{
		Name:    github.String(name),
		Private: github.Bool(private),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create repo: %w", err)
	}

	return created.GetName(), nil
}

func (p *APIProvider) SearchCode(query string) ([]codelist.SearchResult, error) {
	client, err := newGitHubClient()
	if err != nil {
		return nil, err
	}

	resp, _, err := client.Search.Code(context.Background(), query+" org:"+p.owner(), &github.SearchOptions{
		TextMatch:   true,
		ListOptions: github.ListOptions{PerPage: 100},
	})
	if err != nil {
		return nil, fmt.Errorf("code search failed: %w", err)
	}

	results := []codelist.SearchResult{}
	for _, r := range resp.CodeResults {
		snippet := ""
		if len(r.TextMatches) > 0 {
			snippet = r.TextMatches[0].GetFragment()
		}

		results = append(results, codelist.SearchResult{
			Repo:     r.Repository.GetFullName(),
			Filename: r.GetPath(),
			Snippet:  snippet,
			Language: languageOf(r.GetPath()),
		})
	}

	return results, nil
}

func (p *APIProvider) GetWorkflowStatus(repo string, branches []string) []string {
	client, err := newGitHubClient()
	if err != nil {
		return nil
	}

	result := []string{}
	for _, branch := range branches {
		runs, _, err := client.Actions.ListWorkflowRunsByFileName(context.Background(),
			p.owner(), repo, "deploy.yaml", &github.ListWorkflowRunsOptions{Branch: branch})
		if err != nil || len(runs.WorkflowRuns) == 0 {
			continue
		}

		result = append(result, branch+" "+statusEmoji(runs.WorkflowRuns[0].GetConclusion()))
	}

	return result
}

// CLIProvider uses the GitHub CLI (gh)
type CLIProvider struct{ wc config.WorkspaceConfig }

func NewCLIProvider(wc config.WorkspaceConfig) *CLIProvider { return &CLIProvider{wc: wc} }

// gh runs a gh subcommand and returns its stdout.
func (p *CLIProvider) gh(args ...string) ([]byte, error) {
	if _, err := exec.LookPath("gh"); err != nil {
		return nil, fmt.Errorf("gh CLI not found: %w", err)
	}

	cmd := exec.Command("gh", args...)
	out, err := cmd.Output()
	if err != nil {
		var stderr string
		if ee, ok := err.(*exec.ExitError); ok {
			stderr = strings.TrimSpace(string(ee.Stderr))
		}
		return nil, fmt.Errorf("gh %s failed: %w: %s", strings.Join(args, " "), err, stderr)
	}

	return out, nil
}

// GetRepoNames fetches repository names using the GitHub CLI
func (p *CLIProvider) GetRepoNames() ([]string, error) {
	out, err := p.gh("repo", "list", p.owner(), "--json", "name,isArchived", "--limit", "1000")
	if err != nil {
		return nil, err
	}

	var repos []struct {
		Name       string `json:"name"`
		IsArchived bool   `json:"isArchived"`
	}
	if err := json.Unmarshal(out, &repos); err != nil {
		return nil, fmt.Errorf("failed to parse gh CLI output: %w", err)
	}

	repoNames := []string{}
	for _, r := range repos {
		if !p.wc.ShowArchivedRepos && r.IsArchived {
			continue
		}
		// gh CLI returns full names like "owner/repo", we only want "repo"
		name := r.Name
		if i := strings.LastIndex(name, "/"); i != -1 {
			name = name[i+1:]
		}
		repoNames = append(repoNames, name)
	}

	return repoNames, nil
}

func (p *CLIProvider) CreateRepo(name string, private bool) (string, error) {
	target := name
	if p.wc.IsOrg {
		target = p.owner() + "/" + name
	}

	visibility := "--public"
	if private {
		visibility = "--private"
	}

	if _, err := p.gh("repo", "create", target, visibility); err != nil {
		return "", fmt.Errorf("failed to create repo: %w", err)
	}

	return name, nil
}

func (p *CLIProvider) SearchCode(query string) ([]codelist.SearchResult, error) {
	out, err := p.gh("search", "code", query, "--owner", p.owner(), "--limit", "100",
		"--json", "repository,path,textMatches")
	if err != nil {
		return nil, err
	}

	var hits []struct {
		Repository struct {
			NameWithOwner string `json:"nameWithOwner"`
		} `json:"repository"`
		Path        string `json:"path"`
		TextMatches []struct {
			Fragment string `json:"fragment"`
		} `json:"textMatches"`
	}
	if err := json.Unmarshal(out, &hits); err != nil {
		return nil, fmt.Errorf("failed to parse gh CLI output: %w", err)
	}

	results := []codelist.SearchResult{}
	for _, h := range hits {
		snippet := ""
		if len(h.TextMatches) > 0 {
			snippet = h.TextMatches[0].Fragment
		}

		results = append(results, codelist.SearchResult{
			Repo:     h.Repository.NameWithOwner,
			Filename: h.Path,
			Snippet:  snippet,
			Language: languageOf(h.Path),
		})
	}

	return results, nil
}

func (p *CLIProvider) GetWorkflowStatus(repo string, branches []string) []string {
	result := []string{}
	for _, branch := range branches {
		out, err := p.gh("run", "list", "--repo", p.owner()+"/"+repo,
			"--workflow", "deploy.yaml", "--branch", branch, "--limit", "1", "--json", "conclusion")
		if err != nil {
			continue
		}

		var runs []struct {
			Conclusion string `json:"conclusion"`
		}
		if err := json.Unmarshal(out, &runs); err != nil || len(runs) == 0 {
			continue
		}

		result = append(result, branch+" "+statusEmoji(runs[0].Conclusion))
	}

	return result
}

func statusEmoji(conclusion string) string {
	switch conclusion {
	case "success":
		return "🟢"
	// todo: idk if this is the string
	case "in_progress":
		return "🟡"
	default:
		return "🔴"
	}
}

func languageOf(path string) string {
	return strings.TrimPrefix(filepath.Ext(path), ".")
}
