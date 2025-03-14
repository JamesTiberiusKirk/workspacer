package codesearch

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/JamesTiberiusKirk/workspacer/config"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v66/github"
)

// TODO:
// - [ ] clean this up
// - [x] pull out the search logic outside of this package
// - [x] handle line overflow/wrapping
// - [x] highlight search terms
// - [x] fix issue where some of the code snippets (usually markdown or comments) get highlighted on selection
// - [x] make the filter carrot me a / vim style search and only appear when the user presses /
// - [ ] Figure out the line number situation with fragment
// - [ ] Rework the selected sytles, maybe just make the border a different colour
// - [ ] Fix row length so that it doesn't look odd
// - [ ] Add loading spinner
// - [ ] Get file numbers

// BUG:
// - [ ] filter input cursor does not blink
// - [x] enabling the filder blanks the screen out, then existing brings back the normal screen with stuff in the input
// - [ ] work on the partly functioning viewport selection scroll
// - [ ] fix filtering not highlighting over the highlighted search terms
// - [x] sometimes this still breaks the special terminal input "ctrl+c"
// - [ ] fix the arg start search
//	- NOTE: Seems like os.Exist is the culprit here....

var (
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	urlStyle = lipgloss.NewStyle().Foreground(special)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#DDDDDD"})

	selectedStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			Foreground(highlight).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(subtle)

	resultStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(subtle).
			Padding(1, 1, 1, 2)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(highlight)

	highlightStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("205")). // Pink background
			Foreground(lipgloss.Color("#ffffff")).
			Bold(true)

	filterStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#ff2d00")). // Pink background
			Foreground(lipgloss.Color("#ffffff")).
			Bold(true)

	footerBarStyle = lipgloss.NewStyle().
			Foreground(highlight).
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(special)

	titleStyle = lipgloss.NewStyle().
			Foreground(highlight).
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(special)
)

type inputState int

const (
	stateInput inputState = iota
	stateResults
	stateResultsFilter
)

type (
	returnToSearchMsg    struct{}
	resultsFilterEnabled struct{}
	searchResultsMsg     struct {
		results []githubCodeSearchResult
		err     error
	}
)

type Model struct {
	searchInOrg            githubCodeSearchFunc
	startOrSwitchToSession startOrSwitchToSessionFunc

	searchInput textinput.Model
	filterInput textinput.Model
	viewport    viewport.Model

	width, height      int
	query              string
	results            []githubCodeSearchResult
	cursor             int
	filterEnabled      bool
	err                error
	itemHeight         int
	visibleItemCount   int
	state              inputState
	clearFilterConfirm bool
	startSearch        string

	wc config.WorkspaceConfig
	sp map[string]config.SessionConfig
}

