package agent

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// WorkflowResult represents the output from a single agent in the workflow.
type WorkflowResult struct {
	AgentName string
	Output    string
	Error     error
	Duration  time.Duration
}

// WorkflowResult represents the output from a single agent in the workflow.
type WorkflowStats struct {
	TotalDuration time.Duration
	AgentStats    map[string]time.Duration
}

// Workflow orchestrates the execution of multiple agents in a concurrent pipeline.
type Workflow struct {
	apiKey       string
	ctx          context.Context
	statusUpdate func(string)
	startTime    time.Time
	stats        *WorkflowStats
}

// NewWorkflow creates a new workflow instance with the given API key, context, and status update callback.
func NewWorkflow(apiKey string, ctx context.Context, statusUpdate func(string)) *Workflow {
	return &Workflow{
		apiKey:       apiKey,
		ctx:          ctx,
		statusUpdate: statusUpdate,
		stats: &WorkflowStats{
			AgentStats: make(map[string]time.Duration),
		},
	}
}

// Run executes the complete workflow, coordinating all agents and collecting their results.
func (w *Workflow) Run() []WorkflowResult {
	w.startTime = time.Now()
	var results []WorkflowResult
	var mu sync.Mutex

	addResult := func(name, output string, err error, duration time.Duration) {
		mu.Lock()
		defer mu.Unlock()
		results = append(results, WorkflowResult{
			AgentName: name,
			Output:    output,
			Error:     err,
			Duration:  duration,
		})
		w.stats.AgentStats[name] = duration
	}

	// Check if context is already cancelled
	if w.ctx.Err() != nil {
		addResult("System", "", fmt.Errorf("context cancelled before starting: %w", w.ctx.Err()), 0)
		return results
	}

	w.statusUpdate("Starting writer agent...")
	writer := NewWriterAgent(w.apiKey)
	writerOut := make(chan string, 1)

	// Writer doesn't need input, just start it directly
	writerStart := time.Now()
	if err := writer.Start(w.ctx, nil, writerOut); err != nil {
		addResult(WriterAgentName, "", err, time.Since(writerStart))
	}

	select {
	case <-w.ctx.Done():
		addResult(WriterAgentName, "", w.ctx.Err(), time.Since(writerStart))
	case writerResult := <-writerOut:
		writerDuration := time.Since(writerStart)
		if writerResult == "" {
			addResult(WriterAgentName, "", fmt.Errorf("empty output from writer"), writerDuration)
			return results
		}

		addResult(WriterAgentName, writerResult, nil, writerDuration)

		w.statusUpdate("Processing content with analysis agents...")
		summarizer := NewSummarizerAgent(w.apiKey)
		rater := NewStructuredRaterAgent(w.apiKey)
		titler := NewTitlerAgent(w.apiKey)

		var analysisWG sync.WaitGroup
		analysisWG.Add(1)
		go func() {
			defer analysisWG.Done()
			w.statusUpdate("Summarizing the content...")
			start := time.Now()
			summarizerResult, err := w.runSingleAgent(summarizer, SummarizerAgentName, writerResult)
			addResult(SummarizerAgentName, summarizerResult, err, time.Since(start))
		}()
		analysisWG.Add(1)
		go func() {
			defer analysisWG.Done()
			w.statusUpdate("Rating the content...")
			ratingStart := time.Now()
			raterResult, err := w.runStructuredAgent(rater, RaterAgentName, writerResult)
			addResult(RaterAgentName, raterResult, err, time.Since(ratingStart))
		}()
		analysisWG.Add(1)
		go func() {
			defer analysisWG.Done()
			w.statusUpdate("Generating a title for the content...")
			titleStart := time.Now()
			titleResult, err := w.runSingleAgent(titler, TitlerAgentName, writerResult)
			addResult(TitlerAgentName, titleResult, err, time.Since(titleStart))
		}()
		analysisWG.Wait()

	}

	w.stats.TotalDuration = time.Since(w.startTime)
	w.statusUpdate("Workflow complete!")
	return results
}

// runSingleAgent executes a single agent with the given input and returns its output.
func (w *Workflow) runSingleAgent(agent Agent, name, input string) (string, error) {
	agentIn := make(chan string, 1)
	agentOut := make(chan string, 1)

	agentIn <- input
	close(agentIn)

	if err := agent.Start(w.ctx, agentIn, agentOut); err != nil {
		return "", err
	}

	select {
	case result := <-agentOut:
		if result == "" {
			return "", fmt.Errorf("empty response from %s", name)
		}
		return result, nil
	case <-w.ctx.Done():
		return "", w.ctx.Err()
	case <-time.After(30 * time.Second):
		return "", fmt.Errorf("timeout waiting for %s response", name)
	}
}

// runStructuredAgent executes a structured rater agent with the given input and returns its output.
func (w *Workflow) runStructuredAgent(agent *StructuredRaterAgent, name, input string) (string, error) {
	agentIn := make(chan string, 1)
	agentOut := make(chan string, 1)

	agentIn <- input
	close(agentIn)

	if err := agent.Start(w.ctx, agentIn, agentOut); err != nil {
		return "", err
	}

	select {
	case result := <-agentOut:
		if result == "" {
			return "", fmt.Errorf("empty response from %s", name)
		}
		return result, nil
	case <-w.ctx.Done():
		return "", w.ctx.Err()
	case <-time.After(30 * time.Second):
		return "", fmt.Errorf("timeout waiting for %s response", name)
	}
}

// GetStats returns the workflow timing statistics.
func (w *Workflow) GetStats() *WorkflowStats {
	return w.stats
}
