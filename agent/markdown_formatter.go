package agent

import (
	"context"
	"fmt"
)

// MarkdownFormatterAgentName is the identifier for the markdown formatter agent.
const MarkdownFormatterAgentName = "MarkdownFormatter"

// MarkdownFormatterAgent formats content into well-structured markdown documents.
type MarkdownFormatterAgent struct {
	*BaseAgent
}

// NewMarkdownFormatterAgent creates a new MarkdownFormatterAgent with predefined prompts for markdown formatting.
func NewMarkdownFormatterAgent(apiKey string) *MarkdownFormatterAgent {
	config := Config{
		Name:  MarkdownFormatterAgentName,
		Model: "gpt-5",
		Prompt: `You are a markdown formatting expert. 
		Take the provided content and format it into a well-structured markdown document. 
		Use appropriate headers, formatting, and structure to make it readable and professional.`,
	}
	return &MarkdownFormatterAgent{
		BaseAgent: NewBaseAgent(config, apiKey),
	}
}

// Start processes input content and formats it as markdown, sending results to the output channel.
func (m *MarkdownFormatterAgent) Start(ctx context.Context, input <-chan string, output chan<- string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case content, ok := <-input:
			if !ok {
				close(output)
				return nil
			}

			response, err := m.callOpenAI(ctx, fmt.Sprintf("Format this content into a well-structured markdown document: %s", content))
			if err != nil {
				continue
			}

			output <- response
			close(output)
			return nil
		}
	}
}
