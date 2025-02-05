package main

import (
	"fmt"

	"github.com/JamesTiberiusKirk/workspacer/cmd/codesearchv2/data"
	"github.com/JamesTiberiusKirk/workspacer/ui/codelist"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(codelist.New(data.Results, "complex exa uti"), tea.WithAltScreen())

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
