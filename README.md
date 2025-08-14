# Advice Rating Tool - Concurrent Programming in Go with AI Support

This is an agentic REPL tool that rates user-submitted advice using multiple expert AI agents running concurrently. The tool demonstrates concurrent programming patterns in Go while providing practical advice evaluation.

![Concurrent Programming in Go with AI Support][lil-thumbnail-url]

## Features

- **Interactive REPL Interface**: Enter advice and get real-time expert analysis
- **Multiple Expert Agents**: Six specialized agents evaluate advice from different perspectives:
  - **Career Agent**: Rates advice for career impact (0-10 or -1 if not applicable)
  - **BestFriend Agent**: Rates advice for interpersonal relationships
  - **Financial Agent**: Rates advice for financial success
  - **TechSupport Agent**: Rates advice for technology accuracy
  - **Dietician Agent**: Rates advice for health and diet
  - **Lawyer Agent**: Rates advice for legal accuracy
- **Advice Summarizer**: Provides final rating (terrible/bad/neutral/good/fantastic)
- **Concurrent Processing**: All expert agents run in parallel using Go goroutines
- **Structured Output**: Uses OpenAI's structured output with JSON schemas

## Architecture

The application uses a fan-out/fan-in pattern:
1. User input is distributed to all expert agents simultaneously
2. Each agent processes the advice concurrently
3. Results are collected and passed to the summarizer agent
4. Final rating and summary are presented to the user

## Quick Start

1. **Set up your OpenAI API key**:
   ```bash
   export OPENAI_API_KEY="your-api-key-here"
   # OR create a .env file with: OPENAI_API_KEY=your-api-key-here
   ```

2. **Run the application**:
   ```bash
   go run main.go
   ```

3. **Start rating advice**:
   ```
   Enter advice: Always invest in index funds for long-term growth
   ```

4. **View expert analysis** and final rating

5. **Type 'quit' to exit**

## Example Usage

```
🤖 Advice Rating Tool
Enter advice to get it rated by expert agents, or 'quit' to exit.

Enter advice: Drink 8 glasses of water daily
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

Career: -1 - This advice is not related to career development
BestFriend: -1 - This advice doesn't apply to interpersonal relationships
Financial: -1 - This advice is not related to financial matters
TechSupport: -1 - This advice is not technology related
Dietician: 8 - Good general hydration advice, though individual needs vary
Lawyer: -1 - This advice is not related to legal matters

=== FINAL ASSESSMENT ===
Final Rating: GOOD

Summary: Only the dietician provided a rating (8/10) as this advice specifically relates to health and hydration. The advice is generally sound for maintaining proper hydration, though individual water needs can vary based on activity level, climate, and health conditions.

Analysis completed in 3.2s
```

## Instructions

This repository has branches for each of the videos in the course. You can use the branch pop up menu in github to switch to a specific branch and take a look at the course at that stage, or you can add `/tree/BRANCH_NAME` to the URL to go to the branch you want to access.

## Branches

The branches are structured to correspond to the videos in the course. The naming convention is `CHAPTER#_MOVIE#`. As an example, the branch named `02_03` corresponds to the second chapter and the third video in that chapter.
Some branches will have a beginning and an end state. These are marked with the letters `b` for "beginning" and `e` for "end". The `b` branch contains the code as it is at the beginning of the movie. The `e` branch contains the code as it is at the end of the movie. The `main` branch holds the final state of the code when in the course.

When switching from one exercise files branch to the next after making changes to the files, you may get a message like this:

    error: Your local changes to the following files would be overwritten by checkout:        [files]
    Please commit your changes or stash them before you switch branches.
    Aborting

To resolve this issue:
Add changes to git using this command: git add .
Commit changes using this command: git commit -m "some message"

## Installing

1. To use these exercise files, you must have the following installed, and available in your `$PATH`:
   - [golang](https://go.dev/doc/install) (version 1.24)
   - goimports
   - gopls
   - ripgrep (`rg` command)
   - fd-find (`fd` command)
   - [crush](https://charm.land/crush)
2. Clone this repository into your local machine using the terminal (Mac), CMD (Windows), or a GUI tool like SourceTree.
3. Launch `crush` and configure your AI provider API key and model.

[0]: # "Replace these placeholder URLs with actual course URLs"
[lil-course-url]: https://www.linkedin.com/learning/
[lil-thumbnail-url]: https://media.licdn.com/dms/image/v2/D4E0DAQG0eDHsyOSqTA/learning-public-crop_675_1200/B4EZVdqqdwHUAY-/0/1741033220778?e=2147483647&v=beta&t=FxUDo6FA8W8CiFROwqfZKL_mzQhYx9loYLfjN-LNjgA
