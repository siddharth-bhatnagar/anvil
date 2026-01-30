package components

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SpinnerStyle represents different spinner styles
type SpinnerStyle int

const (
	SpinnerDots SpinnerStyle = iota
	SpinnerLine
	SpinnerCircle
	SpinnerBounce
	SpinnerPulse
)

// SpinnerFrames returns the frames for each spinner style
func SpinnerFrames(style SpinnerStyle) []string {
	switch style {
	case SpinnerDots:
		return []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	case SpinnerLine:
		return []string{"|", "/", "-", "\\"}
	case SpinnerCircle:
		return []string{"◐", "◓", "◑", "◒"}
	case SpinnerBounce:
		return []string{"⠁", "⠂", "⠄", "⠂"}
	case SpinnerPulse:
		return []string{"█", "▓", "▒", "░", "▒", "▓"}
	default:
		return []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	}
}

// SpinnerTickMsg is sent when the spinner should update
type SpinnerTickMsg time.Time

// Spinner represents an animated loading spinner
type Spinner struct {
	frames   []string
	index    int
	style    lipgloss.Style
	interval time.Duration
	active   bool
	message  string
}

// NewSpinner creates a new spinner
func NewSpinner(style SpinnerStyle) *Spinner {
	return &Spinner{
		frames:   SpinnerFrames(style),
		index:    0,
		style:    lipgloss.NewStyle().Foreground(lipgloss.Color("86")),
		interval: 80 * time.Millisecond,
		active:   false,
	}
}

// NewSpinnerWithMessage creates a spinner with a message
func NewSpinnerWithMessage(style SpinnerStyle, message string) *Spinner {
	s := NewSpinner(style)
	s.message = message
	return s
}

// Start starts the spinner animation
func (s *Spinner) Start() tea.Cmd {
	s.active = true
	return s.tick()
}

// Stop stops the spinner animation
func (s *Spinner) Stop() {
	s.active = false
}

// IsActive returns whether the spinner is active
func (s *Spinner) IsActive() bool {
	return s.active
}

// SetMessage sets the spinner message
func (s *Spinner) SetMessage(message string) {
	s.message = message
}

// SetStyle sets the lipgloss style for the spinner
func (s *Spinner) SetStyle(style lipgloss.Style) {
	s.style = style
}

// Update handles spinner tick messages
func (s *Spinner) Update(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case SpinnerTickMsg:
		if s.active {
			s.index = (s.index + 1) % len(s.frames)
			return s.tick()
		}
	}
	return nil
}

// View renders the spinner
func (s *Spinner) View() string {
	if !s.active {
		return ""
	}

	frame := s.style.Render(s.frames[s.index])
	if s.message != "" {
		return frame + " " + s.message
	}
	return frame
}

// tick returns a command that sends a tick message after the interval
func (s *Spinner) tick() tea.Cmd {
	return tea.Tick(s.interval, func(t time.Time) tea.Msg {
		return SpinnerTickMsg(t)
	})
}

// ProgressBar represents a progress bar
type ProgressBar struct {
	width     int
	progress  float64 // 0.0 to 1.0
	style     lipgloss.Style
	fillStyle lipgloss.Style
	showPct   bool
}

// NewProgressBar creates a new progress bar
func NewProgressBar(width int) *ProgressBar {
	return &ProgressBar{
		width:     width,
		progress:  0,
		style:     lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		fillStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("82")),
		showPct:   true,
	}
}

// SetProgress sets the progress (0.0 to 1.0)
func (p *ProgressBar) SetProgress(progress float64) {
	if progress < 0 {
		progress = 0
	} else if progress > 1 {
		progress = 1
	}
	p.progress = progress
}

// SetWidth sets the width of the progress bar
func (p *ProgressBar) SetWidth(width int) {
	p.width = width
}

// ShowPercentage sets whether to show the percentage
func (p *ProgressBar) ShowPercentage(show bool) {
	p.showPct = show
}

