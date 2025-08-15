package agent

import (
	"context"
	"fmt"
)

// StoryAgent represents a specialized agent for story writing tasks
type StoryAgent struct {
	*BaseAgent
	agentType string
}

// NewStoryAgent creates a new story agent with the specified type and prompt
func NewStoryAgent(agentType, prompt, apiKey string) *StoryAgent {
	config := Config{
		Name:   agentType,
		Model:  "gpt-4o",
		Prompt: prompt,
	}
	
	return &StoryAgent{
		BaseAgent: NewBaseAgent(config, apiKey),
		agentType: agentType,
	}
}

// Start implements the Agent interface for story agents
func (s *StoryAgent) Start(ctx context.Context, input <-chan string, output chan<- string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case prompt, ok := <-input:
			if !ok {
				return nil
			}
			
			response, err := s.callOpenAI(ctx, prompt)
			if err != nil {
				output <- fmt.Sprintf("Error from %s: %v", s.agentType, err)
				continue
			}
			
			output <- response
		}
	}
}