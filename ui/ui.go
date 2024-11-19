package ui

import (
	"fmt"
	"strings"
	"os"

	"github.com/charmbracelet/bubbletea"
	"go-figure/ai"
	"go-figure/history"
	"go-figure/mode"
)

type teaModel struct {
	query         string
	selectMode    string
	steps         []ai.Step
	currentScreen string
	inputBuffer   string
	historyView   string
}

func InitialModel() teaModel {
	return teaModel{
		query:         "",
		selectMode:    "",
		steps:         nil,
		currentScreen: "menu", // "menu", "query", "steps", "history"
		inputBuffer:   "",
		historyView:   history.GetHistory(),
	}
}

func (m teaModel) Init() tea.Cmd {
	return nil
}

func (m teaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Quit on Ctrl+C or 'q'
			return m, tea.Quit
		case "enter":
			if m.currentScreen == "menu" {
				// Handle menu selection
				if strings.ToLower(m.inputBuffer) == "query" {
					m.currentScreen = "query"
					m.inputBuffer = ""
				} else if strings.ToLower(m.inputBuffer) == "history" {
					m.currentScreen = "history"
					m.inputBuffer = ""
				}
			} else if m.currentScreen == "query" {
				// Process the query input
				m.query = strings.TrimSpace(m.inputBuffer)
				m.inputBuffer = ""
				m.currentScreen = "mode"
			} else if m.currentScreen == "mode" {
				// Handle mode selection
				if strings.ToLower(m.inputBuffer) == "execute" || strings.ToLower(m.inputBuffer) == "write-to-file" {
					m.selectMode = strings.ToLower(m.inputBuffer)
					m.inputBuffer = ""
					m.steps = ai.GetStepsFromAI(m.query)
					m.currentScreen = "steps"
				}
			} else if m.currentScreen == "steps" {
				// Handle steps processing
				if m.selectMode == mode.ModeExecute {
					mode.ExecuteSteps(m.steps)
				} else if m.selectMode == mode.ModeWriteToFile {
					fmt.Println("Enter the file path to save commands:")
					// Write steps to file logic
				}
				return m, tea.Quit
			}
		case "backspace":
			if len(m.inputBuffer) > 0 {
				m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
			}
		default:
			m.inputBuffer += msg.String()
		}
	}

	return m, nil
}

func (m teaModel) View() string {
	switch m.currentScreen {
	case "menu":
		return "Go-Figure\nSelect an option:\n- Query\n- History\n\n" +
			"Type your choice and press Enter:\n" + m.inputBuffer
	case "query":
		return "Enter your query and press Enter:\n" + m.inputBuffer
	case "mode":
		return "Select a mode:\n- execute\n- write-to-file\n\n" +
			"Type your choice and press Enter:\n" + m.inputBuffer
	case "steps":
		var sb strings.Builder
		sb.WriteString("Suggested Steps:\n")
		for _, step := range m.steps {
			sb.WriteString(fmt.Sprintf("Step %d: %s\n", step.StepNumber, step.Description))
			sb.WriteString(fmt.Sprintf("Reason: %s\n", step.Reason))
			if step.Command != "" {
				sb.WriteString(fmt.Sprintf("Command: %s\n", step.Command))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\nPress Enter to execute/write-to-file or 'q' to quit.")
		return sb.String()
	case "history":
		return "History:\n" + m.historyView + "\n\nPress 'q' to quit."
	default:
		return "Unknown state. Press 'q' to quit."
	}
}

func Run() {
	p := tea.NewProgram(InitialModel())
	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}