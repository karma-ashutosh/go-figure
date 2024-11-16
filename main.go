package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sync"

	"github.com/sashabaranov/go-openai"
)

// Structs
type Step struct {
	StepNumber  int    `json:"step"`
	Description string `json:"description"`
	Reason      string `json:"reason"`
	Command     string `json:"command"`
}

type QueryRequest struct {
	Query string `json:"query"`
	Mode  string `json:"mode"` // "execute" or "write"
}

type QueryResponse struct {
	Steps []Step `json:"steps"`
	Error string `json:"error,omitempty"`
}

type CommandRequest struct {
	Command string `json:"command"`
}

type CommandResponse struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

type HistoryEntry struct {
	Query    string `json:"query"`
	Response []Step `json:"response"`
}

var history []HistoryEntry
var historyMutex sync.Mutex

// Modes
const (
	ModeExecute     = "execute"
	ModeWriteToFile = "write-to-file"
)

// Helper Functions
func getStepsFromAI(prompt string) []Step {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENAI_API_KEY environment variable not set.")
		os.Exit(1)
	}

	client := openai.NewClient(apiKey)
	ctx := context.Background()

	aiPrompt := fmt.Sprintf(`
		You are a Linux command assistant. Given the user's input, respond with a series of steps in JSON format.
		Each step should include the step number, a description of the action, the reason for the action, and the command to execute if applicable.
		If no command is needed, omit the "command" field.

		Respond only in JSON format.

		Example 1:
		User input: "Create a file and write 'Hello World' into it."
		Response:
		[
			{"step number": 1, "description": "Create a new file named 'hello.txt'", "reason": "To store the message", "command": "touch hello.txt"},
			{"step number": 2, "description": "Write 'Hello World' into the file", "reason": "To add the content to the file", "command": "echo 'Hello World' > hello.txt"}
		]

		User input: %s`, prompt)

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are an AI assistant helping with Linux commands, returning structured JSON responses.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: aiPrompt,
			},
		},
	})

	if err != nil {
		fmt.Printf("Error communicating with OpenAI API: %v\n", err)
		return []Step{}
	}

	var steps []Step
	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &steps)
	if err != nil {
		fmt.Printf("Error parsing response JSON: %v\n", err)
		return []Step{}
	}

	return steps
}

func executeCommand(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// Handlers
func handleQuery(w http.ResponseWriter, r *http.Request) {
	fmt.Println("in handle query")
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}

	steps := getStepsFromAI(req.Query)

	response := QueryResponse{Steps: steps}
	if len(steps) == 0 {
		response.Error = "Failed to generate steps."
	}

	// Add to history
	historyMutex.Lock()
	history = append(history, HistoryEntry{Query: req.Query, Response: steps})
	historyMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}

	if req.Command == "" {
		http.Error(w, "No command provided", http.StatusBadRequest)
		return
	}

	output, err := executeCommand(req.Command)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CommandResponse{
			Output: output,
			Error:  err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CommandResponse{
		Output: output,
	})
}

func handleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	historyMutex.Lock()
	json.NewEncoder(w).Encode(history)
	historyMutex.Unlock()
}

func main() {
	http.HandleFunc("/api/query", handleQuery)
	http.HandleFunc("/api/command", handleCommand)
	http.HandleFunc("/api/history", handleHistory)

	port := ":8080"
	fmt.Printf("Server running on http://localhost%s\n", port)
	http.ListenAndServe(port, nil)
}
