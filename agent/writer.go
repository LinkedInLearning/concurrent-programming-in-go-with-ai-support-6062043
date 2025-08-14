package agent

import (
	"context"
)

// WriterAgentName is the identifier for the writer agent.
const WriterAgentName = "Writer"

// WriterAgent generates original content about startup companies.
type WriterAgent struct {
	*BaseAgent
}

// NewWriterAgent creates a new WriterAgent with predefined prompts for startup content generation.
func NewWriterAgent(apiKey string) *WriterAgent {
	config := Config{
		Name:   WriterAgentName,
		Model:  "gpt-5",
		Prompt: "You are a business writing expert. Write exactly one well-structured paragraph about building a startup company. Focus on practical advice and key considerations. Keep it informative and engaging.",
	}
	return &WriterAgent{
		BaseAgent: NewBaseAgent(config, apiKey),
	}
}

// Start generates startup content and sends it to the output channel.
func (w *WriterAgent) Start(ctx context.Context, input <-chan string, output chan<- string) error {
	response, err := w.callOpenAI(ctx, "Write a paragraph about building a startup company.")
	if err != nil {
		return err
	}

	output <- response
	close(output)
	return nil
}
