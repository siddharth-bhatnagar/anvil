package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ChangeType represents the type of file change
type ChangeType int

const (
	ChangeCreate ChangeType = iota
	ChangeModify
	ChangeDelete
	ChangeRename
)

// String returns the string representation of a change type
func (ct ChangeType) String() string {
	switch ct {
	case ChangeCreate:
		return "create"
	case ChangeModify:
		return "modify"
	case ChangeDelete:
		return "delete"
	case ChangeRename:
		return "rename"
	default:
		return "unknown"
	}
}

// Icon returns an icon for the change type
func (ct ChangeType) Icon() string {
	switch ct {
	case ChangeCreate:
		return "+"
	case ChangeModify:
		return "~"
	case ChangeDelete:
		return "-"
	case ChangeRename:
		return "→"
	default:
		return "?"
	}
}

// FileChange represents a single file change
type FileChange struct {
	ID           string
	Path         string
	OldPath      string     // For renames
	Type         ChangeType
	OldContent   string     // Original content (for rollback)
	NewContent   string     // New content
	Diff         string     // Unified diff
	Description  string     // Human-readable description
	Applied      bool       // Whether the change has been applied
	AppliedAt    time.Time  // When it was applied
	RolledBack   bool       // Whether it was rolled back
	RolledBackAt time.Time  // When it was rolled back
}

// LinesChanged returns the number of lines added and removed
func (fc *FileChange) LinesChanged() (added, removed int) {
	lines := strings.Split(fc.Diff, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			added++
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			removed++
		}
	}
	return
}

// ChangeSet represents a group of related changes
type ChangeSet struct {
	ID          string
	Name        string
	Description string
	Changes     []*FileChange
	CreatedAt   time.Time
	AppliedAt   time.Time
	RolledBack  bool
}

// Summary returns a summary of the change set
func (cs *ChangeSet) Summary() string {
	var totalAdded, totalRemoved int
	var files []string

	for _, change := range cs.Changes {
		added, removed := change.LinesChanged()
		totalAdded += added
		totalRemoved += removed
		files = append(files, fmt.Sprintf("%s %s", change.Type.Icon(), change.Path))
	}

	return fmt.Sprintf(
		"%s (%d files, +%d -%d)\n%s",
		cs.Name,
		len(cs.Changes),
		totalAdded,
		totalRemoved,
		strings.Join(files, "\n"),
	)
}

// AffectedFiles returns a list of affected file paths
func (cs *ChangeSet) AffectedFiles() []string {
	var files []string
	for _, change := range cs.Changes {
		files = append(files, change.Path)
		if change.OldPath != "" {
			files = append(files, change.OldPath)
		}
	}
	return files
}

// ChangeManager coordinates changes across multiple files
type ChangeManager struct {
	mu         sync.RWMutex
	changeSets map[string]*ChangeSet
	current    *ChangeSet
	history    []*ChangeSet
}

// NewChangeManager creates a new change manager
func NewChangeManager() *ChangeManager {
	return &ChangeManager{
		changeSets: make(map[string]*ChangeSet),
		history:    make([]*ChangeSet, 0),
	}
}

// StartChangeSet starts a new change set
func (cm *ChangeManager) StartChangeSet(name, description string) *ChangeSet {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cs := &ChangeSet{
		ID:          generateChangeSetID(),
		Name:        name,
		Description: description,
		Changes:     make([]*FileChange, 0),
		CreatedAt:   time.Now(),
	}

	cm.current = cs
	cm.changeSets[cs.ID] = cs

	return cs
}

// AddChange adds a change to the current change set
func (cm *ChangeManager) AddChange(change *FileChange) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.current == nil {
		// Auto-create a change set
		cm.current = &ChangeSet{
			ID:          generateChangeSetID(),
			Name:        "Auto-generated",
			Description: "Automatically created change set",
			Changes:     make([]*FileChange, 0),
			CreatedAt:   time.Now(),
		}
		cm.changeSets[cm.current.ID] = cm.current
	}

	if change.ID == "" {
		change.ID = fmt.Sprintf("change_%d", len(cm.current.Changes))
	}

	cm.current.Changes = append(cm.current.Changes, change)
	return nil
}

