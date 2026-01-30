package panels

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DiffPanel displays file diffs
type DiffPanel struct {
	width    int
	height   int
	focused  bool
	viewport viewport.Model
	diff     string
	filePath string
	ready    bool
}

// NewDiffPanel creates a new diff panel
func NewDiffPanel() *DiffPanel {
	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	return &DiffPanel{
		viewport: vp,
	}
}

// Init initializes the panel
func (p *DiffPanel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (p *DiffPanel) Update(msg tea.Msg) (*DiffPanel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if p.IsFocused() {
			p.viewport, cmd = p.viewport.Update(msg)
		}
	case tea.MouseMsg:
		if p.IsFocused() {
			p.viewport, cmd = p.viewport.Update(msg)
		}
	}

	return p, cmd
}

// View renders the panel
func (p *DiffPanel) View() string {
	if !p.ready {
		return "Loading..."
	}

	if p.diff == "" {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("No diff to display")
	}

	// Render diff with syntax highlighting
	var content strings.Builder

	if p.filePath != "" {
		header := lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true).
			Render("File: " + p.filePath)
		content.WriteString(header)
		content.WriteString("\n\n")
	}

	// Split diff into lines and color them
	lines := strings.Split(p.diff, "\n")
	for i, line := range lines {
		var style lipgloss.Style

		if strings.HasPrefix(line, "+") {
			// Added lines in green
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
		} else if strings.HasPrefix(line, "-") {
			// Removed lines in red
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		} else if strings.HasPrefix(line, "@@") {
			// Hunk headers in cyan
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Bold(true)
		} else if strings.HasPrefix(line, "diff") || strings.HasPrefix(line, "index") {
			// Diff headers in yellow
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
		} else {
			// Context lines in default color
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
		}

		content.WriteString(style.Render(line))
		if i < len(lines)-1 {
			content.WriteString("\n")
		}
	}

	p.viewport.SetContent(content.String())
	return p.viewport.View()
}

// SetSize sets the panel dimensions
func (p *DiffPanel) SetSize(width, height int) {
	p.width = width
	p.height = height
	p.viewport.Width = width - 2
	p.viewport.Height = height - 2
	p.ready = true
}

// Focus sets the panel as focused
func (p *DiffPanel) Focus() {
	p.focused = true
}

// Blur removes focus from the panel
func (p *DiffPanel) Blur() {
	p.focused = false
}

// IsFocused returns whether the panel is focused
func (p *DiffPanel) IsFocused() bool {
	return p.focused
}

// Type returns the panel type
func (p *DiffPanel) Type() PanelType {
	return PanelDiff
}

// Title returns the panel title
func (p *DiffPanel) Title() string {
	return "Diff"
}

// SetDiff sets the diff content
func (p *DiffPanel) SetDiff(diff, filePath string) {
	p.diff = diff
	p.filePath = filePath
	p.viewport.GotoTop()
}

// ClearDiff clears the diff content
func (p *DiffPanel) ClearDiff() {
	p.diff = ""
	p.filePath = ""
}

// GetDiff returns the current diff
func (p *DiffPanel) GetDiff() string {
	return p.diff
}

// UpdatePanel is a wrapper that implements the Panel interface
func (p *DiffPanel) UpdatePanel(msg tea.Msg) tea.Cmd {
	_, cmd := p.Update(msg)
	return cmd
}
