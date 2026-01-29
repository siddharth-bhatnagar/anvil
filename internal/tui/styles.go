package tui

import "github.com/charmbracelet/lipgloss"

// Theme colors
var (
	ColorPrimary   = lipgloss.Color("#7C3AED") // Purple
	ColorSecondary = lipgloss.Color("#A78BFA") // Light purple
	ColorAccent    = lipgloss.Color("#10B981") // Green
	ColorError     = lipgloss.Color("#EF4444") // Red
	ColorWarning   = lipgloss.Color("#F59E0B") // Amber
	ColorMuted     = lipgloss.Color("#6B7280") // Gray
	ColorText      = lipgloss.Color("#F9FAFB") // Off-white
	ColorBorder    = lipgloss.Color("#374151") // Dark gray
)

// Base styles
var (
	// StatusBarStyle is the style for the status bar at the bottom
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorPrimary).
			Padding(0, 1)

	// TitleStyle is the style for panel titles
	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true).
			Padding(0, 1)

	// BorderStyle is the style for panel borders
	BorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(1, 2)

	// ErrorStyle is the style for error messages
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	// MutedStyle is the style for muted/secondary text
	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// HighlightStyle is the style for highlighted/selected items
	HighlightStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)
)
