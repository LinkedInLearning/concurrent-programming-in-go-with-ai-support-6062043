# Token Usage & Cost Tracking Implementation Plan

## 🎯 Goal

Add real-time token usage and cost tracking to demonstrate resource monitoring in concurrent agentic systems. Show live token consumption, per-agent costs, and cumulative expenses with persistent storage.

## 📊 Current State Analysis

### Existing Token Infrastructure

- ✅ **tiktoken library** already integrated (`agent/memory.go`)
- ✅ **TokenCounter** struct with GPT-5 encoder
- ✅ **CountTokens()** method available

### Current OpenAI Integration

- ✅ **BaseAgent.callOpenAI()** makes API calls (`agent/base.go`)
- ✅ **GPT-5 model** specified in requests
- ❌ **No token usage capture** from API responses
- ❌ **No cost calculation**

### Current Progress System

- ✅ **Progress channels** for UI updates (`agent/story_workflow.go`)
- ✅ **Real-time status** display in TUI (`main.go`)
- ❌ **No token/cost display** in UI

## 🏗️ Implementation Plan

### Phase 1: Core Token Tracking Infrastructure

#### 1.1 Create Token Usage Tracker

**File**: `agent/token_tracker.go` (NEW)

```go
type TokenUsage struct {
    InputTokens  int     `json:"input_tokens"`
    OutputTokens int     `json:"output_tokens"`
    TotalTokens  int     `json:"total_tokens"`
    Cost         float64 `json:"cost_usd"`
}

type AgentTokenTracker struct {
    AgentName    string     `json:"agent_name"`
    Usage        TokenUsage `json:"usage"`
    CallCount    int        `json:"call_count"`
    LastUpdated  time.Time  `json:"last_updated"`
}

type SystemTokenTracker struct {
    TotalUsage   TokenUsage                        `json:"total_usage"`
    AgentUsage   map[string]*AgentTokenTracker     `json:"agent_usage"`
    SessionStart time.Time                         `json:"session_start"`
    mu           sync.RWMutex
}
```

#### 1.2 Add Cost Calculation

**GPT-4o Pricing (as of 2024)**:

- Input: $2.50 per 1M tokens
- Output: $10.00 per 1M tokens

```go
func (t *TokenUsage) CalculateCost() {
    inputCost := float64(t.InputTokens) * 2.50 / 1000000
    outputCost := float64(t.OutputTokens) * 10.00 / 1000000
    t.Cost = inputCost + outputCost
}
```

### Phase 2: Integrate Token Tracking into API Calls

#### 2.1 Modify BaseAgent.callOpenAI()

**File**: `agent/base.go`

```go
// Add token tracker parameter
func (a *BaseAgent) callOpenAI(ctx context.Context, prompt string, tracker *SystemTokenTracker) (string, error) {
    // Count input tokens
    inputTokens := a.countInputTokens(prompt)

    completion, err := a.client.Chat.Completions.New(callCtx, openai.ChatCompletionNewParams{
        Messages: []openai.ChatCompletionMessageParamUnion{
            openai.SystemMessage(a.Config.Prompt),
            openai.UserMessage(prompt),
        },
        Model: openai.ChatModelGPT4o,
    })

    if err != nil {
        return "", fmt.Errorf("OpenAI API call failed: %w", err)
    }

    // Extract token usage from response
    outputTokens := len(completion.Choices[0].Message.Content) // Estimate or use API response

    // Update tracker
    tracker.RecordUsage(a.Config.Name, inputTokens, outputTokens)

    return completion.Choices[0].Message.Content, nil
}
```

#### 2.2 Thread-Safe Token Updates

```go
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
```

### Phase 3: Persistent Storage

#### 3.1 JSON File Storage

**File**: `workspace/token_usage.json`

```json
{
  "total_usage": {
    "input_tokens": 45230,
    "output_tokens": 12890,
    "total_tokens": 58120,
    "cost_usd": 0.242075
  },
  "agent_usage": {
    "plot_designer": {
      "agent_name": "plot_designer",
      "usage": {
        "input_tokens": 1250,
        "output_tokens": 890,
        "total_tokens": 2140,
        "cost_usd": 0.012025
      },
      "call_count": 1,
      "last_updated": "2024-01-15T10:30:45Z"
    },
    "worldbuilder": {
      "agent_name": "worldbuilder",
      "usage": {
        "input_tokens": 2100,
        "output_tokens": 1200,
        "total_tokens": 3300,
        "cost_usd": 0.01725
      },
      "call_count": 1,
      "last_updated": "2024-01-15T10:32:12Z"
    }
  },
  "session_start": "2024-01-15T10:28:30Z"
}
```

#### 3.2 Auto-Save Functionality

