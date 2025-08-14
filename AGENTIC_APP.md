# Agentic Application

A concurrent Go application that demonstrates an agentic workflow using the OpenAI API. The application coordinates multiple AI agents that work together to generate, analyze, and process content about startup companies.

## Architecture

The application consists of four agents working in a scatter-gather pattern:

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
                    │   Final Output  │
                    │   (Plain Text)  │
                    └─────────────────┘
```

**Flow Description:**

1. **Writer Agent** generates original content about startup companies
2. **Scatter Phase**: Writer output is sent to three agents concurrently:
   - **Summarizer Agent** - Creates a two-sentence summary
   - **Structured Rater Agent** - Provides a 1-10 rating with explanation
   - **Titler Agent** - Generates a compelling title
3. **Gather Phase**: All results are collected and displayed as plain text

The application consists of four agents:

1. **Writer Agent** - Generates a paragraph about building a startup company
2. **Summarizer Agent** - Summarizes the paragraph into two sentences
3. **Structured Rater Agent** - Uses JSON schema to return a structured rating (1-10) with explanation
4. **Titler Agent** - Creates a compelling title for the content

## Features

- **Concurrent Processing**: Agents run concurrently using Go channels and goroutines
- **Structured JSON Output**: Rater agent uses OpenAI's structured output with JSON schema validation
- **Simple Text Output**: Results displayed as plain text without fancy formatting
- **Status Updates**: Shows progress messages while AI agents are processing
- **Error Handling**: Graceful error handling with clear error messages
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
2. Concurrently process the paragraph with three agents:
   - Summarizer: Creates a two-sentence summary
   - Rater: Provides a structured helpfulness/accuracy rating (1-10)
   - Titler: Generates a compelling title
3. Display the results as plain text with timing statistics

## Project Structure

```
.
├── main.go                 # Main application (simple command-line interface)
├── agent/
│   ├── agent.go           # Agent interface and configuration
│   ├── base.go            # Base agent with OpenAI integration and timeouts
│   ├── workflow.go        # Workflow orchestration (handles all complexity)
│   ├── writer.go          # Content writer agent
│   ├── summarizer.go      # Text summarizer agent
│   ├── structured_rater.go # Structured JSON rater agent with schema validation
│   └── titler.go          # Title generator agent
├── go.mod
├── go.sum
└── README.md
```

## Clean Architecture

The application now follows a clean architecture:

- **main.go**: Simple command-line interface that calls the workflow
- **agent/workflow.go**: Contains all orchestration logic, timeouts, and error handling
- **agent/base.go**: Handles OpenAI API calls with built-in timeouts
- **Individual agents**: Focus only on their specific tasks

## Key Improvements

- **Separation of Concerns**: Command-line interface separated from business logic
- **Built-in Timeouts**: All agents have automatic timeout handling
- **Simplified Error Handling**: Centralized in the workflow package
- **Clean Main**: No complex channel orchestration in main.go
- **Reusable Workflow**: The workflow can be used independently of the interface
- **No External UI Dependencies**: Removed Bubble Tea and markdown rendering for simplicity

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
- `github.com/joho/godotenv` - Environment variable loading from .env files
- `github.com/invopop/jsonschema` - JSON schema generation for structured outputs

## Example Output

The application displays results as simple, clean text output:

- **Structured Rating**: The rater agent returns a guaranteed integer 1-10 with explanation using JSON schema validation
- **Plain Text Output**: All results are displayed as readable plain text
- **Status Messages**: Progress updates shown during processing
- **Error Handling**: Individual agent results shown if any errors occur
- **Timing Statistics**: Detailed timing information for each agent and total workflow duration

The final output includes:

- Original startup content from the writer
- Two-sentence summary
- Structured rating (e.g., "8/10 - Comprehensive and practical advice...")
- Compelling title
- Timing statistics for performance analysis

## Error Handling

The application handles various error scenarios:

- Missing OpenAI API key
- API call failures
- Network timeouts
- Agent communication errors

All errors are displayed as clear text messages with specific error details.
