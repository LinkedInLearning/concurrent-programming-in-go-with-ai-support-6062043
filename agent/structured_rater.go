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

// RaterAgentName is the identifier for the structured rater agent.
const RaterAgentName = "Rater"

// RatingResponse represents the structured output from the rating agent.
type RatingResponse struct {
	Rating      int    `json:"rating" jsonschema:"minimum=1,maximum=10" jsonschema_description:"Numeric rating from 1-10 where 10 is extremely helpful and accurate"`
	Explanation string `json:"explanation" jsonschema_description:"Brief explanation of the rating"`
}

// StructuredRaterAgent provides structured ratings with explanations for input text.
type StructuredRaterAgent struct {
	Config Config
	client *openai.Client
	schema any
}

// NewStructuredRaterAgent creates a new StructuredRaterAgent with JSON schema for structured output.
func NewStructuredRaterAgent(apiKey string) *StructuredRaterAgent {
	client := openai.NewClient(option.WithAPIKey(apiKey))

	// Generate JSON schema for structured output
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	schema := reflector.Reflect(RatingResponse{})

	config := Config{
		Name:   RaterAgentName,
		Model:  "gpt-5-nano",
		Prompt: "You are an expert evaluator. Rate the provided text for helpfulness and accuracy on a scale of 1-10, where 10 is extremely helpful and accurate. Provide your rating and a brief explanation.",
	}

	return &StructuredRaterAgent{
		Config: config,
		client: &client,
		schema: schema,
	}
}

// Start processes input text and generates structured ratings, sending results to the output channel.
func (r *StructuredRaterAgent) Start(ctx context.Context, input <-chan string, output chan<- string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case text, ok := <-input:
			if !ok {
				close(output)
				return nil
			}

			result, err := r.callOpenAIStructured(ctx, fmt.Sprintf("Rate this text for helpfulness and accuracy: %s", text))
			if err != nil {
				continue
			}

			output <- fmt.Sprintf("%d/10 - %s", result.Rating, result.Explanation)
			close(output)
			return nil
		}
	}
}

// callOpenAIStructured makes a structured API call to OpenAI and returns a parsed RatingResponse.
func (r *StructuredRaterAgent) callOpenAIStructured(ctx context.Context, prompt string) (*RatingResponse, error) {
	// Add timeout to the context for OpenAI calls
	callCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "rating_response",
		Description: openai.String("Rating and explanation for text evaluation"),
		Schema:      r.schema,
		Strict:      openai.Bool(true),
	}

	completion, err := r.client.Chat.Completions.New(callCtx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(r.Config.Prompt),
			openai.UserMessage(prompt),
		},
		Model: openai.ChatModelGPT4o,
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

	var rating RatingResponse
	if err := json.Unmarshal([]byte(completion.Choices[0].Message.Content), &rating); err != nil {
		return nil, fmt.Errorf("failed to parse structured response: %w", err)
	}

	return &rating, nil
}
