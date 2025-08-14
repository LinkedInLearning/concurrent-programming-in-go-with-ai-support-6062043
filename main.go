package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
)

const (
	EnvOpenAIAPIKey = "OPENAI_API_KEY"
)

// main is the entry point of the application that initializes the environment and starts the TUI.
func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or could not be loaded: %v", err)
	}

	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		fmt.Println("🤖 Agentic Application")
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
		fmt.Println("  • Writer - Generates content about startup companies")
		fmt.Println("  • Summarizer - Creates concise summaries")
		fmt.Println("  • Rater - Provides structured ratings (1-10)")
		fmt.Println("  • Titler - Generates compelling titles")
		fmt.Println("  • MarkdownFormatter - Formats results as markdown")
		fmt.Println()
		fmt.Println("The final output is rendered with beautiful markdown formatting!")
		return
	}

	program := tea.NewProgram(initialModel())
	if _, err := program.Run(); err != nil {
		log.Fatal(err)
	}
}