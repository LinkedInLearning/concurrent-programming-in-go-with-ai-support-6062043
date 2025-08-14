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

const BestFriendAgentName = "BestFriend"

type BestFriendRatingResponse struct {
	Rating      int    `json:"rating" jsonschema:"minimum=-1,maximum=10" jsonschema_description:"Rating from 0-10 for interpersonal advice, or -1 if not applicable"`
	Explanation string `json:"explanation" jsonschema_description:"Brief explanation of the rating"`
}

type BestFriendAgent struct {
	Config Config
	client *openai.Client
	schema any
}

func NewBestFriendAgent(apiKey string) *BestFriendAgent {
	client := openai.NewClient(option.WithAPIKey(apiKey))

	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	schema := reflector.Reflect(BestFriendRatingResponse{})

	config := Config{
		Name:  BestFriendAgentName,
		Model: "gpt-5-mini",
		Prompt: "You are a really good friend, and you know all sorts of interpersonal information. Your best friend has been given some advice, which they will relay to you. You are tasked with thinking about the advice to provide a rating from 0-10 on how good the advice would be for their interpersonal life. If the advice seems to not apply to personal relationships or their personal life, please return -1.",
	}

	return &BestFriendAgent{
		Config: config,
		client: &client,
		schema: schema,
	}
}

func (b *BestFriendAgent) Start(ctx context.Context, input <-chan string, output chan<- string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case text, ok := <-input:
			if !ok {
				close(output)
				return nil
			}

			result, err := b.callOpenAIStructured(ctx, text)
			if err != nil {
				continue
			}

			output <- fmt.Sprintf("%d - %s", result.Rating, result.Explanation)
			close(output)
			return nil
		}
	}
}

func (b *BestFriendAgent) callOpenAIStructured(ctx context.Context, prompt string) (*BestFriendRatingResponse, error) {
	callCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "bestfriend_rating_response",
		Description: openai.String("Interpersonal advice rating and explanation"),
		Schema:      b.schema,
		Strict:      openai.Bool(true),
	}

	completion, err := b.client.Chat.Completions.New(callCtx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(b.Config.Prompt),
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

	var rating BestFriendRatingResponse
	if err := json.Unmarshal([]byte(completion.Choices[0].Message.Content), &rating); err != nil {
		return nil, fmt.Errorf("failed to parse structured response: %w", err)
	}

	return &rating, nil
}
