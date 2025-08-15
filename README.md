# Concurrent Programming in Go with AI Support

This is the repository for the LinkedIn Learning course `Concurrent Programming in Go with AI Support`. The full course is available from [LinkedIn Learning][lil-course-url].

![Concurrent Programming in Go with AI Support][lil-thumbnail-url]

_See the readme file in the main branch for updated instructions and information._

## Applications

This repository contains two AI-powered applications demonstrating concurrent programming patterns:

### 1. Original Demo Application
**Location**: `main.go`
**Description**: Demonstrates basic agentic workflow with startup company content generation.

**Agents**:
- Writer - Generates content about startup companies
- Summarizer - Creates concise summaries  
- Rater - Provides structured ratings (1-10)
- Titler - Generates compelling titles
- MarkdownFormatter - Formats results as markdown

**Usage**:
```bash
go run main.go
```

### 2. Agentic Storywriter
**Location**: `cmd/storywriter/main.go`
**Description**: Advanced multi-agent system for collaborative story writing following a supervisor-agent architecture.

**Agents**:
- **Plot Designer** - Creates 9-point story structure following the hero's journey
- **Worldbuilder** - Develops story world, setting, and rules (fantasy, sci-fi, historical, etc.)
- **Plot Expander** - Expands plot points into detailed paragraphs, checks for plot holes
- **Character Developer** - Creates protagonist, villain, and supporting characters with backstories
- **Author** - Writes story chapters (2 pages each)
- **Story Summarizer** - Creates chapter summaries for memory management
- **Editor** - Reviews and edits chapters for coherence across the story
- **Supervisor Summary** - Manages memory pressure and session compaction

**Features**:
- **Concurrent Processing**: Agents run as persistent background goroutines
- **Memory Management**: Uses tiktoken to monitor context usage and compact sessions at 70% threshold
- **Secure File Operations**: Agents can only create/edit .md files within the workspace folder
- **Parallel Execution**: Independent operations are parallelized for efficiency
- **Persistent Storage**: All outputs saved to workspace folder for review

**Usage**:
```bash
go run ./cmd/storywriter
```

**Output**: All generated content is saved to the `workspace/` folder as markdown files:
- `plot_design.md` - Initial story structure
- `world_building.md` - World and setting details  
- `plot_expansion.md` - Expanded plot details
- `characters.md` - Character profiles and backstories
- `chapter_N.md` - Original chapters
- `chapter_N_edited.md` - Edited chapters
- `story_summary.md` - Complete story summary

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
3. Set your OpenAI API key: `export OPENAI_API_KEY=your_key_here` or create a `.env` file
4. Launch `crush` and configure your AI provider API key and model.

## Environment Setup

Both applications require an OpenAI API key. Set it using one of these methods:

1. **Environment Variable**:
   ```bash
   export OPENAI_API_KEY=your_key_here
   ```

2. **Create .env file**:
   ```
   OPENAI_API_KEY=your_key_here
   ```

[0]: # "Replace these placeholder URLs with actual course URLs"
[lil-course-url]: https://www.linkedin.com/learning/
[lil-thumbnail-url]: https://media.licdn.com/dms/image/v2/D4E0DAQG0eDHsyOSqTA/learning-public-crop_675_1200/B4EZVdqqdwHUAY-/0/1741033220778?e=2147483647&v=beta&t=FxUDo6FA8W8CiFROwqfZKL_mzQhYx9loYLfjN-LNjgA
