package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const AdviceSummarizerAgentName = "AdviceSummarizer"

type AdviceSummaryResponse struct {
	FinalRating string `json:"final_rating" jsonschema:"enum=terrible,enum=bad,enum=neutral,enum=good,enum=fantastic" jsonschema_description:"Final rating: terrible, bad, neutral, good, or fantastic"`
	Summary     string `json:"summary" jsonschema_description:"Summary of all expert opinions and reasoning for final rating"`
}

type AdviceSummarizerAgent struct {
	Config Config
	client *openai.Client
	schema any
}

func NewAdviceSummarizerAgent(apiKey string) *AdviceSummarizerAgent {
	client := openai.NewClient(option.WithAPIKey(apiKey))

	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	schema := reflector.Reflect(AdviceSummaryResponse{})

	config := Config{
		Name:  AdviceSummarizerAgentName,
		Model: "gpt-5-mini",
		Prompt: "You are an expert advice analyst. You will receive ratings and explanations from multiple expert agents about a piece of advice. Your job is to summarize their opinions and provide a final rating. Use this scale: 0-2 = terrible, 3-4 = bad, 5 = neutral, 6-7 = good, 8-10 = fantastic. Average the scores from experts who provided ratings (ignore -1 scores), then convert to the final rating scale.",
	}

	return &AdviceSummarizerAgent{
		Config: config,
		client: &client,
		schema: schema,
	}
}

func (a *AdviceSummarizerAgent) Start(ctx context.Context, input <-chan string, output chan<- string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case text, ok := <-input:
			if !ok {
				close(output)
				return nil
			}

			result, err := a.callOpenAIStructured(ctx, text)
			if err != nil {
				continue
			}

			output <- fmt.Sprintf("Final Rating: %s\n\nSummary: %s", strings.ToUpper(result.FinalRating), result.Summary)
			close(output)
			return nil
		}
	}
}

func (a *AdviceSummarizerAgent) callOpenAIStructured(ctx context.Context, prompt string) (*AdviceSummaryResponse, error) {
	callCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "advice_summary_response",
		Description: openai.String("Final advice rating and summary"),
		Schema:      a.schema,
		Strict:      openai.Bool(true),
	}

	completion, err := a.client.Chat.Completions.New(callCtx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(a.Config.Prompt),
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

	var summary AdviceSummaryResponse
	if err := json.Unmarshal([]byte(completion.Choices[0].Message.Content), &summary); err != nil {
		return nil, fmt.Errorf("failed to parse structured response: %w", err)
	}

	return &summary, nil
}

func CalculateAverageRating(ratings []string) float64 {
	var validRatings []int
	for _, rating := range ratings {
		parts := strings.Split(rating, " - ")
		if len(parts) > 0 {
			if score, err := strconv.Atoi(parts[0]); err == nil && score != -1 {
				validRatings = append(validRatings, score)
			}
		}
	}

	if len(validRatings) == 0 {
		return 0
	}

	sum := 0
	for _, rating := range validRatings {
		sum += rating
	}

	return float64(sum) / float64(len(validRatings))
}
