package panels

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// Message represents a conversation message
type Message struct {
	Role    string // "user" or "assistant"
	Content string
}

// ConversationPanel displays the conversation history
type ConversationPanel struct {
	width    int
	height   int
	focused  bool
	viewport viewport.Model
	messages []Message
	renderer *glamour.TermRenderer
	ready    bool
}

// NewConversationPanel creates a new conversation panel
func NewConversationPanel() *ConversationPanel {
	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	// Create glamour renderer for markdown
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)

	return &ConversationPanel{
		viewport: vp,
		messages: make([]Message, 0),
		renderer: renderer,
	}
}

// Init initializes the panel
func (p *ConversationPanel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (p *ConversationPanel) Update(msg tea.Msg) (*ConversationPanel, tea.Cmd) {
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
func (p *ConversationPanel) View() string {
	if !p.ready {
		return "Loading..."
	}

	// Build the conversation view
	var content strings.Builder

	if len(p.messages) == 0 {
		content.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("No messages yet. Start a conversation!"))
	} else {
		for i, msg := range p.messages {
			// Render message with styling
			var style lipgloss.Style
			var prefix string

			if msg.Role == "user" {
				style = lipgloss.NewStyle().
					Foreground(lipgloss.Color("86")). // Cyan
					Bold(true)
				prefix = "You: "
			} else {
				style = lipgloss.NewStyle().
					Foreground(lipgloss.Color("141")) // Purple
				prefix = "Assistant: "
			}

			// Render markdown content
			rendered, err := p.renderer.Render(msg.Content)
			if err != nil {
				rendered = msg.Content
			}

			content.WriteString(style.Render(prefix))
			content.WriteString("\n")
			content.WriteString(rendered)

			if i < len(p.messages)-1 {
				content.WriteString("\n\n")
			}
		}
	}

	p.viewport.SetContent(content.String())
	return p.viewport.View()
}

// SetSize sets the panel dimensions
func (p *ConversationPanel) SetSize(width, height int) {
	p.width = width
	p.height = height
	p.viewport.Width = width - 2  // Account for borders
	p.viewport.Height = height - 2
	p.ready = true
}

// Focus sets the panel as focused
func (p *ConversationPanel) Focus() {
	p.focused = true
}

// Blur removes focus from the panel
func (p *ConversationPanel) Blur() {
	p.focused = false
}

// IsFocused returns whether the panel is focused
func (p *ConversationPanel) IsFocused() bool {
	return p.focused
}

// Type returns the panel type
func (p *ConversationPanel) Type() PanelType {
	return PanelConversation
}

// Title returns the panel title
func (p *ConversationPanel) Title() string {
	return "Conversation"
}

// AddMessage adds a message to the conversation
func (p *ConversationPanel) AddMessage(role, content string) {
	p.messages = append(p.messages, Message{
		Role:    role,
		Content: content,
	})

	// Scroll to bottom
	p.viewport.GotoBottom()
}

// ClearMessages clears all messages
func (p *ConversationPanel) ClearMessages() {
	p.messages = make([]Message, 0)
}

// GetMessages returns all messages
func (p *ConversationPanel) GetMessages() []Message {
	return p.messages
}

// UpdatePanel is a wrapper that implements the Panel interface
func (p *ConversationPanel) UpdatePanel(msg tea.Msg) tea.Cmd {
	_, cmd := p.Update(msg)
	return cmd
}
