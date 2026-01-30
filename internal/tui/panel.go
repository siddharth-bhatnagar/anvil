package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/siddharth-bhatnagar/anvil/internal/tui/panels"
)

// PanelType is an alias for panels.PanelType
type PanelType = panels.PanelType

// Re-export panel type constants
const (
	PanelConversation = panels.PanelConversation
	PanelFiles        = panels.PanelFiles
	PanelDiff         = panels.PanelDiff
	PanelPlan         = panels.PanelPlan
)

// Panel represents a UI panel that can be displayed
type Panel interface {
	// Init initializes the panel
	Init() tea.Cmd

	// Update handles messages and returns a command
	// The panel updates itself in place
	UpdatePanel(msg tea.Msg) tea.Cmd

	// View renders the panel
	View() string

	// SetSize sets the dimensions of the panel
	SetSize(width, height int)

	// Focus sets whether the panel is focused
	Focus()

	// Blur removes focus from the panel
	Blur()

	// IsFocused returns whether the panel is focused
	IsFocused() bool

	// Type returns the type of panel
	Type() PanelType

	// Title returns the panel title
	Title() string
}

// BasePanelModel provides common functionality for panels
type BasePanelModel struct {
	width   int
	height  int
	focused bool
	pType   PanelType
	title   string
}

// NewBasePanelModel creates a new base panel model
func NewBasePanelModel(pType PanelType, title string) BasePanelModel {
	return BasePanelModel{
		pType: pType,
		title: title,
	}
}

// SetSize sets the panel dimensions
func (m *BasePanelModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Focus sets the panel as focused
func (m *BasePanelModel) Focus() {
	m.focused = true
}

// Blur removes focus from the panel
func (m *BasePanelModel) Blur() {
	m.focused = false
}

// IsFocused returns whether the panel is focused
func (m *BasePanelModel) IsFocused() bool {
	return m.focused
}

// Type returns the panel type
func (m *BasePanelModel) Type() PanelType {
	return m.pType
}

// Title returns the panel title
func (m *BasePanelModel) Title() string {
	return m.title
}

// Width returns the panel width
func (m *BasePanelModel) Width() int {
	return m.width
}

// Height returns the panel height
func (m *BasePanelModel) Height() int {
	return m.height
}
