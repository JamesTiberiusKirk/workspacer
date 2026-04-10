package newwizard

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/JamesTiberiusKirk/workspacer/config"
	"github.com/JamesTiberiusKirk/workspacer/util"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Result is what the wizard returns to the caller.
type Result struct {
	Name      string
	GitHub    bool
	Private   bool
	Cancelled bool
}

// Run starts the wizard and blocks until the user confirms or cancels.
func Run(wc config.WorkspaceConfig) (Result, error) {
	m := newModel(wc)
	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		return Result{}, err
	}
	fm, ok := final.(model)
	if !ok {
		return Result{Cancelled: true}, nil
	}
	if fm.cancelled || !fm.confirmed {
		return Result{Cancelled: true}, nil
	}
	return Result{
		Name:    strings.TrimSpace(fm.nameInput.Value()),
		GitHub:  fm.github,
		Private: fm.private,
	}, nil
}

type step int

const (
	stepName step = iota
	stepGitHub
	stepPrivate
	stepSummary
)

var (
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	danger    = lipgloss.AdaptiveColor{Light: "#C4314B", Dark: "#FF5C7A"}

	titleStyle = lipgloss.NewStyle().
			Foreground(highlight).
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(special)

	questionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#DDDDDD"}).
			Bold(true)

	optionStyle = lipgloss.NewStyle().
			Foreground(subtle).
			PaddingLeft(2)

	selectedOptionStyle = lipgloss.NewStyle().
				Foreground(highlight).
				Bold(true).
				PaddingLeft(2)

	errorStyle = lipgloss.NewStyle().
			Foreground(danger).
			Italic(true)

	hintStyle = lipgloss.NewStyle().
			Foreground(subtle).
			Italic(true)

	summaryKeyStyle = lipgloss.NewStyle().
			Foreground(subtle).
			Width(12)

	summaryValStyle = lipgloss.NewStyle().
			Foreground(special).
			Bold(true)
)

type model struct {
	wc config.WorkspaceConfig

	step step

	github  bool
	private bool
	// privateSet tracks whether the user has made an explicit choice on the
	// private step, so that flipping github no->yes->no->yes restores it.
	privateSet bool

	nameInput textinput.Model
	nameErr   string

	// yesNoCursor is used on the github / private / summary yes-no screens.
	// 0 = yes, 1 = no (or confirm/cancel on summary).
	yesNoCursor int

	confirmed bool
	cancelled bool
}

func newModel(wc config.WorkspaceConfig) model {
	ti := textinput.New()
	ti.Placeholder = "project-name"
	ti.Prompt = "» "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(highlight)
	ti.TextStyle = lipgloss.NewStyle().Foreground(special)
	ti.CharLimit = 100
	ti.Width = 40
	ti.Focus()

	return model{
		wc:          wc,
		step:        stepName,
		yesNoCursor: 0,
		nameInput:   ti,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global cancel.
		if msg.Type == tea.KeyCtrlC {
			m.cancelled = true
			return m, tea.Quit
		}

		switch m.step {
		case stepGitHub:
			return m.updateGitHub(msg)
		case stepPrivate:
			return m.updatePrivate(msg)
		case stepName:
			return m.updateName(msg)
		case stepSummary:
			return m.updateSummary(msg)
		}
	}
	return m, nil
}

func (m model) updateName(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// On the first step esc cancels the whole wizard.
		m.cancelled = true
		return m, tea.Quit
	case "enter":
		name := strings.TrimSpace(m.nameInput.Value())
		if err := m.validateName(name); err != "" {
			m.nameErr = err
			return m, nil
		}
		m.nameErr = ""
		m.nameInput.Blur()
		m.step = stepGitHub
		if m.github {
			m.yesNoCursor = 0
		} else {
			m.yesNoCursor = 1
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(msg)
	// Clear stale error as the user types.
	m.nameErr = ""
	return m, cmd
}

func (m model) updateGitHub(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.yesNoCursor > 0 {
			m.yesNoCursor--
		}
	case "down", "j":
		if m.yesNoCursor < 1 {
			m.yesNoCursor++
		}
	case "y", "Y":
		m.yesNoCursor = 0
	case "n", "N":
		m.yesNoCursor = 1
	case "esc":
		m.step = stepName
		m.nameInput.Focus()
		return m, textinput.Blink
	case "enter":
		m.github = m.yesNoCursor == 0
		if m.github {
			m.step = stepPrivate
			// Restore cursor to previous choice or default to "no".
			if m.privateSet {
				if m.private {
					m.yesNoCursor = 0
				} else {
					m.yesNoCursor = 1
				}
			} else {
				m.yesNoCursor = 1
			}
		} else {
			m.step = stepSummary
			m.yesNoCursor = 0
		}
	}
	return m, nil
}

func (m model) updatePrivate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.yesNoCursor > 0 {
			m.yesNoCursor--
		}
	case "down", "j":
		if m.yesNoCursor < 1 {
			m.yesNoCursor++
		}
	case "y", "Y":
		m.yesNoCursor = 0
	case "n", "N":
		m.yesNoCursor = 1
	case "esc":
		m.step = stepGitHub
		// Restore cursor to previous github choice.
		if m.github {
			m.yesNoCursor = 0
		} else {
			m.yesNoCursor = 1
		}
	case "enter":
		m.private = m.yesNoCursor == 0
		m.privateSet = true
		m.step = stepSummary
		m.yesNoCursor = 0
	}
	return m, nil
}

