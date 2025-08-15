package client

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"github.com/concurrent-programming-in-go/ratelimiter"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/pkoukk/tiktoken-go"
)

const (
	DefaultRequestsPerMinute = 60
	DefaultTokensPerMinute   = 90000
)

var modelPricing = map[string]struct {
	InputCostPer1K  float64
	OutputCostPer1K float64
}{
	"gpt-4o":        {0.0025, 0.01},
	"gpt-4o-mini":   {0.00015, 0.0006},
	"gpt-4-turbo":   {0.01, 0.03},
	"gpt-4":         {0.03, 0.06},
	"gpt-3.5-turbo": {0.0015, 0.002},
	"gpt-5":         {0.005, 0.015},
	"gpt-5-mini":    {0.0003, 0.0012},
	"gpt-5-nano":    {0.0001, 0.0004},
}

type APIClient struct {
	client         openai.Client
	logger         *log.Logger
	requestLimiter *ratelimiter.TokenBucket
	tokenLimiter   *ratelimiter.TokenBucket
}

type APIClientConfig struct {
	APIKey            string
	RequestsPerMinute int
	TokensPerMinute   int
	Logger            *log.Logger
}

func (a APIClient) AvailableTokens() int {
	return a.tokenLimiter.AvailableTokens()
}

func NewAPIClient(config APIClientConfig) *APIClient {
	if config.RequestsPerMinute <= 0 {
		config.RequestsPerMinute = DefaultRequestsPerMinute
	}
	if config.TokensPerMinute <= 0 {
		config.TokensPerMinute = DefaultTokensPerMinute
	}
	if config.Logger == nil {
		config.Logger = log.Default()
	}

	client := openai.NewClient(option.WithAPIKey(config.APIKey))

	requestRefillRate := time.Minute / time.Duration(config.RequestsPerMinute)
	tokenRefillRate := time.Minute / time.Duration(config.TokensPerMinute)

	return &APIClient{
		client:         client,
		logger:         config.Logger,
		requestLimiter: ratelimiter.NewTokenBucket(config.RequestsPerMinute, requestRefillRate),
		tokenLimiter:   ratelimiter.NewTokenBucket(config.TokensPerMinute, tokenRefillRate),
	}
}

func (c *APIClient) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
	startTime := time.Now()

	model := string(req.Model)
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	inputTokens, err := c.countInputTokens(req.Messages, model)
	if err != nil {
		c.logger.Error("Failed to count input tokens", "error", err, "model", model)
		inputTokens = 0
	}

	if err := c.requestLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("request rate limit exceeded: %w", err)
	}

	for i := 0; i < inputTokens; i++ {
		if err := c.tokenLimiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("token rate limit exceeded: %w", err)
		}
	}

	resp, err := c.client.Chat.Completions.New(ctx, req)
	duration := time.Since(startTime)

	if err != nil {
		c.logger.Error("OpenAI API request failed",
			"error", err,
			"model", model,
			"input_tokens", inputTokens,
			"duration", duration,
		)
		return nil, err
	}

	outputTokens := int(resp.Usage.CompletionTokens)
	totalTokens := int(resp.Usage.TotalTokens)

	expectedCost := c.calculateCost(model, inputTokens, outputTokens)

	c.logger.Info("OpenAI API request completed",
		"model", model,
		"input_tokens", inputTokens,
		"output_tokens", outputTokens,
		"total_tokens", totalTokens,
		"expected_cost_usd", expectedCost,
		"duration", duration,
		"request_id", resp.ID,
	)

	return resp, nil
}

func (c *APIClient) countInputTokens(messages []openai.ChatCompletionMessageParamUnion, model string) (int, error) {
	encoding, err := tiktoken.EncodingForModel(model)
	if err != nil {
		encoding, err = tiktoken.GetEncoding("cl100k_base")
		if err != nil {
			return 0, fmt.Errorf("failed to get encoding: %w", err)
		}
	}

	totalTokens := 0
	for _, msg := range messages {
		content := c.extractMessageContent(msg)
		if content != "" {
			tokens := encoding.Encode(content, nil, nil)
			totalTokens += len(tokens)
		}
		totalTokens += 4
	}
	totalTokens += 2

	return totalTokens, nil
}

func (c *APIClient) extractMessageContent(msg openai.ChatCompletionMessageParamUnion) string {
	return ""
}

func (c *APIClient) calculateCost(model string, inputTokens, outputTokens int) float64 {
	pricing, exists := modelPricing[model]
	if !exists {
		pricing = modelPricing["gpt-3.5-turbo"]
	}

	inputCost := float64(inputTokens) / 1000.0 * pricing.InputCostPer1K
	outputCost := float64(outputTokens) / 1000.0 * pricing.OutputCostPer1K

	return inputCost + outputCost
}

func (c *APIClient) Close() {
	if c.requestLimiter != nil {
		c.requestLimiter.Stop()
	}
	if c.tokenLimiter != nil {
		c.tokenLimiter.Stop()
	}
}

func (c *APIClient) GetAvailableRequestTokens() int {
	return c.requestLimiter.AvailableTokens()
}

func (c *APIClient) GetAvailableTokens() int {
	return c.tokenLimiter.AvailableTokens()
}
