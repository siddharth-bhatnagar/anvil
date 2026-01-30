package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// StatusBarItem represents an item in the status bar
type StatusBarItem struct {
	Label   string
	Value   string
	Icon    string
	Style   lipgloss.Style
	Visible bool
}

// StatusBar represents an enhanced status bar
type StatusBar struct {
	width      int
	leftItems  []StatusBarItem
	rightItems []StatusBarItem
	style      lipgloss.Style
	separator  string
}

// NewStatusBar creates a new status bar
func NewStatusBar(width int) *StatusBar {
	return &StatusBar{
		width:      width,
		leftItems:  make([]StatusBarItem, 0),
		rightItems: make([]StatusBarItem, 0),
		style: lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 1),
		separator: " │ ",
	}
}

// SetWidth sets the status bar width
func (sb *StatusBar) SetWidth(width int) {
	sb.width = width
}

// AddLeftItem adds an item to the left side
func (sb *StatusBar) AddLeftItem(item StatusBarItem) {
	if item.Style.Value() == "" {
		item.Style = lipgloss.NewStyle()
	}
	item.Visible = true
	sb.leftItems = append(sb.leftItems, item)
}

// AddRightItem adds an item to the right side
func (sb *StatusBar) AddRightItem(item StatusBarItem) {
	if item.Style.Value() == "" {
		item.Style = lipgloss.NewStyle()
	}
	item.Visible = true
	sb.rightItems = append(sb.rightItems, item)
}

// SetLeftItem sets a left item by label
func (sb *StatusBar) SetLeftItem(label, value string) {
	for i := range sb.leftItems {
		if sb.leftItems[i].Label == label {
			sb.leftItems[i].Value = value
			return
		}
	}
	// Add if not found
	sb.AddLeftItem(StatusBarItem{Label: label, Value: value})
}

// SetRightItem sets a right item by label
func (sb *StatusBar) SetRightItem(label, value string) {
	for i := range sb.rightItems {
		if sb.rightItems[i].Label == label {
			sb.rightItems[i].Value = value
			return
		}
	}
	// Add if not found
	sb.AddRightItem(StatusBarItem{Label: label, Value: value})
}

// HideItem hides an item by label
func (sb *StatusBar) HideItem(label string) {
	for i := range sb.leftItems {
		if sb.leftItems[i].Label == label {
			sb.leftItems[i].Visible = false
			return
		}
	}
	for i := range sb.rightItems {
		if sb.rightItems[i].Label == label {
			sb.rightItems[i].Visible = false
			return
		}
	}
}

// ShowItem shows an item by label
func (sb *StatusBar) ShowItem(label string) {
	for i := range sb.leftItems {
		if sb.leftItems[i].Label == label {
			sb.leftItems[i].Visible = true
			return
		}
	}
	for i := range sb.rightItems {
		if sb.rightItems[i].Label == label {
			sb.rightItems[i].Visible = true
			return
		}
	}
}

// ClearLeft clears left items
func (sb *StatusBar) ClearLeft() {
	sb.leftItems = make([]StatusBarItem, 0)
}

// ClearRight clears right items
func (sb *StatusBar) ClearRight() {
	sb.rightItems = make([]StatusBarItem, 0)
}

// View renders the status bar
func (sb *StatusBar) View() string {
	// Render left items
	var leftParts []string
	for _, item := range sb.leftItems {
		if !item.Visible {
			continue
		}
		part := ""
		if item.Icon != "" {
			part += item.Icon + " "
		}
		if item.Label != "" && item.Value != "" {
			part += item.Label + ": " + item.Style.Render(item.Value)
		} else if item.Value != "" {
			part += item.Style.Render(item.Value)
		} else if item.Label != "" {
			part += item.Style.Render(item.Label)
		}
		if part != "" {
			leftParts = append(leftParts, part)
		}
	}

	// Render right items
	var rightParts []string
	for _, item := range sb.rightItems {
		if !item.Visible {
			continue
		}
		part := ""
		if item.Icon != "" {
			part += item.Icon + " "
		}
		if item.Label != "" && item.Value != "" {
			part += item.Label + ": " + item.Style.Render(item.Value)
		} else if item.Value != "" {
			part += item.Style.Render(item.Value)
		} else if item.Label != "" {
			part += item.Style.Render(item.Label)
		}
		if part != "" {
			rightParts = append(rightParts, part)
		}
	}

	left := strings.Join(leftParts, sb.separator)
	right := strings.Join(rightParts, sb.separator)

	// Calculate spacing
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	spacing := sb.width - leftWidth - rightWidth - 2 // -2 for padding

	if spacing < 1 {
		spacing = 1
	}

	// Build the status bar
	content := left + strings.Repeat(" ", spacing) + right

	return sb.style.Width(sb.width).Render(content)
}

// DefaultStatusBar creates a status bar with common items
func DefaultStatusBar(width int) *StatusBar {
	sb := NewStatusBar(width)

	// Left side - mode and state info
	sb.AddLeftItem(StatusBarItem{
		Label: "Mode",
		Value: "Normal",
		Icon:  "⚡",
		Style: lipgloss.NewStyle().Foreground(lipgloss.Color("86")),
	})

	sb.AddLeftItem(StatusBarItem{
		Label: "Panel",
		Value: "Conversation",
	})

	// Right side - shortcuts and info
	sb.AddRightItem(StatusBarItem{
		Label: "",
		Value: "? Help",
		Style: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
	})

	sb.AddRightItem(StatusBarItem{
		Label: "",
		Value: "q Quit",
		Style: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
	})

	return sb
}

// KeyHint represents a keyboard shortcut hint
type KeyHint struct {
	Key         string
	Description string
}

// KeyHints formats a list of key hints for display
func FormatKeyHints(hints []KeyHint) string {
	var parts []string

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	for _, hint := range hints {
		parts = append(parts, keyStyle.Render(hint.Key)+" "+descStyle.Render(hint.Description))
	}

	return strings.Join(parts, "  ")
}

// ModeIndicator represents the current mode
type ModeIndicator struct {
	mode  string
	style lipgloss.Style
}

// NewModeIndicator creates a new mode indicator
func NewModeIndicator(mode string) *ModeIndicator {
	return &ModeIndicator{
		mode:  mode,
		style: lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true),
	}
}

// SetMode sets the mode
func (m *ModeIndicator) SetMode(mode string) {
	m.mode = mode
}

// SetStyle sets the style based on mode
func (m *ModeIndicator) SetModeStyle(mode string, dangerous bool) {
	m.mode = mode
	if dangerous {
		m.style = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	} else {
		m.style = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
	}
}

// View renders the mode indicator
func (m *ModeIndicator) View() string {
	return m.style.Render(fmt.Sprintf("[%s]", m.mode))
}