func (m model) updateSummary(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.yesNoCursor > 0 {
			m.yesNoCursor--
		}
	case "down", "j":
		if m.yesNoCursor < 1 {
			m.yesNoCursor++
		}
	case "esc":
		// Go back to the most recent prior step.
		if m.github {
			m.step = stepPrivate
			if m.private {
				m.yesNoCursor = 0
			} else {
				m.yesNoCursor = 1
			}
		} else {
			m.step = stepGitHub
			m.yesNoCursor = 1 // we just came from github=no
		}
		return m, nil
	case "enter":
		if m.yesNoCursor == 0 {
			m.confirmed = true
		} else {
			m.cancelled = true
		}
		return m, tea.Quit
	}
	return m, nil
}

// validateName returns an empty string if valid, otherwise a human-readable
// error message.
func (m model) validateName(name string) string {
	if name == "" {
		return "name cannot be empty"
	}
	if strings.ContainsAny(name, `/\`) {
		return "name cannot contain / or \\"
	}
	if strings.Contains(name, "..") {
		return "name cannot contain .."
	}
	if strings.HasPrefix(name, ".") {
		return "name cannot start with ."
	}
	if util.DoesProjectExist(m.wc, name) {
		return fmt.Sprintf("a project named %q already exists in this workspace", name)
	}
	if err := validateGitHubName(name); err != "" {
		return err
	}
	return ""
}

// GitHub repo name rules: alphanumerics, '-', '_', '.', max 100 chars,
// not starting or ending with '.' or '-'.
var githubNameRe = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

func validateGitHubName(name string) string {
	if len(name) > 100 {
		return "github name cannot exceed 100 characters"
	}
	if !githubNameRe.MatchString(name) {
		return "github name may only contain letters, digits, '.', '_', '-'"
	}
	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		return "github name cannot start or end with '-'"
	}
	if strings.HasSuffix(name, ".") {
		return "github name cannot end with '.'"
	}
	return ""
}

// ---------- View ----------

func (m model) View() string {
	var body string
	switch m.step {
	case stepGitHub:
		body = m.viewGitHub()
	case stepPrivate:
		body = m.viewPrivate()
	case stepName:
		body = m.viewName()
	case stepSummary:
		body = m.viewSummary()
	}

	title := titleStyle.Render("workspacer › new project")
	return fmt.Sprintf("\n%s\n\n%s\n", title, body)
}

func (m model) viewGitHub() string {
	q := questionStyle.Render("Create a remote GitHub repository?")
	yes := renderOption("yes", m.yesNoCursor == 0)
	no := renderOption("no", m.yesNoCursor == 1)
	hint := hintStyle.Render("enter: continue • esc: back • ↑/↓ or y/n")
	return fmt.Sprintf("%s\n\n%s\n%s\n\n%s", q, yes, no, hint)
}

func (m model) viewPrivate() string {
	q := questionStyle.Render("Make the GitHub repository private?")
	yes := renderOption("yes", m.yesNoCursor == 0)
	no := renderOption("no", m.yesNoCursor == 1)
	hint := hintStyle.Render("enter: continue • esc: back • ↑/↓ or y/n")
	return fmt.Sprintf("%s\n\n%s\n%s\n\n%s", q, yes, no, hint)
}

func (m model) viewName() string {
	q := questionStyle.Render("Project name")
	input := m.nameInput.View()
	hint := hintStyle.Render("enter: continue • esc: cancel • ctrl+c: cancel")

	var errLine string
	if m.nameErr != "" {
		errLine = "\n" + errorStyle.Render("✗ "+m.nameErr)
	}

	return fmt.Sprintf("%s\n\n  %s%s\n\n%s", q, input, errLine, hint)
}

func (m model) viewSummary() string {
	q := questionStyle.Render("Ready to create project")

	name := strings.TrimSpace(m.nameInput.Value())
	targetPath := ""
	if expanded, err := util.ExpandTilde(m.wc.Path); err == nil {
		targetPath = expanded + "/" + name
	} else {
		targetPath = m.wc.Path + "/" + name
	}

	lines := []string{
		summaryLine("name", name),
		summaryLine("path", targetPath),
		summaryLine("remote", ynLabel(m.github)),
	}
	if m.github {
		lines = append(lines, summaryLine("private", ynLabel(m.private)))
		lines = append(lines, summaryLine("org", m.wc.GithubOrg))
	}

	summary := strings.Join(lines, "\n")

	confirm := renderOption("create", m.yesNoCursor == 0)
	cancel := renderOption("cancel", m.yesNoCursor == 1)
	hint := hintStyle.Render("enter: confirm selection • esc: back • ctrl+c: cancel")

	return fmt.Sprintf("%s\n\n%s\n\n%s\n%s\n\n%s", q, summary, confirm, cancel, hint)
}

func renderOption(label string, selected bool) string {
	if selected {
		return selectedOptionStyle.Render("› " + label)
	}
	return optionStyle.Render("  " + label)
}

func summaryLine(key, val string) string {
	return "  " + summaryKeyStyle.Render(key+":") + " " + summaryValStyle.Render(val)
}

func ynLabel(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
