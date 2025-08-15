package router

import (
	"context"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/openai/openai-go"
)

type mockClient struct {
	id string
}

func (m mockClient) AvailableTokens() int {
	return 10
}

func (m *mockClient) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
	return &openai.ChatCompletion{
		ID: m.id,
	}, nil
}

func TestRouter_RoundRobin(t *testing.T) {
	clients := []Client{
		&mockClient{id: "client1"},
		&mockClient{id: "client2"},
		&mockClient{id: "client3"},
	}

	logger := log.New(nil)
	logger.SetLevel(log.ErrorLevel)
	router := NewRouter(clients, logger)

	ctx := context.Background()
	req := openai.ChatCompletionNewParams{}

	expectedOrder := []string{"client1", "client2", "client3", "client1", "client2", "client3"}

	for i, expectedID := range expectedOrder {
		resp, err := router.CreateChatCompletion(ctx, req)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i+1, err)
		}
		if resp.ID != expectedID {
			t.Errorf("Request %d: expected client %s, got %s", i+1, expectedID, resp.ID)
		}
	}
}

func TestRouter_EmptyClients(t *testing.T) {
	router := NewRouter([]Client{}, nil)

	ctx := context.Background()
	req := openai.ChatCompletionNewParams{}

	_, err := router.CreateChatCompletion(ctx, req)
	if err == nil {
		t.Error("Expected error for empty clients, got nil")
	}
}
