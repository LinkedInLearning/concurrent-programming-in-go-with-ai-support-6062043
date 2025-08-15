package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"concurrent-programming-go-agents/agent"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joho/godotenv"
)

const (
	StoryEnvOpenAIAPIKey = "OPENAI_API_KEY"
)

type AgentStatus struct {
	name       string
	status     string // "waiting", "running", "completed", "error"
	message    string
	spinner    spinner.Model
	startTime  time.Time
	endTime    time.Time
	tokenUsage *agent.TokenUsage // NEW: Token usage for this agent
}

type storyModel struct {
	userInput     string
	workflow      *agent.StoryWorkflow
	status        string
	isProcessing  bool
	err           error
	agentStatuses []AgentStatus
	currentStep   int
	progressChan  chan agent.ProgressUpdate
	totalCost     float64 // NEW: Total cost tracking
	totalTokens   int     // NEW: Total token tracking
}

type storyCompleteMsg struct {
	err error
}

type progressUpdateMsg struct {
	update agent.ProgressUpdate
}

func initialStoryModel() storyModel {
	// Initialize spinners for each agent
	agentNames := []string{
		"Plot Designer",
		"Worldbuilder",
		"Plot Expander",
		"Character Developer",
		"Author (Ch 1)",
		"Author (Ch 2)",
		"Author (Ch 3)",
		"Author (Ch 4)",
		"Author (Ch 5)",
		"Author (Ch 6)",
		"Author (Ch 7)",
		"Author (Ch 8)",
		"Author (Ch 9)",
		"Story Summarizer",
		"Editor",
	}

	agentStatuses := make([]AgentStatus, len(agentNames))
	for i, name := range agentNames {
		s := spinner.New()
		s.Spinner = spinner.Dot
		s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
		agentStatuses[i] = AgentStatus{
			name:    name,
			status:  "waiting",
			message: "Waiting to start...",
			spinner: s,
		}
	}

	model := storyModel{
		status:        "Enter your story prompt and press Enter to begin...",
		agentStatuses: agentStatuses,
		currentStep:   -1,
	}

	// Check if there are existing workspace files to auto-resume
	if hasExistingWorkspace() {
		model.status = "Found existing story in progress. Resuming automatically..."
		model.userInput = "auto-resume" // Special flag for auto-resume
	}

	return model
}

// hasExistingWorkspace checks if there are any story files in the workspace
func hasExistingWorkspace() bool {
	// Only check for story_prompt.md - the first file that gets created
	if _, err := os.Stat("workspace/story_prompt.md"); err == nil {
		return true
	}
	return false
}

func (m storyModel) Init() tea.Cmd {
	// Initialize spinner commands for all agents
	var cmds []tea.Cmd
	for i := range m.agentStatuses {
		cmds = append(cmds, m.agentStatuses[i].spinner.Tick)
	}

	// Auto-start if we have existing workspace
	if m.userInput == "auto-resume" {
		cmds = append(cmds, func() tea.Msg {
			// Small delay to show the resume message
			time.Sleep(time.Millisecond * 500)
			return tea.KeyMsg{Type: tea.KeyEnter}
		})
	}

	return tea.Batch(cmds...)
}

