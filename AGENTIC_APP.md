# Agentic Application

A concurrent Go application that demonstrates an agentic workflow using the OpenAI API. The application coordinates multiple AI agents that work together to generate, analyze, and process content about startup companies.

## Architecture

The application consists of five agents working in a scatter-gather pattern:

```
                    ┌─────────────────┐
                    │  Writer Agent   │
                    │   (Generates    │
                    │   startup       │
                    │   content)      │
                    └─────────┬───────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │   Writer Output │
                    │   (Paragraph    │
                    │   about startup)│
                    └─────────┬───────┘
                              │
                              ▼
              ┌───────────────┼───────────────┐
              │               │               │
              ▼               ▼               ▼
    ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
    │ Summarizer  │ │   Rater     │ │   Titler    │
    │   Agent     │ │   Agent     │ │   Agent     │
    │ (2 sentence │ │ (1-10 with  │ │ (Compelling │
    │  summary)   │ │explanation) │ │   title)    │
    └─────────────┘ └─────────────┘ └─────────────┘
              │               │               │
              └───────────────┼───────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │   All Results   │
                    │   Combined      │
                    └─────────┬───────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │ Markdown        │
                    │ Formatter       │
                    │ Agent           │
                    │ (Final output)  │
                    └─────────────────┘
```

**Flow Description:**

1. **Writer Agent** generates original content about startup companies
2. **Scatter Phase**: Writer output is sent to three agents concurrently:
   - **Summarizer Agent** - Creates a two-sentence summary
   - **Structured Rater Agent** - Provides a 1-10 rating with explanation
   - **Titler Agent** - Generates a compelling title
3. **Gather Phase**: All results are collected and combined
4. **Markdown Formatter Agent** - Formats everything into a beautiful markdown document

The application consists of five agents:

1. **Writer Agent** - Generates a paragraph about building a startup company
2. **Summarizer Agent** - Summarizes the paragraph into two sentences
3. **Structured Rater Agent** - Uses JSON schema to return a structured rating (1-10) with explanation
4. **Titler Agent** - Creates a compelling title for the content
5. **Markdown Formatter Agent** - Formats all results into a beautiful markdown document

## Features

- **Concurrent Processing**: Agents run concurrently using Go channels and goroutines
- **Structured JSON Output**: Rater agent uses OpenAI's structured output with JSON schema validation
- **Beautiful Markdown Rendering**: Final output rendered with Charm's Glamour library
- **Loading Indicators**: Shows spinners while AI agents are processing
- **Error Handling**: Graceful error handling with styled error messages
- **Modular Design**: Each agent is in its own file with a common interface

## Prerequisites

- Go 1.24.5 or later
- OpenAI API key

## Installation

1. Clone the repository
2. Install dependencies:

   ```bash
   go mod tidy
   ```

3. Set your OpenAI API key by creating a `.env` file:

   ```bash
   cp .env.sample .env
   # Edit .env and add your actual API key
   ```

   Or set it as an environment variable:

   ```bash
   export OPENAI_API_KEY="your-api-key-here"
   ```

## Usage

Run the application:

```bash
go run main.go
```

The application will:

1. Generate a paragraph about startup companies using the writer agent
2. Concurrently process the paragraph with four agents:
   - Summarizer: Creates a two-sentence summary
   - Rater: Provides a structured helpfulness/accuracy rating (1-10)
   - Titler: Generates a compelling title
   - MarkdownFormatter: Compiles everything into beautiful markdown
3. Display the final result with stunning terminal markdown rendering

## Project Structure

```
.
├── main.go                 # Main application with UI (much simpler now!)
├── agent/
│   ├── agent.go           # Agent interface and configuration
│   ├── base.go            # Base agent with OpenAI integration and timeouts
│   ├── workflow.go        # Workflow orchestration (handles all complexity)
│   ├── writer.go          # Content writer agent
│   ├── summarizer.go      # Text summarizer agent
│   ├── structured_rater.go # Structured JSON rater agent with schema validation
│   ├── titler.go          # Title generator agent
│   └── markdown_formatter.go # Markdown formatter agent
├── go.mod
├── go.sum
└── README.md
```

## Clean Architecture

The application now follows a much cleaner architecture:

- **main.go**: Only handles UI and calls the workflow
- **agent/workflow.go**: Contains all orchestration logic, timeouts, and error handling
- **agent/base.go**: Handles OpenAI API calls with built-in timeouts
- **Individual agents**: Focus only on their specific tasks

## Key Improvements

- **Separation of Concerns**: UI logic separated from business logic
- **Built-in Timeouts**: All agents have automatic timeout handling
- **Simplified Error Handling**: Centralized in the workflow package
- **Cleaner Main**: No complex channel orchestration in main.go
- **Reusable Workflow**: The workflow can be used independently of the UI

## Agent Interface

All agents implement the `Agent` interface:

```go
type Agent interface {
    Start(ctx context.Context, input <-chan string, output chan<- string) error
}
```

Each agent is configured with:

- **Name**: Agent identifier
- **Model**: OpenAI model to use (gpt-4o)
- **Prompt**: Agent-specific system prompt

## Dependencies

- `github.com/openai/openai-go` - Official OpenAI Go client with structured output support
- `github.com/charmbracelet/glamour` - Terminal markdown renderer
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `github.com/charmbracelet/bubbles` - Terminal UI components
- `github.com/charmbracelet/bubbletea` - Terminal UI framework
- `github.com/joho/godotenv` - Environment variable loading from .env files
- `github.com/invopop/jsonschema` - JSON schema generation for structured outputs

## Example Output

The application now displays results as beautifully rendered markdown instead of styled boxes:

- **Structured Rating**: The rater agent returns a guaranteed integer 1-10 with explanation using JSON schema validation
- **Markdown Formatting**: All results are compiled into a professional markdown document
- **Glamour Rendering**: The final markdown is rendered with syntax highlighting and beautiful formatting
- **Error Handling**: Individual agent results shown if any errors occur
- **Loading Spinner**: Animated spinner while processing (typically 60-90 seconds)

The final output includes:

- Original startup content from the writer
- Two-sentence summary
- Structured rating (e.g., "8/10 - Comprehensive and practical advice...")
- Compelling title
- All formatted as a beautiful markdown document with headers, formatting, and structure

## Error Handling

The application handles various error scenarios:

- Missing OpenAI API key
- API call failures
- Network timeouts
- Agent communication errors

All errors are displayed in styled red boxes with clear error messages.
