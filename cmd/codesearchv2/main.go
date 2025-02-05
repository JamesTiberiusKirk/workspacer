package main

import (
	"fmt"
	"strings"

	"github.com/JamesTiberiusKirk/workspacer/cmd/codesearchv2/data"
	"github.com/JamesTiberiusKirk/workspacer/cmd/codesearchv2/util"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	searchTerms  []string
	results      []data.SearchResult
	cursor       int
	selected     *data.SearchResult
	viewport     viewport.Model
	ready        bool
	resultStarts []int // Stores the starting line index of each result
	itemHeights  []int // Stores the height of each result item
}

func initialModel(results []data.SearchResult, query string) model {
	m := model{
		searchTerms: strings.Split(query, " "),
		results:     results,
		viewport: viewport.Model{
			Width:  80,
			Height: 20,
		},
	}

	m.calcSizes()

	return m
}

func (m *model) calcSizes() {
	resultStarts := make([]int, len(m.results))
	itemHeights := make([]int, len(m.results))
	totalLines := 0
	for i, result := range m.results {
		resultStarts[i] = totalLines
		result.Snippet = util.WrapText(result.Snippet, m.viewport.Width-10)
		snippetLines := strings.Split(result.Snippet, "\n")
		itemHeight := 3 + len(snippetLines) // repo (1) + filename (1) + snippet (n) + separator (1)
		itemHeights[i] = itemHeight
		totalLines += itemHeight
	}

	m.resultStarts = resultStarts
	m.itemHeights = itemHeights
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.ensureCursorVisible()
			}
		case "down", "j":
			if m.cursor < len(m.results)-1 {
				m.cursor++
				m.ensureCursorVisible()
			}
		case "enter":
			m.selected = &m.results[m.cursor]
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-4)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 4
		}
		m.calcSizes()
		m.ensureCursorVisible()
	}

	m.viewport.SetContent(m.renderResults(m.results, m.cursor))
	m.viewport, cmd = m.viewportUpdate(msg)
	return m, cmd
}

func (m *model) ensureCursorVisible() {
	if len(m.results) == 0 {
		return
	}

	cursorStartLine := m.resultStarts[m.cursor]
	itemHeight := m.itemHeights[m.cursor]

	// Calculate the maximum possible offset
	totalContentHeight := m.resultStarts[len(m.results)-1] + m.itemHeights[len(m.results)-1]
	maxOffset := totalContentHeight - m.viewport.Height
	if maxOffset < 0 {
		maxOffset = 0
	}

	// Ensure the entire item is visible
	if cursorStartLine < m.viewport.YOffset {
		// Scroll up
		m.viewport.SetYOffset(cursorStartLine)
	} else if cursorStartLine+itemHeight > m.viewport.YOffset+m.viewport.Height {
		// Scroll down
		newOffset := cursorStartLine + itemHeight - m.viewport.Height
		if newOffset > maxOffset {
			newOffset = maxOffset
		}
		m.viewport.SetYOffset(newOffset)
	}

	// Additional check to ensure the cursor is always visible
	if m.cursor == len(m.results)-1 && m.viewport.YOffset < maxOffset {
		m.viewport.SetYOffset(maxOffset)
	}

	// Ensure we don't scroll past the end of the content
	if m.viewport.YOffset > maxOffset {
		m.viewport.SetYOffset(maxOffset)
	}
}

func (m model) viewportUpdate(msg tea.Msg) (viewport.Model, tea.Cmd) {
	// Ignore mouse wheel events
	if _, ok := msg.(tea.MouseMsg); ok {
		return m.viewport, nil
	}

	if _, ok := msg.(tea.KeyMsg); ok {
		return m.viewport, nil
	}

	// For all other events, use the default viewport update
	return m.viewport.Update(msg)
}

func (m model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	header := headerStyle.Render("GitHub Search Results")

	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	footer := footerStyle.Render("↑/k: Up • ↓/j: Down • Enter: Select • q: Quit")

	return fmt.Sprintf("%s\n\n%s\n\n%s", header, m.viewport.View(), footer)
}

func (m model) renderResults(results []data.SearchResult, cursor int) string {
	var s strings.Builder

	for i, result := range results {
		cursorStr := " "
		if cursor == i {
			cursorStr = ">"
		}

		repoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Bold(true)
		fileStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Italic(true)
		lineNumStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

		highlightedRepo := util.HighlightSearchTerms(result.Repo, m.searchTerms)
		highlightedFilename := util.HighlightSearchTerms(result.Language+" "+result.Filename, m.searchTerms)

		s.WriteString(fmt.Sprintf("%s %s\n", cursorStr, repoStyle.Render(highlightedRepo)))
		s.WriteString(fmt.Sprintf("  %s:%s\n",
			fileStyle.Render(highlightedFilename),
			lineNumStyle.Render(fmt.Sprintf("%d", result.LineNum)),
		))

		highlightedSnippet, _ := util.HighlightCode(result.Snippet, result.Language)
		highlightedSnippet = util.HighlightSearchTerms(highlightedSnippet, m.searchTerms)
		snippetLines := strings.Split(highlightedSnippet, "\n")
		for _, line := range snippetLines {
			s.WriteString(fmt.Sprintf("  %s\n", line))
		}

		s.WriteString("\n")
	}

	return s.String()
}

func main() {
	p := tea.NewProgram(initialModel(data.Results, "repo func dle"), tea.WithAltScreen())

	m, err := p.Run()
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}

	if m, ok := m.(model); ok && m.selected != nil {
		fmt.Printf("Selected: %s - %s\n%s\n", m.selected.Repo, m.selected.Filename, m.selected.Snippet)
	}
}
