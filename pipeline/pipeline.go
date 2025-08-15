package pipeline

import (
	"context"
	"fmt"

	"concurrent-programming-go-agents/embedding"
	"concurrent-programming-go-agents/summarizer"
	"concurrent-programming-go-agents/vectordb"
)

type Pipeline struct {
	embeddingService  *embedding.EmbeddingService
	summarizerService *summarizer.Summarizer
	vectorDB          *vectordb.WeaviateDB
}

func NewPipeline(embeddingService *embedding.EmbeddingService, summarizerService *summarizer.Summarizer, vectorDB *vectordb.WeaviateDB) *Pipeline {
	return &Pipeline{
		embeddingService:  embeddingService,
		summarizerService: summarizerService,
		vectorDB:          vectorDB,
	}
}

func (p *Pipeline) SearchArticles(ctx context.Context, query string, limit int) ([]*vectordb.Article, error) {
	queryVector, err := p.embeddingService.GetEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get query embedding: %w", err)
	}

	articles, err := p.vectorDB.SearchSimilar(ctx, queryVector, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search articles: %w", err)
	}

	return articles, nil
}

func (p *Pipeline) GenerateResponse(ctx context.Context, query string, articles []*vectordb.Article) (string, error) {
	articleTexts := make([]string, len(articles))
	for i, article := range articles {
		articleTexts[i] = fmt.Sprintf("Title: %s\nSummary: %s\nLink: %s", article.Title, article.Summary, article.Link)
	}

	return p.summarizerService.GenerateResponse(ctx, query, articleTexts)
}

func (p *Pipeline) Close(ctx context.Context) error {
	if p.vectorDB != nil {
		return p.vectorDB.Close(ctx)
	}
	return nil
}
