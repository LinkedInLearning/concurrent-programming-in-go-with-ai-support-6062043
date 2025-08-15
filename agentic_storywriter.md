# Agentic Storywriter

An AI-powered story writing application using multiple specialized agents running concurrently in Go.

## Overview

The Agentic Storywriter is a sophisticated multi-agent system that collaboratively creates complete stories from simple user prompts. It follows a supervisor-agent architecture where a central supervisor coordinates multiple specialized AI agents, each responsible for different aspects of story creation.

## Architecture

### Supervisor-Agent Pattern

- **StorySupervisor**: Central coordinator managing all sub-agents
- **Persistent Goroutines**: Each agent runs as a background goroutine
- **Channel Communication**: Agents communicate via Go channels
- **Concurrent Processing**: Independent operations run in parallel for efficiency

### System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           AGENTIC STORYWRITER SYSTEM                            │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─────────────────┐     User Input/Display    ┌─────────────────────────────────────┐
│                 │◄─────────────────────────►│          MAIN UI (TUI)              │
│     USER        │                           │    • BubbleTea Interface            │
│                 │                           │    • Live Progress Display          │
│                 │                           │    • Spinner Animations             │
└─────────────────┘                           │    • Auto-Resume Detection          │
                                              └─────────────────┬───────────────────┘
                                                                │
                                                                ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                            STORY WORKFLOW                                       │
│  ┌──────────────────────────────────────────────────────────────────────────────┤
│  │ PHASE 1: FOUNDATION BUILDING (Sequential)                                    │
│  │                                                                              │
│  │  [Plot Designer] ──► [Worldbuilder] ──► [Plot Expander] ──► [Char Dev]       │
│  │       │                   │                   │                │             │
│  │       ▼                   ▼                   ▼                ▼             │
│  │  plot_design.md    world_building.md   plot_expansion.md  characters.md      │
│  └──────────────────────────────────────────────────────────────────────────────┤
│  │ PHASE 2: CHAPTER WRITING + PARALLEL SUMMARIZATION                            │
│  │                                                                              │
│  │  [Author Ch1] ──► [Author Ch2] ──► ... ──► [Author Ch9]                      │
│  │       │               │                        │                             │
│  │       ▼               ▼                        ▼                             │
│  │  chapter_1.md    chapter_2.md    ...    chapter_9.md                         │
│  │       │               │                        │                             │
│  │       ▼               ▼                        ▼                             │
│  │  ┌─────────────────────────────────────────────────────────────────────┐     │
│  │  │              PARALLEL SUMMARIZATION                                 │     │
│  │  │                                                                     │     │
│  │  │  [Summarizer] ◄──┬──► [Summarizer] ◄──┬──► ... ◄──► [Summarizer]    │     │
│  │  │       │          │         │          │                   │         │     │
│  │  │       ▼          │         ▼          │                   ▼         │     │
│  │  │ ch_1_summary.md  │   ch_2_summary.md  │             ch_9_summary.md │     │
│  │  └─────────────────────────────────────────────────────────────────────┘     │
│  │                     │                    │                                   │
│  │                     ▼                    ▼                                   │
│  │              ┌─────────────────────────────────────┐                         │
│  │              │    WAIT FOR ALL SUMMARIES           │                         │
│  │              └─────────────────────────────────────┘                         │
│  └──────────────────────────────────────────────────────────────────────────────┤
│  │ PHASE 3: EDITING WITH COMPLETE CONTEXT                                       │
│  │                                                                              │
│  │  ┌─────────────────────────────────────────────────────────────────────┐     │
│  │  │                    ALL CHAPTER SUMMARIES                            │     │
│  │  │  Ch1: summary │ Ch2: summary │ ... │ Ch9: summary                   │     │
│  │  └─────────────────────────────────────────────────────────────────────┘     │
│  │                                    │                                         │
│  │                                    ▼                                         │
│  │  [Editor Ch1] ──► [Editor Ch2] ──► ... ──► [Editor Ch9]                      │
│  │       │               │                        │                             │
│  │       ▼               ▼                        ▼                             │
│  │ ch_1_edited.md  ch_2_edited.md  ...    ch_9_edited.md                        │
│  └──────────────────────────────────────────────────────────────────────────────┤
│  │ FINAL: STORY COMPLETION                                                      │
│  │                                                                              │
│  │  [Concatenate All] ──► full_story_complete.md                                │
│  │  [Story Summary]   ──► story_summary.md                                      │
│  └──────────────────────────────────────────────────────────────────────────────┘
└─────────────────────────────────────────────────────────────────────────────────┘
                                              │
                                              ▼
