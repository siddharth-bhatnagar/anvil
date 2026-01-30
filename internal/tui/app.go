package tui

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/siddharth-bhatnagar/anvil/internal/config"
	"github.com/siddharth-bhatnagar/anvil/internal/llm"
	"github.com/siddharth-bhatnagar/anvil/internal/tui/panels"
)

const version = "0.1.0"

// Model represents the application state
type Model struct {
	width         int
	height        int
	ready         bool
	panelManager  *PanelManager
	showHelp      bool
	input         textinput.Model
	llmClient     llm.Client
	tokenTracker  *llm.TokenTracker
	configManager *config.Manager
	streaming     bool
	streamBuffer  string
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.panelManager.Init(),
		textinput.Blink,
	)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.updateLayout()
		return m, nil

	case tea.KeyMsg:
		// If streaming, only allow interrupt
		if m.streaming {
			if msg.String() == "ctrl+c" {
				m.streaming = false
				return m, tea.Quit
			}
			return m, nil
		}

		// Handle input field when focused
		if m.input.Focused() {
			switch msg.String() {
			case "enter":
				// Send message
				userMsg := m.input.Value()
				if userMsg != "" {
					m.input.SetValue("")
					return m, m.sendMessage(userMsg)
				}
				return m, nil
			case "esc":
				m.input.Blur()
				return m, nil
			default:
				var cmd tea.Cmd
				m.input, cmd = m.input.Update(msg)
				return m, cmd
			}
		}

		// Global shortcuts
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "i", "/":
			m.input.Focus()
			return m, textinput.Blink

		case "tab":
			m.panelManager.NextPanel()
			return m, nil

		case "shift+tab":
			m.panelManager.PrevPanel()
			return m, nil

		case "1":
			m.panelManager.SetActivePanelByType(PanelConversation)
			return m, nil

		case "2":
			m.panelManager.SetActivePanelByType(PanelFiles)
			return m, nil

		case "3":
			m.panelManager.SetActivePanelByType(PanelDiff)
			return m, nil

		case "4":
			m.panelManager.SetActivePanelByType(PanelPlan)
			return m, nil

		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		}

	case StreamChunkMsg:
		// Handle streaming chunk
		if msg.Error != nil {
			m.streaming = false
			// Show error
			return m, func() tea.Msg {
				return ErrorMsg{Error: msg.Error}
			}
		}

		if msg.Delta != "" {
			m.streamBuffer += msg.Delta
		}

		if msg.Done {
			m.streaming = false
			// Add complete message to conversation
			convPanel := m.panelManager.GetPanelByType(PanelConversation).(*panels.ConversationPanel)
			convPanel.AddMessage("assistant", m.streamBuffer)

			// Track usage
			if msg.Usage != nil {
				m.tokenTracker.AddUsage(*msg.Usage)
			}

			m.streamBuffer = ""
		}

		return m, nil

	case ErrorMsg:
		// Handle error - could show in status bar or conversation
		m.streaming = false
		return m, nil
	}

	// Update input if not focused on input
	if !m.input.Focused() {
		cmd := m.panelManager.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// sendMessage sends a message to the LLM
func (m *Model) sendMessage(content string) tea.Cmd {
	// Add user message to conversation
	convPanel := m.panelManager.GetPanelByType(PanelConversation).(*panels.ConversationPanel)
	convPanel.AddMessage("user", content)

	// Prepare request
	messages := []llm.Message{
		{Role: llm.RoleUser, Content: content},
	}

	req := llm.Request{
		Messages: messages,
		Stream:   true,
	}

	m.streaming = true
	m.streamBuffer = ""

	// Start streaming
	return func() tea.Msg {
		ctx := context.Background()

		err := m.llmClient.Stream(ctx, req, func(event llm.StreamEvent) {
			// Send each chunk as a message
			// Note: In a real implementation, we'd use a channel
			// For simplicity, we'll just accumulate and send at the end
		})

		if err != nil {
			return StreamChunkMsg{Error: err, Done: true}
		}

		return StreamChunkMsg{Done: true}
	}
}

// View renders the UI
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	if m.showHelp {
		return m.renderHelp()
	}

	// Header
	header := TitleStyle.Render(fmt.Sprintf("⚒  Anvil v%s", version))

	// Calculate dimensions
	headerHeight := lipgloss.Height(header)
	statusBarHeight := 1
	inputHeight := 3 // Input field height
	contentHeight := m.height - headerHeight - statusBarHeight - inputHeight - 2 // -2 for spacing

	// Layout panels in 2x2 grid
	panelWidth := m.width / 2
	panelHeight := contentHeight / 2

	allPanels := m.panelManager.GetPanels()
	if len(allPanels) >= 4 {
		// Top row: Conversation | Files
		convPanel := allPanels[0]
		filesPanel := allPanels[1]

		convView := m.renderPanelWithBorder(convPanel, panelWidth, panelHeight)
		filesView := m.renderPanelWithBorder(filesPanel, panelWidth, panelHeight)

		topRow := lipgloss.JoinHorizontal(lipgloss.Top, convView, filesView)

		// Bottom row: Diff | Plan
		diffPanel := allPanels[2]
		planPanel := allPanels[3]

		diffView := m.renderPanelWithBorder(diffPanel, panelWidth, panelHeight)
		planView := m.renderPanelWithBorder(planPanel, panelWidth, panelHeight)

		bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, diffView, planView)

		// Combine rows
		content := lipgloss.JoinVertical(lipgloss.Left, topRow, bottomRow)

		// Input field
		inputStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Width(m.width - 4).
			Padding(0, 1)

		if m.input.Focused() {
			inputStyle = inputStyle.BorderForeground(ColorAccent)
		}

		inputView := inputStyle.Render(m.input.View())

		// Status bar
		statusBar := m.renderStatusBar()

		// Combine all
		view := lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			content,
			inputView,
			statusBar,
		)

		return view
	}

	return "Initializing panels..."
}

