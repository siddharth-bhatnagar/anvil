package panels

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DiffFile represents a single file's diff
type DiffFile struct {
	Path     string
	OldPath  string // For renames
	Diff     string
	Added    int
	Removed  int
	Language string
}

// DiffPanel displays file diffs
type DiffPanel struct {
	width       int
	height      int
	focused     bool
	viewport    viewport.Model
	diff        string
	filePath    string
	ready       bool
	files       []DiffFile
	currentFile int
	showStats   bool
}

// NewDiffPanel creates a new diff panel
func NewDiffPanel() *DiffPanel {
	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	return &DiffPanel{
		viewport: vp,
	}
}

// Init initializes the panel
func (p *DiffPanel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (p *DiffPanel) Update(msg tea.Msg) (*DiffPanel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if p.IsFocused() {
			switch msg.String() {
			case "n", "tab":
				// Next file in multi-file mode
				if len(p.files) > 1 {
					p.NextFile()
					return p, nil
				}
			case "p", "shift+tab":
				// Previous file in multi-file mode
				if len(p.files) > 1 {
					p.PrevFile()
					return p, nil
				}
			case "s":
				// Toggle stats
				p.ToggleStats()
				return p, nil
			}
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
func (p *DiffPanel) View() string {
	if !p.ready {
		return "Loading..."
	}

	// If we have multiple files, show file list
	if len(p.files) > 0 {
		return p.renderMultiFileDiff()
	}

	if p.diff == "" {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("No diff to display")
	}

	// Render single diff with syntax highlighting
	content := p.renderDiffContent(p.diff, p.filePath)
	p.viewport.SetContent(content)
	return p.viewport.View()
}

// renderMultiFileDiff renders a multi-file diff view
func (p *DiffPanel) renderMultiFileDiff() string {
	var content strings.Builder

	// Show file summary
	if p.showStats {
		totalAdded, totalRemoved := 0, 0
		for _, f := range p.files {
			totalAdded += f.Added
			totalRemoved += f.Removed
		}

		statsStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)
		content.WriteString(statsStyle.Render(fmt.Sprintf(
			"Files: %d | +%d -%d\n",
			len(p.files), totalAdded, totalRemoved,
		)))
		content.WriteString("\n")
	}

	// Show file tabs/list
	for i, f := range p.files {
		var style lipgloss.Style
		prefix := "  "
		if i == p.currentFile {
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")).
				Bold(true)
			prefix = "▸ "
		} else {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
		}

		addedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
		removedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

		stats := fmt.Sprintf(" %s%s",
			addedStyle.Render(fmt.Sprintf("+%d", f.Added)),
			removedStyle.Render(fmt.Sprintf(" -%d", f.Removed)),
		)

		content.WriteString(style.Render(prefix + f.Path))
		content.WriteString(stats)
		content.WriteString("\n")
	}
	content.WriteString("\n")

	// Show current file's diff
	if p.currentFile >= 0 && p.currentFile < len(p.files) {
		currentDiff := p.files[p.currentFile]
		content.WriteString(p.renderDiffContent(currentDiff.Diff, currentDiff.Path))
	}

	p.viewport.SetContent(content.String())
	return p.viewport.View()
}

// renderDiffContent renders diff content with syntax highlighting
func (p *DiffPanel) renderDiffContent(diff, filePath string) string {
	var content strings.Builder

	if filePath != "" {
		header := lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true).
			Render("File: " + filePath)
		content.WriteString(header)
		content.WriteString("\n\n")
	}

	// Define styles
	addedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	removedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	hunkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	contextStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	// Line number styles
	lineNumStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	// Split diff into lines and color them
	lines := strings.Split(diff, "\n")
	oldLine, newLine := 0, 0

	for i, line := range lines {
		var style lipgloss.Style
		var lineNums string

		// Parse hunk header for line numbers
		if strings.HasPrefix(line, "@@") {
			// Parse @@ -start,count +start,count @@
			parts := strings.Split(line, " ")
			for _, part := range parts {
				if strings.HasPrefix(part, "-") && strings.Contains(part, ",") {
					fmt.Sscanf(part, "-%d", &oldLine)
				} else if strings.HasPrefix(part, "+") && strings.Contains(part, ",") {
					fmt.Sscanf(part, "+%d", &newLine)
				}
			}
			style = hunkStyle
		} else if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
			style = headerStyle
		} else if strings.HasPrefix(line, "diff") || strings.HasPrefix(line, "index") {
			style = headerStyle
		} else if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			// Added line
			style = addedStyle
			if newLine > 0 {
				lineNums = lineNumStyle.Render(fmt.Sprintf("   %4d ", newLine))
				newLine++
			}
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			// Removed line
			style = removedStyle
			if oldLine > 0 {
				lineNums = lineNumStyle.Render(fmt.Sprintf("%4d    ", oldLine))
				oldLine++
			}
		} else if len(line) > 0 {
			// Context line
			style = contextStyle
			if oldLine > 0 && newLine > 0 {
				lineNums = lineNumStyle.Render(fmt.Sprintf("%4d %4d ", oldLine, newLine))
				oldLine++
				newLine++
			}
		} else {
			style = contextStyle
		}

		if lineNums != "" {
			content.WriteString(lineNums)
		}
		content.WriteString(style.Render(line))
		if i < len(lines)-1 {
			content.WriteString("\n")
		}
	}

	return content.String()
}

