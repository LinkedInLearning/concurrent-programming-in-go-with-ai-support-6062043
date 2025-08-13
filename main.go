package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const (
	ModelGPT35Turbo = "gpt-3.5-turbo"
	ModelGPT4oMini  = "gpt-4o-mini"
	ModelGPT4       = "gpt-4"
)

type LLMResponse struct {
	Model    string
	Prompt   string
	Response string
	Duration time.Duration
	Error    error
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	prompts := []string{
		"Write a haiku about programming",
		"Explain quantum computing in one sentence",
		"What is the meaning of life?",
		"Tell me a programming joke",
	}

	models := []string{
		ModelGPT35Turbo,
		ModelGPT4oMini,
		ModelGPT4,
	}

	fmt.Println("🚀 Starting concurrent LLM demo...")
	fmt.Printf("Testing %d models with %d prompts each (%d total requests)\n\n", 
		len(models), len(prompts), len(models)*len(prompts))

	responses := make(chan LLMResponse, len(models)*len(prompts))
	var wg sync.WaitGroup

	startTime := time.Now()

	for _, model := range models {
		for _, prompt := range prompts {
			wg.Add(1)
			go func(m, p string) {
				defer wg.Done()
				resp := callLLM(client, m, p)
				responses <- resp
			}(model, prompt)
		}
	}

	go func() {
		wg.Wait()
		close(responses)
	}()

	var allResponses []LLMResponse
	for resp := range responses {
		allResponses = append(allResponses, resp)
		if resp.Error != nil {
			fmt.Printf("❌ Error with %s: %v\n", resp.Model, resp.Error)
		} else {
			fmt.Printf("✅ Completed %s request in %v\n", resp.Model, resp.Duration)
			fmt.Printf("   📝 Prompt: %s\n", truncateString(resp.Prompt, 60))
			fmt.Printf("   🤖 Response: %s\n\n", formatResponse(resp.Response, 150))
		}
	}

	totalTime := time.Since(startTime)
	fmt.Printf("\n🎉 All requests completed in %v\n\n", totalTime)

	displayResults(allResponses)
}

func callLLM(client openai.Client, model, prompt string) LLMResponse {
	start := time.Now()
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	completion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Model: openai.ChatModel(model),
	})

	duration := time.Since(start)

	if err != nil {
		return LLMResponse{
			Model:    model,
			Prompt:   prompt,
			Response: "",
			Duration: duration,
			Error:    err,
		}
	}

	response := ""
	if len(completion.Choices) > 0 {
		response = completion.Choices[0].Message.Content
	}

	return LLMResponse{
		Model:    model,
		Prompt:   prompt,
		Response: response,
		Duration: duration,
		Error:    nil,
	}
}

func displayResults(responses []LLMResponse) {
	t := table.NewWriter()
	t.AppendHeader(table.Row{"Model", "Prompt", "Response", "Duration", "Status"})

	for _, resp := range responses {
		status := "✅ Success"
		response := resp.Response
		if resp.Error != nil {
			status = "❌ Error"
			response = resp.Error.Error()
		}

		if len(response) > 200 {
			response = response[:197] + "..."
		}

		t.AppendRow(table.Row{
			resp.Model,
			truncateString(resp.Prompt, 30),
			response,
			resp.Duration.Round(time.Millisecond),
			status,
		})
	}

	t.SetStyle(table.StyleColoredBright)
	t.SetTitle("🤖 Concurrent LLM Demo Results")
	t.SetCaption("Generated responses from multiple OpenAI models running concurrently")
	
	fmt.Println(t.Render())

	printSummary(responses)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatResponse(response string, maxLen int) string {
	// Replace newlines with spaces for console output
	formatted := strings.ReplaceAll(response, "\n", " ")
	formatted = strings.ReplaceAll(formatted, "\r", " ")
	// Remove extra spaces
	formatted = strings.Join(strings.Fields(formatted), " ")
	
	if len(formatted) <= maxLen {
		return formatted
	}
	return formatted[:maxLen-3] + "..."
}

func printSummary(responses []LLMResponse) {
	fmt.Println("\n📊 Summary Statistics:")
	
	modelStats := make(map[string]struct {
		count    int
		totalDur time.Duration
		errors   int
	})

	for _, resp := range responses {
		stats := modelStats[resp.Model]
		stats.count++
		stats.totalDur += resp.Duration
		if resp.Error != nil {
			stats.errors++
		}
		modelStats[resp.Model] = stats
	}

	summaryTable := table.NewWriter()
	summaryTable.AppendHeader(table.Row{"Model", "Requests", "Avg Duration", "Success Rate"})

	for model, stats := range modelStats {
		avgDuration := stats.totalDur / time.Duration(stats.count)
		successRate := float64(stats.count-stats.errors) / float64(stats.count) * 100
		
		summaryTable.AppendRow(table.Row{
			model,
			stats.count,
			avgDuration.Round(time.Millisecond),
			fmt.Sprintf("%.1f%%", successRate),
		})
	}

	summaryTable.SetStyle(table.StyleLight)
	fmt.Println(summaryTable.Render())
}