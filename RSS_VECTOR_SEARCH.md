# RSS Vector Search

A Go application that processes RSS/Atom feeds, creates vector embeddings using OpenAI, stores them in Weaviate, and provides a REPL interface for semantic search.

## Features

- **RSS/Atom Feed Processing**: Parse feeds from URLs or files
- **Vector Embeddings**: Generate embeddings using OpenAI's text-embedding-ada-002 model
- **Vector Database**: Store and search embeddings using Weaviate (local binary)
- **Semantic Search**: Find relevant articles using natural language queries
- **Interactive REPL**: Command-line interface for searching and browsing articles
- **Browser Integration**: Open articles directly in your default browser

## Architecture

The application consists of several components:

- `rss/`: RSS feed parsing and item extraction
- `embedding/`: OpenAI embedding service
- `vectordb/`: Weaviate database integration with Testcontainers
- `pipeline/`: Main processing pipeline that orchestrates the workflow
- `repl/`: Interactive command-line interface

## Prerequisites

- Go 1.24.5 or later
- OpenAI API key
- Internet connection (for downloading Weaviate binary)

## Setup

1. **Set up environment variables**:
   ```bash
   cp .env.rss-example .env
   # Edit .env and add your OpenAI API key
   ```

2. **Install dependencies**:
   ```bash
   go mod tidy
   ```

## Usage

### Basic Usage

Run the application and it will prompt you to select an RSS feed:

```bash
go run main.go
```

The application will show you popular RSS feed options:

```
📰 Popular RSS Feeds:
  1. O'Reilly Radar: https://feeds.feedburner.com/oreilly/radar
  2. Hacker News: https://hnrss.org/frontpage
  3. TechCrunch: https://techcrunch.com/feed/
  4. Ars Technica: http://feeds.arstechnica.com/arstechnica/index
  5. The Verge: https://www.theverge.com/rss/index.xml

🔗 Enter RSS feed URL (or number 1-5): 
```

You can either:
- Enter a number (1-5) to select a popular feed
- Enter any RSS/Atom feed URL starting with http:// or https://

### Download Progress

On first run, the application will download the Weaviate binary with a progress bar:

```
📥 Downloading: [========================================] 100% (45.2 MB / 45.2 MB)
```

### REPL Commands

Once the application starts, you can interact with it naturally:

- Simply ask questions about the articles in natural language
- Type `help` to show available commands
- Type `quit` to exit the application

### Example Session

```
🔍 RSS Vector Search REPL
Ask questions about the articles and get AI-powered responses!

Commands:
  <your question> - Ask about the articles
  help - Show this help message
  quit - Exit the REPL

❓ Ask me anything: What are the latest trends in AI?
Searching for: machine learning

Found 3 articles:

1. Introduction to Machine Learning with Python
   Description: A comprehensive guide to getting started with machine learning using Python and scikit-learn...
   Link: https://example.com/ml-python-intro
   Published: 2024-01-15 10:30:00

2. Deep Learning Fundamentals
   Description: Understanding the basics of neural networks and deep learning architectures...
   Link: https://example.com/deep-learning-basics
   Published: 2024-01-10 14:20:00

Enter article number to open in browser (or press Enter to continue): 1
Opening https://example.com/ml-python-intro in browser...

> quit
Goodbye!
```

## Architecture Overview

### System Components Diagram

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           RSS Vector Search System                              │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌─────────────┐    ┌──────────────┐    ┌─────────────────┐    ┌─────────────┐  │
│  │    User     │    │     RSS      │    │   Worker Pool   │    │  Weaviate   │  │
│  │  Interface  │    │    Parser    │    │  (10 workers)   │    │   Binary    │  │
│  │   (REPL)    │    │              │    │                 │    │ (v1.32.3)   │  │
│  └─────────────┘    └──────────────┘    └─────────────────┘    └─────────────┘  │
│         │                   │                      │                    │       │
│         │                   │                      │                    │       │
│  ┌─────────────┐    ┌──────────────┐    ┌─────────────────┐    ┌─────────────┐  │
│  │   Pipeline  │    │  Summarizer  │    │   Embedding     │    │   Vector    │  │
│  │ Orchestrator│    │ (GPT-5-mini) │    │   Service       │    │  Database   │  │
│  │             │    │              │    │ (OpenAI Ada-002)│    │ (HNSW Index)│  │
│  └─────────────┘    └──────────────┘    └─────────────────┘    └─────────────┘  │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### Data Processing Flow

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                            Processing Pipeline                                  │
└─────────────────────────────────────────────────────────────────────────────────┘

1. RSS Feed Input
   │
   ▼
┌─────────────────┐
│   RSS Parser    │ ──► Parse feed items (title, description, link, date)
│  (gofeed lib)   │
└─────────────────┘
   │
   ▼ (First 50 items)
┌─────────────────┐
│  Job Channel    │ ──► Distribute work to 10 concurrent workers
│   (buffered)    │
└─────────────────┘
   │
   ▼ (Parallel Processing)