```go
func (st *SystemTokenTracker) SaveToFile(workspace *WorkspaceManager) error {
    st.mu.RLock()
    defer st.mu.RUnlock()

    data, err := json.MarshalIndent(st, "", "  ")
    if err != nil {
        return err
    }

    return workspace.WriteFile("token_usage.json", string(data))
}

func (st *SystemTokenTracker) LoadFromFile(workspace *WorkspaceManager) error {
    content, err := workspace.ReadFile("token_usage.json")
    if err != nil {
        return err // File doesn't exist, start fresh
    }

    return json.Unmarshal([]byte(content), st)
}
```

### Phase 4: Real-Time UI Updates

#### 4.1 Add Token Display to Progress Updates

**File**: `agent/story_workflow.go`

```go
type ProgressUpdate struct {
    AgentName    string
    Status       string
    Message      string
    Error        error
    TokenUsage   *TokenUsage  // NEW
    TotalCost    float64      // NEW
}

func (sw *StoryWorkflow) sendProgress(agentName, status, message string, err error) {
    // Get current token usage
    totalCost := sw.tokenTracker.GetTotalCost()
    agentUsage := sw.tokenTracker.GetAgentUsage(agentName)

    select {
    case sw.progressChan <- ProgressUpdate{
        AgentName:  agentName,
        Status:     status,
        Message:    message,
        Error:      err,
        TokenUsage: agentUsage,
        TotalCost:  totalCost,
    }:
    default:
        // Channel full, skip update
    }
}
```

#### 4.2 Enhanced UI Display

**File**: `main.go`

```go
func (m storyModel) View() string {
    // ... existing code ...

    if m.isProcessing {
        // Add token usage display at top
        s += "💰 Total Cost: $" + fmt.Sprintf("%.4f", m.totalCost) + "\n"
        s += "🔢 Total Tokens: " + fmt.Sprintf("%d", m.totalTokens) + "\n"
        s += "📊 Memory Usage: [████████░░] " + fmt.Sprintf("%.1f%%", m.memoryPercentage) + "\n\n"

        s += "🔄 Story Creation Progress:\n\n"

        // Display each agent with token info
        for _, agent := range m.agentStatuses {
            // ... existing status display ...

            if agent.tokenUsage != nil {
                s += fmt.Sprintf("   💸 $%.4f (%d tokens)\n",
                    agent.tokenUsage.Cost, agent.tokenUsage.TotalTokens)
            }
        }
    }

    return s
}
```

### Phase 5: Integration Points

#### 5.1 Modify StoryWorkflow

**File**: `agent/story_workflow.go`

```go
type StoryWorkflow struct {
    supervisor     *StorySupervisor
    memory         *SessionMemory
    tokenTracker   *SystemTokenTracker  // NEW
    initialPrompt  string
    completedTasks []string
    progressChan   chan ProgressUpdate
    mu             sync.RWMutex
}

func NewStoryWorkflow(apiKey string) (*StoryWorkflow, error) {
    // ... existing code ...

    tokenTracker := NewSystemTokenTracker()

    // Try to load existing token data
    tokenTracker.LoadFromFile(supervisor.workspace)

    return &StoryWorkflow{
        supervisor:     supervisor,
        memory:         memory,
        tokenTracker:   tokenTracker,  // NEW
        completedTasks: make([]string, 0),
        progressChan:   make(chan ProgressUpdate, 100),
    }, nil
}
```

#### 5.2 Update executeAgentTask

```go
func (sw *StoryWorkflow) executeAgentTask(ctx context.Context, agentName, input string) (string, error) {
    agent, exists := sw.supervisor.GetAgent(agentName)
    if !exists {
        return "", fmt.Errorf("agent %s not found", agentName)
    }

    // Pass token tracker to agent
    result, err := agent.CallWithTracking(ctx, input, sw.tokenTracker)
    if err != nil {
        return "", err
    }

    // Auto-save token usage after each call
    sw.tokenTracker.SaveToFile(sw.supervisor.workspace)

    sw.memory.AddEntry(fmt.Sprintf("%s Result: %s", agentName, result))
    return result, nil
}
```

## 🚀 Implementation Order

1. **Create basic TokenTracker struct** in new file
2. **Add cost calculation method** with GPT-5 pricing (yes, I mean gpt 5. if the tiktoken library doesn't seem to support it, update the library!!!)
3. **Modify one agent call** to count tokens (estimate input/output)
4. **Add simple cost display** to UI

### Full Implementation (30 minutes):

1. Complete token tracking infrastructure
2. Integrate with all OpenAI calls
3. Add JSON persistence
4. Enhanced UI with per-agent breakdown
5. Auto-save and resume functionality

## 🎯 Success Metrics

- ✅ Live token count updates during story creation
- ✅ Per-agent cost breakdown in UI
- ✅ Persistent token_usage.json file in workspace
- ✅ Total session cost displayed prominently
- ✅ Memory usage percentage with compaction trigger
- ✅ Resume functionality preserves token history

This enhancement transforms the application into a production-ready system with full resource monitoring and cost awareness.
