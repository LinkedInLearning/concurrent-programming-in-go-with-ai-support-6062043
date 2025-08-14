package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const FinancialAgentName = "Financial"

type FinancialRatingResponse struct {
	Rating      int    `json:"rating" jsonschema:"minimum=-1,maximum=10" jsonschema_description:"Rating from 0-10 for financial advice, or -1 if not applicable"`
	Explanation string `json:"explanation" jsonschema_description:"Brief explanation of the rating"`
}

type FinancialAgent struct {
	Config Config
	client *openai.Client
	schema any
}

func NewFinancialAgent(apiKey string) *FinancialAgent {
	client := openai.NewClient(option.WithAPIKey(apiKey))

	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	schema := reflector.Reflect(FinancialRatingResponse{})

	config := Config{
		Name:  FinancialAgentName,
		Model: "gpt-5-mini",
		Prompt: "You are an expert financial advisor. Your client has been given some advice, and you are tasked with analyzing the advice to provide a rating from 0-10 on how good the advice would be for their financial success. If the advice isn't finance applicable, please return -1.",
	}

	return &FinancialAgent{
		Config: config,
		client: &client,
		schema: schema,
	}
}

func (f *FinancialAgent) Start(ctx context.Context, input <-chan string, output chan<- string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case text, ok := <-input:
			if !ok {
				close(output)
				return nil
			}

			result, err := f.callOpenAIStructured(ctx, text)
			if err != nil {
				continue
			}

			output <- fmt.Sprintf("%d - %s", result.Rating, result.Explanation)
			close(output)
			return nil
		}
	}
}

func (f *FinancialAgent) callOpenAIStructured(ctx context.Context, prompt string) (*FinancialRatingResponse, error) {
	callCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "financial_rating_response",
		Description: openai.String("Financial advice rating and explanation"),
		Schema:      f.schema,
		Strict:      openai.Bool(true),
	}

	completion, err := f.client.Chat.Completions.New(callCtx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(f.Config.Prompt),
			openai.UserMessage(prompt),
		},
		Model: "gpt-5-mini",
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				JSONSchema: schemaParam,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	var rating FinancialRatingResponse
	if err := json.Unmarshal([]byte(completion.Choices[0].Message.Content), &rating); err != nil {
		return nil, fmt.Errorf("failed to parse structured response: %w", err)
	}

	return &rating, nil
}