func New(
	wc config.WorkspaceConfig,
	sp map[string]config.SessionConfig,
	searchInOrg githubCodeSearchFunc,
	startOrSwitchToSession startOrSwitchToSessionFunc,
	argSearchString string,
) Model {
	ti := textinput.New()
	ti.Placeholder = "Enter GitHub search query..."
	ti.PromptStyle = lipgloss.NewStyle().Foreground(highlight)
	ti.TextStyle = lipgloss.NewStyle().Foreground(special)

	state := stateInput
	if argSearchString != "" {
		state = stateResults
		ti.SetValue(argSearchString)
	} else {
		ti.Focus()
	}

	filterInput := textinput.New()
	filterInput.Prompt = ""

	return Model{
		searchInOrg:            searchInOrg,
		startOrSwitchToSession: startOrSwitchToSession,

		searchInput: ti,
		filterInput: filterInput,
		viewport:    viewport.New(0, 0),
		itemHeight:  12, // Adjust based on your actual item height
		startSearch: argSearchString,
		state:       state,

		wc: wc,
		sp: sp,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.startSearch != "" {
		m.searchInput.SetValue(m.startSearch)
		m.query = m.startSearch
		m.state = stateResults
		m.startSearch = ""

		// m.viewport.SetContent(m.viewportContent())
		// m.viewport, cmd = m.viewport.Update(msg)
		m.searchInput, cmd = m.searchInput.Update(msg)

		return m, m.search
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case stateInput:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc":
				return m, tea.Quit
			case "enter":
				m.query = m.searchInput.Value()
				m.state = stateResults
				return m, m.search
			}
		case stateResults:
			switch msg.String() {
			case "esc":
				if m.clearFilterConfirm {
					m.clearFilterConfirm = true
				} else {
					m.clearFilterConfirm = false
					m.filterInput.SetValue("")
				}
			case "ctrl+c":
				return m, tea.Quit
			case "q":
				return m, func() tea.Msg { return returnToSearchMsg{} }
			case "up", "k":
				m.updateCursor(-1)
			case "down", "j":
				m.updateCursor(1)
			case "pgup":
				m.updateCursor(-m.visibleItemCount)
			case "pgdown":
				m.updateCursor(m.visibleItemCount)
			case "G":
				m.cursor = len(m.results) - 1
				m.viewport.GotoBottom()
			case "/":
				m.state = stateResultsFilter
				m.filterEnabled = true
				m.filterInput.Prompt = "/"
				// m.filterInput.Cursor.Blink = true
				// m.filterInput.Cursor.BlinkSpeed = time.Second * 1
				// m.filterInput.Cursor.Focus()
				m.filterInput.Focus()
				return m, textinput.Blink
			case "enter":
				m.startOrSwitchToSession(
					m.wc.Prefix,
					m.wc,
					m.sp,
					strings.TrimPrefix(m.results[m.cursor].repo, m.wc.GithubOrg+"/")+
						":"+m.results[m.cursor].file+": -c '/"+m.query+"'",
				)
				return m, tea.Quit
			}
		case stateResultsFilter:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc", "enter":
				m.state = stateResults
				m.filterEnabled = false
				m.filterInput.Prompt = ""
				m.filterInput.Blur()
			}

		}

	case searchResultsMsg:
		m.results = msg.results
		m.err = msg.err
		m.cursor = 0
		m.viewport.GotoTop()

	case resultsFilterEnabled:
		m.state = stateResultsFilter
		m.filterEnabled = true
		m.filterInput.Focus()

	case returnToSearchMsg:
		m.state = stateInput
		m.results = nil
		m.cursor = 0
		m.viewport.GotoTop()
		m.searchInput.SetValue("")
		m.searchInput.Focus()

	case tea.WindowSizeMsg:
		headerHeight := 4 // Adjust based on your header content
		footerHeight := 4 // Adjust based on your footer content
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 2
		m.viewport.Height = msg.Height - headerHeight - footerHeight
		m.visibleItemCount = m.viewport.Height / m.itemHeight

	}

	switch m.state {
	case stateInput:
		m.searchInput, cmd = m.searchInput.Update(msg)
	case stateResultsFilter:
		m.filterInput, cmd = m.filterInput.Update(msg)
		m.viewport.SetContent(m.viewportContent())
		m.viewport, cmd = m.viewport.Update(msg)
	default:
		m.viewport.SetContent(m.viewportContent())
		m.viewport, cmd = m.viewport.Update(msg)
	}

	return m, cmd
}

func (m Model) View() string {
	switch m.state {
	case stateInput:
		var s strings.Builder
		s.WriteString(m.searchInput.View() + "\n\n")
		s.WriteString(infoStyle.Render("Press Enter to search, Ctrl+C to quit"))
		return lipgloss.JoinVertical(lipgloss.Left, s.String())
	case stateResults, stateResultsFilter:
		searchQueryStyle := lipgloss.NewStyle().
			Foreground(special).
			Italic(true)
		titleBar := titleStyle.Render("GitHub Code Search: ", searchQueryStyle.Render(m.query))
		footerBar := m.createStatusLine()
		mainContent := borderStyle.Render(m.viewport.View())
		return lipgloss.JoinVertical(
			lipgloss.Left,
			titleBar,
			lipgloss.NewStyle().Height(m.viewport.Height-1).Render(mainContent),
			footerBar,
		)
	}
	return ""
}

