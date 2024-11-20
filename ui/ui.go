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
	loading 	  bool
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
	// fmt.Println("Got msg %s", msg)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// Handle quitting the program
		case "ctrl+c", "q":
			return m, tea.Quit

		// Handle Enter key
		case "enter":
			// fmt.Println("handling enter")
			return m.handleEnter()

		// Handle backspace for text inputs
		case "backspace":
			if m.currentScreen == screenQuery || m.currentScreen == screenMode {
				var cmd tea.Cmd
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}

		// Handle navigation keys for text input or list
		default:
			if m.currentScreen == screenQuery || m.currentScreen == screenMode {
				// Update text input
				var cmd tea.Cmd
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			} else if m.currentScreen == screenMenu {
				// Update list navigation
				var cmd tea.Cmd
				m.list, cmd = m.list.Update(msg)
				return m, cmd
			}
		}
	case stepsMsg:
		return m.handleStepsResponse(msg)
	}

	return m, nil
}

func (m *teaModel) handleEnter() (tea.Model, tea.Cmd) {
	fmt.Printf("Handling Enter on screen: %s\n", m.currentScreen)

	switch m.currentScreen {
	case screenMenu:
		// Handle menu selection
		item := m.list.SelectedItem()
		if item == nil {
			m.errorMessage = "No item selected."
			return m, nil
		}
		selected := item.(listItem).name

		switch selected {
		case "Query Assistance":
			m.currentScreen = screenQuery
			m.textInput.Reset()
		case "History":
			m.currentScreen = screenHistory
		default:
			m.errorMessage = "Invalid menu option selected."
		}

	case screenQuery:
		// Handle query input
		m.query = strings.TrimSpace(m.textInput.Value())
		if m.query == "" {
			m.errorMessage = "Query cannot be empty."
			return m, nil
		}
		m.currentScreen = screenMode
		m.textInput.Reset()

	case screenMode:
		// Handle mode selection
		m.selectMode = strings.ToLower(strings.TrimSpace(m.textInput.Value()))
		if m.selectMode != mode.ModeExecute && m.selectMode != mode.ModeWriteToFile {
			m.errorMessage = "Invalid mode. Choose 'execute' or 'write-to-file'."
			return m, nil
		}

		// Set loading state and fetch steps
		m.loading = true
		return m, m.fetchSteps()

	case screenSteps:
		// Handle execution or quitting from steps
		if m.selectMode == mode.ModeExecute {
			mode.ExecuteSteps(m.steps)
		} else if m.selectMode == mode.ModeWriteToFile {
			fmt.Println("Saving steps to a file...")
			// Add file saving logic
		}
		return m, tea.Quit
	}

	return m, nil
}


func (m *teaModel) fetchSteps() tea.Cmd {
	return func() tea.Msg {
		steps := ai.GetStepsFromAI(m.query)
		return stepsMsg{steps: steps}
	}
}
func (m teaModel) View() string {
	if m.loading {
		return "Querying AI... Please wait.\n\nPress 'q' to quit."
	}

	switch m.currentScreen {
	case screenMenu:
		return m.list.View()
	case screenQuery:
		return "Query Assistance\n\nEnter your query below:\n" + m.textInput.View()
	case screenMode:
		return "Choose Execution Mode:\n- execute\n- write-to-file\n\n" + m.textInput.View()
	case screenSteps:
		return m.viewStepsScreen()
	case screenHistory:
		return "Command History:\n" + m.historyView
	default:
		return "Unknown screen. Press 'q' to quit."
	}
}

// Custom message type for handling steps response
type stepsMsg struct {
	steps []ai.Step
}
func (m *teaModel) handleStepsResponse(msg stepsMsg) (tea.Model, tea.Cmd) {
	m.loading = false

	if len(msg.steps) == 0 {
		m.errorMessage = "No steps found for your query."
		return m, nil
	}

	m.steps = msg.steps
	m.currentScreen = screenSteps
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
