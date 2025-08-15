package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// BaseAgent provides common functionality for all OpenAI-based agents.
type BaseAgent struct {
	Config       Config
	client       *openai.Client
	tokenCounter *TokenCounter
}

// NewBaseAgent creates a new BaseAgent with the given configuration and API key.
func NewBaseAgent(config Config, apiKey string) *BaseAgent {
	client := openai.NewClient(option.WithAPIKey(apiKey))
	tokenCounter, err := NewTokenCounter()
	if err != nil {
		// Fallback to nil if token counter fails to initialize
		tokenCounter = nil
	}
	return &BaseAgent{
		Config:       config,
		client:       &client,
		tokenCounter: tokenCounter,
	}
}

// callOpenAIWithTracking makes a chat completion request with token tracking
func (a *BaseAgent) callOpenAIWithTracking(ctx context.Context, prompt string, tracker *SystemTokenTracker) (string, error) {
	return a.callOpenAIWithTrackingName(ctx, prompt, tracker, a.Config.Name)
}

// callOpenAIWithTrackingName makes a chat completion request with token tracking using a custom name
func (a *BaseAgent) callOpenAIWithTrackingName(ctx context.Context, prompt string, tracker *SystemTokenTracker, trackingName string) (string, error) {
	// Add timeout to the context for OpenAI calls
	callCtx, cancel := context.WithTimeout(ctx, 180*time.Second)
	defer cancel()

	// Count input tokens if we have a token counter
	var inputTokens int
	if a.tokenCounter != nil {
		systemPrompt := a.Config.Prompt
		fullInput := systemPrompt + "\n" + prompt
		inputTokens = a.tokenCounter.CountTokens(fullInput)
	}

	completion, err := a.client.Chat.Completions.New(callCtx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(a.Config.Prompt),
			openai.UserMessage(prompt),
		},
		Model: "gpt-4o", // Using gpt-4o instead of gpt-5 for now
	})
	if err != nil {
		return "", fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	response := completion.Choices[0].Message.Content

	// Count output tokens and record usage if we have tracking enabled
	if a.tokenCounter != nil && tracker != nil {
		outputTokens := a.tokenCounter.CountTokens(response)
		tracker.RecordUsage(trackingName, inputTokens, outputTokens)
	}

	return response, nil
}