// AddFileCreate adds a file creation change
func (cm *ChangeManager) AddFileCreate(path, content, description string) (*FileChange, error) {
	change := &FileChange{
		Path:        path,
		Type:        ChangeCreate,
		NewContent:  content,
		Description: description,
		Diff:        generateCreateDiff(path, content),
	}

	err := cm.AddChange(change)
	return change, err
}

// AddFileModify adds a file modification change
func (cm *ChangeManager) AddFileModify(path, oldContent, newContent, description string) (*FileChange, error) {
	change := &FileChange{
		Path:        path,
		Type:        ChangeModify,
		OldContent:  oldContent,
		NewContent:  newContent,
		Description: description,
		Diff:        generateModifyDiff(path, oldContent, newContent),
	}

	err := cm.AddChange(change)
	return change, err
}

// AddFileDelete adds a file deletion change
func (cm *ChangeManager) AddFileDelete(path, oldContent, description string) (*FileChange, error) {
	change := &FileChange{
		Path:        path,
		Type:        ChangeDelete,
		OldContent:  oldContent,
		Description: description,
		Diff:        generateDeleteDiff(path, oldContent),
	}

	err := cm.AddChange(change)
	return change, err
}

// AddFileRename adds a file rename change
func (cm *ChangeManager) AddFileRename(oldPath, newPath, description string) (*FileChange, error) {
	change := &FileChange{
		Path:        newPath,
		OldPath:     oldPath,
		Type:        ChangeRename,
		Description: description,
		Diff:        fmt.Sprintf("rename %s -> %s", oldPath, newPath),
	}

	err := cm.AddChange(change)
	return change, err
}

// GetCurrentChangeSet returns the current change set
func (cm *ChangeManager) GetCurrentChangeSet() *ChangeSet {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.current
}

// ApplyChangeSet applies all changes in the current change set
func (cm *ChangeManager) ApplyChangeSet() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.current == nil {
		return fmt.Errorf("no active change set")
	}

	// Apply each change
	for _, change := range cm.current.Changes {
		if err := applyChange(change); err != nil {
			// Rollback applied changes
			cm.rollbackApplied(cm.current)
			return fmt.Errorf("failed to apply change %s: %w", change.Path, err)
		}
		change.Applied = true
		change.AppliedAt = time.Now()
	}

	cm.current.AppliedAt = time.Now()
	cm.history = append(cm.history, cm.current)
	cm.current = nil

	return nil
}

// RollbackChangeSet rolls back the most recent change set
func (cm *ChangeManager) RollbackChangeSet() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if len(cm.history) == 0 {
		return fmt.Errorf("no change sets to rollback")
	}

	// Get the most recent change set
	cs := cm.history[len(cm.history)-1]

	if cs.RolledBack {
		return fmt.Errorf("change set already rolled back")
	}

	// Rollback in reverse order
	cm.rollbackApplied(cs)

	cs.RolledBack = true
	return nil
}

// rollbackApplied rolls back applied changes in a change set
func (cm *ChangeManager) rollbackApplied(cs *ChangeSet) {
	// Rollback in reverse order
	for i := len(cs.Changes) - 1; i >= 0; i-- {
		change := cs.Changes[i]
		if change.Applied && !change.RolledBack {
			rollbackChange(change)
			change.RolledBack = true
			change.RolledBackAt = time.Now()
		}
	}
}

// DiscardCurrentChangeSet discards the current change set without applying
func (cm *ChangeManager) DiscardCurrentChangeSet() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.current != nil {
		delete(cm.changeSets, cm.current.ID)
		cm.current = nil
	}
}

// GetHistory returns the change set history
func (cm *ChangeManager) GetHistory() []*ChangeSet {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	history := make([]*ChangeSet, len(cm.history))
	copy(history, cm.history)
	return history
}

