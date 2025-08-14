package agent

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

type WorkflowResult struct {
	AgentName string
	Output    string
	Error     error
	Duration  time.Duration
}

type WorkflowStats struct {
	TotalDuration time.Duration
	AgentStats    map[string]time.Duration
}

type AdviceWorkflow struct {
	apiKey       string
	ctx          context.Context
	statusUpdate func(string)
	startTime    time.Time
	stats        *WorkflowStats
}

func NewAdviceWorkflow(apiKey string, ctx context.Context, statusUpdate func(string)) *AdviceWorkflow {
	return &AdviceWorkflow{
		apiKey:       apiKey,
		ctx:          ctx,
		statusUpdate: statusUpdate,
		stats: &WorkflowStats{
			AgentStats: make(map[string]time.Duration),
		},
	}
}

func (w *AdviceWorkflow) RateAdvice(advice string) []WorkflowResult {
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

	if w.ctx.Err() != nil {
		addResult("System", "", fmt.Errorf("context cancelled before starting: %w", w.ctx.Err()), 0)
		return results
	}

	w.statusUpdate("Analyzing advice with expert agents...")

	careerAgent := NewCareerAgent(w.apiKey)
	bestFriendAgent := NewBestFriendAgent(w.apiKey)
	financialAgent := NewFinancialAgent(w.apiKey)
	techSupportAgent := NewTechSupportAgent(w.apiKey)
	dieticianAgent := NewDieticianAgent(w.apiKey)
	lawyerAgent := NewLawyerAgent(w.apiKey)

	var expertWG sync.WaitGroup
	var expertRatings []string
	var ratingsMu sync.Mutex

	addRating := func(rating string) {
		ratingsMu.Lock()
		defer ratingsMu.Unlock()
		expertRatings = append(expertRatings, rating)
	}

	experts := []struct {
		agent Agent
		name  string
	}{
		{careerAgent, CareerAgentName},
		{bestFriendAgent, BestFriendAgentName},
		{financialAgent, FinancialAgentName},
		{techSupportAgent, TechSupportAgentName},
		{dieticianAgent, DieticianAgentName},
		{lawyerAgent, LawyerAgentName},
	}

	for _, expert := range experts {
		expertWG.Add(1)
		go func(agent Agent, name string) {
			defer expertWG.Done()
			w.statusUpdate(fmt.Sprintf("Getting %s opinion...", name))
			start := time.Now()
			rating, err := w.runSingleAgent(agent, name, advice)
			addResult(name, rating, err, time.Since(start))
			if err == nil {
				addRating(rating)
			}
		}(expert.agent, expert.name)
	}

	expertWG.Wait()

	w.statusUpdate("Summarizing expert opinions...")
	summarizerStart := time.Now()
	summarizer := NewAdviceSummarizerAgent(w.apiKey)

	ratingsMu.Lock()
	allRatings := strings.Join(expertRatings, "\n")
	ratingsMu.Unlock()

	summaryInput := fmt.Sprintf("Original advice: %s\n\nExpert ratings:\n%s", advice, allRatings)
	summary, err := w.runSingleAgent(summarizer, AdviceSummarizerAgentName, summaryInput)
	addResult(AdviceSummarizerAgentName, summary, err, time.Since(summarizerStart))

	w.stats.TotalDuration = time.Since(w.startTime)
	w.statusUpdate("Analysis complete!")
	return results
}

func (w *AdviceWorkflow) runSingleAgent(agent Agent, name, input string) (string, error) {
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

func (w *AdviceWorkflow) GetStats() *WorkflowStats {
	return w.stats
}