┌──────────────────────────────────────────────────────────────────────────────────┐
│                         INFRASTRUCTURE LAYER                                     │
├──────────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐    ┌──────────────────┐    ┌─────────────────────────────┐  │
│  │ STORY SUPERVISOR│    │ SESSION MEMORY   │    │    WORKSPACE MANAGER        │  │
│  │                 │    │                  │    │                             │  │
│  │ • Agent Factory │    │ • Token Counting │    │ • Secure File Operations    │  │
│  │ • Agent Registry│    │ • Auto Compaction│    │ • .md Files Only            │  │
│  │ • Resource Mgmt │    │ • Memory Pressure│    │ • Path Validation           │  │
│  │                 │    │   Detection      │    │ • Resume Functionality      │  │
│  └─────────────────┘    └──────────────────┘    └─────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────────────────┘
                                              │
                                              ▼
┌──────────────────────────────────────────────────────────────────────────────────┐
│                            AI AGENTS LAYER                                       │
├──────────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────────────────────┐ │
│  │Plot Designer│ │Worldbuilder │ │Plot Expander│ │    Character Developer      │ │
│  │             │ │             │ │             │ │                             │ │
│  │• 9-pt Hero's│ │• Fantasy/   │ │• Expand     │ │• Protagonist & Antagonist   │ │
│  │  Journey    │ │  Sci-Fi/    │ │  Plot Points│ │• Supporting Characters      │ │
│  │• Story Arc  │ │  Historical │ │• Fix Holes  │ │• Backstories & Lore         │ │
│  │• Foundation │ │• Magic Rules│ │• Avoid      │ │• Physical Descriptions      │ │
│  │             │ │• World Logic│ │  Clichés    │ │• Names Fitting World        │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────────────────────┘ │
│                                                                                  │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐                                 │
│  │   Author    │ │Story Summary│ │   Editor    │                                 │
│  │             │ │             │ │             │                                 │
│  │• Write      │ │• Chapter    │ │• Review     │                                 │
│  │  Chapters   │ │  Summaries  │ │  Coherence  │                                 │
│  │• 2 Pages    │ │• Single     │ │• Fix Plot   │                                 │
│  │  Each       │ │  Paragraphs │ │  Holes      │                                 │
│  │• Narrative  │ │• Memory     │ │• Consistency│                                 │
│  │  Flow       │ │  Management │ │  Check      │                                 │
│  └─────────────┘ └─────────────┘ └─────────────┘                                 │
└──────────────────────────────────────────────────────────────────────────────────┘
                                              │
                                              ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           EXTERNAL SERVICES                                     │
├─────────────────────────────────────────────────────────────────────────────────┤
│  ┌────────────────────────────────────────────────────────────────────────────┐ │
│  │                          OPENAI API                                        │ │
│  │                                                                            │ │
│  │  • GPT-4o Model Calls                                                      │ │
│  │  • Specialized Agent Prompts                                               │ │
│  │  • Context Management                                                      │ │
│  │  • Token Usage Monitoring                                                  │ │
│  └────────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────────┐
│                              DATA FLOW                                          │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  User Prompt ──► Plot Design ──► World Building ──► Plot Expansion ──►          │
│                                                                                 │
│  Character Development ──► Chapter Writing (1-9) ──► Parallel Summarization ──► │
│                                                                                 │
│  Wait for All Summaries ──► Sequential Editing (with all summaries) ──►         │
│                                                                                 │
│  Story Completion ──► File Concatenation ──►  COMPLETE STORY                    │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘

