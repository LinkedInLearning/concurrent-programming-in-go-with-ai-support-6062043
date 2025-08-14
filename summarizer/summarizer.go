package summarizer

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type Summarizer struct {
	client openai.Client
}

func NewSummarizer(apiKey string) *Summarizer {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &Summarizer{
		client: client,
	}
}

func (s *Summarizer) Summarize(ctx context.Context, title, description string) (string, error) {
	prompt := fmt.Sprintf(`You are a summarization assistant. Create a concise summary of the following article using ONLY the information provided. Do not add external knowledge or assumptions.

Title: %s
Description: %s

Provide a brief, factual summary based solely on the information above:`, title, description)

	resp, err := s.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Model: "gpt-5-mini",
	})
	if err != nil {
		return "", fmt.Errorf("failed to get summary: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no summary generated")
	}

	return resp.Choices[0].Message.Content, nil
}

func (s *Summarizer) GenerateResponse(ctx context.Context, query string, articles []string) (string, error) {
	articlesText := ""
	for i, article := range articles {
		articlesText += fmt.Sprintf("%d. %s\n", i+1, article)
	}

	prompt := fmt.Sprintf(`You are a helpful assistant that answers questions based ONLY on the provided articles. Do not use external knowledge or make assumptions beyond what is explicitly stated in the articles.

User Query: "%s"

Relevant Articles:
%s

Instructions:
- First, determine if any of the articles are actually relevant to the user's query
- If none of the articles match or relate to the query, respond with: "None of the retrieved articles matched your query. Please try a different search term."
- If the articles are relevant but don't contain enough specific information to fully answer the query, say: "The retrieved articles are related to your query but don't contain enough specific information to provide a complete answer."
- Only if the articles are relevant AND contain sufficient information, provide a helpful answer using ONLY information from the articles
- Do not add external facts or knowledge not present in the articles
- Be concise and factual

Response:`, query, articlesText)

	resp, err := s.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Model: "gpt-5-mini",
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response generated")
	}

	return resp.Choices[0].Message.Content, nil
}