// renderPanelWithBorder renders a panel with a border and title
func (m Model) renderPanelWithBorder(panel Panel, width, height int) string {
	// Create border style
	borderColor := ColorBorder
	if panel.IsFocused() {
		borderColor = ColorAccent
	}

	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(width - 2).
		Height(height - 2)

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(panel.IsFocused())

	title := titleStyle.Render(panel.Title())

	// Panel content
	content := panel.View()

	// Combine
	return style.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			"",
			content,
		),
	)
}

// renderStatusBar renders the status bar
func (m Model) renderStatusBar() string {
	activePanel := m.panelManager.GetActivePanel()
	activeName := "None"
	if activePanel != nil {
		activeName = activePanel.Title()
	}

	// Get token stats
	stats := m.tokenTracker.GetStats()
	tokenInfo := ""
	if stats.TotalTokens > 0 {
		tokenInfo = fmt.Sprintf(" | Tokens: %d", stats.TotalTokens)
	}

	// Show streaming indicator
	streamingIndicator := ""
	if m.streaming {
		streamingIndicator = " | ⚡ Streaming..."
	}

	left := fmt.Sprintf("Panel: %s%s%s", activeName, tokenInfo, streamingIndicator)
	right := fmt.Sprintf("i Input | ? Help | Tab Switch | q Quit | %dx%d", m.width, m.height)

	// Calculate spacing
	spacing := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if spacing < 0 {
		spacing = 0
	}

	statusContent := left + lipgloss.NewStyle().Width(spacing).Render("") + right

	return StatusBarStyle.
		Width(m.width).
		Render(statusContent)
}

// renderHelp renders the help overlay
func (m Model) renderHelp() string {
	help := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorAccent).
		Padding(1, 2).
		Width(60).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				HighlightStyle.Render("Anvil Keyboard Shortcuts"),
				"",
				"Navigation:",
				"  Tab / Shift+Tab - Switch between panels",
				"  1-4             - Jump to specific panel",
				"",
				"Panels:",
				"  1 - Conversation",
				"  2 - Files (j/k to navigate, r to refresh)",
				"  3 - Diff",
				"  4 - Plan",
				"",
				"Input:",
				"  i or / - Focus input field",
				"  Enter  - Send message",
				"  Esc    - Exit input mode",
				"",
				"General:",
				"  ?      - Toggle this help",
				"  q      - Quit",
				"  Ctrl+C - Quit",
				"",
				MutedStyle.Render("Press ? to close this help"),
			),
		)

	// Center the help dialog
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		help,
	)
}

