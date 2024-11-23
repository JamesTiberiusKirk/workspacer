package telescope

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/JamesTiberiusKirk/workspacer/config"
	"github.com/JamesTiberiusKirk/workspacer/workspacer"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v66/github"
	"github.com/joho/godotenv"

	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

// TODO:
// - [ ] clean this up
// - [ ] pull out the search logic outside of this package
// - [x] handle line overflow/wrapping
// - [x] highlight search terms
// - [x] fix issue where some of the code snippets (usually markdown or comments) get highlighted on selection

// BUG:
// - [ ] using the highlight style breaks the rest of the code highlighting
// - [ ] enabling the filder blanks the screen out, then existing brings back the normal screen with stuff in the input

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

type Model struct {
	query              string
	results            []searchResult
	cursor             int
	searchInput        textinput.Model
	resultsFilterInput textinput.Model
	filterEnabled      bool
	err                error
	viewport           viewport.Model
	itemHeight         int
	visibleItemCount   int
	state              inputState
}

type searchResult struct {
	repo     string
	file     string
	content  string
	language string
}

func New() Model {
	ti := textinput.New()
	ti.Placeholder = "Enter GitHub search query..."
	ti.Focus()

	ti.PromptStyle = lipgloss.NewStyle().Foreground(highlight)
	ti.TextStyle = lipgloss.NewStyle().Foreground(special)

	return Model{
		searchInput:        ti,
		resultsFilterInput: textinput.New(),
		viewport:           viewport.New(0, 0),
		itemHeight:         12, // Adjust based on your actual item height
		state:              stateInput,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

type returnToSearchMsg struct{}
type resultsFilterEnabled struct{}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case stateInput:
			switch msg.String() {
			case "ctrl+c", "esc":
				return m, tea.Quit
			case "enter":
				m.query = m.searchInput.Value()
				m.state = stateResults
				return m, m.search
			default:
				m.searchInput, cmd = m.searchInput.Update(msg)
			}
		case stateResults:
			switch msg.String() {
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
				m.resultsFilterInput.Focus()
				return m, textinput.Blink
				// return m, func() tea.Msg { return resultsFilterEnabled{} }
			case "enter":
				workspacer.StartOrSwitchToSession(
					"av",
					config.DefaultGlobalConfig.Workspaces["av"],
					config.DefaultGlobalConfig.SessionPresets,
					strings.TrimPrefix(m.results[m.cursor].repo, "aviva-verde/")+":"+m.results[m.cursor].file+":22",
				)
				return m, tea.Quit
			}
		case stateResultsFilter:
			switch msg.String() {
			case "ctrl+c", "esc", "enter":
				m.state = stateResults
				m.filterEnabled = false
				m.resultsFilterInput.Blur()
			default:
				m.resultsFilterInput.Update(msg)
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
		m.resultsFilterInput.Focus()

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
		m.viewport.Width = msg.Width - 2
		m.viewport.Height = msg.Height - headerHeight - footerHeight
		m.visibleItemCount = m.viewport.Height / m.itemHeight

	}

	switch m.state {
	case stateInput:
		m.searchInput, cmd = m.searchInput.Update(msg)
	// case stateResults:
	case stateResultsFilter:
		m.resultsFilterInput, cmd = m.resultsFilterInput.Update(msg)
	default:
		m.viewport.SetContent(m.viewportContent())
		m.viewport, cmd = m.viewport.Update(msg)
	}

	// if m.state == stateInput {
	// 	m.searchInput, cmd = m.searchInput.Update(msg)
	// } else {
	// 	m.viewport.SetContent(m.viewportContent())
	// 	m.viewport, cmd = m.viewport.Update(msg)
	// }

	return m, cmd
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

func (m Model) View() string {

	width, height := m.viewport.Width, m.viewport.Height

	switch m.state {
	case stateInput:
		var s strings.Builder
		s.WriteString(m.searchInput.View() + "\n\n")
		s.WriteString(infoStyle.Render("Press Enter to search, Ctrl+C to quit"))
		return lipgloss.JoinVertical(lipgloss.Left, s.String())
	case stateResults:

		searchQueryStyle := lipgloss.NewStyle().
			Foreground(special).
			Italic(true)

		titleBar := titleStyle.Render("GitHub Code Search: ", searchQueryStyle.Render(m.query))
		footerBar := m.createStatusLine(width)
		mainContent := borderStyle.Render(m.viewport.View())

		return lipgloss.JoinVertical(
			lipgloss.Left,
			titleBar,
			lipgloss.NewStyle().Height(height-1).Render(mainContent),
			footerBar,
		)
	}

	return ""
}

func (m Model) createStatusLine(width int) string {
	width = width - 2
	left := "Press ↑/↓ to navigate, q to search again, Ctrl+C to quit " + m.resultsFilterInput.View()
	right := ""

	// if m.filterEnabled {
	// right =
	// }

	// Create status line content here
	left = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render(left)
	right = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render(right)

	// Calculate remaining space for the center section
	remainingSpace := width - lipgloss.Width(left) - lipgloss.Width(right)
	center := strings.Repeat(" ", remainingSpace)

	return footerBarStyle.Render(lipgloss.JoinHorizontal(lipgloss.Left, left, center, right))
}

func newGitHubClient() *github.Client {
	godotenv.Load()
	return github.NewClient(nil).WithAuthToken(os.Getenv("GITHUB_AUTH"))
}

func (m Model) search() tea.Msg {
	client := newGitHubClient()

	searchResp, githubResp, err := client.Search.Code(context.Background(), m.query+" org:aviva-verde", &github.SearchOptions{
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

	var searchResults []searchResult
	for _, result := range searchResp.CodeResults {
		language := ""
		if result.Path != nil {
			language = strings.TrimPrefix(filepath.Ext(*result.Path), ".")
		}

		searchResults = append(searchResults, searchResult{
			repo:     *result.Repository.FullName,
			file:     *result.Path,
			content:  *result.TextMatches[0].Fragment,
			language: language,
		})
	}

	return searchResultsMsg{results: searchResults}
}

func (m Model) viewportContent() string {
	var s strings.Builder
	if m.err != nil {
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(fmt.Sprintf("Error: %v\n\n", m.err)))
	}
	if len(m.results) > 0 {
		for i, result := range m.results {
			style := normalStyle
			if m.cursor == i {
				style = selectedStyle
			}
			repoLine := urlStyle.Render(result.repo)
			fileLine := infoStyle.Render("File: " + result.file)
			wrappedCode := wrapText(result.content, m.viewport.Width-10) // Subtract padding
			highlightedCode := highlightCode(m.query, wrappedCode, result.language)
			contentLine := style.Render(highlightedCode)
			resultBox := lipgloss.JoinVertical(lipgloss.Left,
				repoLine,
				fileLine,
				contentLine,
			)
			s.WriteString(resultStyle.Render(resultBox) + "\n\n")
		}
	}
	return s.String()
}

func wrapText(text string, width int) string {
	lines := strings.Split(text, "\n")
	var wrappedLines []string

	for _, line := range lines {
		if len(line) <= width {
			wrappedLines = append(wrappedLines, line)
			continue
		}

		wrappedLines = append(wrappedLines, line[0:width-1])
		wl := line[width:]
		for {
			if len(wl) < width {
				wrappedLines = append(wrappedLines, wl)
				break
			}
			wl = wl[:width-1]
			wrappedLines = append(wrappedLines, wl)
		}

	}

	return strings.Join(wrappedLines, "\n")
}

func highlightCode(query, code, language string) string {
	hightlightTokens := strings.Split(query, " ")

	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code
	}
	var buf strings.Builder
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return code
	}
	syntaxHightlited := buf.String()

	for _, token := range hightlightTokens {
		re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(token))
		syntaxHightlited = re.ReplaceAllStringFunc(syntaxHightlited, func(match string) string {
			return highlightStyle.Render(match)
		})
	}

	return syntaxHightlited
}

type searchResultsMsg struct {
	results []searchResult
	err     error
}
