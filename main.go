package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"go-figure/ui"
	"go-figure/ai"
)

func main() {
	// Check if input is piped
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Handle piped input
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		handleQuery(strings.TrimSpace(input))
		return
	}

	// Check for direct query argument
	if len(os.Args) > 1 {
		query := strings.Join(os.Args[1:], " ")
		handleQuery(query)
		return
	}

	// Default to TUI mode
	ui.Run()
}

func handleQuery(query string) {
	if query == "" {
		fmt.Println("No query provided. Exiting.")
		return
	}

	// Process query with AI
	steps := ai.GetStepsFromAI(query)
	if len(steps) == 0 {
		fmt.Println("No suggestions found. Exiting.")
		return
	}

	// Display suggestions
	fmt.Println("Suggestions:")
	for _, step := range steps {
		fmt.Printf("Step %d: %s\n", step.StepNumber, step.Description)
		fmt.Printf("Reason: %s\n", step.Reason)
		if step.Command != "" {
			fmt.Printf("Command: %s\n", step.Command)
		}
		fmt.Println()
	}
}
