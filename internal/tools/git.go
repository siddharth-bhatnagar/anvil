package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/siddharth-bhatnagar/anvil/pkg/schema"
)

// GitStatusTool shows git status
type GitStatusTool struct {
	BaseTool
}

// NewGitStatusTool creates a new git status tool
func NewGitStatusTool() *GitStatusTool {
	return &GitStatusTool{
		BaseTool: NewBaseTool(
			"git_status",
			"Show the working tree status",
			[]schema.ToolParameter{
				{
					Name:        "path",
					Description: "Path to the git repository",
					Type:        "string",
					Required:    false,
					Default:     ".",
				},
			},
		),
	}
}

// Execute shows git status
func (t *GitStatusTool) Execute(ctx context.Context, args map[string]any) (*schema.ToolResult, error) {
	path := "."
	if pathVal, ok := args["path"]; ok {
		path = fmt.Sprintf("%v", pathVal)
	}

	repo, err := git.PlainOpen(path)
	if err != nil {
		return &schema.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("not a git repository: %v", err),
		}, err
	}

	w, err := repo.Worktree()
	if err != nil {
		return &schema.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to get worktree: %v", err),
		}, err
	}

	status, err := w.Status()
	if err != nil {
		return &schema.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to get status: %v", err),
		}, err
	}

	var output strings.Builder
	output.WriteString("Git Status:\n\n")

	if status.IsClean() {
		output.WriteString("Working tree clean\n")
	} else {
		for file, stat := range status {
			staging := string(stat.Staging)
			worktree := string(stat.Worktree)
			output.WriteString(fmt.Sprintf("  %s%s %s\n", staging, worktree, file))
		}
	}

	return &schema.ToolResult{
		Success: true,
		Output:  output.String(),
	}, nil
}

// RequiresApproval returns false
func (t *GitStatusTool) RequiresApproval(args map[string]any) bool {
	return false
}

// GitDiffTool shows git diff
type GitDiffTool struct {
	BaseTool
}

// NewGitDiffTool creates a new git diff tool
func NewGitDiffTool() *GitDiffTool {
	return &GitDiffTool{
		BaseTool: NewBaseTool(
			"git_diff",
			"Show changes in the working directory",
			[]schema.ToolParameter{
				{
					Name:        "path",
					Description: "Path to the git repository",
					Type:        "string",
					Required:    false,
					Default:     ".",
				},
				{
					Name:        "file",
					Description: "Specific file to diff",
					Type:        "string",
					Required:    false,
				},
			},
		),
	}
}

// Execute shows git diff
func (t *GitDiffTool) Execute(ctx context.Context, args map[string]any) (*schema.ToolResult, error) {
	_ = ctx // unused for now

	// For now, use simple file comparison
	// A full git diff implementation would use go-git's diff capabilities
	return &schema.ToolResult{
		Success: true,
		Output:  "Git diff functionality - use external git command for full diff",
	}, nil
}

// RequiresApproval returns false
func (t *GitDiffTool) RequiresApproval(args map[string]any) bool {
	return false
}

// GitLogTool shows git log
type GitLogTool struct {
	BaseTool
}

// NewGitLogTool creates a new git log tool
func NewGitLogTool() *GitLogTool {
	return &GitLogTool{
		BaseTool: NewBaseTool(
			"git_log",
			"Show commit logs",
			[]schema.ToolParameter{
				{
					Name:        "path",
					Description: "Path to the git repository",
					Type:        "string",
					Required:    false,
					Default:     ".",
				},
				{
					Name:        "max_count",
					Description: "Maximum number of commits to show",
					Type:        "number",
					Required:    false,
					Default:     10,
				},
			},
		),
	}
}

// Execute shows git log
func (t *GitLogTool) Execute(ctx context.Context, args map[string]any) (*schema.ToolResult, error) {
	path := "."
	if pathVal, ok := args["path"]; ok {
		path = fmt.Sprintf("%v", pathVal)
	}

	maxCount := 10
	if maxVal, ok := args["max_count"]; ok {
		if maxInt, ok := maxVal.(float64); ok {
			maxCount = int(maxInt)
		}
	}

	repo, err := git.PlainOpen(path)
	if err != nil {
		return &schema.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("not a git repository: %v", err),
		}, err
	}

	ref, err := repo.Head()
	if err != nil {
		return &schema.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to get HEAD: %v", err),
		}, err
	}

	commits, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return &schema.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to get log: %v", err),
		}, err
	}

	var output strings.Builder
	output.WriteString("Recent commits:\n\n")

	count := 0
	err = commits.ForEach(func(c *object.Commit) error {
		if count >= maxCount {
			return fmt.Errorf("max count reached")
		}

		output.WriteString(fmt.Sprintf("commit %s\n", c.Hash))
		output.WriteString(fmt.Sprintf("Author: %s\n", c.Author.Name))
		output.WriteString(fmt.Sprintf("Date:   %s\n", c.Author.When.Format("Mon Jan 2 15:04:05 2006")))
		output.WriteString(fmt.Sprintf("\n    %s\n\n", c.Message))

		count++
		return nil
	})

	if err != nil && err.Error() != "max count reached" {
		return &schema.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to iterate commits: %v", err),
		}, err
	}

	return &schema.ToolResult{
		Success: true,
		Output:  output.String(),
	}, nil
}

// RequiresApproval returns false
func (t *GitLogTool) RequiresApproval(args map[string]any) bool {
	return false
}