// View renders the progress bar
func (p *ProgressBar) View() string {
	// Calculate filled width
	fillWidth := int(float64(p.width) * p.progress)
	if fillWidth > p.width {
		fillWidth = p.width
	}

	// Build the bar
	filled := ""
	for i := 0; i < fillWidth; i++ {
		filled += "█"
	}

	empty := ""
	for i := fillWidth; i < p.width; i++ {
		empty += "░"
	}

	bar := p.fillStyle.Render(filled) + p.style.Render(empty)

	if p.showPct {
		pct := int(p.progress * 100)
		return bar + " " + p.style.Render(string(rune('0'+pct/10))+string(rune('0'+pct%10))+"%")
	}

	return bar
}

// StatusIndicator represents a status indicator
type StatusIndicator struct {
	status  Status
	message string
	style   lipgloss.Style
}

// Status represents the status type
type Status int

const (
	StatusIdle Status = iota
	StatusLoading
	StatusSuccess
	StatusWarning
	StatusError
)

// NewStatusIndicator creates a new status indicator
func NewStatusIndicator() *StatusIndicator {
	return &StatusIndicator{
		status: StatusIdle,
		style:  lipgloss.NewStyle(),
	}
}

// SetStatus sets the status and message
func (s *StatusIndicator) SetStatus(status Status, message string) {
	s.status = status
	s.message = message
}

// View renders the status indicator
func (s *StatusIndicator) View() string {
	var icon string
	var style lipgloss.Style

	switch s.status {
	case StatusIdle:
		icon = "○"
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	case StatusLoading:
		icon = "◐"
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	case StatusSuccess:
		icon = "●"
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	case StatusWarning:
		icon = "◐"
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	case StatusError:
		icon = "✗"
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	}

	if s.message != "" {
		return style.Render(icon) + " " + s.message
	}
	return style.Render(icon)
}

// Toast represents a temporary notification message
type Toast struct {
	message   string
	style     lipgloss.Style
	visible   bool
	duration  time.Duration
	createdAt time.Time
}

// ToastExpiredMsg is sent when a toast expires
type ToastExpiredMsg struct{}

// NewToast creates a new toast notification
func NewToast(message string, duration time.Duration) *Toast {
	return &Toast{
		message:   message,
		style:     lipgloss.NewStyle().Background(lipgloss.Color("240")).Padding(0, 1),
		visible:   true,
		duration:  duration,
		createdAt: time.Now(),
	}
}

// NewSuccessToast creates a success toast
func NewSuccessToast(message string) *Toast {
	t := NewToast(message, 3*time.Second)
	t.style = lipgloss.NewStyle().
		Background(lipgloss.Color("22")).
		Foreground(lipgloss.Color("82")).
		Padding(0, 1)
	return t
}

// NewErrorToast creates an error toast
func NewErrorToast(message string) *Toast {
	t := NewToast(message, 5*time.Second)
	t.style = lipgloss.NewStyle().
		Background(lipgloss.Color("52")).
		Foreground(lipgloss.Color("196")).
		Padding(0, 1)
	return t
}

// NewWarningToast creates a warning toast
func NewWarningToast(message string) *Toast {
	t := NewToast(message, 4*time.Second)
	t.style = lipgloss.NewStyle().
		Background(lipgloss.Color("58")).
		Foreground(lipgloss.Color("214")).
		Padding(0, 1)
	return t
}

// Show shows the toast and returns a command to hide it after duration
func (t *Toast) Show() tea.Cmd {
	t.visible = true
	t.createdAt = time.Now()
	return tea.Tick(t.duration, func(time.Time) tea.Msg {
		return ToastExpiredMsg{}
	})
}

// Hide hides the toast
func (t *Toast) Hide() {
	t.visible = false
}

// IsVisible returns whether the toast is visible
func (t *Toast) IsVisible() bool {
	return t.visible
}

// View renders the toast
func (t *Toast) View() string {
	if !t.visible {
		return ""
	}
	return t.style.Render(t.message)
}
