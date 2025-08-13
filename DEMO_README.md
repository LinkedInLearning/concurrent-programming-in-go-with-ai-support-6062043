# Concurrent LLM Demo

A Go demonstration program showcasing concurrent programming by sending multiple prompts to different OpenAI models simultaneously and displaying the results in a beautiful table format.

## Features

- **Concurrent API Calls**: Uses goroutines to send requests to multiple OpenAI models simultaneously
- **Multiple Models**: Tests GPT-3.5-turbo, GPT-4o-mini, and GPT-4 models
- **Beautiful Tables**: Uses the `go-pretty` library to display results in colorful, formatted tables
- **Environment Variables**: Automatically loads API keys from `.env` files using `godotenv`
- **Performance Metrics**: Shows response times and success rates for each model
- **Error Handling**: Gracefully handles API errors and timeouts

## Prerequisites

- Go 1.19 or later
- OpenAI API key

## Installation

1. Clone this repository
2. Install dependencies:
   ```bash
   go mod tidy
   ```
3. Create a `.env` file with your OpenAI API key:
   ```
   OPENAI_API_KEY=your_openai_api_key_here
   ```

## Usage

Run the demo:
```bash
go run main.go
```

The program will:
1. Load environment variables from the `.env` file
2. Create concurrent goroutines for each model-prompt combination
3. Send requests to the OpenAI API simultaneously
4. Display real-time progress as requests complete
5. Show a detailed results table with responses, durations, and status
6. Provide summary statistics for each model

## Example Output

```
🚀 Starting concurrent LLM demo...
Testing 3 models with 4 prompts each (12 total requests)

✅ Completed gpt-3.5-turbo request in 941ms
✅ Completed gpt-4o-mini request in 1.267s
✅ Completed gpt-4 request in 1.512s
...

🎉 All requests completed in 3.899s

[Colorful table with results]

📊 Summary Statistics:
[Summary table with performance metrics]
```

## Architecture

The program demonstrates several Go concurrency patterns:

- **Goroutines**: Each API call runs in its own goroutine
- **Channels**: Used to collect results from concurrent operations
- **WaitGroups**: Ensures all goroutines complete before processing results
- **Context with Timeout**: Prevents hanging requests

## Dependencies

- [`github.com/openai/openai-go`](https://github.com/openai/openai-go) - Official OpenAI Go client
- [`github.com/jedib0t/go-pretty/v6/table`](https://github.com/jedib0t/go-pretty) - Beautiful table formatting
- [`github.com/joho/godotenv`](https://github.com/joho/godotenv) - Environment variable loading

## Configuration

The program tests these prompts by default:
- "Write a haiku about programming"
- "Explain quantum computing in one sentence"
- "What is the meaning of life?"
- "Tell me a programming joke"

And these models:
- `gpt-3.5-turbo`
- `gpt-4o-mini`
- `gpt-4`

You can modify these in the `main.go` file to test different prompts or models.

## License

This project is open source and available under the [MIT License](LICENSE).