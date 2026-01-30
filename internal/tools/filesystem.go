package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/siddharth-bhatnagar/anvil/pkg/schema"
)

// Sensitive file patterns that should not be read
var sensitivePatterns = []string{
	".env",
	".env.local",
	".env.production",
	"credentials.json",
	"secrets.yaml",
	"secrets.yml",
	"id_rsa",
	"id_ed25519",
	".pem",
	".key",
}

// ReadFileTool reads a file from the filesystem
type ReadFileTool struct {
	BaseTool
}

// NewReadFileTool creates a new read file tool
func NewReadFileTool() *ReadFileTool {
	return &ReadFileTool{
		BaseTool: NewBaseTool(
			"read_file",
			"Read the contents of a file",
			[]schema.ToolParameter{
				{
					Name:        "path",
					Description: "Path to the file to read",
					Type:        "string",
					Required:    true,
				},
			},
		),
	}
}

// Execute reads a file
func (t *ReadFileTool) Execute(ctx context.Context, args map[string]any) (*schema.ToolResult, error) {
	pathVal, ok := args["path"]
	if !ok {
		return &schema.ToolResult{
			Success: false,
			Error:   "missing required parameter: path",
		}, fmt.Errorf("missing required parameter: path")
	}

	path := fmt.Sprintf("%v", pathVal)

	// Security check: prevent reading sensitive files
	if isSensitiveFile(path) {
		return &schema.ToolResult{
			Success: false,
			Error:   "cannot read sensitive file: " + path,
		}, fmt.Errorf("cannot read sensitive file: %s", path)
	}

	// Read the file
	content, err := os.ReadFile(path)
	if err != nil {
		return &schema.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to read file: %v", err),
		}, err
	}

	return &schema.ToolResult{
		Success: true,
		Output:  string(content),
		Data: map[string]any{
			"path": path,
			"size": len(content),
		},
	}, nil
}

// RequiresApproval returns false for read operations
func (t *ReadFileTool) RequiresApproval(args map[string]any) bool {
	return false
}

// WriteFileTool writes content to a file
type WriteFileTool struct {
	BaseTool
}

// NewWriteFileTool creates a new write file tool
func NewWriteFileTool() *WriteFileTool {
	return &WriteFileTool{
		BaseTool: NewBaseTool(
			"write_file",
			"Write content to a file (creates or overwrites)",
			[]schema.ToolParameter{
				{
					Name:        "path",
					Description: "Path to the file to write",
					Type:        "string",
					Required:    true,
				},
				{
					Name:        "content",
					Description: "Content to write to the file",
					Type:        "string",
					Required:    true,
				},
			},
		),
	}
}

// Execute writes to a file
func (t *WriteFileTool) Execute(ctx context.Context, args map[string]any) (*schema.ToolResult, error) {
	pathVal, ok := args["path"]
	if !ok {
		return &schema.ToolResult{
			Success: false,
			Error:   "missing required parameter: path",
		}, fmt.Errorf("missing required parameter: path")
	}

	contentVal, ok := args["content"]
	if !ok {
		return &schema.ToolResult{
			Success: false,
			Error:   "missing required parameter: content",
		}, fmt.Errorf("missing required parameter: content")
	}

	path := fmt.Sprintf("%v", pathVal)
	content := fmt.Sprintf("%v", contentVal)

	// Security check
	if isSensitiveFile(path) {
		return &schema.ToolResult{
			Success: false,
			Error:   "cannot write to sensitive file: " + path,
		}, fmt.Errorf("cannot write to sensitive file: %s", path)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return &schema.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to create directory: %v", err),
		}, err
	}

	// Write the file
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return &schema.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to write file: %v", err),
		}, err
	}

	return &schema.ToolResult{
		Success: true,
		Output:  fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), path),
		Data: map[string]any{
			"path": path,
			"size": len(content),
		},
	}, nil
}

// RequiresApproval returns true for write operations
func (t *WriteFileTool) RequiresApproval(args map[string]any) bool {
	return true
}

// SearchFilesTool searches for files matching a pattern
type SearchFilesTool struct {
	BaseTool
}

// NewSearchFilesTool creates a new search files tool
func NewSearchFilesTool() *SearchFilesTool {
	return &SearchFilesTool{
		BaseTool: NewBaseTool(
			"search_files",
			"Search for files matching a glob pattern",
			[]schema.ToolParameter{
				{
					Name:        "pattern",
					Description: "Glob pattern to match (e.g., '**/*.go', 'src/**/*.ts')",
					Type:        "string",
					Required:    true,
				},
				{
					Name:        "max_results",
					Description: "Maximum number of results to return",
					Type:        "number",
					Required:    false,
					Default:     100,
				},
			},
		),
	}
}

