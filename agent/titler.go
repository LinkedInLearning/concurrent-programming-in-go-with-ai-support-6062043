package agent

import (
	"context"
	"fmt"
)

// TitlerAgentName is the identifier for the titler agent.
const TitlerAgentName = "Titler"

// TitlerAgent generates compelling titles for input text.
type TitlerAgent struct {
	*BaseAgent
}

// NewTitlerAgent creates a new TitlerAgent with predefined prompts for title generation.
func NewTitlerAgent(apiKey string) *TitlerAgent {
	config := Config{
		Name:   TitlerAgentName,
		Model:  "gpt-5-nano",
		Prompt: "You are a title generation expert. Create a compelling, concise title for the provided text that captures its essence. Provide only the title, nothing else.",
	}
	return &TitlerAgent{
		BaseAgent: NewBaseAgent(config, apiKey),
	}
}

// Start processes input text and generates titles, sending results to the output channel.
func (t *TitlerAgent) Start(ctx context.Context, input <-chan string, output chan<- string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case text, ok := <-input:
			if !ok {
				close(output)
				return nil
			}

			response, err := t.callOpenAI(ctx, fmt.Sprintf("Create a title for this text: %s", text))
			if err != nil {
				continue
			}

			output <- response
			close(output)
			return nil
		}
	}
}
