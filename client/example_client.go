package client

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/openai/openai-go"
)

func ExampleAPIClient() {
	logger := log.New(os.Stdout)

	config := APIClientConfig{
		APIKey:            "test-key",
		RequestsPerMinute: 60,
		TokensPerMinute:   90000,
		Logger:            logger,
	}

	client := NewAPIClient(config)
	defer client.Close()

	ctx := context.Background()

	req := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Hello, world!"),
		},
		Model: openai.ChatModelGPT4o,
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Response: %s\n", resp.Choices[0].Message.Content)
}
