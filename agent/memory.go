package agent

import (
	"fmt"
	"strings"

	"github.com/tiktoken-go/tokenizer"
)

// TokenCounter handles token counting for memory management
type TokenCounter struct {
	encoder tokenizer.Codec
}

// NewTokenCounter creates a new token counter for GPT-4o
func NewTokenCounter() (*TokenCounter, error) {
	encoder, err := tokenizer.ForModel(tokenizer.GPT4o)
	if err != nil {
		return nil, fmt.Errorf("failed to create tokenizer: %w", err)
	}
	
	return &TokenCounter{
		encoder: encoder,
	}, nil
}

// CountTokens counts the number of tokens in the given text
func (tc *TokenCounter) CountTokens(text string) int {
	tokens, _, _ := tc.encoder.Encode(text)
	return len(tokens)
}

// IsAtThreshold checks if the text is at 70% of the context threshold (128k tokens for GPT-4o)
func (tc *TokenCounter) IsAtThreshold(text string) bool {
	const maxTokens = 128000
	const threshold = 0.7
	
	tokenCount := tc.CountTokens(text)
	return float64(tokenCount) >= float64(maxTokens)*threshold
}

// SessionMemory manages conversation history and compaction
type SessionMemory struct {
	history      []string
	tokenCounter *TokenCounter
}

// NewSessionMemory creates a new session memory manager
func NewSessionMemory() (*SessionMemory, error) {
	tc, err := NewTokenCounter()
	if err != nil {
		return nil, err
	}
	
	return &SessionMemory{
		history:      make([]string, 0),
		tokenCounter: tc,
	}, nil
}

// AddEntry adds a new entry to the session history
func (sm *SessionMemory) AddEntry(entry string) {
	sm.history = append(sm.history, entry)
}

// GetFullHistory returns the complete history as a single string
func (sm *SessionMemory) GetFullHistory() string {
	return strings.Join(sm.history, "\n\n")
}

// NeedsCompaction checks if the session needs to be compacted
func (sm *SessionMemory) NeedsCompaction() bool {
	fullHistory := sm.GetFullHistory()
	return sm.tokenCounter.IsAtThreshold(fullHistory)
}

// CompactSession reduces the session to essential information
func (sm *SessionMemory) CompactSession(initialPrompt string, completedTasks []string, mostRecentInteraction string) {
	sm.history = []string{
		fmt.Sprintf("Initial Prompt: %s", initialPrompt),
		fmt.Sprintf("Completed Tasks:\n%s", strings.Join(completedTasks, "\n• ")),
		fmt.Sprintf("Most Recent Interaction: %s", mostRecentInteraction),
	}
}