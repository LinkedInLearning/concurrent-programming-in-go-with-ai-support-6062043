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

const CareerAgentName = "Career"

type CareerRatingResponse struct {
	Rating      int    `json:"rating" jsonschema:"minimum=-1,maximum=10" jsonschema_description:"Rating from 0-10 for career advice, or -1 if not applicable"`
	Explanation string `json:"explanation" jsonschema_description:"Brief explanation of the rating"`
}

type CareerAgent struct {
	Config Config
	client *openai.Client
	schema any
}

func NewCareerAgent(apiKey string) *CareerAgent {
	client := openai.NewClient(option.WithAPIKey(apiKey))

	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	schema := reflector.Reflect(CareerRatingResponse{})

	config := Config{
		Name:  CareerAgentName,
		Model: "gpt-5-mini",
		Prompt: "You are an expert career counselor. Your client has been given some advice, and you are tasked with analyzing the advice to provide a rating from 0-10 on how good the advice would be for their career. If the advice isn't career applicable, please return -1.",
	}

	return &CareerAgent{
		Config: config,
		client: &client,
		schema: schema,
	}
}

func (c *CareerAgent) Start(ctx context.Context, input <-chan string, output chan<- string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case text, ok := <-input:
			if !ok {
				close(output)
				return nil
			}

			result, err := c.callOpenAIStructured(ctx, text)
			if err != nil {
				continue
			}

			output <- fmt.Sprintf("%d - %s", result.Rating, result.Explanation)
			close(output)
			return nil
		}
	}
}

func (c *CareerAgent) callOpenAIStructured(ctx context.Context, prompt string) (*CareerRatingResponse, error) {
	callCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "career_rating_response",
		Description: openai.String("Career advice rating and explanation"),
		Schema:      c.schema,
		Strict:      openai.Bool(true),
	}

	completion, err := c.client.Chat.Completions.New(callCtx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(c.Config.Prompt),
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

	var rating CareerRatingResponse
	if err := json.Unmarshal([]byte(completion.Choices[0].Message.Content), &rating); err != nil {
		return nil, fmt.Errorf("failed to parse structured response: %w", err)
	}

	return &rating, nil
}
