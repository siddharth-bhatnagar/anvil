package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/siddharth-bhatnagar/anvil/pkg/schema"
)

func TestNewRegistry(t *testing.T) {
	reg := NewRegistry()

	if reg == nil {
		t.Fatal("NewRegistry returned nil")
	}

	tools := reg.List()
	if len(tools) != 0 {
		t.Errorf("New registry should be empty, got %d tools", len(tools))
	}
}

func TestRegistryRegister(t *testing.T) {
	reg := NewRegistry()

	tool := NewReadFileTool()
	err := reg.Register(tool)

	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Try to register same tool again
	err = reg.Register(tool)
	if err == nil {
		t.Error("Should not allow registering duplicate tool")
	}
}

func TestRegistryGet(t *testing.T) {
	reg := NewRegistry()
	reg.Register(NewReadFileTool())

	tool, err := reg.Get("read_file")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if tool.Name() != "read_file" {
		t.Errorf("Expected 'read_file', got '%s'", tool.Name())
	}

	// Get non-existent tool
	_, err = reg.Get("non_existent")
	if err == nil {
		t.Error("Should return error for non-existent tool")
	}
}

func TestRegistryList(t *testing.T) {
	reg := NewRegistry()
	reg.Register(NewReadFileTool())
	reg.Register(NewWriteFileTool())

	tools := reg.List()
	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}
}

func TestRegistryListDefinitions(t *testing.T) {
	reg := NewRegistry()
	reg.Register(NewReadFileTool())

	defs := reg.ListDefinitions()
	if len(defs) != 1 {
		t.Fatalf("Expected 1 definition, got %d", len(defs))
	}

	if defs[0].Name != "read_file" {
		t.Errorf("Expected 'read_file', got '%s'", defs[0].Name)
	}
}

func TestDefaultRegistry(t *testing.T) {
	reg, err := DefaultRegistry()
	if err != nil {
		t.Fatalf("DefaultRegistry failed: %v", err)
	}

	// Should have all default tools registered
	expectedTools := []string{
		"read_file",
		"write_file",
		"search_files",
		"grep_files",
		"list_directory",
		"git_status",
		"git_diff",
		"git_log",
		"shell_command",
		"analyze_file",
		"find_symbol",
	}

	for _, name := range expectedTools {
		_, err := reg.Get(name)
		if err != nil {
			t.Errorf("Expected tool '%s' to be registered", name)
		}
	}
}

func TestReadFileTool(t *testing.T) {
	// Create a temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	content := "Hello, World!"
	os.WriteFile(tmpFile, []byte(content), 0644)

	tool := NewReadFileTool()

	// Test reading the file
	result, err := tool.Execute(context.Background(), map[string]any{
		"path": tmpFile,
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}

	if result.Output != content {
		t.Errorf("Expected '%s', got '%s'", content, result.Output)
	}
}

func TestReadFileToolMissingPath(t *testing.T) {
	tool := NewReadFileTool()

	result, err := tool.Execute(context.Background(), map[string]any{})

	if err == nil {
		t.Error("Should return error for missing path")
	}

	if result.Success {
		t.Error("Should not succeed without path")
	}
}

func TestReadFileToolSensitiveFile(t *testing.T) {
	tool := NewReadFileTool()

	// Try to read a sensitive file pattern
	result, err := tool.Execute(context.Background(), map[string]any{
		"path": "/some/path/.env",
	})

	if err == nil {
		t.Error("Should return error for sensitive file")
	}

	if result.Success {
		t.Error("Should not succeed for sensitive file")
	}
}

func TestWriteFileTool(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "output.txt")

	tool := NewWriteFileTool()

	// Write should require approval
	if !tool.RequiresApproval(map[string]any{"path": tmpFile}) {
		t.Error("Write should require approval")
	}

	// Execute the write
	result, err := tool.Execute(context.Background(), map[string]any{
		"path":    tmpFile,
		"content": "Test content",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}

	// Verify the file was written
	content, _ := os.ReadFile(tmpFile)
	if string(content) != "Test content" {
		t.Errorf("File content mismatch")
	}
}

func TestWriteFileToolCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "subdir", "nested", "file.txt")

	tool := NewWriteFileTool()

	result, err := tool.Execute(context.Background(), map[string]any{
		"path":    tmpFile,
		"content": "Nested content",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}

	// Verify directory was created
	if _, err := os.Stat(filepath.Dir(tmpFile)); os.IsNotExist(err) {
		t.Error("Directory should have been created")
	}
}

