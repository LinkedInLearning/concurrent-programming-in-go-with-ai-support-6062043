package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"concurrent-programming-go-agents/agent"
	"github.com/joho/godotenv"
)

const (
	EnvOpenAIAPIKey = "OPENAI_API_KEY"
)

// main is the entry point of the application that initializes the environment and runs the workflow.
func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or could not be loaded: %v", err)
	}

	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		fmt.Println("Agentic Application")
		fmt.Println()
		fmt.Println("A concurrent Go application demonstrating an agentic workflow using OpenAI.")
		fmt.Println()
		fmt.Println("SETUP:")
		fmt.Println("  Set OPENAI_API_KEY environment variable or create a .env file")
		fmt.Println()
		fmt.Println("USAGE:")
		fmt.Println("  go run main.go")
		fmt.Println("  ./agentic-app")
		fmt.Println()
		fmt.Println("AGENTS:")
		fmt.Println("  - Writer: Generates content about startup companies")
		fmt.Println("  - Summarizer: Creates concise summaries")
		fmt.Println("  - Rater: Provides structured ratings (1-10)")
		fmt.Println("  - Titler: Generates compelling titles")
		fmt.Println()
		return
	}

	// Check for API key
	apiKey := os.Getenv(EnvOpenAIAPIKey)
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set")
	}

	fmt.Println("Starting agentic workflow...")
	fmt.Println()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create and run workflow
	workflow := agent.NewWorkflow(apiKey, ctx, func(status string) {
		fmt.Printf("Status: %s\n", status)
	})

	results := workflow.Run()
	stats := workflow.GetStats()

	// Display results
	fmt.Println()
	fmt.Println("=== WORKFLOW RESULTS ===")
	fmt.Println()

	// Check for errors first
	hasErrors := false
	for _, result := range results {
		if result.Error != nil {
			hasErrors = true
			fmt.Printf("ERROR in %s: %v\n", result.AgentName, result.Error)
		}
	}

	if hasErrors {
		fmt.Println()
		fmt.Println("Workflow completed with errors.")
		return
	}

	// Display successful results
	for _, result := range results {
		if result.Error == nil {
			fmt.Printf("%s:\n%s\n\n", result.AgentName, result.Output)
		}
	}

	// Display timing statistics
	fmt.Println("=== TIMING STATISTICS ===")
	fmt.Printf("Total Duration: %v\n\n", stats.TotalDuration)
	fmt.Println("Individual Agent Timings:")

	agentOrder := []string{
		agent.WriterAgentName,
		agent.SummarizerAgentName,
		agent.RaterAgentName,
		agent.TitlerAgentName,
	}

	for _, agentName := range agentOrder {
		if duration, exists := stats.AgentStats[agentName]; exists {
			fmt.Printf("  %s: %v\n", agentName, duration)
		}
	}

	fmt.Println()
	fmt.Println("Workflow completed successfully!")
}