package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbletea"
)

type screen string

const (
	screenMenu screen = "menu"
	screenList screen = "list"
	screenDone screen = "done"
)

type teaModel struct {
	currentScreen screen
	list          list.Model
	selectedItem  string
}

func InitialModel() teaModel {
	// Create a list with predefined items
	items := []list.Item{
		listItem{name: "Query"},
		listItem{name: "History"},
	}

	// Configure the list
	l := list.New(items, list.NewDefaultDelegate(), 30, 10)
	l.Title = "Choose an Option"

	return teaModel{
		currentScreen: screenMenu,
		list:          l,
	}
}

func (m teaModel) Init() tea.Cmd {
	return nil
}

func (m teaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.currentScreen == screenMenu {
				// Save the selected item and switch to the next screen
				m.selectedItem = m.list.SelectedItem().(listItem).name
				m.currentScreen = screenDone
			}
		}

		// Update the list state
		if m.currentScreen == screenMenu {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m teaModel) View() string {
	switch m.currentScreen {
	case screenMenu:
		return m.list.View()
	case screenDone:
		return fmt.Sprintf("You selected: %s\nPress 'q' to quit.", m.selectedItem)
	default:
		return "Unknown state"
	}
}

// listItem implements the list.Item interface
type listItem struct {
	name string
}

func (i listItem) Title() string       { return i.name }
func (i listItem) Description() string { return "" }
func (i listItem) FilterValue() string { return i.name }

func Run() {
	p := tea.NewProgram(InitialModel())
	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
	}
}
