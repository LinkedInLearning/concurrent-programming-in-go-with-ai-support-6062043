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
	spinner  spinner.Model
	viewport viewport.Model
	results  []agent.WorkflowResult
	finished bool
	status   string
	quitting bool
	ready    bool
}

// resultsMsg is a message type that carries workflow results.
type resultsMsg []agent.WorkflowResult

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
				m.viewport.LineUp(1)
			case "down", "j":
				m.viewport.LineDown(1)
			case "pgup", "b":
				m.viewport.HalfViewUp()
			case "pgdown", "f":
				m.viewport.HalfViewDown()
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
		m.status = string(msg)

	case resultsMsg:
		m.results = []agent.WorkflowResult(msg)
		m.finished = true
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
		return "\nGoodbye! 👋\n"
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
		} else if result.AgentName == "MarkdownFormatter" {
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
			if result.Error == nil && result.AgentName != "MarkdownFormatter" {
				output += fmt.Sprintf("**%s:**\n%s\n\n", result.AgentName, result.Output)
			}
		}
	}

	// Show individual agent results for debugging if there are errors
	if hasErrors {
		output += "\n**Individual Agent Results:**\n"
		for _, result := range m.results {
			if result.Error == nil && result.AgentName != "MarkdownFormatter" {
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

// runWorkflow creates and executes the agent workflow, returning a command that produces workflow results.
func runWorkflow() tea.Cmd {
	return tea.Sequence(
		func() tea.Msg { return statusMsg("Starting writer agent...") },
		tea.Tick(2*time.Second, func(time.Time) tea.Msg { return statusMsg("Processing content with analysis agents...") }),
		tea.Tick(4*time.Second, func(time.Time) tea.Msg { return statusMsg("Formatting results as markdown...") }),
		func() tea.Msg {
			apiKey := os.Getenv(EnvOpenAIAPIKey)
			if apiKey == "" {
				return resultsMsg{
					{AgentName: "System", Error: fmt.Errorf("OPENAI_API_KEY environment variable not set")},
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			workflow := agent.NewWorkflow(apiKey, ctx, func(status string) {
				// Status updates are handled by the sequence above for simplicity
			})
			results := workflow.Run()

			return resultsMsg(results)
		},
	)
}