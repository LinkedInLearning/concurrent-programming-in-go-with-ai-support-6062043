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

const TechSupportAgentName = "TechSupport"

type TechSupportRatingResponse struct {
	Rating      int    `json:"rating" jsonschema:"minimum=-1,maximum=10" jsonschema_description:"Rating from 0-10 for tech advice, or -1 if not applicable"`
	Explanation string `json:"explanation" jsonschema_description:"Brief explanation of the rating"`
}

type TechSupportAgent struct {
	Config Config
	client *openai.Client
	schema any
}

func NewTechSupportAgent(apiKey string) *TechSupportAgent {
	client := openai.NewClient(option.WithAPIKey(apiKey))

	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	schema := reflector.Reflect(TechSupportRatingResponse{})

	config := Config{
		Name:  TechSupportAgentName,
		Model: "gpt-5-mini",
		Prompt: "You are the best tech support engineer at a large fortune 100 company, an expert in all things computers. Your coworkers came across some tips online, and they want to ask you if they are good advice. You are tasked with analyzing the advice to provide a rating from 0-10 on how accurate the advice is. If the advice isn't technology or IT applicable, please return -1.",
	}

	return &TechSupportAgent{
		Config: config,
		client: &client,
		schema: schema,
	}
}

func (t *TechSupportAgent) Start(ctx context.Context, input <-chan string, output chan<- string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case text, ok := <-input:
			if !ok {
				close(output)
				return nil
			}

			result, err := t.callOpenAIStructured(ctx, text)
			if err != nil {
				continue
			}

			output <- fmt.Sprintf("%d - %s", result.Rating, result.Explanation)
			close(output)
			return nil
		}
	}
}

func (t *TechSupportAgent) callOpenAIStructured(ctx context.Context, prompt string) (*TechSupportRatingResponse, error) {
	callCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "techsupport_rating_response",
		Description: openai.String("Tech advice rating and explanation"),
		Schema:      t.schema,
		Strict:      openai.Bool(true),
	}

	completion, err := t.client.Chat.Completions.New(callCtx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(t.Config.Prompt),
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

	var rating TechSupportRatingResponse
	if err := json.Unmarshal([]byte(completion.Choices[0].Message.Content), &rating); err != nil {
		return nil, fmt.Errorf("failed to parse structured response: %w", err)
	}

	return &rating, nil
}
