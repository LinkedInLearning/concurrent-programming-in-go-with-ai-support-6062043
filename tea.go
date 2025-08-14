package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"concurrent-programming-go-agents/agent"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")).
			MarginBottom(1)

	errorStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#EF4444")).
			Padding(1).
			MarginBottom(1)
)

// model represents the application state for the bubbletea TUI.
type model struct {
	spinner       spinner.Model
	viewport      viewport.Model
	results       []agent.WorkflowResult
	workflowStats *agent.WorkflowStats
	finished      bool
	status        string
	quitting      bool
	ready         bool
}

// resultsMsg is a message type that carries workflow results.
type resultsMsg struct {
	results []agent.WorkflowResult
	stats   *agent.WorkflowStats
}

// statusMsg is a message type that carries status updates.
type statusMsg string

// initialModel creates and returns the initial model state for the application.
func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED"))
	return model{
		spinner: s,
		results: make([]agent.WorkflowResult, 0),
		status:  "Initializing workflow...",
	}
}

// Init initializes the model and returns the initial command to run.
func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, runWorkflow())
}

// Update handles incoming messages and updates the model state accordingly.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-4)
			m.viewport.YPosition = 1
			m.ready = true
		}
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 4

	case tea.KeyMsg:
		if m.finished {
			switch msg.String() {
			case "ctrl+c", "q":
				m.quitting = true
				return m, tea.Quit
			case "up", "k":
				m.viewport.ScrollUp(1)
			case "down", "j":
				m.viewport.ScrollDown(1)
			case "pgup", "b":
				m.viewport.HalfPageUp()
			case "pgdown", "f":
				m.viewport.HalfPageDown()
			case "home", "g":
				m.viewport.GotoTop()
			case "end", "G":
				m.viewport.GotoBottom()
			}
		} else {
			switch msg.String() {
			case "ctrl+c", "q":
				m.quitting = true
				return m, tea.Quit
			}
		}

	case spinner.TickMsg:
		if !m.finished && !m.quitting {
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case statusMsg:
		if !m.finished {
			m.status = string(msg)
			// Continue polling for status updates
			return m, statusUpdater()
		}

	case resultsMsg:
		m.results = msg.results
		m.workflowStats = msg.stats
		m.finished = true
		m.status = "Workflow complete!"
		// Update viewport content when results are ready
		if m.ready {
			m.viewport.SetContent(m.renderContent())
		}
	}

	// Update viewport
	if m.finished && m.ready {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the current state of the model as a string for display.
func (m model) View() string {
	if m.quitting {
		goodbyeMsg := "\nGoodbye! 👋\n"
		if m.workflowStats != nil {
			goodbyeMsg += m.formatTimingStats()
		}
		return goodbyeMsg
	}

	if !m.finished {
		return fmt.Sprintf("\n%s %s\n\n", m.spinner.View(), m.status)
	}

	if !m.ready {
		return "\nInitializing...\n"
	}

	header := titleStyle.Render("🤖 Agentic Application Results")
	footer := "\n" + lipgloss.NewStyle().Faint(true).Render("↑/↓: scroll • q/ctrl+c: quit • g/G: top/bottom • pgup/pgdn: page up/down")

	return header + "\n" + m.viewport.View() + footer
}

// renderContent formats and renders the workflow results as markdown content.
func (m model) renderContent() string {
	var output string

	// Check if we have a markdown formatter result
	var markdownContent string
	var hasErrors bool

	for _, result := range m.results {
		if result.Error != nil {
			hasErrors = true
			content := fmt.Sprintf("Agent: %s\nError: %s", result.AgentName, result.Error.Error())
			output += errorStyle.Render(content) + "\n"
		} else if result.AgentName == agent.MarkdownFormatterAgentName {
			markdownContent = result.Output
		}
	}

	// If we have markdown content and no errors, render it with glamour
	if markdownContent != "" && !hasErrors {
		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(m.viewport.Width-4),
		)
		if err != nil {
			output += errorStyle.Render(fmt.Sprintf("Failed to create markdown renderer: %v", err)) + "\n"
		} else {
			rendered, err := renderer.Render(markdownContent)
			if err != nil {
				output += errorStyle.Render(fmt.Sprintf("Failed to render markdown: %v", err)) + "\n"
			} else {
				// Remove trailing newlines to prevent extra spacing
				output += strings.TrimRight(rendered, "\n")
			}
		}
	} else if !hasErrors {
		// Fallback to showing individual results if no markdown formatter
		for _, result := range m.results {
			if result.Error == nil && result.AgentName != agent.MarkdownFormatterAgentName {
				output += fmt.Sprintf("**%s:**\n%s\n\n", result.AgentName, result.Output)
			}
		}
	}

	// Show individual agent results for debugging if there are errors
	if hasErrors {
		output += "\n**Individual Agent Results:**\n"
		for _, result := range m.results {
			if result.Error == nil && result.AgentName != agent.MarkdownFormatterAgentName {
				output += fmt.Sprintf("- **%s:** %s\n", result.AgentName, truncateString(result.Output, 100))
			}
		}
	}

	return output
}

// truncateString truncates a string to the specified maximum length and adds ellipsis if needed.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Global variable to track workflow status (simple solution)
var currentWorkflowStatus = "Initializing workflow..."

// runWorkflow creates and executes the agent workflow, returning a command that produces workflow results.
func runWorkflow() tea.Cmd {
	return tea.Batch(
		// Start a status updater that polls the global status
		statusUpdater(),
		// Run the actual workflow
		func() tea.Msg {
			apiKey := os.Getenv(EnvOpenAIAPIKey)
			if apiKey == "" {
				return resultsMsg{
					results: []agent.WorkflowResult{
						{AgentName: "System", Error: fmt.Errorf("OPENAI_API_KEY environment variable not set")},
					},
					stats: nil,
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			workflow := agent.NewWorkflow(apiKey, ctx, func(status string) {
				// Update the global status variable
				currentWorkflowStatus = status
			})
			
			results := workflow.Run()
			stats := workflow.GetStats()
			currentWorkflowStatus = "Workflow complete!"
			
			return resultsMsg{
				results: results,
				stats:   stats,
			}
		},
	)
}

// statusUpdater creates a command that periodically checks for status updates
func statusUpdater() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
		return statusMsg(currentWorkflowStatus)
	})
}

// formatTimingStats formats and returns the timing statistics as a string.
func (m model) formatTimingStats() string {
	if m.workflowStats == nil {
		return ""
	}

	var output strings.Builder
	output.WriteString("\n📊 Workflow Timing Statistics:\n")
	output.WriteString(fmt.Sprintf("Total Duration: %v\n\n", m.workflowStats.TotalDuration))
	
	output.WriteString("Individual Agent Timings:\n")
	
	// Define the order of agents to display
	agentOrder := []string{
		agent.WriterAgentName,
		agent.SummarizerAgentName,
		agent.RaterAgentName,
		agent.TitlerAgentName,
		agent.MarkdownFormatterAgentName,
	}
	
	for _, agentName := range agentOrder {
		if duration, exists := m.workflowStats.AgentStats[agentName]; exists {
			output.WriteString(fmt.Sprintf("  • %s: %v\n", agentName, duration))
		}
	}
	
	output.WriteString("\n")
	return output.String()
}
