package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"concurrent-programming-go-agents/embedding"
	"concurrent-programming-go-agents/pipeline"
	"concurrent-programming-go-agents/repl"
	"concurrent-programming-go-agents/rss"
	"concurrent-programming-go-agents/summarizer"
	"concurrent-programming-go-agents/vectordb"

	"github.com/joho/godotenv"
)

const (
	MaxItems    = 50
	WorkerCount = 10
)

type ProcessingJob struct {
	Item  *rss.FeedItem
	Index int
}

type ProcessingResult struct {
	Article *vectordb.Article
	Index   int
	Error   error
	Skipped bool
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	fmt.Println("🔍 RSS Vector Search Pipeline")
	fmt.Println("=============================")
	fmt.Println()
	fmt.Println("This application will:")
	fmt.Println("  📥 Parse your RSS feed")
	fmt.Println("  🤖 Generate summaries using GPT-5-mini")
	fmt.Println("  🧠 Create vector embeddings")
	fmt.Println("  💾 Store in local Weaviate database")
	fmt.Println("  🔍 Enable semantic search via REPL")
	fmt.Println()

	// Prompt for RSS feed URL
	feedURL := promptForFeedURL()

	ctx := context.Background()

	log.Println("🚀 Starting RSS Vector Search Pipeline...")
	log.Printf("📡 Processing RSS feed: %s", feedURL)
	log.Printf("📊 Processing first %d items with %d workers", MaxItems, WorkerCount)

	// Initialize services
	feedProcessor := rss.NewFeedProcessor()
	embeddingService := embedding.NewEmbeddingService(openaiAPIKey)
	summarizerService := summarizer.NewSummarizer(openaiAPIKey)

	vectorDB, err := vectordb.NewWeaviateDB(ctx)
	if err != nil {
		log.Fatalf("❌ Failed to initialize vector database: %v", err)
	}
	defer func() {
		if err := vectorDB.Close(ctx); err != nil {
			log.Printf("⚠️ Error closing vector database: %v", err)
		}
	}()

	// Parse RSS feed
	log.Println("📖 Parsing RSS feed...")
	feedItems, err := feedProcessor.ParseFeedFromURL(feedURL)
	if err != nil {
		log.Fatalf("❌ Failed to parse RSS feed: %v", err)
	}

	// Limit to first 50 items
	if len(feedItems) > MaxItems {
		feedItems = feedItems[:MaxItems]
	}

	log.Printf("✅ Found %d items to process", len(feedItems))

	// Process items with worker pool
	if err := processItemsWithWorkerPool(ctx, feedItems, embeddingService, summarizerService, vectorDB); err != nil {
		log.Fatalf("❌ Failed to process items: %v", err)
	}

	log.Println("🎉 Processing complete! Starting REPL...")

	// Start REPL
	p := pipeline.NewPipeline(embeddingService, summarizerService, vectorDB)

	r := repl.NewREPL(p)
	r.Start(ctx)
}

func processItemsWithWorkerPool(ctx context.Context, items []*rss.FeedItem, embeddingService *embedding.EmbeddingService, summarizerService *summarizer.Summarizer, vectorDB *vectordb.WeaviateDB) error {
	jobs := make(chan ProcessingJob, len(items))
	results := make(chan ProcessingResult, len(items))

	// Start workers
	var wg sync.WaitGroup
	for i := range WorkerCount {
		wg.Add(1)
		go worker(ctx, i+1, jobs, results, embeddingService, summarizerService, vectorDB, &wg)
	}

	// Send jobs
	go func() {
		defer close(jobs)
		for i, item := range items {
			jobs <- ProcessingJob{Item: item, Index: i + 1}
		}
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(results)
	}()

	successCount := 0
	errorCount := 0
	skippedCount := 0

	for result := range results {
		if result.Error != nil {
			log.Printf("❌ Worker failed to process item %d: %v", result.Index, result.Error)
			errorCount++
		} else if result.Skipped {
			skippedCount++
		} else {
			log.Printf("✅ Successfully processed item %d/%d: %s", result.Index, len(items), result.Article.Title)
			successCount++
		}
	}

	log.Printf("📈 Processing summary: %d successful, %d skipped (duplicates), %d failed", successCount, skippedCount, errorCount)
	return nil
}

