package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// PanelManager manages multiple panels and handles focus
type PanelManager struct {
	panels      []Panel
	activeIndex int
}

// NewPanelManager creates a new panel manager
func NewPanelManager() *PanelManager {
	return &PanelManager{
		panels:      make([]Panel, 0),
		activeIndex: 0,
	}
}

// AddPanel adds a panel to the manager
func (pm *PanelManager) AddPanel(panel Panel) {
	pm.panels = append(pm.panels, panel)
}

// Init initializes all panels
func (pm *PanelManager) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, panel := range pm.panels {
		cmds = append(cmds, panel.Init())
	}

	// Focus the first panel
	if len(pm.panels) > 0 {
		pm.panels[pm.activeIndex].Focus()
	}

	return tea.Batch(cmds...)
}

// Update updates all panels
func (pm *PanelManager) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	for _, panel := range pm.panels {
		cmd := panel.UpdatePanel(msg)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

// NextPanel moves focus to the next panel
func (pm *PanelManager) NextPanel() {
	if len(pm.panels) == 0 {
		return
	}

	pm.panels[pm.activeIndex].Blur()
	pm.activeIndex = (pm.activeIndex + 1) % len(pm.panels)
	pm.panels[pm.activeIndex].Focus()
}

// PrevPanel moves focus to the previous panel
func (pm *PanelManager) PrevPanel() {
	if len(pm.panels) == 0 {
		return
	}

	pm.panels[pm.activeIndex].Blur()
	pm.activeIndex--
	if pm.activeIndex < 0 {
		pm.activeIndex = len(pm.panels) - 1
	}
	pm.panels[pm.activeIndex].Focus()
}

// SetActivePanel sets the active panel by index
func (pm *PanelManager) SetActivePanel(index int) {
	if index < 0 || index >= len(pm.panels) {
		return
	}

	pm.panels[pm.activeIndex].Blur()
	pm.activeIndex = index
	pm.panels[pm.activeIndex].Focus()
}

// SetActivePanelByType sets the active panel by type
func (pm *PanelManager) SetActivePanelByType(pType PanelType) {
	for i, panel := range pm.panels {
		if panel.Type() == pType {
			pm.SetActivePanel(i)
			return
		}
	}
}

// GetActivePanel returns the currently active panel
func (pm *PanelManager) GetActivePanel() Panel {
	if len(pm.panels) == 0 {
		return nil
	}
	return pm.panels[pm.activeIndex]
}

// GetPanel returns a panel by index
func (pm *PanelManager) GetPanel(index int) Panel {
	if index < 0 || index >= len(pm.panels) {
		return nil
	}
	return pm.panels[index]
}

// GetPanelByType returns a panel by type
func (pm *PanelManager) GetPanelByType(pType PanelType) Panel {
	for _, panel := range pm.panels {
		if panel.Type() == pType {
			return panel
		}
	}
	return nil
}

// GetPanels returns all panels
func (pm *PanelManager) GetPanels() []Panel {
	return pm.panels
}

// ActiveIndex returns the index of the active panel
func (pm *PanelManager) ActiveIndex() int {
	return pm.activeIndex
}

// Count returns the number of panels
func (pm *PanelManager) Count() int {
	return len(pm.panels)
}

// SetSize sets the size for all panels
func (pm *PanelManager) SetSize(width, height int) {
	for _, panel := range pm.panels {
		panel.SetSize(width, height)
	}
}
