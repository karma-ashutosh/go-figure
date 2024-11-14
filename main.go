package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Mock function to simulate AI API response
func getCommandFromAI(prompt string) string {
	// Replace this with actual API call logic to Claude, OpenAI, etc.
	// Here we're mocking the response
	return fmt.Sprintf("echo %s", prompt)
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

		// Get command suggestion from AI
		suggestedCommand := getCommandFromAI(query)
		fmt.Printf("Suggested command: %s\n", suggestedCommand)

		// Prompt user to execute
		fmt.Print("Do you want to execute this command? (yes/no): ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(strings.ToLower(choice))

		if choice == "yes" {
			// Execute the command
			cmd := exec.Command("bash", "-c", suggestedCommand)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Printf("Error executing command: %s\n", err)
			}
		} else {
			fmt.Println("Command not executed.")
		}
	}
}