func worker(ctx context.Context, workerID int, jobs <-chan ProcessingJob, results chan<- ProcessingResult, embeddingService *embedding.EmbeddingService, summarizerService *summarizer.Summarizer, vectorDB *vectordb.WeaviateDB, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Printf("🔧 Worker %d started", workerID)

	for job := range jobs {
		log.Printf("🔄 Worker %d processing item %d: %s", workerID, job.Index, job.Item.Title)

		// Check for duplicates first
		existingArticle, err := vectorDB.SearchByTitle(ctx, job.Item.Title)
		if err != nil {
			results <- ProcessingResult{Index: job.Index, Error: fmt.Errorf("failed to check for duplicates: %w", err)}
			continue
		}

		if existingArticle != nil {
			log.Printf("⏭️  Worker %d skipping duplicate item %d: %s", workerID, job.Index, job.Item.Title)
			results <- ProcessingResult{Index: job.Index, Skipped: true}
			continue
		}

		// Generate summary
		summary, err := summarizerService.Summarize(ctx, job.Item.Title, job.Item.Description)
		if err != nil {
			results <- ProcessingResult{Index: job.Index, Error: fmt.Errorf("failed to summarize: %w", err)}
			continue
		}

		// Generate embedding
		vector, err := embeddingService.GetEmbeddingForArticle(ctx, job.Item.Title, job.Item.Description, summary)
		if err != nil {
			results <- ProcessingResult{Index: job.Index, Error: fmt.Errorf("failed to get embedding: %w", err)}
			continue
		}

		// Create article
		article := &vectordb.Article{
			Title:           job.Item.Title,
			Description:     job.Item.Description,
			Summary:         summary,
			Link:            job.Item.Link,
			PublicationDate: job.Item.PublicationDate,
			Vector:          vector,
		}

		// Store in vector database
		if err := vectorDB.StoreArticle(ctx, article); err != nil {
			results <- ProcessingResult{Index: job.Index, Error: fmt.Errorf("failed to store article: %w", err)}
			continue
		}

		results <- ProcessingResult{Article: article, Index: job.Index, Error: nil}
	}

	log.Printf("🏁 Worker %d finished", workerID)
}

func promptForFeedURL() string {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("📰 Popular RSS Feeds:")
	fmt.Println("  1. O'Reilly Radar: https://feeds.feedburner.com/oreilly/radar")
	fmt.Println("  2. Hacker News: https://hnrss.org/frontpage")
	fmt.Println("  3. TechCrunch: https://techcrunch.com/feed/")
	fmt.Println("  4. Ars Technica: http://feeds.arstechnica.com/arstechnica/index")
	fmt.Println("  5. The Verge: https://www.theverge.com/rss/index.xml")
	fmt.Println()

	for {
		fmt.Print("🔗 Enter RSS feed URL (or number 1-5): ")

		if !scanner.Scan() {
			fmt.Println("\n👋 Goodbye!")
			os.Exit(0)
		}

		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			fmt.Println("⚠️  Please enter a valid URL or number.")
			continue
		}

		// Handle numbered options
		switch input {
		case "1":
			return "https://feeds.feedburner.com/oreilly/radar"
		case "2":
			return "https://hnrss.org/frontpage"
		case "3":
			return "https://techcrunch.com/feed/"
		case "4":
			return "http://feeds.arstechnica.com/arstechnica/index"
		case "5":
			return "https://www.theverge.com/rss/index.xml"
		default:
			// Validate URL format
			if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
				return input
			}
			fmt.Println("⚠️  Please enter a valid URL starting with http:// or https://")
		}
	}
}
