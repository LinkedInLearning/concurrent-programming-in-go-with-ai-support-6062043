package agent

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// ProgressUpdate represents a status update from an agent
type ProgressUpdate struct {
	AgentName   string
	Status      string // "started", "completed", "error"
	Message     string
	Error       error
	TokenUsage  *TokenUsage // NEW: Token usage for this agent
	TotalCost   float64     // NEW: Total cost across all agents
	TotalTokens int         // NEW: Total tokens across all agents
}

// StoryWorkflow manages the complete story writing process
type StoryWorkflow struct {
	supervisor     *StorySupervisor
	memory         *SessionMemory
	tokenTracker   *SystemTokenTracker // NEW: Token tracking
	initialPrompt  string
	completedTasks []string
	progressChan   chan ProgressUpdate
	mu             sync.RWMutex
}

// NewStoryWorkflow creates a new story workflow
func NewStoryWorkflow(apiKey string) (*StoryWorkflow, error) {
	supervisor, err := NewStorySupervisor(apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create supervisor: %w", err)
	}

	memory, err := NewSessionMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to create session memory: %w", err)
	}

	// Initialize token tracker
	tokenTracker := NewSystemTokenTracker()

	supervisor.InitializeAgents()

	// Try to load existing token data
	tokenTracker.LoadFromFile(supervisor.workspace)

	return &StoryWorkflow{
		supervisor:     supervisor,
		memory:         memory,
		tokenTracker:   tokenTracker,
		completedTasks: make([]string, 0),
		progressChan:   make(chan ProgressUpdate, 100),
	}, nil
}

// GetProgressChan returns the progress channel for UI updates
func (sw *StoryWorkflow) GetProgressChan() <-chan ProgressUpdate {
	return sw.progressChan
}

// GetTotalCost returns the current total cost from the token tracker
func (sw *StoryWorkflow) GetTotalCost() float64 {
	return sw.tokenTracker.GetTotalCost()
}

// GetTotalTokens returns the current total tokens from the token tracker
func (sw *StoryWorkflow) GetTotalTokens() int {
	return sw.tokenTracker.GetTotalTokens()
}

// GetAgentUsage returns the token usage for a specific agent
func (sw *StoryWorkflow) GetAgentUsage(agentName string) *TokenUsage {
	return sw.tokenTracker.GetAgentUsage(agentName)
}

// GetTokenTracker returns the token tracker (for testing)
func (sw *StoryWorkflow) GetTokenTracker() *SystemTokenTracker {
	return sw.tokenTracker
}

// GetSupervisor returns the supervisor (for testing)
func (sw *StoryWorkflow) GetSupervisor() *StorySupervisor {
	return sw.supervisor
}

// sendProgress sends a progress update
func (sw *StoryWorkflow) sendProgress(agentName, status, message string, err error) {
	// Get current token usage
	totalCost := sw.tokenTracker.GetTotalCost()
	totalTokens := sw.tokenTracker.GetTotalTokens()
	agentUsage := sw.tokenTracker.GetAgentUsage(agentName)

	select {
	case sw.progressChan <- ProgressUpdate{
		AgentName:   agentName,
		Status:      status,
		Message:     message,
		Error:       err,
		TokenUsage:  agentUsage,
		TotalCost:   totalCost,
		TotalTokens: totalTokens,
	}:
	default:
		// Channel full, skip update
	}
}

