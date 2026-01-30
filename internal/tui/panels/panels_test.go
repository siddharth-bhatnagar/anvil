package panels

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// ConversationPanel tests
func TestNewConversationPanel(t *testing.T) {
	p := NewConversationPanel()

	if p == nil {
		t.Error("NewConversationPanel returned nil")
	}
	if len(p.messages) != 0 {
		t.Error("New panel should have no messages")
	}
	if p.focused {
		t.Error("New panel should not be focused")
	}
}

func TestConversationPanelAddMessage(t *testing.T) {
	p := NewConversationPanel()

	p.AddMessage("user", "Hello")
	if len(p.messages) != 1 {
		t.Error("Message not added")
	}
	if p.messages[0].Role != "user" || p.messages[0].Content != "Hello" {
		t.Error("Message content incorrect")
	}

	p.AddMessage("assistant", "Hi there!")
	if len(p.messages) != 2 {
		t.Error("Second message not added")
	}
}

func TestConversationPanelClearMessages(t *testing.T) {
	p := NewConversationPanel()

	p.AddMessage("user", "Hello")
	p.AddMessage("assistant", "Hi")

	p.ClearMessages()
	if len(p.messages) != 0 {
		t.Error("Messages not cleared")
	}
}

func TestConversationPanelGetMessages(t *testing.T) {
	p := NewConversationPanel()

	p.AddMessage("user", "Test")
	messages := p.GetMessages()

	if len(messages) != 1 {
		t.Error("GetMessages returned wrong count")
	}
}

func TestConversationPanelFocus(t *testing.T) {
	p := NewConversationPanel()

	if p.IsFocused() {
		t.Error("Panel should not be focused initially")
	}

	p.Focus()
	if !p.IsFocused() {
		t.Error("Panel should be focused after Focus()")
	}

	p.Blur()
	if p.IsFocused() {
		t.Error("Panel should not be focused after Blur()")
	}
}

func TestConversationPanelSetSize(t *testing.T) {
	p := NewConversationPanel()

	p.SetSize(100, 50)

	if p.width != 100 || p.height != 50 {
		t.Error("SetSize did not set dimensions correctly")
	}
	if !p.ready {
		t.Error("Panel should be ready after SetSize")
	}
}

func TestConversationPanelType(t *testing.T) {
	p := NewConversationPanel()

	if p.Type() != PanelConversation {
		t.Error("Type() returned wrong value")
	}
}

func TestConversationPanelTitle(t *testing.T) {
	p := NewConversationPanel()

	if p.Title() != "Conversation" {
		t.Error("Title() returned wrong value")
	}
}

