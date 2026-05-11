package list

import (
	"strings"

	"github.com/JamesTiberiusKirk/workspacer/ui/list/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Item struct {
	Display, Subtitle, Value string
	IsActive                 bool
}

func (i Item) Title() string       { return i.Display }
func (i Item) Description() string { return i.Subtitle }
func (i Item) FilterValue() string {
	return i.Display + i.Subtitle + i.Value
}

type model struct {
	list         list.Model
	choise       *Item
	width        int
	height       int
	bottomStatus string
	onRefresh    func() ([]Item, string)
	useCard      bool
}

type refreshResultMsg struct {
	items  []Item
	status string
}

var (
	bottomBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 2)

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2)
)

const (
	maxCardContentHeight = 30
	minCardHeight        = 24
)

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case refreshResultMsg:
		m.bottomStatus = msg.status
		ii := make([]list.Item, len(msg.items))
		for i, item := range msg.items {
			ii[i] = item
		}
		return m, m.list.SetItems(ii)
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "ctrl+r" {
			if m.onRefresh != nil {
				m.bottomStatus = "refreshing..."
				refreshFn := m.onRefresh
				return m, func() tea.Msg {
					items, status := refreshFn()
					return refreshResultMsg{items: items, status: status}
				}
			}
			return m, nil
		}
		if msg.String() == "ctrl+shift+r" {
			m.choise = &Item{Value: "root:workspace"}
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
		m.width = msg.Width
		m.height = msg.Height

		availHeight := msg.Height - 1
		if availHeight < 5 {
			availHeight = 5
		}

		listWidth := msg.Width
		if listWidth > 80 {
			listWidth = 80
		}
		listWidth -= cardStyle.GetHorizontalFrameSize()
		if listWidth < 10 {
			listWidth = 10
		}

		m.useCard = availHeight > minCardHeight+cardStyle.GetVerticalFrameSize()
		if m.useCard {
			listHeight := availHeight - cardStyle.GetVerticalFrameSize()
			if listHeight > maxCardContentHeight {
				listHeight = maxCardContentHeight
			}
			m.list.SetSize(listWidth, listHeight)
		} else {
			m.list.SetSize(listWidth, availHeight)
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	listView := m.list.View()
	if m.useCard {
		listView = cardStyle.Render(listView)
	}

	leftStr := "ctrl+r  refresh  |  ctrl+shift+r  root"
	rightStr := m.bottomStatus

	innerWidth := m.width - 4
	if innerWidth < 10 {
		innerWidth = 10
	}

	leftW := lipgloss.Width(leftStr)
	rightW := lipgloss.Width(rightStr)
	gap := innerWidth - leftW - rightW
	if gap < 1 {
		gap = 1
	}

	barContent := leftStr + strings.Repeat(" ", gap) + rightStr
	bottomBar := bottomBarStyle.Render(barContent)
	bottomHeight := lipgloss.Height(bottomBar)

	availHeight := m.height - bottomHeight

	var content string
	if m.useCard {
		content = lipgloss.Place(m.width, availHeight, lipgloss.Center, lipgloss.Center, listView)
	} else {
		content = listView
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		content,
		bottomBar,
	)
}

// orderPreservingFilter keeps items in their input order while filtering
func orderPreservingFilter(term string, targets []string) []list.Rank {
	if term == "" {
		// Return all items in order
		result := make([]list.Rank, len(targets))
		for i := range targets {
			result[i] = list.Rank{Index: i}
		}
		return result
	}

	// Use fuzzy matching but preserve input order
	matches := []list.Rank{}
	for i, target := range targets {
		// Simple case-insensitive substring matching
		// You can use fuzzy matching here if needed, but preserve order
		targetLower := strings.ToLower(target)
		termLower := strings.ToLower(term)

		if strings.Contains(targetLower, termLower) {
			matches = append(matches, list.Rank{Index: i})
		}
	}

	return matches
}

func NewList(title string, Items []Item, bottomStatus string, onRefresh func() ([]Item, string)) (Item, bool, error) {
	// Sort items with active first to preserve order during filtering
	sortedItems := make([]Item, len(Items))
	copy(sortedItems, Items)

	// Separate into active and inactive, preserving order within each group
	activeItems := []Item{}
	inactiveItems := []Item{}
	for _, item := range sortedItems {
		if item.IsActive {
			activeItems = append(activeItems, item)
		} else {
			inactiveItems = append(inactiveItems, item)
		}
	}

	// Combine: active first, then inactive
	sortedItems = append(activeItems, inactiveItems...)

	// Convert to list.Item interface
	ii := make([]list.Item, len(sortedItems))
	for i, item := range sortedItems {
		ii[i] = item
	}

	m := model{
		list:         list.New(ii, list.NewDefaultDelegate(), 0, 0),
		bottomStatus: bottomStatus,
		onRefresh:    onRefresh,
	}
	m.list.Title = title
	m.list.SetShowHelp(false)

	// Use custom order-preserving filter
	m.list.Filter = orderPreservingFilter

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