// ExecuteStoryCreation runs the complete story creation workflow
func (sw *StoryWorkflow) ExecuteStoryCreation(ctx context.Context, userPrompt string) error {
	sw.initialPrompt = userPrompt
	sw.memory.AddEntry(fmt.Sprintf("User Request: %s", userPrompt))

	// Check for existing progress
	resumeFrom, existingData := sw.checkWorkspaceProgress()

	// Load original prompt if resuming, otherwise use provided prompt
	originalPrompt := userPrompt
	if resumeFrom != "start" {
		if data, exists := existingData["story_prompt"]; exists {
			originalPrompt = data
		}
	}

	// Variables to hold story components
	var plotDesign, worldBuilding, plotExpansion, characters string
	var err error

	if resumeFrom != "start" && len(existingData) > 0 {
		sw.sendProgress("System", "started", "Resuming from existing progress...", nil)
		
		// Send progress updates for all completed agents to show their token usage
		if _, exists := existingData["plot_design"]; exists {
			sw.sendProgress("Plot Designer", "completed", "Plot design completed", nil)
		}
		if _, exists := existingData["world_building"]; exists {
			sw.sendProgress("Worldbuilder", "completed", "World building completed", nil)
		}
		if _, exists := existingData["plot_expansion"]; exists {
			sw.sendProgress("Plot Expander", "completed", "Plot expansion completed", nil)
		}
		if _, exists := existingData["characters"]; exists {
			sw.sendProgress("Character Developer", "completed", "Character development completed", nil)
		}
		
		// Send progress for completed chapters
		for i := 1; i <= 9; i++ {
			if _, exists := existingData[fmt.Sprintf("chapter_%d", i)]; exists {
				sw.sendProgress(fmt.Sprintf("Author (Ch %d)", i), "completed", fmt.Sprintf("Chapter %d completed", i), nil)
			}
			if _, exists := existingData[fmt.Sprintf("chapter_%d_summary", i)]; exists {
				sw.sendProgress("Story Summarizer", "completed", fmt.Sprintf("Chapter %d summarized", i), nil)
			}
			if _, exists := existingData[fmt.Sprintf("chapter_%d_edited", i)]; exists {
				sw.sendProgress("Editor", "completed", fmt.Sprintf("Chapter %d edited", i), nil)
			}
		}
	}

	// Save the initial story prompt (if not resuming)
	if resumeFrom == "start" {
		if err := sw.saveToWorkspace("story_prompt.md", fmt.Sprintf("# Story Prompt\n\n%s", originalPrompt)); err != nil {
			return fmt.Errorf("failed to save story prompt: %w", err)
		}
	}

	// Step 1: Plot Design
	if resumeFrom == "start" || resumeFrom == "plot_design" {
		sw.sendProgress("Plot Designer", "started", "Creating story structure...", nil)
		plotDesign, err = sw.executeAgentTask(ctx, "plot_designer", originalPrompt)
		if err != nil {
			sw.sendProgress("Plot Designer", "error", "Failed to create plot", err)
			return fmt.Errorf("plot design failed: %w", err)
		}
		sw.sendProgress("Plot Designer", "completed", "Story structure created", nil)
		sw.addCompletedTask("Plot Design")

		// Save plot design immediately
		if err := sw.saveToWorkspace("plot_design.md", fmt.Sprintf("# Plot Design\n\n%s", plotDesign)); err != nil {
			return fmt.Errorf("failed to save plot design: %w", err)
		}
	} else {
		// Load existing plot design
		if data, exists := existingData["plot_design"]; exists {
			plotDesign = data
			sw.sendProgress("Plot Designer", "completed", "Loaded existing plot design", nil)
		}
	}

	// Step 2: World Building
	if resumeFrom == "start" || resumeFrom == "plot_design" || resumeFrom == "world_building" {
		sw.sendProgress("Worldbuilder", "started", "Designing story world...", nil)
		worldBuildingInput := fmt.Sprintf("Original Story Prompt:\n%s\n\nPlot Design:\n%s", originalPrompt, plotDesign)
		worldBuilding, err = sw.executeAgentTask(ctx, "worldbuilder", worldBuildingInput)
		if err != nil {
			sw.sendProgress("Worldbuilder", "error", "Failed to create world", err)
			return fmt.Errorf("world building failed: %w", err)
		}
		sw.sendProgress("Worldbuilder", "completed", "Story world designed", nil)
		sw.addCompletedTask("World Building")

		// Save world building immediately
		if err := sw.saveToWorkspace("world_building.md", fmt.Sprintf("# World Building\n\n%s", worldBuilding)); err != nil {
			return fmt.Errorf("failed to save world building: %w", err)
		}
	} else {
		// Load existing world building
		if data, exists := existingData["world_building"]; exists {
			worldBuilding = data
			sw.sendProgress("Worldbuilder", "completed", "Loaded existing world building", nil)
		}
	}

	// Step 3: Plot Expansion
	if resumeFrom == "start" || resumeFrom == "plot_design" || resumeFrom == "world_building" || resumeFrom == "plot_expansion" {
		sw.sendProgress("Plot Expander", "started", "Expanding plot details...", nil)
		plotExpansionInput := fmt.Sprintf("Original Story Prompt:\n%s\n\nPlot Design:\n%s\n\nWorld Building:\n%s", originalPrompt, plotDesign, worldBuilding)
		plotExpansion, err = sw.executeAgentTask(ctx, "plot_expander", plotExpansionInput)
		if err != nil {
			sw.sendProgress("Plot Expander", "error", "Failed to expand plot", err)
			return fmt.Errorf("plot expansion failed: %w", err)
		}
		sw.sendProgress("Plot Expander", "completed", "Plot details expanded", nil)
		sw.addCompletedTask("Plot Expansion")

		// Save plot expansion immediately
		if err := sw.saveToWorkspace("plot_expansion.md", fmt.Sprintf("# Plot Expansion\n\n%s", plotExpansion)); err != nil {
			return fmt.Errorf("failed to save plot expansion: %w", err)
		}
	} else {
		// Load existing plot expansion
		if data, exists := existingData["plot_expansion"]; exists {
			plotExpansion = data
			sw.sendProgress("Plot Expander", "completed", "Loaded existing plot expansion", nil)
		}
	}

	// Step 4: Character Development
	if resumeFrom == "start" || resumeFrom == "plot_design" || resumeFrom == "world_building" || resumeFrom == "plot_expansion" || resumeFrom == "characters" {
		sw.sendProgress("Character Developer", "started", "Creating characters...", nil)
		characterInput := fmt.Sprintf("Plot:\n%s\n\nWorld:\n%s\n\nExpanded Plot:\n%s", plotDesign, worldBuilding, plotExpansion)
		characters, err = sw.executeAgentTask(ctx, "character_developer", characterInput)
		if err != nil {
			sw.sendProgress("Character Developer", "error", "Failed to create characters", err)
			return fmt.Errorf("character development failed: %w", err)
		}
		sw.sendProgress("Character Developer", "completed", "Characters created", nil)
		sw.addCompletedTask("Character Development")

		// Save characters immediately
		if err := sw.saveToWorkspace("characters.md", fmt.Sprintf("# Characters\n\n%s", characters)); err != nil {
			return fmt.Errorf("failed to save characters: %w", err)
		}
	} else {
		// Load existing characters
		if data, exists := existingData["characters"]; exists {
			characters = data
			sw.sendProgress("Character Developer", "completed", "Loaded existing characters", nil)
		}
	}

	// Step 5: Write chapters and summarize in parallel
	authorInput := fmt.Sprintf("Characters:\n%s\n\nWorld:\n%s\n\nExpanded Plot:\n%s", characters, worldBuilding, plotExpansion)

	// Phase 1: Write all chapters and start summarization in parallel
	chapters := make([]string, 9)
	chapterSummaries := make([]string, 9)

	// Check if we should skip to editing phase
	if resumeFrom == "editing" {
		// Load all existing chapters and summaries for editing
		for i := 1; i <= 9; i++ {
			if chapterData, exists := existingData[fmt.Sprintf("chapter_%d", i)]; exists {
				// Extract chapter content
				lines := strings.Split(chapterData, "\n")
				chapter := ""
				for j, line := range lines {
					if strings.HasPrefix(line, "# Chapter") {
						// Skip header and empty lines
						for k := j + 1; k < len(lines) && strings.TrimSpace(lines[k]) == ""; k++ {
							j = k
						}
						if j+1 < len(lines) {
							chapter = strings.Join(lines[j+1:], "\n")
						}
						break
					}
				}
				if chapter == "" {
					chapter = chapterData // Fallback
				}
				chapters[i-1] = chapter
			}

			// Load existing summaries
			if summaryData, exists := existingData[fmt.Sprintf("chapter_%d_summary", i)]; exists {
				lines := strings.Split(summaryData, "\n")
				summary := ""
				for j, line := range lines {
					if strings.HasPrefix(line, "# Chapter") && strings.Contains(line, "Summary") {
						// Skip header and empty lines
						for k := j + 1; k < len(lines) && strings.TrimSpace(lines[k]) == ""; k++ {
							j = k
						}
						if j+1 < len(lines) {
							summary = strings.Join(lines[j+1:], "\n")
						}
						break
					}
				}
				if summary == "" {
					summary = summaryData // Fallback
				}
				chapterSummaries[i-1] = summary
			}
		}
	} else {
		// Normal chapter writing and summarization flow
		// Channel to collect summary results
		summaryResults := make(chan struct {
			index   int
			summary string
			err     error
		}, 9)

		summariesStarted := 0

		for i := 1; i <= 9; i++ {
			var chapter string
			var err error

			// Check if chapter already exists (resume logic)
			if chapterData, exists := existingData[fmt.Sprintf("chapter_%d", i)]; exists {
				// Chapter exists, extract content
				lines := strings.Split(chapterData, "\n")
				for j, line := range lines {
					if strings.HasPrefix(line, "# Chapter") {
						// Skip header and empty lines
						for k := j + 1; k < len(lines) && strings.TrimSpace(lines[k]) == ""; k++ {
							j = k
						}
						if j+1 < len(lines) {
							chapter = strings.Join(lines[j+1:], "\n")
						}
						break
					}
				}
				if chapter == "" {
					chapter = chapterData // Fallback
				}
				sw.sendProgress(fmt.Sprintf("Author (Ch %d)", i), "completed", fmt.Sprintf("Loaded existing chapter %d", i), nil)
			} else {
				// Write new chapter
				agentName := fmt.Sprintf("Author (Ch %d)", i)
				sw.sendProgress(agentName, "started", fmt.Sprintf("Writing chapter %d...", i), nil)
				chapterPrompt := fmt.Sprintf("%s\n\nWrite Chapter %d of the story.", authorInput, i)
				chapter, err = sw.executeAgentTaskWithName(ctx, "author", agentName, chapterPrompt)
				if err != nil {
					sw.sendProgress(agentName, "error", fmt.Sprintf("Failed to write chapter %d", i), err)
					return fmt.Errorf("chapter %d writing failed: %w", i, err)
				}
				sw.sendProgress(agentName, "completed", fmt.Sprintf("Chapter %d written", i), nil)

				// Save chapter
				chapterFilename := fmt.Sprintf("chapter_%d.md", i)
				if err := sw.saveToWorkspace(chapterFilename, fmt.Sprintf("# Chapter %d\n\n%s", i, chapter)); err != nil {
					return fmt.Errorf("failed to save chapter %d: %w", i, err)
				}
			}

			chapters[i-1] = chapter
			sw.addCompletedTask(fmt.Sprintf("Chapter %d Written", i))

			// Start summarization in parallel (if not already exists)
			if summaryData, exists := existingData[fmt.Sprintf("chapter_%d_summary", i)]; exists {
				// Summary already exists, extract content
				lines := strings.Split(summaryData, "\n")
				summary := ""
				for j, line := range lines {
					if strings.HasPrefix(line, "# Chapter") && strings.Contains(line, "Summary") {
						// Skip header and empty lines
						for k := j + 1; k < len(lines) && strings.TrimSpace(lines[k]) == ""; k++ {
							j = k
						}
						if j+1 < len(lines) {
							summary = strings.Join(lines[j+1:], "\n")
						}
						break
					}
				}
				if summary == "" {
					summary = summaryData // Fallback
				}
				chapterSummaries[i-1] = fmt.Sprintf("Chapter %d: %s", i, strings.TrimSpace(summary))
				sw.sendProgress("Story Summarizer", "completed", fmt.Sprintf("Loaded existing chapter %d summary", i), nil)

				// Send result to channel
				go func(idx int, sum string) {
					summaryResults <- struct {
						index   int
						summary string
						err     error
					}{idx, sum, nil}
				}(i-1, chapterSummaries[i-1])
				summariesStarted++
			} else {
				// Start summarization in background
				go func(chapterIndex int, chapterContent string) {
					sw.sendProgress("Story Summarizer", "started", fmt.Sprintf("Summarizing chapter %d...", chapterIndex+1), nil)
					summary, err := sw.executeAgentTask(ctx, "story_summarizer", chapterContent)
					if err != nil {
						sw.sendProgress("Story Summarizer", "error", fmt.Sprintf("Failed to summarize chapter %d", chapterIndex+1), err)
						summaryResults <- struct {
							index   int
							summary string
							err     error
						}{chapterIndex, "", err}
						return
					}

					sw.sendProgress("Story Summarizer", "completed", fmt.Sprintf("Chapter %d summarized", chapterIndex+1), nil)

					// Save chapter summary
					summaryFilename := fmt.Sprintf("chapter_%d_summary.md", chapterIndex+1)
					if err := sw.saveToWorkspace(summaryFilename, fmt.Sprintf("# Chapter %d Summary\n\n%s", chapterIndex+1, summary)); err != nil {
						summaryResults <- struct {
							index   int
							summary string
							err     error
						}{chapterIndex, "", err}
						return
					}

					formattedSummary := fmt.Sprintf("Chapter %d: %s", chapterIndex+1, strings.TrimSpace(summary))
					summaryResults <- struct {
						index   int
						summary string
						err     error
					}{chapterIndex, formattedSummary, nil}
				}(i-1, chapter)
				summariesStarted++
			}

			// Check memory pressure and compact if needed
			if sw.memory.NeedsCompaction() {
				sw.memory.CompactSession(sw.initialPrompt, sw.completedTasks, chapter)
			}
		}

		// Phase 2: Wait for all summaries to complete
		for i := 0; i < summariesStarted; i++ {
			result := <-summaryResults
			if result.err != nil {
				return fmt.Errorf("chapter %d summarization failed: %w", result.index+1, result.err)
			}
			chapterSummaries[result.index] = result.summary
			sw.addCompletedTask(fmt.Sprintf("Chapter %d Summarized", result.index+1))
		}
	} // Close the else block

	// Phase 3: Edit all chapters using all summaries
	allSummariesText := strings.Join(chapterSummaries, "\n")
	for i := 1; i <= 9; i++ {
		var editedChapter string
		var err error

		// Check if edited chapter already exists (resume logic)
		if _, exists := existingData[fmt.Sprintf("chapter_%d_edited", i)]; exists {
			// Chapter is already edited, skip
			sw.sendProgress("Editor", "completed", fmt.Sprintf("Loaded existing chapter %d edit", i), nil)
			continue
		}

		// Edit chapter with all summaries
		sw.sendProgress("Editor", "started", fmt.Sprintf("Editing chapter %d...", i), nil)
		editInput := fmt.Sprintf("All Chapter Summaries:\n%s\n\nCurrent Chapter to Edit:\n%s",
			allSummariesText, chapters[i-1])
		editedChapter, err = sw.executeAgentTask(ctx, "editor", editInput)
		if err != nil {
			sw.sendProgress("Editor", "error", fmt.Sprintf("Failed to edit chapter %d", i), err)
			return fmt.Errorf("chapter %d editing failed: %w", i, err)
		}
		sw.sendProgress("Editor", "completed", fmt.Sprintf("Chapter %d edited", i), nil)

		// Save edited chapter
		editedFilename := fmt.Sprintf("chapter_%d_edited.md", i)
		if err := sw.saveToWorkspace(editedFilename, fmt.Sprintf("# Chapter %d (Edited)\n\n%s", i, editedChapter)); err != nil {
			return fmt.Errorf("failed to save edited chapter %d: %w", i, err)
		}

		sw.addCompletedTask(fmt.Sprintf("Chapter %d Edited", i))

		// Check memory pressure and compact if needed
		if sw.memory.NeedsCompaction() {
			sw.memory.CompactSession(sw.initialPrompt, sw.completedTasks, editedChapter)
		}
	}

	// Save chapter summaries
	allSummaries := fmt.Sprintf("# Story Summary\n\n%s", strings.Join(chapterSummaries, "\n"))
	if err := sw.saveToWorkspace("story_summary.md", allSummaries); err != nil {
		return fmt.Errorf("failed to save story summary: %w", err)
	}

	// Create complete story by concatenating all edited chapters
	if err := sw.CreateCompleteStory(); err != nil {
		return fmt.Errorf("failed to create complete story: %w", err)
	}

	sw.addCompletedTask("Complete Story Created")
	return nil
}