CONCURRENCY PATTERNS:
• Sequential: Foundation building (plot → world → expansion → characters)
• Parallel: Chapter summarization (as soon as each chapter is written)
• Sequential: Chapter editing (after ALL summaries complete)
• Goroutines: Each agent runs in persistent background goroutine
• Channels: Progress updates and agent communication
• Mutex: Thread-safe access to shared resources
```

### Memory Management

- **Token Counting**: Uses tiktoken with GPT-4o counter to monitor context usage
- **Automatic Compaction**: Compacts session history when reaching 70% of context threshold
- **Session Persistence**: Maintains conversation history throughout the workflow

### Secure File Operations

- **Workspace Isolation**: All file operations restricted to `workspace/` folder
- **Markdown Only**: Agents can only create/edit `.md` files
- **Path Validation**: Prevents directory traversal attacks

## Agents

### 1. Plot Designer

**Role**: Creates the foundational story structure
**Output**: 9-point hero's journey plot structure
**Details**:

- Establishes background scene and worldbuilding rules
- Defines protagonist and reader engagement
- Maps the complete hero's journey arc
- Ensures each plot point is thoroughly detailed

### 2. Worldbuilder

**Role**: Develops the story's setting and universe
**Output**: Comprehensive world description
**Details**:

- Determines genre (fantasy, sci-fi, historical, contemporary)
- Establishes magic systems, technology levels, or historical context
- Creates unique world rules and constraints
- Enhances plot potential through creative world design

### 3. Plot Expander

**Role**: Transforms plot points into detailed narratives
**Output**: Expanded plot with full paragraphs for each point
**Details**:

- Takes initial plot structure and world information
- Expands each plot point from sentences to paragraphs
- Identifies and resolves potential plot holes
- Avoids common clichés like deus ex machina

### 4. Character Developer

**Role**: Creates compelling characters for the story
**Output**: Detailed character profiles
**Details**:

- Develops protagonist, antagonist, and supporting characters
- Creates relevant backstories and character lore
- Provides physical descriptions and personality traits
- Ensures character names fit the established world

### 5. Author

**Role**: Writes the actual story content
**Output**: Story chapters (2 pages each, 9 chapters total)
**Details**:

- Transforms plot, world, and character information into narrative
- Writes engaging, readable chapters
- Maintains consistency with established elements
- Creates compelling dialogue and descriptions

### 6. Story Summarizer

**Role**: Creates concise chapter summaries
**Output**: Single paragraph summaries per chapter
**Details**:

- Distills each chapter to essential plot points
- Enables memory management for the supervisor
- Maintains story coherence across chapters
- Provides quick reference for editing process

### 7. Editor

**Role**: Reviews and refines story content
**Output**: Edited chapters with improved coherence
**Details**:

- Reviews chapters against all previous summaries
- Identifies and resolves plot inconsistencies
- Eliminates contradictions between chapters
- Ensures story flows logically from beginning to end

### 8. Supervisor Summary

**Role**: Manages memory pressure and session compaction
**Output**: Compressed session summaries
**Details**:

- Monitors token usage across the entire workflow
- Compacts session history when memory threshold is reached
- Preserves essential information while reducing context size
- Maintains workflow continuity during long story creation sessions

## Usage

### Prerequisites

- Go 1.24 or later
- OpenAI API key

### Setup

1. Set your OpenAI API key:

   ```bash
   export OPENAI_API_KEY=your_key_here
   ```

   Or create a `.env` file:

   ```
   OPENAI_API_KEY=your_key_here
   ```

### Running the Application

```bash
# Run directly
go run .

# Or build and run
go build -o storywriter .
./storywriter

# View help
go run . --help
```

### Using the Interface

1. Launch the application
2. Enter your story prompt (a few sentences describing the story you want)
3. Press Enter to begin the story creation process
4. Wait for the agents to complete their work (this may take several minutes)
5. Check the `workspace/` folder for all generated files

## Output Files

All generated content is saved to the `workspace/` directory as markdown files:

- **`plot_design.md`** - Initial 9-point story structure
- **`world_building.md`** - World setting, rules, and background
- **`plot_expansion.md`** - Detailed plot with expanded paragraphs
- **`characters.md`** - Character profiles, backstories, and descriptions
- **`chapter_1.md`** through **`chapter_9.md`** - Original story chapters
- **`chapter_1_edited.md`** through **`chapter_9_edited.md`** - Editor-reviewed chapters
- **`story_summary.md`** - Complete story overview with chapter summaries

## Technical Features

### Concurrency

- Agents run as persistent background goroutines
- Independent operations are parallelized where possible
- Channel-based communication ensures thread safety
- Timeout handling prevents hanging operations

### Error Handling

- Comprehensive error handling throughout the workflow
- Graceful degradation when individual agents fail
- User-friendly error messages in the TUI
- Automatic retry mechanisms for transient failures

### User Interface

- Beautiful terminal user interface built with Bubble Tea
- Real-time status updates during story creation
- Progress indicators showing which agents are active
- Responsive design that works in various terminal sizes

### Memory Efficiency

- Intelligent token counting prevents context overflow
- Automatic session compaction preserves essential information
- Efficient string handling and memory management
- Optimized for long-running story creation sessions

## Example Workflow

1. **User Input**: "A young wizard discovers they have the power to control time, but each use ages them rapidly."

2. **Plot Designer**: Creates 9-point structure around time magic and aging consequences

3. **Worldbuilder**: Develops magical world with time magic rules and societal implications

4. **Plot Expander**: Expands each plot point with detailed scenarios and character development

5. **Character Developer**: Creates the young wizard protagonist, mentors, antagonists, and supporting characters

6. **Author**: Writes 9 chapters telling the complete story

7. **Story Summarizer**: Creates summaries of each chapter for coherence checking

8. **Editor**: Reviews and refines each chapter for consistency and flow

9. **Output**: Complete story with all supporting materials saved to workspace

## Customization

The system is designed to be easily extensible:

- **Agent Prompts**: Modify agent prompts in `agent/supervisor.go`
- **Chapter Count**: Adjust the number of chapters in `agent/story_workflow.go`
- **Memory Thresholds**: Modify token limits in `agent/memory.go`
- **File Operations**: Extend workspace functionality in `agent/workspace.go`

## Performance

- **Parallel Processing**: Independent agents run concurrently
- **Efficient Memory Usage**: Automatic compaction prevents memory bloat
- **Optimized API Calls**: Batched operations where possible
- **Responsive UI**: Non-blocking interface during long operations

The Agentic Storywriter demonstrates advanced concurrent programming patterns in Go while creating engaging, coherent stories through collaborative AI agents.
