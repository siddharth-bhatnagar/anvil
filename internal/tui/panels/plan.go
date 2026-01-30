package panels

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// StepStatus represents the status of a plan step
type StepStatus int

const (
	StepPending StepStatus = iota
	StepInProgress
	StepCompleted
	StepFailed
)

// String returns the string representation of the step status
func (s StepStatus) String() string {
	switch s {
	case StepPending:
		return "pending"
	case StepInProgress:
		return "in progress"
	case StepCompleted:
		return "completed"
	case StepFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// Step represents a single step in a plan
type Step struct {
	ID          int
	Description string
	Status      StepStatus
	Details     string
}

// PlanPanel displays a plan with steps
type PlanPanel struct {
	width    int
	height   int
	focused  bool
	viewport viewport.Model
	steps    []Step
	ready    bool
}

// NewPlanPanel creates a new plan panel
func NewPlanPanel() *PlanPanel {
	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	return &PlanPanel{
		viewport: vp,
		steps:    make([]Step, 0),
	}
}

// Init initializes the panel
func (p *PlanPanel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (p *PlanPanel) Update(msg tea.Msg) (*PlanPanel, tea.Cmd) {
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
func (p *PlanPanel) View() string {
	if !p.ready {
		return "Loading..."
	}

	if len(p.steps) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("No plan yet")
	}

	var content strings.Builder

	for i, step := range p.steps {
		// Icon and color based on status
		var icon string
		var style lipgloss.Style

		switch step.Status {
		case StepPending:
			icon = "○"
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		case StepInProgress:
			icon = "◐"
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")).
				Bold(true)
		case StepCompleted:
			icon = "●"
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
		case StepFailed:
			icon = "✗"
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Bold(true)
		}

		// Format the step
		line := fmt.Sprintf("%s %d. %s", icon, step.ID, step.Description)
		content.WriteString(style.Render(line))

		// Add details if present
		if step.Details != "" {
			details := lipgloss.NewStyle().
				Foreground(lipgloss.Color("244")).
				Render(fmt.Sprintf("   %s", step.Details))
			content.WriteString("\n")
			content.WriteString(details)
		}

		if i < len(p.steps)-1 {
			content.WriteString("\n")
		}
	}

	p.viewport.SetContent(content.String())
	return p.viewport.View()
}

// SetSize sets the panel dimensions
func (p *PlanPanel) SetSize(width, height int) {
	p.width = width
	p.height = height
	p.viewport.Width = width - 2
	p.viewport.Height = height - 2
	p.ready = true
}

// Focus sets the panel as focused
func (p *PlanPanel) Focus() {
	p.focused = true
}

// Blur removes focus from the panel
func (p *PlanPanel) Blur() {
	p.focused = false
}

// IsFocused returns whether the panel is focused
func (p *PlanPanel) IsFocused() bool {
	return p.focused
}

// Type returns the panel type
func (p *PlanPanel) Type() PanelType {
	return PanelPlan
}

// Title returns the panel title
func (p *PlanPanel) Title() string {
	return "Plan"
}

// AddStep adds a step to the plan
func (p *PlanPanel) AddStep(description string) {
	p.steps = append(p.steps, Step{
		ID:          len(p.steps) + 1,
		Description: description,
		Status:      StepPending,
	})
}

// UpdateStep updates a step's status and details
func (p *PlanPanel) UpdateStep(id int, status StepStatus, details string) {
	for i := range p.steps {
		if p.steps[i].ID == id {
			p.steps[i].Status = status
			p.steps[i].Details = details
			return
		}
	}
}

// ClearSteps clears all steps
func (p *PlanPanel) ClearSteps() {
	p.steps = make([]Step, 0)
}

// GetSteps returns all steps
func (p *PlanPanel) GetSteps() []Step {
	return p.steps
}

// GetProgress returns the current progress (completed / total)
func (p *PlanPanel) GetProgress() (completed, total int) {
	total = len(p.steps)
	for _, step := range p.steps {
		if step.Status == StepCompleted {
			completed++
		}
	}
	return
}

// UpdatePanel is a wrapper that implements the Panel interface
func (p *PlanPanel) UpdatePanel(msg tea.Msg) tea.Cmd {
	_, cmd := p.Update(msg)
	return cmd
}
