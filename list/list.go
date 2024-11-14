package list

import (
	"github.com/JamesTiberiusKirk/workspacer/list/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TODO: at some point figure this out
// Preferably would like it to just be a viewport centered both horizontally and vertically

// var docStyle = lipgloss.NewStyle().Margin(1, 2)
var docStyle = lipgloss.NewStyle().
	// Height(40).
	// Width(40).
	// Margin(1, 2).
	Align(lipgloss.Left, lipgloss.Center)

type Item struct {
	Display, Subtitle, Value string
}

func (i Item) Title() string       { return i.Display }
func (i Item) Description() string { return i.Subtitle }
func (i Item) FilterValue() string { return i.Display + i.Subtitle + i.Value }

type model struct {
	list   list.Model
	choise *Item
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "enter" {
			item, ok := m.list.SelectedItem().(Item)
			if ok {
				m.choise = &item
			}
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}

func NewList(title string, Items []Item) (Item, bool, error) {
	ii := make([]list.Item, len(Items))
	for i, item := range Items {
		ii[i] = item
	}

	m := model{list: list.New(ii, list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = title
	// This does not display the whole list to begin with
	// m.list.SetFilterState(list.Filtering)
	m.list.StartFiltering()

	p := tea.NewProgram(m, tea.WithAltScreen())

	mraw, err := p.Run()
	if err != nil {
		return Item{}, false, err
	}

	m, ok := mraw.(model)
	if !ok {
		return Item{}, false, nil
	}
	if m.choise == nil {
		return Item{}, false, nil
	}

	return *m.choise, true, nil
}
