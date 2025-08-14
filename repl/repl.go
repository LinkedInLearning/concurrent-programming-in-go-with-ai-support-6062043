package repl

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"concurrent-programming-go-agents/pipeline"
)

type REPL struct {
	pipeline *pipeline.Pipeline
	scanner  *bufio.Scanner
}

func NewREPL(pipeline *pipeline.Pipeline) *REPL {
	return &REPL{
		pipeline: pipeline,
		scanner:  bufio.NewScanner(os.Stdin),
	}
}

func (r *REPL) Start(ctx context.Context) {
	fmt.Println("🔍 RSS Vector Search REPL")
	fmt.Println("Ask questions about the articles and get AI-powered responses!")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  <your question> - Ask about the articles")
	fmt.Println("  help - Show this help message")
	fmt.Println("  quit - Exit the REPL")
	fmt.Println()

	for {
		fmt.Print("❓ Ask me anything: ")
		if !r.scanner.Scan() {
			break
		}

		input := strings.TrimSpace(r.scanner.Text())
		if input == "" {
			continue
		}

		switch strings.ToLower(input) {
		case "help":
			r.showHelp()
		case "quit", "exit", "q":
			fmt.Println("👋 Goodbye!")
			return
		default:
			r.handleQuery(ctx, input)
		}
	}
}

func (r *REPL) handleQuery(ctx context.Context, query string) {
	fmt.Printf("🔍 Searching for articles related to: %s\n", query)
	
	// Search for relevant articles
	articles, err := r.pipeline.SearchArticles(ctx, query, 5)
	if err != nil {
		fmt.Printf("❌ Error searching articles: %v\n", err)
		return
	}

	if len(articles) == 0 {
		fmt.Println("😔 No relevant articles found.")
		return
	}

	fmt.Printf("📚 Found %d relevant articles:\n\n", len(articles))
	
	// Display articles
	for i, article := range articles {
		fmt.Printf("%d. 📰 %s\n", i+1, article.Title)
		fmt.Printf("   📝 %s\n", truncateString(article.Summary, 150))
		fmt.Printf("   🔗 %s\n", article.Link)
		fmt.Printf("   📅 %s\n", article.PublicationDate.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}

	// Generate AI response
	fmt.Println("🤖 Generating AI response...")
	response, err := r.pipeline.GenerateResponse(ctx, query, articles)
	if err != nil {
		fmt.Printf("❌ Error generating response: %v\n", err)
	} else {
		fmt.Println("💡 AI Response:")
		fmt.Println("─────────────────────────────────────────────────────────────")
		fmt.Println(response)
		fmt.Println("─────────────────────────────────────────────────────────────")
	}

	fmt.Print("\n📖 Enter article number to open in browser (or press Enter to continue): ")
	if r.scanner.Scan() {
		input := strings.TrimSpace(r.scanner.Text())
		if input != "" {
			if num, err := strconv.Atoi(input); err == nil && num > 0 && num <= len(articles) {
				r.openInBrowser(articles[num-1].Link)
			} else {
				fmt.Println("❌ Invalid article number.")
			}
		}
	}
	fmt.Println()
}

func (r *REPL) showHelp() {
	fmt.Println("🆘 Available commands:")
	fmt.Println("  Ask any question about the articles - Get AI-powered responses with relevant articles")
	fmt.Println("  help - Show this help message")
	fmt.Println("  quit - Exit the REPL")
	fmt.Println()
	fmt.Println("💡 Examples:")
	fmt.Println("  What are the latest trends in AI?")
	fmt.Println("  Tell me about machine learning developments")
	fmt.Println("  What programming languages are mentioned?")
	fmt.Println("  Summarize the key points about technology")
}

func (r *REPL) openInBrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)

	if err := exec.Command(cmd, args...).Start(); err != nil {
		fmt.Printf("❌ Failed to open browser: %v\n", err)
		fmt.Printf("🔗 Please open this URL manually: %s\n", url)
	} else {
		fmt.Printf("🌐 Opening %s in browser...\n", url)
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}