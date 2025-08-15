package agent

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ProgressUpdate represents a status update from an agent
type ProgressUpdate struct {
	AgentName string
	Status    string // "started", "completed", "error"
	Message   string
	Error     error
}

// StoryWorkflow manages the complete story writing process
type StoryWorkflow struct {
	supervisor     *StorySupervisor
	memory         *SessionMemory
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
	
	supervisor.InitializeAgents()
	
	return &StoryWorkflow{
		supervisor:     supervisor,
		memory:         memory,
		completedTasks: make([]string, 0),
		progressChan:   make(chan ProgressUpdate, 100),
	}, nil
}

// GetProgressChan returns the progress channel for UI updates
func (sw *StoryWorkflow) GetProgressChan() <-chan ProgressUpdate {
	return sw.progressChan
}

// sendProgress sends a progress update
func (sw *StoryWorkflow) sendProgress(agentName, status, message string, err error) {
	select {
	case sw.progressChan <- ProgressUpdate{
		AgentName: agentName,
		Status:    status,
		Message:   message,
		Error:     err,
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
	
	// Variables to hold story components
	var plotDesign, worldBuilding, plotExpansion, characters string
	var err error
	
	if resumeFrom != "start" && len(existingData) > 0 {
		sw.sendProgress("System", "started", "Resuming from existing progress...", nil)
	}
	
	// Save the initial story prompt (if not resuming)
	if resumeFrom == "start" {
		if err := sw.saveToWorkspace("story_prompt.md", fmt.Sprintf("# Story Prompt\n\n%s", userPrompt)); err != nil {
			return fmt.Errorf("failed to save story prompt: %w", err)
		}
	}
	
	// Step 1: Plot Design
	if resumeFrom == "start" || resumeFrom == "plot_design" {
		sw.sendProgress("Plot Designer", "started", "Creating story structure...", nil)
		plotDesign, err = sw.executeAgentTask(ctx, "plot_designer", userPrompt)
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
		worldBuilding, err = sw.executeAgentTask(ctx, "worldbuilder", plotDesign)
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
		plotExpansionInput := fmt.Sprintf("Plot Design:\n%s\n\nWorld Building:\n%s", plotDesign, worldBuilding)
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
	
	// Step 5: Write chapters
	chapterSummaries := make([]string, 0)
	authorInput := fmt.Sprintf("Characters:\n%s\n\nWorld:\n%s\n\nExpanded Plot:\n%s", characters, worldBuilding, plotExpansion)
	
	// Write 5 chapters (can be adjusted)
	for i := 1; i <= 5; i++ {
		// Write chapter
		sw.sendProgress(fmt.Sprintf("Author (Ch %d)", i), "started", fmt.Sprintf("Writing chapter %d...", i), nil)
		chapterPrompt := fmt.Sprintf("%s\n\nWrite Chapter %d of the story.", authorInput, i)
		chapter, err := sw.executeAgentTask(ctx, "author", chapterPrompt)
		if err != nil {
			sw.sendProgress(fmt.Sprintf("Author (Ch %d)", i), "error", fmt.Sprintf("Failed to write chapter %d", i), err)
			return fmt.Errorf("chapter %d writing failed: %w", i, err)
		}
		sw.sendProgress(fmt.Sprintf("Author (Ch %d)", i), "completed", fmt.Sprintf("Chapter %d written", i), nil)
		
		// Save chapter
		chapterFilename := fmt.Sprintf("chapter_%d.md", i)
		if err := sw.saveToWorkspace(chapterFilename, fmt.Sprintf("# Chapter %d\n\n%s", i, chapter)); err != nil {
			return fmt.Errorf("failed to save chapter %d: %w", i, err)
		}
		
		// Summarize chapter (as per spec: once per chapter)
		sw.sendProgress("Story Summarizer", "started", fmt.Sprintf("Summarizing chapter %d...", i), nil)
		summary, err := sw.executeAgentTask(ctx, "story_summarizer", chapter)
		if err != nil {
			sw.sendProgress("Story Summarizer", "error", fmt.Sprintf("Failed to summarize chapter %d", i), err)
			return fmt.Errorf("chapter %d summarization failed: %w", i, err)
		}
		sw.sendProgress("Story Summarizer", "completed", fmt.Sprintf("Chapter %d summarized", i), nil)
		chapterSummaries = append(chapterSummaries, fmt.Sprintf("Chapter %d: %s", i, summary))
		
		// Save chapter summary immediately
		summaryFilename := fmt.Sprintf("chapter_%d_summary.md", i)
		if err := sw.saveToWorkspace(summaryFilename, fmt.Sprintf("# Chapter %d Summary\n\n%s", i, summary)); err != nil {
			return fmt.Errorf("failed to save chapter %d summary: %w", i, err)
		}
		
		// Edit chapter (as per spec: once per chapter with all summaries + current chapter)
		sw.sendProgress("Editor", "started", fmt.Sprintf("Editing chapter %d...", i), nil)
		editInput := fmt.Sprintf("Chapter Summaries:\n%s\n\nCurrent Chapter:\n%s", 
			strings.Join(chapterSummaries, "\n"), chapter)
		editedChapter, err := sw.executeAgentTask(ctx, "editor", editInput)
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
		
		sw.addCompletedTask(fmt.Sprintf("Chapter %d Written and Edited", i))
		
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
	agent, exists := sw.supervisor.GetAgent(agentName)
	if !exists {
		return "", fmt.Errorf("agent %s not found", agentName)
	}
	
	inputChan := make(chan string, 1)
	outputChan := make(chan string, 1)
	
	// Start agent in goroutine
	agentCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	
	go func() {
		defer close(outputChan)
		agent.Start(agentCtx, inputChan, outputChan)
	}()
	
	// Send input
	inputChan <- input
	close(inputChan)
	
	// Wait for output
	select {
	case result := <-outputChan:
		sw.memory.AddEntry(fmt.Sprintf("%s Result: %s", agentName, result))
		return result, nil
	case <-agentCtx.Done():
		return "", fmt.Errorf("agent %s timed out", agentName)
	}
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
	
	// Check chapters
	chaptersComplete := 0
	for i := 1; i <= 5; i++ {
		if content, exists, _ := sw.loadFromWorkspace(fmt.Sprintf("chapter_%d_edited.md", i)); exists {
			existingData[fmt.Sprintf("chapter_%d", i)] = content
			chaptersComplete = i
		} else {
			break
		}
	}
	
	if chaptersComplete < 5 {
		return fmt.Sprintf("chapter_%d", chaptersComplete+1), existingData
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
	for i := 1; i <= 5; i++ {
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