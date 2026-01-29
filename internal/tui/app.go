package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const version = "0.1.0"

// Model represents the application state
type Model struct {
	width  int
	height int
	ready  bool
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the UI
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Header
	header := TitleStyle.Render(fmt.Sprintf("⚒  Anvil v%s", version))

	// Content
	content := BorderStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			"Welcome to Anvil!",
			"",
			MutedStyle.Render("A terminal-based agentic coding CLI"),
			"",
			HighlightStyle.Render("Press 'q' or 'Ctrl+C' to quit"),
		),
	)

	// Status bar
	statusBar := StatusBarStyle.
		Width(m.width).
		Render(fmt.Sprintf("Ready | Terminal: %dx%d", m.width, m.height))

	// Combine all parts
	mainContent := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		content,
	)

	// Calculate available height for content
	contentHeight := m.height - lipgloss.Height(statusBar) - 1

	// Vertically position content
	availableHeight := max(0, contentHeight-lipgloss.Height(mainContent))

	view := lipgloss.JoinVertical(
		lipgloss.Left,
		mainContent,
		lipgloss.NewStyle().Height(availableHeight).Render(""),
		statusBar,
	)

	return view
}

// NewModel creates a new application model
func NewModel() Model {
	return Model{}
}

// Run starts the TUI application
func Run() error {
	p := tea.NewProgram(
		NewModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err := p.Run()
	return err
}
