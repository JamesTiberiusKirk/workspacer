package workspacer

import (
	"github.com/JamesTiberiusKirk/workspacer/config"
	"github.com/JamesTiberiusKirk/workspacer/log"
	"github.com/JamesTiberiusKirk/workspacer/ui/codelist"
	tea "github.com/charmbracelet/bubbletea"
)

// GetRepoNames fetches repository names using the configured GitHub provider
func GetRepoNames(wc config.WorkspaceConfig) ([]string, error) {
	return GetProvider(wc).GetRepoNames()
}

func GetWorkFlowsStatus(wc config.WorkspaceConfig, repo string, branches ...string) []string {
	return GetProvider(wc).GetWorkflowStatus(repo, branches)
}

func CreateGitHubRepo(wc config.WorkspaceConfig, repoName string, isPrivate bool) (string, error) {
	return GetProvider(wc).CreateRepo(repoName, isPrivate)
}

func SearchGithubInUserOrOrg(wc config.WorkspaceConfig, search string) {
	searchResults, err := GetProvider(wc).SearchCode(search)
	if err != nil {
		log.Info("Unable to do code search on github: %s", err.Error())
		return
	}

	if len(searchResults) > 10 {
		searchResults = searchResults[:10]
	}

	p := tea.NewProgram(codelist.New(searchResults, search))
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