// applyChange applies a single file change
func applyChange(change *FileChange) error {
	switch change.Type {
	case ChangeCreate:
		// Create parent directories if needed
		dir := filepath.Dir(change.Path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		return os.WriteFile(change.Path, []byte(change.NewContent), 0o644)

	case ChangeModify:
		return os.WriteFile(change.Path, []byte(change.NewContent), 0o644)

	case ChangeDelete:
		return os.Remove(change.Path)

	case ChangeRename:
		return os.Rename(change.OldPath, change.Path)

	default:
		return fmt.Errorf("unknown change type: %v", change.Type)
	}
}

// rollbackChange rolls back a single file change
func rollbackChange(change *FileChange) error {
	switch change.Type {
	case ChangeCreate:
		// Delete the created file
		return os.Remove(change.Path)

	case ChangeModify:
		// Restore old content
		return os.WriteFile(change.Path, []byte(change.OldContent), 0o644)

	case ChangeDelete:
		// Restore deleted file
		dir := filepath.Dir(change.Path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		return os.WriteFile(change.Path, []byte(change.OldContent), 0o644)

	case ChangeRename:
		// Rename back
		return os.Rename(change.Path, change.OldPath)

	default:
		return fmt.Errorf("unknown change type: %v", change.Type)
	}
}

// generateChangeSetID generates a unique change set ID
func generateChangeSetID() string {
	return fmt.Sprintf("cs_%d", time.Now().UnixNano())
}

// generateCreateDiff generates a diff for a file creation
func generateCreateDiff(path, content string) string {
	var diff strings.Builder

	diff.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", path, path))
	diff.WriteString("new file mode 100644\n")
	diff.WriteString(fmt.Sprintf("--- /dev/null\n"))
	diff.WriteString(fmt.Sprintf("+++ b/%s\n", path))

	lines := strings.Split(content, "\n")
	diff.WriteString(fmt.Sprintf("@@ -0,0 +1,%d @@\n", len(lines)))

	for _, line := range lines {
		diff.WriteString(fmt.Sprintf("+%s\n", line))
	}

	return diff.String()
}

// generateDeleteDiff generates a diff for a file deletion
func generateDeleteDiff(path, content string) string {
	var diff strings.Builder

	diff.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", path, path))
	diff.WriteString("deleted file mode 100644\n")
	diff.WriteString(fmt.Sprintf("--- a/%s\n", path))
	diff.WriteString("+++ /dev/null\n")

	lines := strings.Split(content, "\n")
	diff.WriteString(fmt.Sprintf("@@ -1,%d +0,0 @@\n", len(lines)))

	for _, line := range lines {
		diff.WriteString(fmt.Sprintf("-%s\n", line))
	}

	return diff.String()
}

// generateModifyDiff generates a simple unified diff for a modification
func generateModifyDiff(path, oldContent, newContent string) string {
	var diff strings.Builder

	diff.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", path, path))
	diff.WriteString(fmt.Sprintf("--- a/%s\n", path))
	diff.WriteString(fmt.Sprintf("+++ b/%s\n", path))

	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	// Simple diff - show all old lines as removed, new lines as added
	// A real implementation would use a proper diff algorithm
	diff.WriteString(fmt.Sprintf("@@ -1,%d +1,%d @@\n", len(oldLines), len(newLines)))

	for _, line := range oldLines {
		diff.WriteString(fmt.Sprintf("-%s\n", line))
	}

	for _, line := range newLines {
		diff.WriteString(fmt.Sprintf("+%s\n", line))
	}

	return diff.String()
}

// PreviewChanges returns a preview of all changes in the current set
func (cm *ChangeManager) PreviewChanges() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.current == nil || len(cm.current.Changes) == 0 {
		return "No pending changes"
	}

	var preview strings.Builder

	preview.WriteString(fmt.Sprintf("Change Set: %s\n", cm.current.Name))
	preview.WriteString(fmt.Sprintf("Description: %s\n", cm.current.Description))
	preview.WriteString(fmt.Sprintf("Files: %d\n\n", len(cm.current.Changes)))

	for i, change := range cm.current.Changes {
		added, removed := change.LinesChanged()
		preview.WriteString(fmt.Sprintf("%d. %s %s (+%d -%d)\n",
			i+1, change.Type.Icon(), change.Path, added, removed))
		if change.Description != "" {
			preview.WriteString(fmt.Sprintf("   %s\n", change.Description))
		}
	}

	return preview.String()
}
