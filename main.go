package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"concurrent-programming-go-agents/agent"
	"github.com/joho/godotenv"
)

const (
	EnvOpenAIAPIKey = "OPENAI_API_KEY"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or could not be loaded: %v", err)
	}

	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		fmt.Println("Advice Rating Tool")
		fmt.Println()
		fmt.Println("An agentic REPL tool that rates user-submitted advice using multiple expert agents.")
		fmt.Println()
		fmt.Println("SETUP:")
		fmt.Println("  Set OPENAI_API_KEY environment variable or create a .env file")
		fmt.Println()
		fmt.Println("USAGE:")
		fmt.Println("  go run main.go")
		fmt.Println("  ./advice-rater")
		fmt.Println()
		fmt.Println("EXPERT AGENTS:")
		fmt.Println("  - Career: Rates advice for career impact")
		fmt.Println("  - BestFriend: Rates advice for interpersonal relationships")
		fmt.Println("  - Financial: Rates advice for financial success")
		fmt.Println("  - TechSupport: Rates advice for technology accuracy")
		fmt.Println("  - Dietician: Rates advice for health and diet")
		fmt.Println("  - Lawyer: Rates advice for legal accuracy")
		fmt.Println("  - AdviceSummarizer: Provides final rating (terrible/bad/neutral/good/fantastic)")
		fmt.Println()
		return
	}

	apiKey := os.Getenv(EnvOpenAIAPIKey)
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set")
	}

	fmt.Println("🤖 Advice Rating Tool")
	fmt.Println("Enter advice to get it rated by expert agents, or 'quit' to exit.")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Enter advice: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if strings.ToLower(input) == "quit" || strings.ToLower(input) == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

		workflow := agent.NewAdviceWorkflow(apiKey, ctx, func(status string) {
			fmt.Printf("Status: %s\n", status)
		})

		results := workflow.RateAdvice(input)
		stats := workflow.GetStats()

		fmt.Println()
		fmt.Println("=== EXPERT RATINGS ===")
		fmt.Println()

		hasErrors := false
		var finalSummary string

		for _, result := range results {
			if result.Error != nil {
				hasErrors = true
				fmt.Printf("ERROR in %s: %v\n", result.AgentName, result.Error)
			} else {
				if result.AgentName == agent.AdviceSummarizerAgentName {
					finalSummary = result.Output
				} else {
					fmt.Printf("%s: %s\n", result.AgentName, result.Output)
				}
			}
		}

		if hasErrors {
			fmt.Println()
			fmt.Println("Analysis completed with some errors.")
		}

		if finalSummary != "" {
			fmt.Println()
			fmt.Println("=== FINAL ASSESSMENT ===")
			fmt.Println(finalSummary)
		}

		fmt.Printf("\nAnalysis completed in %v\n", stats.TotalDuration)
		fmt.Println()
		fmt.Println(strings.Repeat("-", 50))
		fmt.Println()

		cancel()
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading input: %v", err)
	}
}