package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ErrorSeverity represents the severity of an error
type ErrorSeverity int

const (
	SeverityInfo ErrorSeverity = iota
	SeverityWarning
	SeverityError
	SeverityCritical
)

// String returns the string representation of severity
func (s ErrorSeverity) String() string {
	switch s {
	case SeverityInfo:
		return "info"
	case SeverityWarning:
		return "warning"
	case SeverityError:
		return "error"
	case SeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// Icon returns the icon for severity
func (s ErrorSeverity) Icon() string {
	switch s {
	case SeverityInfo:
		return "ℹ"
	case SeverityWarning:
		return "⚠"
	case SeverityError:
		return "✗"
	case SeverityCritical:
		return "☠"
	default:
		return "?"
	}
}

// Color returns the color for severity
func (s ErrorSeverity) Color() lipgloss.Color {
	switch s {
	case SeverityInfo:
		return lipgloss.Color("86")
	case SeverityWarning:
		return lipgloss.Color("214")
	case SeverityError:
		return lipgloss.Color("196")
	case SeverityCritical:
		return lipgloss.Color("196")
	default:
		return lipgloss.Color("252")
	}
}

// ErrorDisplay represents an error message display
type ErrorDisplay struct {
	severity    ErrorSeverity
	title       string
	message     string
	details     string
	suggestions []string
	width       int
}

// NewErrorDisplay creates a new error display
func NewErrorDisplay(severity ErrorSeverity, title, message string) *ErrorDisplay {
	return &ErrorDisplay{
		severity:    severity,
		title:       title,
		message:     message,
		suggestions: make([]string, 0),
		width:       60,
	}
}

// SetDetails sets additional details
func (e *ErrorDisplay) SetDetails(details string) {
	e.details = details
}

// AddSuggestion adds a suggestion for resolving the error
func (e *ErrorDisplay) AddSuggestion(suggestion string) {
	e.suggestions = append(e.suggestions, suggestion)
}

// SetWidth sets the display width
func (e *ErrorDisplay) SetWidth(width int) {
	e.width = width
}

// View renders the error display
func (e *ErrorDisplay) View() string {
	var content strings.Builder

	// Header with icon and title
	headerStyle := lipgloss.NewStyle().
		Foreground(e.severity.Color()).
		Bold(true)

	content.WriteString(headerStyle.Render(e.severity.Icon() + " " + e.title))
	content.WriteString("\n\n")

	// Message
	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Width(e.width - 4)

	content.WriteString(messageStyle.Render(e.message))
	content.WriteString("\n")

	// Details if present
	if e.details != "" {
		content.WriteString("\n")
		detailsStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true).
			Width(e.width - 4)
		content.WriteString(detailsStyle.Render(e.details))
		content.WriteString("\n")
	}

	// Suggestions if present
	if len(e.suggestions) > 0 {
		content.WriteString("\n")
		suggestionHeader := lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)
		content.WriteString(suggestionHeader.Render("Suggestions:"))
		content.WriteString("\n")

		for _, suggestion := range e.suggestions {
			content.WriteString("  • ")
			content.WriteString(suggestion)
			content.WriteString("\n")
		}
	}

	// Wrap in a border
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(e.severity.Color()).
		Padding(1, 2).
		Width(e.width)

	return borderStyle.Render(content.String())
}

// QuickError creates a simple error display
func QuickError(message string) string {
	return NewErrorDisplay(SeverityError, "Error", message).View()
}

// QuickWarning creates a simple warning display
func QuickWarning(message string) string {
	return NewErrorDisplay(SeverityWarning, "Warning", message).View()
}

// QuickInfo creates a simple info display
func QuickInfo(message string) string {
	return NewErrorDisplay(SeverityInfo, "Info", message).View()
}

// APIError creates an error display for API errors
func APIError(provider string, statusCode int, message string) *ErrorDisplay {
	e := NewErrorDisplay(SeverityError, "API Error", message)
	e.SetDetails(fmt.Sprintf("Provider: %s | Status: %d", provider, statusCode))

	switch statusCode {
	case 401:
		e.AddSuggestion("Check that your API key is correct")
		e.AddSuggestion("Verify the API key has not expired")
	case 403:
		e.AddSuggestion("Check that your API key has the required permissions")
	case 429:
		e.AddSuggestion("Wait a moment and try again (rate limit)")
		e.AddSuggestion("Consider upgrading your API plan")
	case 500, 502, 503:
		e.AddSuggestion("The API service may be experiencing issues")
		e.AddSuggestion("Try again in a few moments")
	}

	return e
}

// FileError creates an error display for file operations
func FileError(operation, path string, err error) *ErrorDisplay {
	e := NewErrorDisplay(SeverityError, "File Error", fmt.Sprintf("Failed to %s: %s", operation, path))
	e.SetDetails(err.Error())

	if strings.Contains(err.Error(), "permission denied") {
		e.AddSuggestion("Check file permissions")
		e.AddSuggestion("Try running with appropriate access rights")
	} else if strings.Contains(err.Error(), "no such file") {
		e.AddSuggestion("Verify the file path is correct")
		e.AddSuggestion("Check if the file was moved or deleted")
	}

	return e
}

// ValidationError creates an error display for validation errors
func ValidationError(field string, issue string) *ErrorDisplay {
	e := NewErrorDisplay(SeverityWarning, "Validation Error", fmt.Sprintf("Invalid %s: %s", field, issue))
	return e
}

// ConnectionError creates an error display for connection errors
func ConnectionError(target string, err error) *ErrorDisplay {
	e := NewErrorDisplay(SeverityError, "Connection Error", fmt.Sprintf("Failed to connect to %s", target))
	e.SetDetails(err.Error())
	e.AddSuggestion("Check your internet connection")
	e.AddSuggestion("Verify the service URL is correct")
	e.AddSuggestion("Check if a firewall is blocking the connection")
	return e
}

// ConfigError creates an error display for configuration errors
func ConfigError(setting string, issue string) *ErrorDisplay {
	e := NewErrorDisplay(SeverityError, "Configuration Error", fmt.Sprintf("Invalid configuration for %s", setting))
	e.SetDetails(issue)
	e.AddSuggestion("Check your configuration file")
	e.AddSuggestion("Run 'anvil config --help' for available options")
	return e
}
