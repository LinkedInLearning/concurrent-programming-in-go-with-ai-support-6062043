package embedding

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type EmbeddingService struct {
	client openai.Client
}

func NewEmbeddingService(apiKey string) *EmbeddingService {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &EmbeddingService{
		client: client,
	}
}

func (e *EmbeddingService) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	input := openai.EmbeddingNewParamsInputUnion{
		OfArrayOfStrings: []string{text},
	}

	resp, err := e.client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Input: input,
		Model: openai.EmbeddingModelTextEmbeddingAda002,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get embedding: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}

	embedding := resp.Data[0].Embedding
	result := make([]float32, len(embedding))
	for i, v := range embedding {
		result[i] = float32(v)
	}

	return result, nil
}

func (e *EmbeddingService) GetEmbeddingForArticle(ctx context.Context, title, description, summary string) ([]float32, error) {
	text := fmt.Sprintf("Title: %s\nDescription: %s\nSummary: %s", title, description, summary)
	return e.GetEmbedding(ctx, text)
}
