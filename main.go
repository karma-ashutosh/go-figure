package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"go-figure/ai"
	"go-figure/history"
	"go-figure/mode"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "history" {
		history.Display()
		return
	}

	fmt.Println("Enter your query:")
	reader := bufio.NewReader(os.Stdin)
	query, _ := reader.ReadString('\n')
	query = strings.TrimSpace(query)

	if query == "" {
		fmt.Println("No query provided. Exiting.")
		return
	}

	selectMode := mode.Select()
	fmt.Printf("Selected Mode: %s\n", selectMode)

	steps := ai.GetStepsFromAI(query)
	if len(steps) == 0 {
		fmt.Println("No steps generated. Exiting.")
		return
	}

	fmt.Println("Suggested Steps:")
	for _, step := range steps {
		fmt.Printf("Step %d: %s\n", step.StepNumber, step.Description)
		fmt.Printf("Reason: %s\n", step.Reason)
		if step.Command != "" {
			fmt.Printf("Command: %s\n", step.Command)
		}
		fmt.Println()
	}

	history.Append(query, steps)

	if selectMode == mode.ModeExecute {
		mode.ExecuteSteps(steps)
	} else if selectMode == mode.ModeWriteToFile {
		fmt.Println("Enter the file path to save commands:")
		filePath, _ := reader.ReadString('\n')
		filePath = strings.TrimSpace(filePath)
		mode.WriteStepsToFile(steps, filePath)
	}
}