func TestConversationPanelView(t *testing.T) {
	p := NewConversationPanel()

	// Not ready
	view := p.View()
	if !strings.Contains(view, "Loading") {
		t.Error("Should show loading when not ready")
	}

	// Ready but empty
	p.SetSize(80, 24)
	view = p.View()
	if !strings.Contains(view, "No messages") {
		t.Error("Should show empty message")
	}

	// With messages
	p.AddMessage("user", "Hello world")
	view = p.View()
	// View is rendered through viewport, just check it's not empty
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestConversationPanelInit(t *testing.T) {
	p := NewConversationPanel()
	cmd := p.Init()
	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestConversationPanelUpdate(t *testing.T) {
	p := NewConversationPanel()
	p.SetSize(80, 24)
	p.Focus()

	// Test key message
	newP, cmd := p.Update(tea.KeyMsg{Type: tea.KeyDown})
	if newP == nil {
		t.Error("Update should return panel")
	}
	_ = cmd // May be nil or command
}

// DiffPanel tests
func TestNewDiffPanel(t *testing.T) {
	p := NewDiffPanel()

	if p == nil {
		t.Error("NewDiffPanel returned nil")
	}
	if p.diff != "" {
		t.Error("New panel should have no diff")
	}
}

func TestDiffPanelSetDiff(t *testing.T) {
	p := NewDiffPanel()

	diff := "+added line\n-removed line"
	p.SetDiff(diff, "test.go")

	if p.diff != diff {
		t.Error("Diff not set correctly")
	}
	if p.filePath != "test.go" {
		t.Error("FilePath not set correctly")
	}
}

func TestDiffPanelSetMultiFileDiff(t *testing.T) {
	p := NewDiffPanel()

	files := []DiffFile{
		{Path: "file1.go", Diff: "+line1", Added: 1, Removed: 0},
		{Path: "file2.go", Diff: "-line2", Added: 0, Removed: 1},
	}

	p.SetMultiFileDiff(files)

	if len(p.files) != 2 {
		t.Error("Files not set correctly")
	}
	if p.currentFile != 0 {
		t.Error("currentFile should be 0")
	}
}

func TestDiffPanelNavigation(t *testing.T) {
	p := NewDiffPanel()
	p.SetMultiFileDiff([]DiffFile{
		{Path: "file1.go"},
		{Path: "file2.go"},
		{Path: "file3.go"},
	})

	// Next file
	p.NextFile()
	if p.currentFile != 1 {
		t.Errorf("Expected currentFile 1, got %d", p.currentFile)
	}

	// Next file wraps
	p.NextFile()
	p.NextFile()
	if p.currentFile != 0 {
		t.Error("Should wrap to 0")
	}

	// Prev file
	p.PrevFile()
	if p.currentFile != 2 {
		t.Errorf("Expected currentFile 2, got %d", p.currentFile)
	}

	// Set specific file
	p.SetCurrentFile(1)
	if p.currentFile != 1 {
		t.Error("SetCurrentFile failed")
	}
}

func TestDiffPanelGetCurrentFile(t *testing.T) {
	p := NewDiffPanel()

	// No files
	if p.GetCurrentFile() != nil {
		t.Error("Should return nil when no files")
	}

	// With files
	p.SetMultiFileDiff([]DiffFile{{Path: "test.go"}})
	file := p.GetCurrentFile()
	if file == nil || file.Path != "test.go" {
		t.Error("GetCurrentFile returned wrong file")
	}
}

func TestDiffPanelClearDiff(t *testing.T) {
	p := NewDiffPanel()
	p.SetDiff("diff content", "file.go")

	p.ClearDiff()

	if p.diff != "" || p.filePath != "" {
		t.Error("ClearDiff did not clear")
	}
}

func TestDiffPanelFocus(t *testing.T) {
	p := NewDiffPanel()

	p.Focus()
	if !p.IsFocused() {
		t.Error("Should be focused")
	}

	p.Blur()
	if p.IsFocused() {
		t.Error("Should not be focused")
	}
}

func TestDiffPanelType(t *testing.T) {
	p := NewDiffPanel()
	if p.Type() != PanelDiff {
		t.Error("Wrong type")
	}
}

func TestDiffPanelTitle(t *testing.T) {
	p := NewDiffPanel()
	if p.Title() != "Diff" {
		t.Error("Wrong title")
	}
}

func TestDiffPanelView(t *testing.T) {
	p := NewDiffPanel()

	// Not ready
	view := p.View()
	if !strings.Contains(view, "Loading") {
		t.Error("Should show loading")
	}

	// Ready but empty
	p.SetSize(80, 24)
	view = p.View()
	if !strings.Contains(view, "No diff") {
		t.Error("Should show no diff message")
	}

	// With diff
	p.SetDiff("+added\n-removed", "test.go")
	view = p.View()
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestDiffPanelToggleStats(t *testing.T) {
	p := NewDiffPanel()
	p.SetMultiFileDiff([]DiffFile{{Path: "test.go"}})

	initial := p.showStats
	p.ToggleStats()
	if p.showStats == initial {
		t.Error("ToggleStats should toggle")
	}
}

func TestCountDiffLines(t *testing.T) {
	diff := `--- a/file.go
+++ b/file.go
@@ -1,3 +1,4 @@
 unchanged
+added line 1
+added line 2
-removed line
 unchanged`

	added, removed := countDiffLines(diff)

	if added != 2 {
		t.Errorf("Expected 2 added, got %d", added)
	}
	if removed != 1 {
		t.Errorf("Expected 1 removed, got %d", removed)
	}
}

// FilesPanel tests
func TestNewFilesPanel(t *testing.T) {
	// Use temp dir
	tmpDir := t.TempDir()
	p := NewFilesPanel(tmpDir)

	if p == nil {
		t.Error("NewFilesPanel returned nil")
	}
	if p.rootPath != tmpDir {
		t.Error("Root path not set correctly")
	}
}

func TestNewFilesPanelEmptyPath(t *testing.T) {
	p := NewFilesPanel("")

	if p == nil {
		t.Error("NewFilesPanel returned nil")
	}
	// Should use current working directory
	cwd, _ := os.Getwd()
	if p.rootPath != cwd {
		t.Error("Should use cwd when path is empty")
	}
}

func TestFilesPanelFocus(t *testing.T) {
	p := NewFilesPanel(t.TempDir())

	p.Focus()
	if !p.IsFocused() {
		t.Error("Should be focused")
	}

	p.Blur()
	if p.IsFocused() {
		t.Error("Should not be focused")
	}
}

func TestFilesPanelType(t *testing.T) {
	p := NewFilesPanel(t.TempDir())
	if p.Type() != PanelFiles {
		t.Error("Wrong type")
	}
}

func TestFilesPanelTitle(t *testing.T) {
	p := NewFilesPanel(t.TempDir())
	if p.Title() != "Files" {
		t.Error("Wrong title")
	}
}

func TestFilesPanelView(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewFilesPanel(tmpDir)

	// Not ready
	view := p.View()
	if !strings.Contains(view, "Loading") {
		t.Error("Should show loading")
	}

	// Ready but empty
	p.SetSize(80, 24)
	view = p.View()
	if !strings.Contains(view, "No files") {
		t.Error("Should show no files message")
	}

	// With files
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("content"), 0644)
	p.loadFiles()
	view = p.View()
	// Just verify it renders
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestFilesPanelNavigation(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file3.txt"), []byte(""), 0644)

	p := NewFilesPanel(tmpDir)
	p.SetSize(80, 24)
	p.Focus()

	// Down
	p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if p.selectedIdx != 1 {
		t.Errorf("Expected selectedIdx 1, got %d", p.selectedIdx)
	}

	// Up
	p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if p.selectedIdx != 0 {
		t.Errorf("Expected selectedIdx 0, got %d", p.selectedIdx)
	}

	// Go to end
	p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	if p.selectedIdx != len(p.files)-1 {
		t.Error("Should be at end")
	}

	// Go to start
	p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if p.selectedIdx != 0 {
		t.Error("Should be at start")
	}
}

func TestFilesPanelGetSelectedFile(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewFilesPanel(tmpDir)

	// No files
	if p.GetSelectedFile() != nil && len(p.files) > 0 {
		// There might be files, just test it doesn't panic
	}

	// With files
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte(""), 0644)
	p.loadFiles()
	if len(p.files) > 0 {
		file := p.GetSelectedFile()
		if file == nil {
			t.Error("Should return selected file")
		}
	}
}

// PlanPanel tests
func TestNewPlanPanel(t *testing.T) {
	p := NewPlanPanel()

	if p == nil {
		t.Error("NewPlanPanel returned nil")
	}
	if len(p.GetSteps()) != 0 {
		t.Error("New panel should have no steps")
	}
}

func TestPlanPanelAddStep(t *testing.T) {
	p := NewPlanPanel()

	p.AddStep("Step 1")
	p.AddStep("Step 2")

	if p.StepCount() != 2 {
		t.Errorf("Expected 2 steps, got %d", p.StepCount())
	}

	steps := p.GetSteps()
	if steps[0].Description != "Step 1" {
		t.Error("Step 1 description wrong")
	}
	if steps[0].ID != 1 {
		t.Error("Step 1 ID should be 1")
	}
}

func TestPlanPanelStepOperations(t *testing.T) {
	p := NewPlanPanel()
	p.AddStep("Step 1")
	p.AddStep("Step 2")

	// Update step status
	p.UpdateStep(1, StepInProgress, "Working...")
	steps := p.GetSteps()
	if steps[0].Status != StepInProgress {
		t.Error("Step status not updated")
	}
	if steps[0].Details != "Working..." {
		t.Error("Step details not updated")
	}

	// Mark complete
	p.UpdateStep(1, StepCompleted, "Done")
	completed, total := p.GetProgress()
	if completed != 1 || total != 2 {
		t.Errorf("Expected 1/2 progress, got %d/%d", completed, total)
	}
}

func TestPlanPanelClearSteps(t *testing.T) {
	p := NewPlanPanel()
	p.AddStep("Step 1")

	p.ClearSteps()

	if p.StepCount() != 0 {
		t.Error("Steps not cleared")
	}
}

func TestPlanPanelFocus(t *testing.T) {
	p := NewPlanPanel()

	p.Focus()
	if !p.IsFocused() {
		t.Error("Should be focused")
	}

	p.Blur()
	if p.IsFocused() {
		t.Error("Should not be focused")
	}
}

func TestPlanPanelType(t *testing.T) {
	p := NewPlanPanel()
	if p.Type() != PanelPlan {
		t.Error("Wrong type")
	}
}

func TestPlanPanelTitle(t *testing.T) {
	p := NewPlanPanel()
	if p.Title() != "Plan" {
		t.Error("Wrong title")
	}
}

func TestPlanPanelView(t *testing.T) {
	p := NewPlanPanel()

	// Not ready
	view := p.View()
	if !strings.Contains(view, "Loading") {
		t.Error("Should show loading")
	}

	// Ready but empty
	p.SetSize(80, 24)
	view = p.View()
	if !strings.Contains(view, "No plan") {
		t.Error("Should show no plan message")
	}

	// With steps
	p.AddStep("First step")
	p.AddStep("Second step")
	p.UpdateStep(1, StepCompleted, "")
	p.UpdateStep(2, StepInProgress, "")
	view = p.View()
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestStepStatusString(t *testing.T) {
	statuses := []StepStatus{StepPending, StepInProgress, StepCompleted, StepFailed}

	for _, s := range statuses {
		if s.String() == "" {
			t.Errorf("Status %d has empty string", s)
		}
	}
}

// PanelType tests
func TestPanelTypes(t *testing.T) {
	types := []PanelType{PanelConversation, PanelDiff, PanelFiles, PanelPlan}

	for _, pt := range types {
		if pt.String() == "" {
			t.Errorf("PanelType %d has empty string", pt)
		}
	}
}
