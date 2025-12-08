package spinner

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	timerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

type tickMsg time.Time

type Model struct {
	spinner   spinner.Model
	startTime time.Time
	message   string
	quitting  bool
}

func New(message string) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	return Model{
		spinner:   s,
		startTime: time.Now(),
		message:   message,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tickCmd(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil

	case tickMsg:
		// Update timer every tick
		if !m.quitting {
			return m, tickCmd()
		}
		return m, nil

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	elapsed := time.Since(m.startTime)
	timer := timerStyle.Render(fmt.Sprintf("(%.1fs)", elapsed.Seconds()))

	return fmt.Sprintf("\n  %s %s %s\n", m.spinner.View(), m.message, timer)
}
