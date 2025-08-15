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
	Config Config
	client *openai.Client
}

// NewBaseAgent creates a new BaseAgent with the given configuration and API key.
func NewBaseAgent(config Config, apiKey string) *BaseAgent {
	client := openai.NewClient(option.WithAPIKey(apiKey))
	return &BaseAgent{
		Config: config,
		client: &client,
	}
}

// callOpenAI makes a chat completion request to the OpenAI API with the given prompt.
func (a *BaseAgent) callOpenAI(ctx context.Context, prompt string) (string, error) {
	// Add timeout to the context for OpenAI calls
	callCtx, cancel := context.WithTimeout(ctx, 180*time.Second)
	defer cancel()

	completion, err := a.client.Chat.Completions.New(callCtx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(a.Config.Prompt),
			openai.UserMessage(prompt),
		},
		Model: "gpt-5",
	})
	if err != nil {
		return "", fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return completion.Choices[0].Message.Content, nil
}
