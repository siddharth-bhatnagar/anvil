package panels

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FileEntry represents a file or directory
type FileEntry struct {
	Name      string
	Path      string
	IsDir     bool
	Level     int
	GitStatus string // "", "M", "A", "D", "?", etc.
}

// FilesPanel displays a file browser
type FilesPanel struct {
	width       int
	height      int
	focused     bool
	viewport    viewport.Model
	files       []FileEntry
	selectedIdx int
	rootPath    string
	ready       bool
}

// NewFilesPanel creates a new files panel
func NewFilesPanel(rootPath string) *FilesPanel {
	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	if rootPath == "" {
		rootPath, _ = os.Getwd()
	}

	panel := &FilesPanel{
		viewport:    vp,
		files:       make([]FileEntry, 0),
		selectedIdx: 0,
		rootPath:    rootPath,
	}

	// Load files
	panel.loadFiles()

	return panel
}

// Init initializes the panel
func (p *FilesPanel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (p *FilesPanel) Update(msg tea.Msg) (*FilesPanel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if p.IsFocused() {
			switch msg.String() {
			case "j", "down":
				if p.selectedIdx < len(p.files)-1 {
					p.selectedIdx++
					p.updateViewport()
				}
			case "k", "up":
				if p.selectedIdx > 0 {
					p.selectedIdx--
					p.updateViewport()
				}
			case "g":
				p.selectedIdx = 0
				p.updateViewport()
			case "G":
				p.selectedIdx = len(p.files) - 1
				p.updateViewport()
			case "r":
				p.loadFiles()
				p.updateViewport()
			}
		}
	}

	return p, cmd
}

// View renders the panel
func (p *FilesPanel) View() string {
	if !p.ready {
		return "Loading..."
	}

	var content strings.Builder

	if len(p.files) == 0 {
		content.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("No files found"))
	} else {
		for i, file := range p.files {
			// Build the line
			indent := strings.Repeat("  ", file.Level)
			icon := "📄"
			if file.IsDir {
				icon = "📁"
			}

			var style lipgloss.Style
			if i == p.selectedIdx && p.IsFocused() {
				style = lipgloss.NewStyle().
					Background(lipgloss.Color("240")).
					Foreground(lipgloss.Color("255")).
					Bold(true)
			} else if file.GitStatus != "" {
				// Color based on git status
				color := lipgloss.Color("240")
				switch file.GitStatus {
				case "M":
					color = lipgloss.Color("214") // Orange for modified
				case "A":
					color = lipgloss.Color("82") // Green for added
				case "D":
					color = lipgloss.Color("196") // Red for deleted
				case "?":
					color = lipgloss.Color("244") // Gray for untracked
				}
				style = lipgloss.NewStyle().Foreground(color)
			} else {
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
			}

			line := fmt.Sprintf("%s%s %s", indent, icon, file.Name)
			if file.GitStatus != "" {
				line = fmt.Sprintf("%s [%s]", line, file.GitStatus)
			}

			content.WriteString(style.Render(line))
			if i < len(p.files)-1 {
				content.WriteString("\n")
			}
		}
	}

	p.viewport.SetContent(content.String())
	return p.viewport.View()
}

// SetSize sets the panel dimensions
func (p *FilesPanel) SetSize(width, height int) {
	p.width = width
	p.height = height
	p.viewport.Width = width - 2
	p.viewport.Height = height - 2
	p.ready = true
}

// Focus sets the panel as focused
func (p *FilesPanel) Focus() {
	p.focused = true
}

// Blur removes focus from the panel
func (p *FilesPanel) Blur() {
	p.focused = false
}

// IsFocused returns whether the panel is focused
func (p *FilesPanel) IsFocused() bool {
	return p.focused
}

// Type returns the panel type
func (p *FilesPanel) Type() PanelType {
	return PanelFiles
}

// Title returns the panel title
func (p *FilesPanel) Title() string {
	return "Files"
}

// loadFiles loads files from the root path
func (p *FilesPanel) loadFiles() {
	p.files = make([]FileEntry, 0)

	err := filepath.WalkDir(p.rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip hidden files and common ignore patterns
		name := d.Name()
		if strings.HasPrefix(name, ".") && name != "." {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		// Skip common directories
		if d.IsDir() && (name == "node_modules" || name == "vendor" || name == "dist" || name == "build") {
			return fs.SkipDir
		}

		// Calculate relative path and level
		relPath, _ := filepath.Rel(p.rootPath, path)
		level := strings.Count(relPath, string(os.PathSeparator))
		if relPath == "." {
			return nil // Skip root
		}

		p.files = append(p.files, FileEntry{
			Name:   name,
			Path:   path,
			IsDir:  d.IsDir(),
			Level:  level,
		})

		return nil
	})

	if err != nil {
		// Log error but don't fail
	}
}

// updateViewport updates the viewport position to show the selected item
func (p *FilesPanel) updateViewport() {
	if p.selectedIdx < p.viewport.YOffset {
		p.viewport.YOffset = p.selectedIdx
	} else if p.selectedIdx >= p.viewport.YOffset+p.viewport.Height {
		p.viewport.YOffset = p.selectedIdx - p.viewport.Height + 1
	}
}

// GetSelectedFile returns the currently selected file
func (p *FilesPanel) GetSelectedFile() *FileEntry {
	if p.selectedIdx >= 0 && p.selectedIdx < len(p.files) {
		return &p.files[p.selectedIdx]
	}
	return nil
}

// UpdatePanel is a wrapper that implements the Panel interface
func (p *FilesPanel) UpdatePanel(msg tea.Msg) tea.Cmd {
	_, cmd := p.Update(msg)
	return cmd
}
