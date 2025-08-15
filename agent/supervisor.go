package agent

import (
	"fmt"
	"sync"
)

// StorySupervisor manages the story writing workflow
type StorySupervisor struct {
	workspace *WorkspaceManager
	agents    map[string]*StoryAgent
	apiKey    string
	mu        sync.RWMutex
}

// NewStorySupervisor creates a new story supervisor
func NewStorySupervisor(apiKey string) (*StorySupervisor, error) {
	workspace, err := NewWorkspaceManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}
	
	return &StorySupervisor{
		workspace: workspace,
		agents:    make(map[string]*StoryAgent),
		apiKey:    apiKey,
	}, nil
}

// InitializeAgents creates all the story writing agents
func (s *StorySupervisor) InitializeAgents() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Plot Designer Agent
	s.agents["plot_designer"] = NewStoryAgent("plot_designer", 
		`You are a plot designer agent. Create a story plot following this exact structure:
1. Background scene and worldbuilding rules
2. Reader's reason for caring and protagonist background
3. Hero's journey begins - something wanted/needed, life upside down
4. Hero learns Something Horrible will happen if goals not achieved
5. Protagonist fails and suffers, but all is not lost
6. Hero learns from mistakes, emerges stronger
7. Hero learns villain/obstacle is more powerful than perceived
8. Hero overcomes anyway, pays steep price but achieves victory
9. Hero returns, adjusts to new life, changes evident

Provide detailed answers for each of the 9 plot points.`, s.apiKey)

	// Worldbuilder Agent
	s.agents["worldbuilder"] = NewStoryAgent("worldbuilder",
		`You are a worldbuilder agent. Look at the plot and develop a creative world where it can take place. Consider: Real world vs fantasy? Magic? Time period? Alternate history? Be creative and enhance the plot.`, s.apiKey)

	// Plot Expander Agent
	s.agents["plot_expander"] = NewStoryAgent("plot_expander",
		`You are a plot expander agent. Take the initial plot entries and worldbuilding info, then expand each plot point from sentences into full paragraphs. Check for plot holes and avoid cliches like deus ex machina.`, s.apiKey)

	// Character Developer Agent
	s.agents["character_developer"] = NewStoryAgent("character_developer",
		`You are a character developer agent. Create special characters: protagonist, villain (if applicable), and supporting characters. For each character provide: relevant backstory, lore, physical description, and names that fit the world.`, s.apiKey)

	// Author Agent
	s.agents["author"] = NewStoryAgent("author",
		`You are an author agent. Write the actual story one chapter at a time. Chapters should be short (no longer than 2 pages). Use the character info, worldbuilding, and expanded plot to bring the story to life.`, s.apiKey)

	// Story Summarizer Agent
	s.agents["story_summarizer"] = NewStoryAgent("story_summarizer",
		`You are a story summarizer agent. Take a completed chapter and summarize it down to a single paragraph so the editor and supervisor can keep the whole story in memory.`, s.apiKey)

	// Editor Agent
	s.agents["editor"] = NewStoryAgent("editor",
		`You are an editor agent. Review chapters for coherence across the story. You'll receive summaries of all chapters plus the current full chapter. Edit the text to eliminate plot holes and contradictions.`, s.apiKey)

	// Supervisor Summary Agent
	s.agents["supervisor_summary"] = NewStoryAgent("supervisor_summary",
		`You are a supervisor summary agent. Take the full agentic history and summarize it down to: the initial prompt, a bulleted list of all completed tasks, and the full text of the most recent interaction.`, s.apiKey)
}

// GetAgent returns an agent by name
func (s *StorySupervisor) GetAgent(name string) (*StoryAgent, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	agent, exists := s.agents[name]
	return agent, exists
}