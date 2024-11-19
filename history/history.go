package history

import (
	"fmt"
	"strings"
	"sync"

	"go-figure/ai"
)

type HistoryEntry struct {
	Query    string    `json:"query"`
	Response []ai.Step `json:"response"`
}

var (
	history      []HistoryEntry
	historyMutex sync.Mutex
)

// Append adds a new entry to the history
func Append(query string, steps []ai.Step) {
	historyMutex.Lock()
	defer historyMutex.Unlock()
	history = append(history, HistoryEntry{Query: query, Response: steps})
}

// GetHistory formats the history for display
func GetHistory() string {
	historyMutex.Lock()
	defer historyMutex.Unlock()

	if len(history) == 0 {
		return "No history available."
	}

	var sb strings.Builder
	for i, entry := range history {
		sb.WriteString(fmt.Sprintf("Query %d: %s\n", i+1, entry.Query))
		for _, step := range entry.Response {
			sb.WriteString(fmt.Sprintf("  Step %d: %s\n", step.StepNumber, step.Description))
			sb.WriteString(fmt.Sprintf("    Reason: %s\n", step.Reason))
			if step.Command != "" {
				sb.WriteString(fmt.Sprintf("    Command: %s\n", step.Command))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// Display prints the history to the terminal (for CLI mode)
func Display() {
	fmt.Println(GetHistory())
}
