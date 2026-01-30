package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModel(t *testing.T) {
	m := NewModel()

	if m.ready {
		t.Error("Model should not be ready on initialization")
	}

	if m.width != 0 || m.height != 0 {
		t.Error("Width and height should be 0 on initialization")
	}
}

func TestModelInit(t *testing.T) {
	m := NewModel()
	cmd := m.Init()

	// Init returns a batch of commands for panel initialization and input blinking
	// So cmd should not be nil in the current implementation
	if cmd == nil {
		t.Error("Init should return commands for panel initialization")
	}
}

func TestModelUpdate_WindowSize(t *testing.T) {
	m := NewModel()

	msg := tea.WindowSizeMsg{
		Width:  100,
		Height: 50,
	}

	updated, _ := m.Update(msg)
	updatedModel := updated.(Model)

	if !updatedModel.ready {
		t.Error("Model should be ready after WindowSizeMsg")
	}

	if updatedModel.width != 100 {
		t.Errorf("Width = %d, want 100", updatedModel.width)
	}

	if updatedModel.height != 50 {
		t.Errorf("Height = %d, want 50", updatedModel.height)
	}
}

func TestModelUpdate_Quit(t *testing.T) {
	m := NewModel()
	m.ready = true

	tests := []struct {
		name string
		key  string
	}{
		{"quit with q", "q"},
		{"quit with ctrl+c", "ctrl+c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tea.KeyMsg{Type: tea.KeyRunes}
			if tt.key == "q" {
				msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
			} else if tt.key == "ctrl+c" {
				msg = tea.KeyMsg{Type: tea.KeyCtrlC}
			}

			_, cmd := m.Update(msg)
			if cmd == nil {
				t.Error("Quit keys should return quit command")
			}
		})
	}
}

func TestModelView_NotReady(t *testing.T) {
	m := NewModel()
	view := m.View()

	if !strings.Contains(view, "Initializing") {
		t.Error("View should show 'Initializing' when not ready")
	}
}

func TestModelView_Ready(t *testing.T) {
	m := NewModel()
	m.ready = true
	m.width = 100
	m.height = 50

	// Need to update layout after setting dimensions
	m.updateLayout()

	view := m.View()

	if !strings.Contains(view, "Anvil") {
		t.Error("View should contain 'Anvil'")
	}

	// Check for panel titles that should be visible
	if !strings.Contains(view, "Conversation") {
		t.Error("View should contain 'Conversation' panel")
	}
}
