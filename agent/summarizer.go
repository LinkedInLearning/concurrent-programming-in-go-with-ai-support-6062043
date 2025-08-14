package agent

import (
	"context"
	"fmt"
)

// SummarizerAgent creates concise summaries of input text.
type SummarizerAgent struct {
	*BaseAgent
}

// NewSummarizerAgent creates a new SummarizerAgent with predefined prompts for text summarization.
func NewSummarizerAgent(apiKey string) *SummarizerAgent {
	config := Config{
		Name:   "Summarizer",
		Model:  "gpt-5-mini",
		Prompt: "You are a summarization expert. Take the provided text and summarize it into exactly two clear, concise sentences that capture the main points.",
	}
	return &SummarizerAgent{
		BaseAgent: NewBaseAgent(config, apiKey),
	}
}

// Start processes input text and generates summaries, sending results to the output channel.
func (s *SummarizerAgent) Start(ctx context.Context, input <-chan string, output chan<- string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case text, ok := <-input:
			if !ok {
				close(output)
				return nil
			}

			response, err := s.callOpenAI(ctx, fmt.Sprintf("Summarize this text into exactly two sentences: %s", text))
			if err != nil {
				continue
			}

			output <- response
			close(output)
			return nil
		}
	}
}