// Execute searches for files
func (t *SearchFilesTool) Execute(ctx context.Context, args map[string]any) (*schema.ToolResult, error) {
	patternVal, ok := args["pattern"]
	if !ok {
		return &schema.ToolResult{
			Success: false,
			Error:   "missing required parameter: pattern",
		}, fmt.Errorf("missing required parameter: pattern")
	}

	pattern := fmt.Sprintf("%v", patternVal)

	maxResults := 100
	if maxVal, ok := args["max_results"]; ok {
		if maxInt, ok := maxVal.(float64); ok {
			maxResults = int(maxInt)
		}
	}

	// Search for files
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return &schema.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("invalid pattern: %v", err),
		}, err
	}

	// Limit results
	if len(matches) > maxResults {
		matches = matches[:maxResults]
	}

	output := fmt.Sprintf("Found %d files matching '%s':\n", len(matches), pattern)
	for _, match := range matches {
		output += "  " + match + "\n"
	}

	return &schema.ToolResult{
		Success: true,
		Output:  output,
		Data: map[string]any{
			"matches": matches,
			"count":   len(matches),
		},
	}, nil
}

// RequiresApproval returns false for search operations
func (t *SearchFilesTool) RequiresApproval(args map[string]any) bool {
	return false
}

// GrepFilesTool searches for content in files
type GrepFilesTool struct {
	BaseTool
}

// NewGrepFilesTool creates a new grep files tool
func NewGrepFilesTool() *GrepFilesTool {
	return &GrepFilesTool{
		BaseTool: NewBaseTool(
			"grep_files",
			"Search for text content in files",
			[]schema.ToolParameter{
				{
					Name:        "pattern",
					Description: "Text pattern to search for",
					Type:        "string",
					Required:    true,
				},
				{
					Name:        "path",
					Description: "Path to search in (file or directory)",
					Type:        "string",
					Required:    false,
					Default:     ".",
				},
			},
		),
	}
}

// Execute searches file contents
func (t *GrepFilesTool) Execute(ctx context.Context, args map[string]any) (*schema.ToolResult, error) {
	patternVal, ok := args["pattern"]
	if !ok {
		return &schema.ToolResult{
			Success: false,
			Error:   "missing required parameter: pattern",
		}, fmt.Errorf("missing required parameter: pattern")
	}

	pattern := fmt.Sprintf("%v", patternVal)
	searchPath := "."
	if pathVal, ok := args["path"]; ok {
		searchPath = fmt.Sprintf("%v", pathVal)
	}

	var matches []string
	var matchCount int

	// Walk the directory
	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip directories and sensitive files
		if info.IsDir() || isSensitiveFile(path) {
			return nil
		}

		// Read file
		content, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip unreadable files
		}

		// Search for pattern
		if strings.Contains(string(content), pattern) {
			matches = append(matches, path)
			matchCount++
			if matchCount >= 50 { // Limit results
				return filepath.SkipAll
			}
		}

		return nil
	})

	if err != nil {
		return &schema.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("search failed: %v", err),
		}, err
	}

	output := fmt.Sprintf("Found '%s' in %d files:\n", pattern, len(matches))
	for _, match := range matches {
		output += "  " + match + "\n"
	}

	return &schema.ToolResult{
		Success: true,
		Output:  output,
		Data: map[string]any{
			"matches": matches,
			"count":   len(matches),
		},
	}, nil
}

// RequiresApproval returns false for grep operations
func (t *GrepFilesTool) RequiresApproval(args map[string]any) bool {
	return false
}

// ListDirectoryTool lists files in a directory
type ListDirectoryTool struct {
	BaseTool
}

// NewListDirectoryTool creates a new list directory tool
func NewListDirectoryTool() *ListDirectoryTool {
	return &ListDirectoryTool{
		BaseTool: NewBaseTool(
			"list_directory",
			"List files and directories in a path",
			[]schema.ToolParameter{
				{
					Name:        "path",
					Description: "Path to the directory to list",
					Type:        "string",
					Required:    false,
					Default:     ".",
				},
			},
		),
	}
}

// Execute lists directory contents
func (t *ListDirectoryTool) Execute(ctx context.Context, args map[string]any) (*schema.ToolResult, error) {
	path := "."
	if pathVal, ok := args["path"]; ok {
		path = fmt.Sprintf("%v", pathVal)
	}

	// Read directory
	entries, err := os.ReadDir(path)
	if err != nil {
		return &schema.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to read directory: %v", err),
		}, err
	}

	var files []string
	var dirs []string

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name()+"/")
		} else {
			files = append(files, entry.Name())
		}
	}

	output := fmt.Sprintf("Directory: %s\n\n", path)
	output += fmt.Sprintf("Directories (%d):\n", len(dirs))
	for _, dir := range dirs {
		output += "  " + dir + "\n"
	}
	output += fmt.Sprintf("\nFiles (%d):\n", len(files))
	for _, file := range files {
		output += "  " + file + "\n"
	}

	return &schema.ToolResult{
		Success: true,
		Output:  output,
		Data: map[string]any{
			"path":  path,
			"dirs":  dirs,
			"files": files,
		},
	}, nil
}

// RequiresApproval returns false for list operations
func (t *ListDirectoryTool) RequiresApproval(args map[string]any) bool {
	return false
}

// isSensitiveFile checks if a file path matches sensitive patterns
func isSensitiveFile(path string) bool {
	base := filepath.Base(path)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(base, pattern) {
			return true
		}
	}
	return false
}
