package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/sashabaranov/go-openai"
)

type Step struct {
	StepNumber  int    `json:"step"`
	Description string `json:"description"`
	Reason      string `json:"reason"`
	Command     string `json:"command"`
}

func GetStepsFromAI(prompt string) []Step {
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