func (m *Model) updateCursor(direction int) {
	newCursor := m.cursor + direction

	// Ensure the new cursor position is within bounds
	if newCursor < 0 {
		newCursor = 0
	} else if newCursor >= len(m.results) {
		newCursor = len(m.results) - 1
	}

	// Calculate the current viewport boundaries
	topBoundary := m.viewport.YOffset / m.itemHeight
	bottomBoundary := topBoundary + m.visibleItemCount - 1

	// Define a buffer zone (e.g., 2 items from top and bottom)
	bufferZone := 2

	// Adjust viewport if necessary
	if newCursor <= topBoundary+bufferZone {
		// Scroll up, keeping the cursor bufferZone items from the top
		newOffset := (newCursor - bufferZone) * m.itemHeight
		if newOffset < 0 {
			newOffset = 0
		}
		m.viewport.SetYOffset(newOffset)
	} else if newCursor >= bottomBoundary-bufferZone {
		// Scroll down, keeping the cursor bufferZone items from the bottom
		newOffset := (newCursor - m.visibleItemCount + bufferZone + 1) * m.itemHeight
		if newOffset < 0 {
			newOffset = 0
		}
		m.viewport.SetYOffset(newOffset)
	}

	m.cursor = newCursor
}

func (m *Model) createStatusLine() string {
	width := m.width - 2
	left := "Press ↑/↓ to navigate, q to search again, Ctrl+C to quit " + m.filterInput.View()
	right := ""

	// Create status line content here
	left = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render(left)
	right = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render(right)

	// Calculate remaining space for the center section
	remainingSpace := width - (lipgloss.Width(left) + lipgloss.Width(right))

	if remainingSpace < 0 {
		remainingSpace = 0
	}
	center := strings.Repeat(" ", remainingSpace)

	return footerBarStyle.Render(lipgloss.JoinHorizontal(lipgloss.Left, left, center, right))
}

func (m *Model) viewportContent() string {
	var s strings.Builder
	if m.err != nil {
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(fmt.Sprintf("Error: %v\n\n", m.err)))
	}

	filterText := ""

	if len(m.results) == 0 {
		s.WriteString(infoStyle.Render("No results. Press 'q' to search again."))
		return s.String()
	}

	filterText = strings.ToLower(m.filterInput.Value())
	for i, result := range m.results {
		if filterText != "" && !strings.Contains(strings.ToLower(result.ToFilterString()), filterText) {
			continue
		}

		style := normalStyle
		if m.cursor == i {
			style = selectedStyle
		}

		repoLine := urlStyle.Render(result.repo)
		repoLine = highlightFilterText(repoLine, filterText)
		repoLine = highlightGHQuery(repoLine, m.query)

		fileLine := infoStyle.Render("File: " + result.file)
		fileLine = highlightFilterText(fileLine, filterText)
		fileLine = highlightGHQuery(fileLine, m.query)

		codeText := ""
		codeText = wrapText(result.content, m.width-10)
		codeText = highlightCode(codeText, result.language)
		codeText = highlightFilterText(codeText, filterText)
		codeText = highlightGHQuery(codeText, m.query)
		// codeText = style.Render(codeText)

		resultBox := lipgloss.JoinVertical(lipgloss.Left,
			repoLine,
			fileLine,
			codeText,
		)
		s.WriteString(style.Render(resultStyle.Render(resultBox)) + "\n\n")

	}

	return s.String()
}

func (m *Model) search() tea.Msg {
	// searchResp, githubResp, err := client.Search.Code(context.Background(), m.query+" org:aviva-verde", &github.SearchOptions{
	searchResp, githubResp, err := m.searchInOrg(context.Background(), m.query+" org:"+m.wc.GithubOrg, &github.SearchOptions{
		TextMatch: true,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	})
	if err != nil {
		return searchResultsMsg{err: err}
	}
	if githubResp.StatusCode != 200 {
		return searchResultsMsg{err: fmt.Errorf("GitHub API error: %s", githubResp.Status)}
	}

	var searchResults []githubCodeSearchResult
	for _, result := range searchResp.CodeResults {
		language := ""
		if result.Path != nil {
			language = strings.TrimPrefix(filepath.Ext(*result.Path), ".")
		}

		searchResults = append(searchResults, githubCodeSearchResult{
			repo:     *result.Repository.FullName,
			file:     *result.Path,
			content:  *result.TextMatches[0].Fragment,
			language: language,
		})
	}

	return searchResultsMsg{results: searchResults}
}
