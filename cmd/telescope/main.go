package main

import (
	"fmt"
	"os"

	"github.com/JamesTiberiusKirk/workspacer/ui/telescope"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(telescope.New())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
