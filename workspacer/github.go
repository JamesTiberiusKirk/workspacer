package workspacer

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/JamesTiberiusKirk/workspacer/config"
	"github.com/JamesTiberiusKirk/workspacer/log"
	"github.com/JamesTiberiusKirk/workspacer/ui/codelist"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/go-github/v66/github"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

var ghClient *github.Client

func newGitHubClient() *github.Client {
	if ghClient != nil {
		return ghClient
	}
	ghAuth := os.Getenv("GITHUB_AUTH")
	if ghAuth != "" {
		ghClient = github.NewClient(nil).WithAuthToken(ghAuth)
	} else {
		ghClient = github.NewClient(nil)
	}
	return ghClient
}

var ghGraphQlClient *githubv4.Client

func newGitHubGraphQlClient() *githubv4.Client {
	if ghGraphQlClient != nil {
		return ghGraphQlClient
	}

	token := os.Getenv("GITHUB_AUTH")

	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(context.Background(), src)
	ghGraphQlClient = githubv4.NewClient(httpClient)

	return ghGraphQlClient
}

func isNetworkError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return true
	}

	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}
	return false
}

func GetRepoNames(login string, isOrg bool) ([]string, error) {
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

func GetReposByOrg(userOrOrg string, isOrg bool) ([]*github.Repository, error) {
	client := newGitHubClient()
	ctx := context.Background()

	var allRepos []*github.Repository
	page := 1
	perPage := 100

	for {
		var (
			repos []*github.Repository
			resp  *github.Response
			err   error
		)

		if isOrg {
			opts := &github.RepositoryListByOrgOptions{
				ListOptions: github.ListOptions{Page: page, PerPage: perPage},
			}
			repos, resp, err = client.Repositories.ListByOrg(ctx, userOrOrg, opts)
		} else {
			opts := &github.RepositoryListByUserOptions{
				ListOptions: github.ListOptions{Page: page, PerPage: perPage},
			}
			repos, resp, err = client.Repositories.ListByUser(ctx, userOrOrg, opts)
		}

		if err != nil {
			if isNetworkError(err) {
				return []*github.Repository{}, nil
			}
			return nil, fmt.Errorf("GitHub API error: %w", err)
		}

		allRepos = append(allRepos, repos...)

		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	// Filter out archived repos
	var filteredRepos []*github.Repository
	for _, repo := range allRepos {
		if !repo.GetArchived() {
			filteredRepos = append(filteredRepos, repo)
		}
	}

	return filteredRepos, nil
}

func GetMyRepos() ([]*github.Repository, error) {
	client := newGitHubClient()
	ctx := context.Background()

	opts := &github.RepositoryListByAuthenticatedUserOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allRepos []*github.Repository
	page := 1

	for {
		opts.Page = page
		repos, resp, err := client.Repositories.ListByAuthenticatedUser(ctx, opts)
		if err != nil {
			if isNetworkError(err) {
				return []*github.Repository{}, nil
			}
			return nil, fmt.Errorf("failed to list repos: %w", err)
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	// Filter out archived repos
	var filteredRepos []*github.Repository
	for _, repo := range allRepos {
		if !repo.GetArchived() {
			filteredRepos = append(filteredRepos, repo)
		}
	}

	return filteredRepos, nil
}

func SearchGithubInUserOrOrg(userOrOrg, search string) {
	wc := config.DefaultGlobalConfig.Workspaces[userOrOrg]
	client := newGitHubClient()

	searchResp, githubResp, err := client.Search.Code(context.Background(), search+" org:"+wc.GithubOrg, &github.SearchOptions{
		TextMatch: true,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	})
	if err != nil {
		log.Info("Unable to do code search on github: %s", err.Error())
		return
	}
	if githubResp.StatusCode != 200 {
		log.Info("Non 200 status code")
		return
	}

	searchResults := []codelist.SearchResult{}
	for _, result := range searchResp.CodeResults {
		language := ""
		if result.Path != nil {
			language = strings.TrimPrefix(filepath.Ext(*result.Path), ".")
		}

		searchResults = append(searchResults, codelist.SearchResult{
			Repo:     *result.Repository.FullName,
			Filename: *result.Path,
			Snippet:  *result.TextMatches[0].Fragment,
			Language: language,
		})
	}

	p := tea.NewProgram(codelist.New(searchResults[:10], search))
	m, err := p.Run()
	if err != nil {
		log.Error("Error: %v", err)
		return
	}

	if m, ok := m.(codelist.Model); ok {
		selected := m.Selected()
		if selected == nil {
			log.Error("No selection")
			return
		}
		log.Error("Selected: %s - %s\n%s\n", selected.Repo, selected.Filename, selected.Snippet)
	}
}

func GetWorkFlowsStatus(workspace string, repo string, branches ...string) []string {
	result := []string{}
	for _, branch := range branches {
		wc, ok := config.DefaultGlobalConfig.Workspaces[workspace]
		if !ok {
			continue
		}
		client := newGitHubClient()
		owner := wc.GithubOrg

		// Get the workflow runs
		workflowRuns, _, err := client.Actions.ListWorkflowRunsByFileName(context.Background(), owner, repo, "deploy.yaml", &github.ListWorkflowRunsOptions{
			Branch: branch,
		})
		if err != nil {
			continue
		}

		if len(workflowRuns.WorkflowRuns) == 0 {
			continue
		}

		// Get the latest run
		latestRun := workflowRuns.WorkflowRuns[0]

		emoji := "🔴"
		if latestRun.GetConclusion() == "success" {
			emoji = "🟢"
		}

		// todo: idk if this is the string
		if latestRun.GetConclusion() == "in_progress" {
			emoji = "🟡"
		}

		r := branch + " " + emoji
		result = append(result, r)
	}

	return result
}

func GetOpenPullRequestsByBranch(ws config.WorkspaceConfig, project, branch string) ([]*github.PullRequest, error) {
	client := newGitHubClient()
	opts := &github.PullRequestListOptions{
		State: "open",
		Head:  branch,
	}

	prs, _, err := client.PullRequests.List(context.Background(), ws.GithubOrg, project, opts)
	if err != nil {
		return nil, err
	}

	return prs, nil
}

func CreateGitHubRepo(ws config.WorkspaceConfig, repoName string, isPrivate bool) (string, error) {
	client := newGitHubClient()
	ctx := context.Background()

	repo := &github.Repository{
		Name:    github.String(repoName),
		Private: github.Bool(isPrivate),
	}

	var createdRepo *github.Repository
	var resp *github.Response
	var err error

	if ws.IsOrg {
		// Create under an organization
		createdRepo, resp, err = client.Repositories.Create(ctx, ws.GithubOrg, repo)
	} else {
		// Create under the authenticated user
		createdRepo, resp, err = client.Repositories.Create(ctx, "", repo)
	}

	if err != nil {
		return "", fmt.Errorf("failed to create repo: %w (status: %d)", err, resp.StatusCode)
	}

	// sshURL := createdRepo.GetSSHURL()
	// if sshURL == "" {
	// 	return "", fmt.Errorf("repository created, but SSH URL is empty")
	// }

	return *createdRepo.Name, nil
}
