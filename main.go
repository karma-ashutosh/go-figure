package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// Step represents a single step in the response
type Step struct {
	StepNumber  int    `json:"step number"`
	Description string `json:"description"`
	Reason      string `json:"reason"`
	Command     string `json:"command"`
}

// Function to fetch AI response using OpenAI GPT
func getStepsFromAI(prompt string) []Step {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENAI_API_KEY environment variable not set.")
		os.Exit(1)
	}

	client := openai.NewClient(apiKey)
	ctx := context.Background()

	// Updated prompt with examples
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

		Example 2:
		User input: "List all files in the current directory and display the disk usage."
		Response:
		[
			{"step number": 1, "description": "List all files in the current directory", "reason": "To view the contents of the directory", "command": "ls"},
			{"step number": 2, "description": "Display disk usage for the files", "reason": "To understand how much space each file uses", "command": "du -h"}
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
		os.Exit(1)
	}

	// Parse the JSON response into a list of Step structs
	var steps []Step
	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &steps)
	if err != nil {
		fmt.Printf("Error parsing response JSON: %v\n", err)
		os.Exit(1)
	}

	return steps
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Welcome to CLI AI Assistant!")
	fmt.Println("Type 'exit' to quit.")
	for {
		fmt.Print("\nEnter your query: ")
		query, _ := reader.ReadString('\n')
		query = strings.TrimSpace(query)

		if strings.ToLower(query) == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		// Get steps suggestion from AI
		steps := getStepsFromAI(query)

		// Display the steps to the user
		fmt.Println("\nSuggested Steps:")
		for _, step := range steps {
			fmt.Printf("Step %d: %s\n", step.StepNumber, step.Description)
			fmt.Printf("Reason: %s\n", step.Reason)
			if step.Command != "" {
				fmt.Printf("Command: %s\n", step.Command)
			}
			fmt.Println()
		}

		// Prompt user to execute commands
		for _, step := range steps {
			if step.Command != "" {
				fmt.Printf("Do you want to execute the command for Step %d? (yes/no): ", step.StepNumber)
				choice, _ := reader.ReadString('\n')
				choice = strings.TrimSpace(strings.ToLower(choice))

				if choice == "yes" {
					cmd := exec.Command("bash", "-c", step.Command)
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					if err := cmd.Run(); err != nil {
						fmt.Printf("Error executing command: %s\n", err)
					}
				} else {
					fmt.Printf("Command for Step %d not executed.\n", step.StepNumber)
				}
			}
		}
	}
}
