package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

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

func executeCommand(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func handleQuery(w http.ResponseWriter, r *http.Request) {
	fmt.Println("in handle query")
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Println("Error decoding JSON input:", err)
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}

	fmt.Println("Query received:", req.Query)

	steps := getStepsFromAI(req.Query)

	response := QueryResponse{Steps: steps}
	if len(steps) == 0 {
		fmt.Println("No steps generated for query:", req.Query)
		response.Error = "Failed to generate steps."
	}

	// Add to history
	historyMutex.Lock()
	history = append(history, HistoryEntry{Query: req.Query, Response: steps})
	historyMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Println("Error encoding response JSON:", err)
	}
}

func getStepsFromAI(prompt string) []Step {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENAI_API_KEY environment variable not set.")
		os.Exit(1)
	}

	client := openai.NewClient(apiKey)

	// Use context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

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

	fmt.Println("Sending prompt to OpenAI API...")
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

	fmt.Println("OpenAI API response received. Parsing...")

	// Log the raw response for debugging
	rawResponse := resp.Choices[0].Message.Content
	fmt.Printf("Raw API Response: %s\n", rawResponse)

	var steps []Step
	err = json.Unmarshal([]byte(rawResponse), &steps)
	if err != nil {
		fmt.Printf("Error parsing response JSON: %v\n", err)
		return []Step{}
	}

	fmt.Println("Parsed steps successfully:", steps)
	return steps
}

// Handlers

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
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow CORS from any origin
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/query", handleQuery)
	mux.HandleFunc("/api/command", handleCommand)
	mux.HandleFunc("/api/history", handleHistory)

	// Wrap the mux with the CORS middleware
	corsMux := enableCORS(mux)

	port := ":8080"
	fmt.Printf("Server running on http://localhost%s\n", port)
	http.ListenAndServe(port, corsMux)
}
