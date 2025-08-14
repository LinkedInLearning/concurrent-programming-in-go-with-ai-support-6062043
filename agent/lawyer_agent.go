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

const LawyerAgentName = "Lawyer"

type LawyerRatingResponse struct {
	Rating      int    `json:"rating" jsonschema:"minimum=-1,maximum=10" jsonschema_description:"Rating from 0-10 for legal advice, or -1 if not applicable"`
	Explanation string `json:"explanation" jsonschema_description:"Brief explanation of the rating"`
}

type LawyerAgent struct {
	Config Config
	client *openai.Client
	schema any
}

func NewLawyerAgent(apiKey string) *LawyerAgent {
	client := openai.NewClient(option.WithAPIKey(apiKey))

	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	schema := reflector.Reflect(LawyerRatingResponse{})

	config := Config{
		Name:  LawyerAgentName,
		Model: "gpt-5-mini",
		Prompt: "You are a legal scholar, world famous for your expertise in the law. Your client is coming to you to ask about some of the advice they were given. You are tasked with analyzing the advice to provide a rating from 0-10 on how accurate the advice is. If the advice isn't applicable to matter of the law, please return -1.",
	}

	return &LawyerAgent{
		Config: config,
		client: &client,
		schema: schema,
	}
}

func (l *LawyerAgent) Start(ctx context.Context, input <-chan string, output chan<- string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case text, ok := <-input:
			if !ok {
				close(output)
				return nil
			}

			result, err := l.callOpenAIStructured(ctx, text)
			if err != nil {
				continue
			}

			output <- fmt.Sprintf("%d - %s", result.Rating, result.Explanation)
			close(output)
			return nil
		}
	}
}

func (l *LawyerAgent) callOpenAIStructured(ctx context.Context, prompt string) (*LawyerRatingResponse, error) {
	callCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "lawyer_rating_response",
		Description: openai.String("Legal advice rating and explanation"),
		Schema:      l.schema,
		Strict:      openai.Bool(true),
	}

	completion, err := l.client.Chat.Completions.New(callCtx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(l.Config.Prompt),
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

	var rating LawyerRatingResponse
	if err := json.Unmarshal([]byte(completion.Choices[0].Message.Content), &rating); err != nil {
		return nil, fmt.Errorf("failed to parse structured response: %w", err)
	}

	return &rating, nil
}
