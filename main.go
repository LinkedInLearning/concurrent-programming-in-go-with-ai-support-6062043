package main

import (
	"context"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/concurrent-programming-in-go/client"
	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Warn("Error loading .env file, using system environment variables")
	}

	logger := log.New(os.Stdout)
	logger.SetLevel(log.InfoLevel)

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		logger.Warn("OPENAI_API_KEY not set, using dummy key for demonstration")
		apiKey = "dummy-key-for-demo"
	}

	config := client.APIClientConfig{
		APIKey:            apiKey,
		RequestsPerMinute: 10,   // Very low limit to see rate limiting in action
		TokensPerMinute:   1000, // Also low to demonstrate token limiting
		Logger:            logger,
	}

	apiClient := client.NewAPIClient(config)
	defer apiClient.Close()

	ctx := context.Background()

	logger.Info("Starting rate limiter demonstration",
		"requests_per_minute", config.RequestsPerMinute,
		"tokens_per_minute", config.TokensPerMinute,
	)

	req := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Hello! Please respond with a short greeting."),
		},
		Model: openai.ChatModelGPT4oMini,
	}

	for i := 1; i <= 5; i++ {
		start := time.Now()
		resp, err := apiClient.CreateChatCompletion(ctx, req)
		duration := time.Since(start)

		if err != nil {
			logger.Error("Request failed",
				"request_number", i,
				"error", err,
				"duration", duration,
			)
		} else {
			logger.Info("Request succeeded",
				"request_number", i,
				"response_id", resp.ID,
				"duration", duration,
			)
		}
	}

	logger.Info("Rate limiter demonstration completed")
}
