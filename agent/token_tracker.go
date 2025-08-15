package agent

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// TokenUsage represents token consumption and cost for a single operation
type TokenUsage struct {
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	TotalTokens  int     `json:"total_tokens"`
	Cost         float64 `json:"cost_usd"`
}

// CalculateCost calculates the cost based on GPT-4o pricing
func (t *TokenUsage) CalculateCost() {
	// GPT-4o pricing (as of 2024)
	// Input: $2.50 per 1M tokens
	// Output: $10.00 per 1M tokens
	inputCost := float64(t.InputTokens) * 2.50 / 1000000
	outputCost := float64(t.OutputTokens) * 10.00 / 1000000
	t.Cost = inputCost + outputCost
}

// AgentTokenTracker tracks token usage for a specific agent
type AgentTokenTracker struct {
	AgentName   string     `json:"agent_name"`
	Usage       TokenUsage `json:"usage"`
	CallCount   int        `json:"call_count"`
	LastUpdated time.Time  `json:"last_updated"`
}

// SystemTokenTracker tracks token usage across all agents in the system
type SystemTokenTracker struct {
	TotalUsage   TokenUsage                    `json:"total_usage"`
	AgentUsage   map[string]*AgentTokenTracker `json:"agent_usage"`
	SessionStart time.Time                     `json:"session_start"`
	mu           sync.RWMutex
}

// NewSystemTokenTracker creates a new system-wide token tracker
func NewSystemTokenTracker() *SystemTokenTracker {
	return &SystemTokenTracker{
		TotalUsage:   TokenUsage{},
		AgentUsage:   make(map[string]*AgentTokenTracker),
		SessionStart: time.Now(),
	}
}

// RecordUsage records token usage for a specific agent in a thread-safe manner
func (st *SystemTokenTracker) RecordUsage(agentName string, inputTokens, outputTokens int) {
	st.mu.Lock()
	defer st.mu.Unlock()

	// Update agent-specific usage
	if st.AgentUsage[agentName] == nil {
		st.AgentUsage[agentName] = &AgentTokenTracker{
			AgentName: agentName,
		}
	}

	agent := st.AgentUsage[agentName]
	agent.Usage.InputTokens += inputTokens
	agent.Usage.OutputTokens += outputTokens
	agent.Usage.TotalTokens += inputTokens + outputTokens
	agent.Usage.CalculateCost()
	agent.CallCount++
	agent.LastUpdated = time.Now()

	// Update system totals
	st.TotalUsage.InputTokens += inputTokens
	st.TotalUsage.OutputTokens += outputTokens
	st.TotalUsage.TotalTokens += inputTokens + outputTokens
	st.TotalUsage.CalculateCost()
}

// GetTotalCost returns the total cost across all agents
func (st *SystemTokenTracker) GetTotalCost() float64 {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.TotalUsage.Cost
}

// GetTotalTokens returns the total token count across all agents
func (st *SystemTokenTracker) GetTotalTokens() int {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.TotalUsage.TotalTokens
}

// GetAgentUsage returns the token usage for a specific agent
func (st *SystemTokenTracker) GetAgentUsage(agentName string) *TokenUsage {
	st.mu.RLock()
	defer st.mu.RUnlock()

	if agent, exists := st.AgentUsage[agentName]; exists {
		// Return a copy to avoid race conditions
		return &TokenUsage{
			InputTokens:  agent.Usage.InputTokens,
			OutputTokens: agent.Usage.OutputTokens,
			TotalTokens:  agent.Usage.TotalTokens,
			Cost:         agent.Usage.Cost,
		}
	}
	return &TokenUsage{}
}

// SaveToFile saves the token tracker data to a JSON file in the workspace
func (st *SystemTokenTracker) SaveToFile(workspace *WorkspaceManager) error {
	st.mu.RLock()
	defer st.mu.RUnlock()

	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token data: %w", err)
	}

	return workspace.WriteFile("token_usage.json", string(data))
}

// LoadFromFile loads token tracker data from a JSON file in the workspace
func (st *SystemTokenTracker) LoadFromFile(workspace *WorkspaceManager) error {
	content, err := workspace.ReadFile("token_usage.json")
	if err != nil {
		// File doesn't exist, start fresh
		return nil
	}

	st.mu.Lock()
	defer st.mu.Unlock()

	return json.Unmarshal([]byte(content), st)
}

// GetSessionDuration returns how long the current session has been running
func (st *SystemTokenTracker) GetSessionDuration() time.Duration {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return time.Since(st.SessionStart)
}

// GetAgentStats returns a summary of all agent statistics
func (st *SystemTokenTracker) GetAgentStats() map[string]*AgentTokenTracker {
	st.mu.RLock()
	defer st.mu.RUnlock()

	// Return a copy to avoid race conditions
	stats := make(map[string]*AgentTokenTracker)
	for name, agent := range st.AgentUsage {
		stats[name] = &AgentTokenTracker{
			AgentName:   agent.AgentName,
			Usage:       agent.Usage,
			CallCount:   agent.CallCount,
			LastUpdated: agent.LastUpdated,
		}
	}
	return stats
}