func (m storyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			if !m.isProcessing && m.userInput != "" {
				return m.startStoryCreation()
			}
		case "backspace":
			if len(m.userInput) > 0 && !m.isProcessing {
				m.userInput = m.userInput[:len(m.userInput)-1]
			}
		case "ctrl+h":
			if len(m.userInput) > 0 && !m.isProcessing {
				m.userInput = m.userInput[:len(m.userInput)-1]
			}
		default:
			if !m.isProcessing {
				// Handle pasting and regular input
				input := msg.String()
				for _, char := range input {
					// Only add printable characters
					if char >= 32 && char <= 126 {
						m.userInput += string(char)
					}
				}
			}
		}

	case progressUpdateMsg:
		// Update agent status based on progress update
		if msg.update.AgentName != "" {
			for i := range m.agentStatuses {
				if m.agentStatuses[i].name == msg.update.AgentName {
					m.agentStatuses[i].status = msg.update.Status
					m.agentStatuses[i].message = msg.update.Message
					m.agentStatuses[i].tokenUsage = msg.update.TokenUsage // NEW: Update token usage
					if msg.update.Status == "started" {
						m.agentStatuses[i].startTime = time.Now()
						// Start the spinner for this agent
						cmds = append(cmds, m.agentStatuses[i].spinner.Tick)
					} else if msg.update.Status == "completed" || msg.update.Status == "error" {
						m.agentStatuses[i].endTime = time.Now()
					}
					break
				}
			}
		}
		// Update total cost and tokens
		m.totalCost = msg.update.TotalCost
		m.totalTokens = msg.update.TotalTokens
		// Continue listening for progress updates
		if m.isProcessing {
			cmds = append(cmds, m.listenForProgress())
		}

	case spinner.TickMsg:
		// Update spinners for running agents
		if m.isProcessing {
			for i := range m.agentStatuses {
				if m.agentStatuses[i].status == "started" {
					var cmd tea.Cmd
					m.agentStatuses[i].spinner, cmd = m.agentStatuses[i].spinner.Update(msg)
					if cmd != nil {
						cmds = append(cmds, cmd)
					}
				}
			}
		}

	case storyCompleteMsg:
		m.isProcessing = false
		if msg.err != nil {
			m.err = msg.err
			m.status = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.status = "🎉 STORY COMPLETE! 🎉\n\nYour complete story has been generated and saved to the 'workspace' folder.\nCheck 'full_story_complete.md' for the complete story.\n\nPress any key to exit..."
		}
		// Exit after showing completion message
		return m, tea.Batch(
			tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
				return tea.Quit()
			}),
		)
	}

	return m, tea.Batch(cmds...)
}

// listenForProgress creates a command to listen for progress updates
func (m storyModel) listenForProgress() tea.Cmd {
	return func() tea.Msg {
		if m.workflow != nil {
			select {
			case update := <-m.workflow.GetProgressChan():
				return progressUpdateMsg{update: update}
			case <-time.After(50 * time.Millisecond):
				// Continue listening by returning a special message
				return progressUpdateMsg{update: agent.ProgressUpdate{}}
			}
		}
		return nil
	}
}

func (m storyModel) startStoryCreation() (tea.Model, tea.Cmd) {
	apiKey := os.Getenv(StoryEnvOpenAIAPIKey)
	if apiKey == "" {
		m.err = fmt.Errorf("OPENAI_API_KEY environment variable not set")
		m.status = "Error: OPENAI_API_KEY not set"
		return m, nil
	}

	// Handle auto-resume case
	var userPrompt = m.userInput

	workflow, err := agent.NewStoryWorkflow(apiKey)
	if err != nil {
		m.err = err
		m.status = fmt.Sprintf("Error creating workflow: %v", err)
		return m, nil
	}

	m.workflow = workflow
	m.isProcessing = true
	
	// Initialize UI with current token data
	m.totalCost = workflow.GetTotalCost()
	m.totalTokens = workflow.GetTotalTokens()
	
	if m.userInput == "auto-resume" {
		m.status = "Resuming story creation from existing progress..."
	} else {
		m.status = "Creating your story... This may take several minutes."
	}

	// Start the first agent (Plot Designer) immediately
	if len(m.agentStatuses) > 0 {
		m.agentStatuses[0].status = "started"
		m.agentStatuses[0].message = "Creating story structure..."
		m.agentStatuses[0].startTime = time.Now()
	}

	return m, tea.Batch(
		// Start the story creation workflow
		func() tea.Msg {
			ctx := context.WithoutCancel(context.Background())
			err := workflow.ExecuteStoryCreation(ctx, userPrompt)
			return storyCompleteMsg{err: err}
		},
		// Start listening for progress updates
		m.listenForProgress(),
	)
}