// executeAgentTask runs a single agent task
func (sw *StoryWorkflow) executeAgentTask(ctx context.Context, agentName, input string) (string, error) {
	return sw.executeAgentTaskWithName(ctx, agentName, agentName, input)
}

// executeAgentTaskWithName runs a single agent task with a custom tracking name
func (sw *StoryWorkflow) executeAgentTaskWithName(ctx context.Context, agentName, trackingName, input string) (string, error) {
	agent, exists := sw.supervisor.GetAgent(agentName)
	if !exists {
		return "", fmt.Errorf("agent %s not found", agentName)
	}

	// Use direct call with token tracking, but use trackingName for token attribution
	result, err := agent.CallWithTrackingName(ctx, input, sw.tokenTracker, trackingName)
	if err != nil {
		return "", err
	}

	// Auto-save token usage after each call
	if err := sw.tokenTracker.SaveToFile(sw.supervisor.workspace); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: Failed to save token usage: %v\n", err)
	}

	sw.memory.AddEntry(fmt.Sprintf("%s Result: %s", trackingName, result))
	return result, nil
}

// saveToWorkspace saves content to the workspace
func (sw *StoryWorkflow) saveToWorkspace(filename, content string) error {
	return sw.supervisor.workspace.WriteFile(filename, content)
}