┌─────────────────────────────────────────────────────────────────────────────────┐
│                            Worker Pool (10 workers)                             │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  Worker 1    Worker 2    Worker 3    ...    Worker 10                           │
│     │           │           │                   │                               │
│     ▼           ▼           ▼                   ▼                               │
│  ┌─────────────────────────────────────────────────────────────────────────┐    │
│  │                    Per-Worker Process                                   │    │
│  │                                                                         │    │
│  │  1. Receive Job ──► Article {title, description, link, date}            │    │
│  │     │                                                                   │    │
│  │     ▼                                                                   │    │
│  │  2. Summarize ──► GPT-5-mini API Call                                   │    │
│  │     │              "Create concise summary using ONLY provided info"    │    │
│  │     ▼                                                                   │    │
│  │  3. Vectorize ──► OpenAI Embedding API                                  │    │
│  │     │              text-embedding-ada-002 model                         │    │
│  │     │              Input: "Title: X\nDescription: Y\nSummary: Z"        │    │
│  │     │              Output: [1536 float32 values]                        │    │
│  │     ▼                                                                   │    │
│  │  4. Store ────► Weaviate HTTP API                                       │    │
│  │                 POST /v1/objects                                        │    │
│  │                 {properties: {...}, vector: [...]}                      │    │
│  └─────────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────────┘
   │
   ▼ (Results Channel)
┌─────────────────┐
│ Result Collector│ ──► Track success/failure, display progress
│                 │
└─────────────────┘
   │
   ▼
┌─────────────────┐
│  REPL Ready     │ ──► Interactive search interface
│                 │
└─────────────────┘
```

### Search Query Flow

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              Search Process                                     │
└─────────────────────────────────────────────────────────────────────────────────┘

User Query: "What are the latest AI trends?"
   │
   ▼
┌─────────────────┐
│   REPL Input    │ ──► Capture user query
│                 │
└─────────────────┘
   │
   ▼
┌─────────────────┐
│   Pipeline      │ ──► Orchestrate search process
│  SearchArticles │
└─────────────────┘
   │
   ▼
┌─────────────────┐
│ Query Embedding │ ──► OpenAI API Call
│ Service         │     text-embedding-ada-002
│                 │     Input: "What are the latest AI trends?"
│                 │     Output: [1536 float32 query vector]
└─────────────────┘
   │
   ▼
┌─────────────────┐
│   Weaviate      │ ──► Vector Similarity Search
│ SearchSimilar   │     POST /v1/graphql
│                 │     {
│                 │       Get {
│                 │         Article(
│                 │           nearVector: {
│                 │             vector: [query_vector]
│                 │             certainty: 0.7
│                 │           }
│                 │           limit: 5
│                 │         ) { title, summary, link, ... }
│                 │       }
│                 │     }
└─────────────────┘
   │
   ▼
┌─────────────────┐
│ Vector Database │ ──► HNSW Index Search
│ (Local Binary)  │     - Cosine similarity calculation
│                 │     - Rank by similarity score
│                 │     - Filter by certainty > 0.7
│                 │     - Return top 5 matches
└─────────────────┘
   │
   ▼
┌─────────────────┐
│ Matching        │ ──► [
│ Articles        │       {title: "AI Breakthrough...", summary: "...", link: "..."},
│                 │       {title: "Machine Learning...", summary: "...", link: "..."},
│                 │       ...
│                 │     ]
└─────────────────┘
   │
   ▼
┌─────────────────┐
│   Response      │ ──► GPT-5-mini API Call
│  Generation     │     "Answer query using ONLY these articles:"
│ (GPT-5-mini)    │     - Check article relevance
│                 │     - Generate contextual response
│                 │     - Or say "no matches found"
└─────────────────┘
   │
   ▼
┌─────────────────┐
│   REPL Output   │ ──► Display:
│                 │     - AI-generated response
│                 │     - List of relevant articles
│                 │     - Option to open in browser
└─────────────────┘
```

## How It Works

1. **Feed Processing**: The application parses RSS/Atom feeds and extracts article metadata (title, description, summary, link, publication date)

2. **Embedding Generation**: Each article's content is sent to OpenAI's embedding API to generate a 1536-dimensional vector representation

3. **Vector Storage**: The embeddings and metadata are stored in a local Weaviate vector database binary

4. **Semantic Search**: When you search, your query is also converted to an embedding and compared against stored articles using cosine similarity

5. **Results Display**: Matching articles are ranked by similarity and displayed with options to open in your browser

## Configuration

The application uses the following environment variables:

- `OPENAI_API_KEY`: Your OpenAI API key (required)

## Dependencies

Key dependencies include:

- `github.com/mmcdole/gofeed`: RSS/Atom feed parsing
- `github.com/openai/openai-go`: OpenAI API client
- Standard Go libraries for HTTP, tar/gzip extraction, and process management

## Troubleshooting

### Binary Download Issues
- Ensure you have internet connectivity
- The application automatically downloads Weaviate v1.32.3 binary on first run
- Binary is cached in `./weaviate-data/` directory for subsequent runs

### OpenAI API Issues
- Verify your API key is correct and has sufficient credits
- Check your internet connection for API access

### Feed Parsing Issues
- Ensure the RSS/Atom feed URL is accessible
- Some feeds may require specific user agents or headers (not currently supported)

## Extending the Application

You can extend this application by:

- Adding support for multiple feed sources
- Implementing feed refresh/update mechanisms  
- Adding more sophisticated text preprocessing
- Supporting different embedding models
- Adding web interface instead of REPL
- Implementing user authentication and multi-tenancy