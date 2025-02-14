package workspacer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/JamesTiberiusKirk/workspacer/config"
	"github.com/JamesTiberiusKirk/workspacer/log"
	"github.com/JamesTiberiusKirk/workspacer/ui/codelist"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/go-github/v66/github"
)

var ghClient *github.Client

func newGitHubClient() *github.Client {
	if ghClient != nil {
		fmt.Println("Returning existing client")
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
		fmt.Printf("Error: %v", err)
		return
	}

	if m, ok := m.(codelist.Model); ok {
		selected := m.Selected()
		if selected == nil {
			fmt.Println("No selection")
			return
		}
		fmt.Printf("Selected: %s - %s\n%s\n", selected.Repo, selected.Filename, selected.Snippet)
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
