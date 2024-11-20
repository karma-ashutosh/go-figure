package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"go-figure/ai"
	"go-figure/history"
	"go-figure/mode"
)

type screen string

const (
	screenMenu    screen = "menu"
	screenQuery   screen = "query"
	screenMode    screen = "mode"
	screenSteps   screen = "steps"
	screenHistory screen = "history"
)

type teaModel struct {
	currentScreen screen
	textInput     textinput.Model
	list          list.Model
	query         string
	selectMode    string
	steps         []ai.Step
	historyView   string
	errorMessage  string
}

func InitialModel() teaModel {
	// Text input setup for queries
	ti := textinput.New()
	ti.Placeholder = "Type your query here..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	// List setup for the menu
	items := []list.Item{
		listItem{name: "Query Assistance"},
		listItem{name: "History"},
	}
	menuList := list.New(items, list.NewDefaultDelegate(), 30, 10)
	menuList.Title = "Main Menu"

	// History view
	historyContent := history.GetHistory()

	return teaModel{
		currentScreen: screenMenu,
		textInput:     ti,
		list:          menuList,
		historyView:   historyContent,
	}
}

func (m teaModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m teaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			return m.handleEnter()
		case "backspace":
			if m.currentScreen == screenQuery || m.currentScreen == screenMode {
				var cmd tea.Cmd
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
		default:
			if m.currentScreen == screenQuery || m.currentScreen == screenMode {
				var cmd tea.Cmd
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
		}

		// Allow list navigation for the menu
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
	case screenQuery:
		return m.viewQueryScreen()
	case screenMode:
		return m.viewModeScreen()
	case screenSteps:
		return m.viewStepsScreen()
	case screenHistory:
		return m.viewHistoryScreen()
	default:
		return "Unknown screen. Press 'q' to quit."
	}
}

func (m *teaModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.currentScreen {
	case screenQuery:
		m.query = strings.TrimSpace(m.textInput.Value())
		if m.query == "" {
			m.errorMessage = "Query cannot be empty."
			return m, nil
		}
		m.currentScreen = screenMode
		m.textInput.Reset()
	case screenMode:
		m.selectMode = strings.ToLower(strings.TrimSpace(m.textInput.Value()))
		if m.selectMode != mode.ModeExecute && m.selectMode != mode.ModeWriteToFile {
			m.errorMessage = "Invalid mode. Choose 'execute' or 'write-to-file'."
			return m, nil
		}
		m.steps = ai.GetStepsFromAI(m.query)
		if len(m.steps) == 0 {
			m.errorMessage = "No steps found for your query."
			return m, nil
		}
		// Add the query and steps to history
		history.Append(m.query, m.steps)
		m.currentScreen = screenSteps
	}
	return m, nil
}

func (m teaModel) viewQueryScreen() string {
	return fmt.Sprintf(
		"Query Assistance\n\nEnter your query below:\n%s\n\nPress Enter to submit, or 'q' to quit.\n%s",
		m.textInput.View(),
		m.errorMessage,
	)
}

func (m teaModel) viewModeScreen() string {
	return fmt.Sprintf(
		"Choose Execution Mode\n\n- execute\n- write-to-file\n\nEnter your choice:\n%s\n\nPress Enter to continue, or 'q' to quit.\n%s",
		m.textInput.View(),
		m.errorMessage,
	)
}

func (m teaModel) viewStepsScreen() string {
	var sb strings.Builder
	sb.WriteString("Suggested Steps:\n\n")
	for _, step := range m.steps {
		sb.WriteString(fmt.Sprintf("Step %d: %s\n", step.StepNumber, step.Description))
		sb.WriteString(fmt.Sprintf("Reason: %s\n", step.Reason))
		if step.Command != "" {
			sb.WriteString(fmt.Sprintf("Command: %s\n", step.Command))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\nPress Enter to execute/write-to-file, or 'q' to quit.")
	return sb.String()
}

func (m teaModel) viewHistoryScreen() string {
	return fmt.Sprintf(
		"Command History\n\n%s\n\nPress 'q' to return to the main menu.",
		m.historyView,
	)
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