// updateLayout updates the layout when terminal size changes
func (m *Model) updateLayout() {
	headerHeight := 1
	statusBarHeight := 1
	contentHeight := m.height - headerHeight - statusBarHeight - 2

	panelWidth := m.width / 2
	panelHeight := contentHeight / 2

	for _, panel := range m.panelManager.GetPanels() {
		panel.SetSize(panelWidth-4, panelHeight-4) // Account for borders
	}
}

// NewModel creates a new application model
func NewModel() Model {
	// Get current working directory for file panel
	cwd, _ := os.Getwd()

	// Create panel manager
	pm := NewPanelManager()

	// Create and add panels
	convPanel := panels.NewConversationPanel()
	filesPanel := panels.NewFilesPanel(cwd)
	diffPanel := panels.NewDiffPanel()
	planPanel := panels.NewPlanPanel()

	pm.AddPanel(convPanel)
	pm.AddPanel(filesPanel)
	pm.AddPanel(diffPanel)
	pm.AddPanel(planPanel)

	// Add some demo content
	convPanel.AddMessage("user", "Hello! Can you help me with my code?")
	convPanel.AddMessage("assistant", "Of course! I'd be happy to help. What would you like to work on?\n\n```go\nfunc main() {\n    fmt.Println(\"Hello, Anvil!\")\n}\n```")

	planPanel.AddStep("Understand the user's request")
	planPanel.AddStep("Analyze the codebase")
	planPanel.AddStep("Generate a plan")
	planPanel.AddStep("Execute the changes")
	planPanel.UpdateStep(1, panels.StepCompleted, "")
	planPanel.UpdateStep(2, panels.StepInProgress, "Scanning files...")

	diffPanel.SetDiff(`diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,5 +1,6 @@
 package main

+import "fmt"
+
 func main() {
-    println("old")
+    fmt.Println("new")
 }`, "main.go")

	// Initialize text input
	ti := textinput.New()
	ti.Placeholder = "Type your message here..."
	ti.CharLimit = 500
	ti.Width = 80

	// Initialize token tracker
	tokenTracker := llm.NewTokenTracker()

	return Model{
		panelManager:  pm,
		input:         ti,
		tokenTracker:  tokenTracker,
		streamBuffer:  "",
		streaming:     false,
	}
}

// NewModelWithConfig creates a new application model with config
func NewModelWithConfig(configMgr *config.Manager) (Model, error) {
	m := NewModel()
	m.configManager = configMgr

	// Initialize LLM client
	cfg := configMgr.GetConfig()

	// Try to get API key for the configured provider
	apiKey, err := configMgr.GetAPIKey(cfg.Provider)
	if err != nil {
		// No API key configured - app will work but can't send messages
		// We'll show an error when the user tries to send a message
		return m, nil
	}

	// Create LLM client
	llmConfig := llm.ClientConfig{
		Provider:    llm.ProviderType(cfg.Provider),
		APIKey:      apiKey,
		Model:       cfg.Model,
		MaxTokens:   cfg.MaxTokens,
		Temperature: cfg.Temperature,
		MaxRetries:  3,
	}

	client, err := llm.NewClient(llmConfig)
	if err != nil {
		return m, err
	}

	// Wrap with retry logic
	m.llmClient = llm.NewRetryableClient(client, llm.DefaultRetryConfig())

	return m, nil
}

// Run starts the TUI application
func Run() error {
	return RunWithConfig(nil)
}

// RunWithConfig starts the TUI application with a config manager
func RunWithConfig(configMgr *config.Manager) error {
	var m Model
	var err error

	if configMgr != nil {
		m, err = NewModelWithConfig(configMgr)
		if err != nil {
			return fmt.Errorf("failed to initialize model: %w", err)
		}
	} else {
		m = NewModel()
	}

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err = p.Run()
	return err
}
