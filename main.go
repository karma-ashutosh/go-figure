package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type Step struct {
	StepNumber  int    `json:"step number"`
	Description string `json:"description"`
	Reason      string `json:"reason"`
	Command     string `json:"command"`
}

// Supported modes
const (
	ModeExecute     = "execute"
	ModeWriteToFile = "write-to-file"
)

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

	var steps []Step
	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &steps)
	if err != nil {
		fmt.Printf("Error parsing response JSON: %v\n", err)
		os.Exit(1)
	}

	return steps
}

func handleCommand(mode, command string, outputFile *os.File) {
	switch mode {
	case ModeExecute:
		fmt.Println("Executing:", command)
		cmd := exec.Command("bash", "-c", command)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error executing command: %s\n", err)
		}
	case ModeWriteToFile:
		if outputFile != nil {
			fmt.Fprintf(outputFile, "%s\n", command)
		} else {
			fmt.Println("Error: No file specified for write-to-file mode.")
		}
	default:
		fmt.Println("Invalid mode. Skipping command.")
	}
}

func main() {
	// Parse flags for mode and file
	mode := flag.String("mode", ModeExecute, "Mode of operation: execute or write-to-file")
	outputFilePath := flag.String("file", "", "File path for write-to-file mode (optional)")
	flag.Parse()

	// Open the file if in write-to-file mode
	var outputFile *os.File
	var err error
	if *mode == ModeWriteToFile {
		if *outputFilePath == "" {
			fmt.Println("Error: File path is required for write-to-file mode.")
			os.Exit(1)
		}
		outputFile, err = os.OpenFile(*outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Error opening file: %s\n", err)
			os.Exit(1)
		}
		defer outputFile.Close()
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Welcome to CLI AI Assistant!")
	fmt.Printf("Mode: %s\n", *mode)
	fmt.Println("Type 'exit' to quit or 'cancel' during execution to submit a new query.")

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

		// Execute commands loop
		for _, step := range steps {
			if step.Command != "" {
				fmt.Printf("Do you want to execute the command for Step %d? (yes/no/cancel): ", step.StepNumber)
				choice, _ := reader.ReadString('\n')
				choice = strings.TrimSpace(strings.ToLower(choice))

				if choice == "cancel" {
					fmt.Println("Command execution canceled. Returning to query submission.")
					break
				} else if choice == "yes" {
					handleCommand(*mode, step.Command, outputFile)
				} else {
					fmt.Printf("Command for Step %d not executed.\n", step.StepNumber)
				}
			}
		}
	}
}