// loadFromWorkspace loads content from workspace if it exists
func (sw *StoryWorkflow) loadFromWorkspace(filename string) (string, bool, error) {
	content, err := sw.supervisor.workspace.ReadFile(filename)
	if err != nil {
		return "", false, nil // File doesn't exist or can't be read
	}
	return content, true, nil
}

// checkWorkspaceProgress checks which files exist to determine resume point
func (sw *StoryWorkflow) checkWorkspaceProgress() (resumeFrom string, existingData map[string]string) {
	existingData = make(map[string]string)

	// Check foundation files
	if content, exists, _ := sw.loadFromWorkspace("story_prompt.md"); exists {
		existingData["story_prompt"] = content
	} else {
		return "start", existingData
	}

	if content, exists, _ := sw.loadFromWorkspace("plot_design.md"); exists {
		existingData["plot_design"] = content
	} else {
		return "plot_design", existingData
	}

	if content, exists, _ := sw.loadFromWorkspace("world_building.md"); exists {
		existingData["world_building"] = content
	} else {
		return "world_building", existingData
	}

	if content, exists, _ := sw.loadFromWorkspace("plot_expansion.md"); exists {
		existingData["plot_expansion"] = content
	} else {
		return "plot_expansion", existingData
	}

	if content, exists, _ := sw.loadFromWorkspace("characters.md"); exists {
		existingData["characters"] = content
	} else {
		return "characters", existingData
	}

	// Check chapters and summaries
	originalChaptersComplete := 0
	allEditedComplete := true

	for i := 1; i <= 9; i++ {
		// Check for original chapter
		if content, exists, _ := sw.loadFromWorkspace(fmt.Sprintf("chapter_%d.md", i)); exists {
			existingData[fmt.Sprintf("chapter_%d", i)] = content
			originalChaptersComplete = i
		}

		// Check for edited chapter
		if content, exists, _ := sw.loadFromWorkspace(fmt.Sprintf("chapter_%d_edited.md", i)); exists {
			existingData[fmt.Sprintf("chapter_%d_edited", i)] = content
		} else if originalChaptersComplete >= i {
			// Original chapter exists but not edited yet
			allEditedComplete = false
		}

		// Load chapter summary if it exists
		if content, exists, _ := sw.loadFromWorkspace(fmt.Sprintf("chapter_%d_summary.md", i)); exists {
			existingData[fmt.Sprintf("chapter_%d_summary", i)] = content
		}
	}

	// Determine resume point based on what's missing
	if originalChaptersComplete < 9 {
		// Still need to write original chapters
		return fmt.Sprintf("chapter_%d", originalChaptersComplete+1), existingData
	} else if !allEditedComplete {
		// All original chapters exist, but editing is incomplete
		// Resume from editing phase
		return "editing", existingData
	}

	return "complete", existingData
}