func (m storyModel) View() string {
	var s string

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		MarginBottom(1)

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1).
		Width(80)

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		MarginTop(1)

	if m.isProcessing {
		statusStyle = statusStyle.Foreground(lipgloss.Color("#04B575"))
	}

	if m.err != nil {
		statusStyle = statusStyle.Foreground(lipgloss.Color("#FF5F87"))
	}

	s += titleStyle.Render("🤖 Agentic Storywriter")
	s += "\n\n"

	if !m.isProcessing {
		s += "Enter your story prompt (a couple of sentences describing the story you want):\n\n"
		s += inputStyle.Render(m.userInput + "│")
		s += "\n\n"
		s += "Press Enter to start story creation, or Ctrl+C to quit\n"
	}

	s += "\n"
	s += statusStyle.Render(m.status)

	if m.isProcessing {
		s += "\n\n"

		// Add cost and token tracking display
		costStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Bold(true)
		tokenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))

		s += costStyle.Render(fmt.Sprintf("💰 Total Cost: $%.4f", m.totalCost)) + "  "
		s += tokenStyle.Render(fmt.Sprintf("🔢 Total Tokens: %s", formatNumber(m.totalTokens))) + "\n\n"

		s += "🔄 Story Creation Progress:\n\n"

		// Display each agent's status with spinners and token info
		for _, agent := range m.agentStatuses {
			var statusIcon string
			var statusColor lipgloss.Color

			switch agent.status {
			case "waiting":
				statusIcon = "⏳"
				statusColor = lipgloss.Color("#626262")
			case "started", "running":
				statusIcon = agent.spinner.View()
				statusColor = lipgloss.Color("#7D56F4")
			case "completed":
				statusIcon = "✅"
				statusColor = lipgloss.Color("#04B575")
			case "error":
				statusIcon = "❌"
				statusColor = lipgloss.Color("#FF5F87")
			default:
				statusIcon = "⏳"
				statusColor = lipgloss.Color("#626262")
			}

			agentStyle := lipgloss.NewStyle().Foreground(statusColor)
			s += fmt.Sprintf("%s %s - %s\n",
				statusIcon,
				agentStyle.Render(agent.name),
				agentStyle.Render(agent.message))

			// Add token usage info for completed agents
			if agent.tokenUsage != nil && agent.tokenUsage.TotalTokens > 0 {
				tokenInfo := lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render(
					fmt.Sprintf("   💸 $%.4f (%s tokens)", agent.tokenUsage.Cost, formatNumber(agent.tokenUsage.TotalTokens)))
				s += tokenInfo + "\n"
			}
		}

		s += "\n💡 Tip: This process may take several minutes as each agent works on your story."
	}

	return s
}

// formatNumber formats a number with commas for better readability
func formatNumber(n int) string {
	str := strconv.Itoa(n)
	if len(str) <= 3 {
		return str
	}

	var result strings.Builder
	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteRune(digit)
	}
	return result.String()
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or could not be loaded: %v", err)
	}

	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		fmt.Println("🤖 Agentic Storywriter")
		fmt.Println()
		fmt.Println("An AI-powered story writing application using multiple specialized agents.")
		fmt.Println()
		fmt.Println("SETUP:")
		fmt.Println("  Set OPENAI_API_KEY environment variable or create a .env file")
		fmt.Println()
		fmt.Println("USAGE:")
		fmt.Println("  go run main.go")
		fmt.Println("  go run .")
		fmt.Println()
		fmt.Println("AGENTS:")
		fmt.Println("  • Plot Designer - Creates 9-point story structure")
		fmt.Println("  • Worldbuilder - Develops story world and setting")
		fmt.Println("  • Plot Expander - Expands plot points into paragraphs")
		fmt.Println("  • Character Developer - Creates protagonist, villain, and supporting characters")
		fmt.Println("  • Author - Writes story chapters (2 pages each)")
		fmt.Println("  • Story Summarizer - Creates chapter summaries")
		fmt.Println("  • Editor - Reviews and edits for coherence")
		fmt.Println("  • Supervisor Summary - Manages memory and compaction")
		fmt.Println()
		fmt.Println("OUTPUT:")
		fmt.Println("  All generated content is saved to the 'workspace' folder as .md files")
		return
	}

	program := tea.NewProgram(initialStoryModel())
	if _, err := program.Run(); err != nil {
		log.Fatal(err)
	}
}
