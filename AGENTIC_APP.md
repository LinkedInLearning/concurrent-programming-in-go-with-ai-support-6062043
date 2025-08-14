# Advice Rating Tool - Agentic Application

A concurrent Go application that demonstrates an agentic workflow using the OpenAI API. The application is an interactive REPL tool that rates user-submitted advice using multiple expert AI agents running concurrently.

## Architecture

The application uses a fan-out/fan-in pattern with six expert agents and one summarizer:

```
                    ┌─────────────────┐
                    │   User Input    │
                    │   (Advice to    │
                    │   be rated)     │
                    └─────────┬───────┘
                              │
                              ▼
              ┌───────────────┼───────────────┐
              │               │               │
              ▼               ▼               ▼
    ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
    │   Career    │ │ BestFriend  │ │ Financial   │
    │   Agent     │ │   Agent     │ │   Agent     │
    │ (0-10 or    │ │ (0-10 or    │ │ (0-10 or    │
    │   -1)       │ │   -1)       │ │   -1)       │
    └─────────────┘ └─────────────┘ └─────────────┘
              │               │               │
              └───────────────┼───────────────┘
                              │
              ┌───────────────┼───────────────┐
              │               │               │
              ▼               ▼               ▼
    ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
    │TechSupport  │ │ Dietician   │ │   Lawyer    │
    │   Agent     │ │   Agent     │ │   Agent     │
    │ (0-10 or    │ │ (0-10 or    │ │ (0-10 or    │
    │   -1)       │ │   -1)       │ │   -1)       │
    └─────────────┘ └─────────────┘ └─────────────┘
              │               │               │
              └───────────────┼───────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │ Advice Summary  │
                    │     Agent       │
                    │ (terrible/bad/  │
                    │neutral/good/    │
                    │  fantastic)     │
                    └─────────┬───────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │   Final Rating  │
                    │   & Summary     │
                    └─────────────────┘
```

**Flow Description:**

1. **User Input**: User enters advice through the REPL interface
2. **Fan-Out Phase**: Advice is sent to six expert agents concurrently:
   - **Career Agent** - Rates advice for career impact
   - **BestFriend Agent** - Rates advice for interpersonal relationships  
   - **Financial Agent** - Rates advice for financial success
   - **TechSupport Agent** - Rates advice for technology accuracy
   - **Dietician Agent** - Rates advice for health and diet
   - **Lawyer Agent** - Rates advice for legal accuracy
3. **Fan-In Phase**: All expert ratings are collected
4. **Summarization**: Advice Summarizer Agent processes all ratings and provides final assessment
5. **Output**: Final rating (terrible/bad/neutral/good/fantastic) and summary displayed

## Expert Agents

Each expert agent provides:
- **Rating**: 0-10 scale for applicable advice, or -1 if not applicable to their domain
- **Explanation**: Brief reasoning for the rating
- **Concurrent Processing**: All agents run simultaneously using Go goroutines

### Agent Specializations

1. **Career Agent**: "You are an expert career counselor. Your client has been given some advice, and you are tasked with analyzing the advice to provide a rating from 0-10 on how good the advice would be for their career. If the advice isn't career applicable, please return -1."

2. **BestFriend Agent**: "You are a really good friend, and you know all sorts of interpersonal information. Your best friend has been given some advice, which they will relay to you. You are tasked with thinking about the advice to provide a rating from 0-10 on how good the advice would be for their interpersonal life. If the advice seems to not apply to personal relationships or their personal life, please return -1."

3. **Financial Agent**: "You are an expert financial advisor. Your client has been given some advice, and you are tasked with analyzing the advice to provide a rating from 0-10 on how good the advice would be for their financial success. If the advice isn't finance applicable, please return -1."

4. **TechSupport Agent**: "You are the best tech support engineer at a large fortune 100 company, an expert in all things computers. Your coworkers came across some tips online, and they want to ask you if they are good advice. You are tasked with analyzing the advice to provide a rating from 0-10 on how accurate the advice is. If the advice isn't technology or IT applicable, please return -1."

