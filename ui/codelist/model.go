package codelist

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SearchResult struct {
	Language string
	Repo     string
	Filename string
	Snippet  string
	LineNum  int
}

type Model struct {
	searchTerms       []string
	results           []SearchResult
	filteredResults   []SearchResult
	cursor            int
	selected          *SearchResult
	viewport          viewport.Model
	ready             bool
	resultStarts      []int // Stores the starting line index of each result
	itemHeights       []int // Stores the height of each result item
	filterActive      bool
	filterInput       string
	lastAppliedFilter string
}

func New(results []SearchResult, query string) Model {
	m := Model{
		searchTerms:     strings.Split(query, " "),
		results:         results,
		filteredResults: results,
		viewport: viewport.Model{
			Width:  80,
			Height: 20,
		},
	}

	m.calcSizes()

	return m
}

func (m *Model) calcSizes() {
	resultStarts := make([]int, len(m.filteredResults))
	itemHeights := make([]int, len(m.filteredResults))
	totalLines := 0
	for i, result := range m.filteredResults {
		resultStarts[i] = totalLines
		result.Snippet = wrapText(result.Snippet, m.viewport.Width-10)
		snippetLines := strings.Split(result.Snippet, "\n")
		itemHeight := 3 + len(snippetLines) // repo (1) + filename (1) + snippet (n) + separator (1)
		itemHeights[i] = itemHeight
		totalLines += itemHeight
	}

	m.resultStarts = resultStarts
	m.itemHeights = itemHeights
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.filterActive {
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.filterActive = false
				m.lastAppliedFilter = ""
				m.filteredResults = m.results
				m.cursor = 0
				m.calcSizes()
			case "enter":
				m.filterActive = false
				m.lastAppliedFilter = m.filterInput
				m.filteredResults = m.filterResults(m.results, m.lastAppliedFilter)
				m.cursor = 0
				m.calcSizes()
			case "backspace":
				if len(m.filterInput) > 0 {
					m.filterInput = m.filterInput[:len(m.filterInput)-1]
					m.filteredResults = m.filterResults(m.results, m.filterInput)
					m.cursor = 0
					m.calcSizes()
				}
			default:
				m.filterInput += msg.String()
				m.filteredResults = m.filterResults(m.results, m.filterInput)
				m.cursor = 0
				m.calcSizes()
			}
		} else {
			switch msg.String() {
			case "/":
				m.filterActive = true
				m.filterInput = m.lastAppliedFilter
			case "q", "ctrl+c":
				return m, tea.Quit
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
					m.ensureCursorVisible()
				}
			case "down", "j":
				if m.cursor < len(m.filteredResults)-1 {
					m.cursor++
					m.ensureCursorVisible()
				}
			case "enter":
				if len(m.filteredResults) > 0 {
					m.selected = &m.filteredResults[m.cursor]
					return m, tea.Quit
				}
			}
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

	m.viewport.SetContent(m.renderResults(m.filteredResults, m.cursor))
	m.viewport, cmd = m.viewportUpdate(msg)
	return m, cmd
}

func (m *Model) Selected() *SearchResult {
	return m.selected
}

func (m *Model) filterResults(results []SearchResult, filter string) []SearchResult {
	if filter == "" {
		return results
	}
	var filtered []SearchResult
	filterLower := strings.ToLower(filter)
	for _, result := range results {
		if strings.Contains(strings.ToLower(result.Repo), filterLower) ||
			strings.Contains(strings.ToLower(result.Filename), filterLower) ||
			strings.Contains(strings.ToLower(result.Snippet), filterLower) {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

func (m *Model) ensureCursorVisible() {
	if len(m.filteredResults) == 0 {
		return
	}

	cursorStartLine := m.resultStarts[m.cursor]
	itemHeight := m.itemHeights[m.cursor]

	// Calculate the maximum possible offset
	totalContentHeight := m.resultStarts[len(m.filteredResults)-1] + m.itemHeights[len(m.filteredResults)-1]
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
	if m.cursor == len(m.filteredResults)-1 && m.viewport.YOffset < maxOffset {
		m.viewport.SetYOffset(maxOffset)
	}

	// Ensure we don't scroll past the end of the content
	if m.viewport.YOffset > maxOffset {
		m.viewport.SetYOffset(maxOffset)
	}
}

func (m Model) viewportUpdate(msg tea.Msg) (viewport.Model, tea.Cmd) {
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

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	header := headerStyle.Render("GitHub Search Results")

	var footer string
	if m.filterActive {
		filterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
		footer = filterStyle.Render(fmt.Sprintf("Filter (%d results): %s", len(m.filteredResults), m.filterInput))
	} else {
		footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		footer = footerStyle.Render("↑/k: Up • ↓/j: Down • Enter: Select • /: Filter • q: Quit")
	}

	return fmt.Sprintf("%s\n\n%s\n\n%s", header, m.viewport.View(), footer)
}

func (m Model) renderResults(results []SearchResult, cursor int) string {
	var s strings.Builder

	// Determine which filter to use for highlighting
	activeFilter := m.lastAppliedFilter
	if m.filterActive {
		activeFilter = m.filterInput
	}

	for i, result := range results {
		cursorStr := " "
		if cursor == i {
			cursorStr = ">"
		}

		repoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Bold(true)
		fileStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Italic(true)
		lineNumStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

		highlightedRepo := highlightFilteredText(repoStyle.Render(result.Repo), m.searchTerms, activeFilter)
		highlightedFilename := highlightFilteredText(fileStyle.Render(result.Language+" "+result.Filename), m.searchTerms, activeFilter)

		s.WriteString(fmt.Sprintf("%s %s\n", cursorStr, highlightedRepo))
		s.WriteString(fmt.Sprintf("  %s:%s\n",
			highlightedFilename,
			lineNumStyle.Render(fmt.Sprintf("%d", result.LineNum)),
		))

		highlightedSnippet, _ := highlightCode(result.Snippet, result.Language)
		highlightedSnippet = highlightFilteredText(highlightedSnippet, m.searchTerms, activeFilter)
		snippetLines := strings.Split(highlightedSnippet, "\n")
		for _, line := range snippetLines {
			s.WriteString(fmt.Sprintf("  %s\n", line))
		}

		s.WriteString("\n")
	}

	return s.String()
}
