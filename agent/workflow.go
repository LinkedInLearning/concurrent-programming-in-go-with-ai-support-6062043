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
}

// Workflow orchestrates the execution of multiple agents in a concurrent pipeline.
type Workflow struct {
	apiKey       string
	ctx          context.Context
	statusUpdate func(string)
}

// NewWorkflow creates a new workflow instance with the given API key, context, and status update callback.
func NewWorkflow(apiKey string, ctx context.Context, statusUpdate func(string)) *Workflow {
	return &Workflow{
		apiKey:       apiKey,
		ctx:          ctx,
		statusUpdate: statusUpdate,
	}
}

// Run executes the complete workflow, coordinating all agents and collecting their results.
func (w *Workflow) Run() []WorkflowResult {
	var results []WorkflowResult
	var mu sync.Mutex

	addResult := func(name, output string, err error) {
		mu.Lock()
		defer mu.Unlock()
		results = append(results, WorkflowResult{
			AgentName: name,
			Output:    output,
			Error:     err,
		})
	}

	// Check if context is already cancelled
	if w.ctx.Err() != nil {
		addResult("System", "", fmt.Errorf("context cancelled before starting: %w", w.ctx.Err()))
		return results
	}

	w.statusUpdate("Starting writer agent...")
	writer := NewWriterAgent(w.apiKey)
	writerOut := make(chan string, 1)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		// Writer doesn't need input, just start it directly
		if err := writer.Start(w.ctx, nil, writerOut); err != nil {
			addResult("Writer", "", err)
		}
	}()

	wg.Wait()

	select {
	case writerResult := <-writerOut:
		if writerResult == "" {
			addResult("Writer", "", fmt.Errorf("empty output from writer"))
			return results
		}

		addResult("Writer", writerResult, nil)

		w.statusUpdate("Processing content with analysis agents...")
		summarizer := NewSummarizerAgent(w.apiKey)
		rater := NewStructuredRaterAgent(w.apiKey)
		titler := NewTitlerAgent(w.apiKey)
		formatter := NewMarkdownFormatterAgent(w.apiKey)

		var wg2 sync.WaitGroup

		wg2.Add(1)
		go func() {
			defer wg2.Done()
			result, err := w.runSingleAgent(summarizer, "Summarizer", writerResult)
			addResult("Summarizer", result, err)
		}()

		wg2.Add(1)
		go func() {
			defer wg2.Done()
			result, err := w.runStructuredAgent(rater, "Rater", writerResult)
			addResult("Rater", result, err)
		}()

		wg2.Add(1)
		go func() {
			defer wg2.Done()
			result, err := w.runSingleAgent(titler, "Titler", writerResult)
			addResult("Titler", result, err)
		}()

		wg2.Wait()

		// Format all results into markdown
		w.statusUpdate("Formatting results as markdown...")
		allContent := fmt.Sprintf("Title: %s\n\nSummary: %s\n\nRating: %s\n\nOriginal Content: %s",
			w.getResultByName(results, "Titler"),
			w.getResultByName(results, "Summarizer"),
			w.getResultByName(results, "Rater"),
			writerResult)

		wg2.Add(1)
		go func() {
			defer wg2.Done()
			result, err := w.runSingleAgent(formatter, "MarkdownFormatter", allContent)
			addResult("MarkdownFormatter", result, err)
		}()

		wg2.Wait()

	case <-w.ctx.Done():
		addResult("Writer", "", w.ctx.Err())
	case <-time.After(30 * time.Second):
		addResult("Writer", "", fmt.Errorf("timeout waiting for writer response"))
	}

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

// getResultByName retrieves the output of a specific agent by name from the results slice.
func (w *Workflow) getResultByName(results []WorkflowResult, name string) string {
	for _, result := range results {
		if result.AgentName == name && result.Error == nil {
			return result.Output
		}
	}
	return "N/A"
}