// SetSize sets the panel dimensions
func (p *DiffPanel) SetSize(width, height int) {
	p.width = width
	p.height = height
	p.viewport.Width = width - 2
	p.viewport.Height = height - 2
	p.ready = true
}

// Focus sets the panel as focused
func (p *DiffPanel) Focus() {
	p.focused = true
}

// Blur removes focus from the panel
func (p *DiffPanel) Blur() {
	p.focused = false
}

// IsFocused returns whether the panel is focused
func (p *DiffPanel) IsFocused() bool {
	return p.focused
}

// Type returns the panel type
func (p *DiffPanel) Type() PanelType {
	return PanelDiff
}

// Title returns the panel title
func (p *DiffPanel) Title() string {
	return "Diff"
}

// SetDiff sets the diff content
func (p *DiffPanel) SetDiff(diff, filePath string) {
	p.diff = diff
	p.filePath = filePath
	p.files = nil // Clear multi-file mode
	p.currentFile = 0
	p.viewport.GotoTop()
}

// SetMultiFileDiff sets multiple file diffs
func (p *DiffPanel) SetMultiFileDiff(files []DiffFile) {
	p.files = files
	p.currentFile = 0
	p.showStats = true
	p.diff = ""
	p.filePath = ""
	p.viewport.GotoTop()
}

// AddFile adds a file to the multi-file diff
func (p *DiffPanel) AddFile(file DiffFile) {
	// Count added/removed lines if not set
	if file.Added == 0 && file.Removed == 0 {
		file.Added, file.Removed = countDiffLines(file.Diff)
	}
	p.files = append(p.files, file)
	p.showStats = true
}

// countDiffLines counts added and removed lines in a diff
func countDiffLines(diff string) (added, removed int) {
	lines := strings.Split(diff, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			added++
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			removed++
		}
	}
	return
}

// NextFile switches to the next file
func (p *DiffPanel) NextFile() {
	if len(p.files) > 0 {
		p.currentFile = (p.currentFile + 1) % len(p.files)
		p.viewport.GotoTop()
	}
}

// PrevFile switches to the previous file
func (p *DiffPanel) PrevFile() {
	if len(p.files) > 0 {
		p.currentFile = (p.currentFile - 1 + len(p.files)) % len(p.files)
		p.viewport.GotoTop()
	}
}

// SetCurrentFile sets the current file by index
func (p *DiffPanel) SetCurrentFile(index int) {
	if index >= 0 && index < len(p.files) {
		p.currentFile = index
		p.viewport.GotoTop()
	}
}

// GetCurrentFile returns the current file diff
func (p *DiffPanel) GetCurrentFile() *DiffFile {
	if p.currentFile >= 0 && p.currentFile < len(p.files) {
		return &p.files[p.currentFile]
	}
	return nil
}

// FileCount returns the number of files
func (p *DiffPanel) FileCount() int {
	return len(p.files)
}

// ClearDiff clears the diff content
func (p *DiffPanel) ClearDiff() {
	p.diff = ""
	p.filePath = ""
	p.files = nil
	p.currentFile = 0
}

// GetDiff returns the current diff
func (p *DiffPanel) GetDiff() string {
	return p.diff
}

// ToggleStats toggles the stats display
func (p *DiffPanel) ToggleStats() {
	p.showStats = !p.showStats
}

// UpdatePanel is a wrapper that implements the Panel interface
func (p *DiffPanel) UpdatePanel(msg tea.Msg) tea.Cmd {
	_, cmd := p.Update(msg)
	return cmd
}