5. **Dietician Agent**: "You are an expert dietician, renowned the world over. Your client is coming to you to ask about some of the advice they were given. You are tasked with analyzing the advice to provide a rating from 0-10 on how accurate the advice is. If the advice isn't applicable to health and dieting, please return -1."

6. **Lawyer Agent**: "You are a legal scholar, world famous for your expertise in the law. Your client is coming to you to ask about some of the advice they were given. You are tasked with analyzing the advice to provide a rating from 0-10 on how accurate the advice is. If the advice isn't applicable to matter of the law, please return -1."

## Features

- **Interactive REPL Interface**: Continuous advice rating session
- **Concurrent Processing**: All expert agents run in parallel using Go channels and goroutines
- **Structured JSON Output**: All agents use OpenAI's structured output with JSON schema validation
- **Domain-Specific Expertise**: Six specialized agents cover different life domains
- **Intelligent Summarization**: Final agent averages applicable ratings and provides qualitative assessment
- **Real-time Status Updates**: Shows progress as agents process advice
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

The application will start an interactive REPL where you can:

1. Enter advice to be rated
2. View real-time status updates as agents process
3. See individual expert ratings
4. Get a final qualitative assessment
5. Type 'quit' to exit

### Example Session

```
🤖 Advice Rating Tool
Enter advice to get it rated by expert agents, or 'quit' to exit.

Enter advice: Always invest in index funds for long-term growth
Status: Analyzing advice with expert agents...
Status: Getting Career opinion...
Status: Getting BestFriend opinion...
Status: Getting Financial opinion...
Status: Getting TechSupport opinion...
Status: Getting Dietician opinion...
Status: Getting Lawyer opinion...
Status: Summarizing expert opinions...
Status: Analysis complete!

=== EXPERT RATINGS ===

Career: -1 - This advice is not directly related to career development
BestFriend: -1 - This advice doesn't apply to interpersonal relationships
Financial: 9 - Excellent long-term investment strategy with low fees and diversification
TechSupport: -1 - This advice is not technology related
Dietician: -1 - This advice is not related to health or diet
Lawyer: -1 - This advice is not related to legal matters

=== FINAL ASSESSMENT ===
Final Rating: FANTASTIC

Summary: Only the financial expert provided a rating (9/10) as this advice specifically relates to investment strategy. The advice is excellent - index funds are widely recommended by financial experts for long-term wealth building due to their low costs, broad diversification, and historical performance.

Analysis completed in 2.8s
```

## Project Structure

```
.
├── main.go                    # REPL interface
├── agent/
│   ├── agent.go              # Agent interface and configuration
│   ├── workflow.go           # Advice workflow orchestration
│   ├── career_agent.go       # Career advice expert
│   ├── bestfriend_agent.go   # Interpersonal advice expert
│   ├── financial_agent.go    # Financial advice expert
│   ├── techsupport_agent.go  # Technology advice expert
│   ├── dietician_agent.go    # Health/diet advice expert
│   ├── lawyer_agent.go       # Legal advice expert
│   └── advice_summarizer.go  # Final rating summarizer
├── go.mod
├── go.sum
└── README.md
```

## Rating Scale

The final assessment uses this scale:
- **0-2 average**: terrible
- **3-4 average**: bad  
- **5 average**: neutral
- **6-7 average**: good
- **8-10 average**: fantastic

Only ratings from applicable experts (not -1) are included in the average.

## Agent Interface

All agents implement the `Agent` interface:

```go
type Agent interface {
    Start(ctx context.Context, input <-chan string, output chan<- string) error
}
```

Each agent uses structured JSON output with specific response schemas for consistent, validated responses.

## Dependencies

- `github.com/openai/openai-go` - Official OpenAI Go client with structured output support
- `github.com/joho/godotenv` - Environment variable loading from .env files
- `github.com/invopop/jsonschema` - JSON schema generation for structured outputs

## Error Handling

The application handles various error scenarios:

- Missing OpenAI API key
- API call failures
- Network timeouts
- Agent communication errors
- Invalid JSON responses

All errors are displayed as clear text messages with specific error details, and the application continues running for additional advice rating.