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
		Model:  "gpt-5",
		Prompt: prompt,
	}

	return &StoryAgent{
		BaseAgent: NewBaseAgent(config, apiKey),
		agentType: agentType,
	}
}

// Start implements the Agent interface for story agents
func (s *StoryAgent) Start(ctx context.Context, input <-chan string, output chan<- string) error {
	return s.StartWithTracking(ctx, input, output, nil)
}

// StartWithTracking implements the Agent interface with token tracking support
func (s *StoryAgent) StartWithTracking(ctx context.Context, input <-chan string, output chan<- string, tracker *SystemTokenTracker) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case prompt, ok := <-input:
			if !ok {
				return nil
			}

			response, err := s.callOpenAIWithTracking(ctx, prompt, tracker)
			if err != nil {
				output <- fmt.Sprintf("Error from %s: %v", s.agentType, err)
				continue
			}

			output <- response
		}
	}
}

// CallWithTracking provides a direct way to call the agent with token tracking
func (s *StoryAgent) CallWithTracking(ctx context.Context, prompt string, tracker *SystemTokenTracker) (string, error) {
	return s.CallWithTrackingName(ctx, prompt, tracker, s.agentType)
}

// CallWithTrackingName provides a direct way to call the agent with token tracking using a custom name
func (s *StoryAgent) CallWithTrackingName(ctx context.Context, prompt string, tracker *SystemTokenTracker, trackingName string) (string, error) {
	return s.callOpenAIWithTrackingName(ctx, prompt, tracker, trackingName)
}