// CreateCompleteStory concatenates all edited chapters into a single complete story file
func (sw *StoryWorkflow) CreateCompleteStory() error {
	var completeStory strings.Builder

	// Add title and metadata
	completeStory.WriteString("# Complete Story\n\n")
	completeStory.WriteString("*Generated by Agentic Storywriter*\n\n")
	completeStory.WriteString("---\n\n")

	// Read and concatenate all edited chapters
	for i := 1; i <= 9; i++ {
		chapterFilename := fmt.Sprintf("chapter_%d_edited.md", i)
		chapterContent, exists, err := sw.loadFromWorkspace(chapterFilename)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", chapterFilename, err)
		}
		if !exists {
			// If edited chapter doesn't exist, try original chapter
			chapterFilename = fmt.Sprintf("chapter_%d.md", i)
			chapterContent, exists, err = sw.loadFromWorkspace(chapterFilename)
			if err != nil {
				return fmt.Errorf("failed to read %s: %w", chapterFilename, err)
			}
			if !exists {
				continue // Skip missing chapters
			}
		}

		// Remove the markdown header from individual chapters to avoid duplication
		lines := strings.Split(chapterContent, "\n")
		contentStart := 0
		for j, line := range lines {
			if strings.HasPrefix(line, "# Chapter") {
				contentStart = j + 1
				// Skip empty lines after header
				for contentStart < len(lines) && strings.TrimSpace(lines[contentStart]) == "" {
					contentStart++
				}
				break
			}
		}

		// Add chapter header and content
		completeStory.WriteString(fmt.Sprintf("## Chapter %d\n\n", i))
		if contentStart < len(lines) {
			completeStory.WriteString(strings.Join(lines[contentStart:], "\n"))
		}
		completeStory.WriteString("\n\n---\n\n")
	}

	// Add story summary at the end
	summaryContent, exists, err := sw.loadFromWorkspace("story_summary.md")
	if err == nil && exists {
		completeStory.WriteString("## Story Summary\n\n")
		// Remove the header from summary
		summaryLines := strings.Split(summaryContent, "\n")
		summaryStart := 0
		for j, line := range summaryLines {
			if strings.HasPrefix(line, "# Story Summary") {
				summaryStart = j + 1
				for summaryStart < len(summaryLines) && strings.TrimSpace(summaryLines[summaryStart]) == "" {
					summaryStart++
				}
				break
			}
		}
		if summaryStart < len(summaryLines) {
			completeStory.WriteString(strings.Join(summaryLines[summaryStart:], "\n"))
		}
	}

	// Save the complete story
	return sw.saveToWorkspace("full_story_complete.md", completeStory.String())
}

// addCompletedTask adds a task to the completed tasks list
func (sw *StoryWorkflow) addCompletedTask(task string) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.completedTasks = append(sw.completedTasks, task)
}
