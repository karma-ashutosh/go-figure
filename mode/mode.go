package mode

import (
	"bufio"
	"fmt"
	"os"

	"go-figure/ai"
	"go-figure/utils"
)

const (
	ModeExecute     = "execute"
	ModeWriteToFile = "write-to-file"
)

func Select() string {
	fmt.Println("Select mode:")
	fmt.Println("1. Execute commands")
	fmt.Println("2. Write commands to a file")

	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	switch choice {
	case "1\n":
		return ModeExecute
	case "2\n":
		return ModeWriteToFile
	default:
		fmt.Println("Invalid choice. Defaulting to 'execute'.")
		return ModeExecute
	}
}

func ExecuteSteps(steps []ai.Step) {
	for _, step := range steps {
		if step.Command != "" {
			fmt.Printf("Executing: %s\n", step.Command)
			output := utils.ExecuteCommand(step.Command)
			fmt.Println("Output:", output)
		}
	}
}

func WriteStepsToFile(steps []ai.Step, filePath string) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	for _, step := range steps {
		if step.Command != "" {
			_, err := file.WriteString(step.Command + "\n")
			if err != nil {
				fmt.Printf("Error writing to file: %v\n", err)
				return
			}
		}
	}
	fmt.Println("Commands saved successfully.")
}
