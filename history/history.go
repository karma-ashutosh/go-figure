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
	filePath     = "history.json"
)

func Append(query string, steps []ai.Step) {
	historyMutex.Lock()
	defer historyMutex.Unlock()
	history = append(history, HistoryEntry{Query: query, Response: steps})
	saveToFile()
}

func GetHistory() string {
	historyMutex.Lock()
	defer historyMutex.Unlock()

	if len(history) == 0 {
		return "No history available."
	}

	var result string
	for i, entry := range history {
		result += fmt.Sprintf("Query %d: %s\n", i+1, entry.Query)
		for _, step := range entry.Response {
			result += fmt.Sprintf("  Step %d: %s\n", step.StepNumber, step.Description)
		}
		result += "\n"
	}
	return result
}

func saveToFile() {
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("Error saving history: %v\n", err)
		return
	}
	defer file.Close()

	data, _ := json.MarshalIndent(history, "", "  ")
	file.Write(data)
}

func loadFromFile() {
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	json.NewDecoder(file).Decode(&history)
}

func init() {
	loadFromFile()
}
