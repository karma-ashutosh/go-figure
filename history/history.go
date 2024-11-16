package history

import (
	"encoding/json"
	"fmt"
	"os"
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

func Append(query string, steps []ai.Step) {
	historyMutex.Lock()
	defer historyMutex.Unlock()
	history = append(history, HistoryEntry{Query: query, Response: steps})
}

func Display() {
	historyMutex.Lock()
	defer historyMutex.Unlock()

	if len(history) == 0 {
		fmt.Println("No history available.")
		return
	}

	for i, entry := range history {
		fmt.Printf("Query %d: %s\n", i+1, entry.Query)
		for _, step := range entry.Response {
			fmt.Printf("  Step %d: %s\n", step.StepNumber, step.Description)
		}
		fmt.Println()
	}
}

func SaveToFile(filePath string) error {
	historyMutex.Lock()
	defer historyMutex.Unlock()

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(history)
}