func TestListDirectoryTool(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some files and directories
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.go"), []byte(""), 0644)

	tool := NewListDirectoryTool()

	result, err := tool.Execute(context.Background(), map[string]any{
		"path": tmpDir,
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}

	// Check data contains expected info
	data := result.Data
	if dirs, ok := data["dirs"].([]string); ok {
		if len(dirs) != 1 {
			t.Errorf("Expected 1 directory, got %d", len(dirs))
		}
	}

	if files, ok := data["files"].([]string); ok {
		if len(files) != 2 {
			t.Errorf("Expected 2 files, got %d", len(files))
		}
	}
}

func TestSearchFilesTool(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some files
	os.WriteFile(filepath.Join(tmpDir, "test1.go"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "test2.go"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "other.txt"), []byte(""), 0644)

	tool := NewSearchFilesTool()

	result, err := tool.Execute(context.Background(), map[string]any{
		"pattern": filepath.Join(tmpDir, "*.go"),
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}

	if count, ok := result.Data["count"].(int); ok {
		if count != 2 {
			t.Errorf("Expected 2 matches, got %d", count)
		}
	}
}

func TestGrepFilesTool(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files with content
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("Hello World"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("Goodbye World"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file3.txt"), []byte("Nothing here"), 0644)

	tool := NewGrepFilesTool()

	result, err := tool.Execute(context.Background(), map[string]any{
		"pattern": "World",
		"path":    tmpDir,
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}

	if count, ok := result.Data["count"].(int); ok {
		if count != 2 {
			t.Errorf("Expected 2 matches, got %d", count)
		}
	}
}

func TestShellCommandTool(t *testing.T) {
	tool := NewShellCommandTool()

	// Simple echo command
	result, err := tool.Execute(context.Background(), map[string]any{
		"command": "echo hello",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}

	if result.Output != "hello\n" {
		t.Errorf("Expected 'hello\\n', got '%s'", result.Output)
	}
}

func TestShellCommandToolRequiresApproval(t *testing.T) {
	tool := NewShellCommandTool()

	tests := []struct {
		command  string
		approval bool
	}{
		{"ls", false},
		{"echo hello", false},
		{"rm file.txt", true},
		{"rm -rf /", true},
		{"chmod 777 file", true},
		{"mv old new", true},
		{"cat file", false},
		{"grep pattern file", false},
		{"ls -f", true}, // Contains -f flag
	}

	for _, tt := range tests {
		args := map[string]any{"command": tt.command}
		if tool.RequiresApproval(args) != tt.approval {
			t.Errorf("Command '%s': expected approval=%v, got %v", tt.command, tt.approval, !tt.approval)
		}
	}
}

func TestRegistryExecute(t *testing.T) {
	reg := NewRegistry()
	reg.Register(NewListDirectoryTool())

	toolCall := schema.ToolCall{
		ID:   "test-1",
		Name: "list_directory",
		Arguments: map[string]any{
			"path": ".",
		},
	}

	result, err := reg.Execute(context.Background(), toolCall)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.ToolCallID != "test-1" {
		t.Errorf("Expected ToolCallID 'test-1', got '%s'", result.ToolCallID)
	}

	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}
}

func TestRegistryExecuteNonExistentTool(t *testing.T) {
	reg := NewRegistry()

	toolCall := schema.ToolCall{
		ID:   "test-1",
		Name: "non_existent",
	}

	result, err := reg.Execute(context.Background(), toolCall)

	if err == nil {
		t.Error("Should return error for non-existent tool")
	}

	if result.Success {
		t.Error("Should not succeed for non-existent tool")
	}
}

func TestIsSensitiveFile(t *testing.T) {
	tests := []struct {
		path      string
		sensitive bool
	}{
		{"/path/to/.env", true},
		{"/path/to/.env.local", true},
		{"/path/to/credentials.json", true},
		{"/path/to/id_rsa", true},
		{"/path/to/server.key", true},
		{"/path/to/cert.pem", true},
		{"/path/to/main.go", false},
		{"/path/to/config.yaml", false},
		{"/path/to/README.md", false},
	}

	for _, tt := range tests {
		result := isSensitiveFile(tt.path)
		if result != tt.sensitive {
			t.Errorf("Path '%s': expected sensitive=%v, got %v", tt.path, tt.sensitive, result)
		}
	}
}

func TestIsDestructiveCommand(t *testing.T) {
	tests := []struct {
		command     string
		destructive bool
	}{
		{"ls -la", false},
		{"cat file.txt", false},
		{"rm file.txt", true},
		{"rm -rf /", true},
		{"mv old new", true},
		{"dd if=/dev/zero of=/dev/sda", true},
		{"chmod 755 file", true},
		{"chown user:group file", true},
		{"echo hello > /dev/null", true},
		{"git status", false},
		{"npm install", false},
		{"npm install --force", true},
	}

	for _, tt := range tests {
		result := isDestructiveCommand(tt.command)
		if result != tt.destructive {
			t.Errorf("Command '%s': expected destructive=%v, got %v", tt.command, tt.destructive, result)
		}
	}
}
