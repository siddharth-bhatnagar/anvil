package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/siddharth-bhatnagar/anvil/pkg/schema"
)

// Destructive command patterns
var destructiveCommands = []string{
	"rm",
	"mv",
	"dd",
	"mkfs",
	"format",
	":(){:|:&};:", // Fork bomb
	"chmod",
	"chown",
	"> /dev/",
}

// ShellCommandTool executes shell commands
type ShellCommandTool struct {
	BaseTool
}

// NewShellCommandTool creates a new shell command tool
func NewShellCommandTool() *ShellCommandTool {
	return &ShellCommandTool{
		BaseTool: NewBaseTool(
			"shell_command",
			"Execute a shell command (requires approval for destructive commands)",
			[]schema.ToolParameter{
				{
					Name:        "command",
					Description: "The shell command to execute",
					Type:        "string",
					Required:    true,
				},
				{
					Name:        "timeout_seconds",
					Description: "Timeout in seconds (default: 30)",
					Type:        "number",
					Required:    false,
					Default:     30,
				},
			},
		),
	}
}

// Execute runs a shell command
func (t *ShellCommandTool) Execute(ctx context.Context, args map[string]any) (*schema.ToolResult, error) {
	commandVal, ok := args["command"]
	if !ok {
		return &schema.ToolResult{
			Success: false,
			Error:   "missing required parameter: command",
		}, fmt.Errorf("missing required parameter: command")
	}

	command := fmt.Sprintf("%v", commandVal)

	// Get timeout
	timeout := 30 * time.Second
	if timeoutVal, ok := args["timeout_seconds"]; ok {
		if timeoutInt, ok := timeoutVal.(float64); ok {
			timeout = time.Duration(timeoutInt) * time.Second
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute command
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return &schema.ToolResult{
			Success: false,
			Output:  string(output),
			Error:   fmt.Sprintf("command failed: %v", err),
		}, err
	}

	return &schema.ToolResult{
		Success: true,
		Output:  string(output),
		Data: map[string]any{
			"command":  command,
			"exit_code": cmd.ProcessState.ExitCode(),
		},
	}, nil
}

// RequiresApproval returns true if the command is destructive
func (t *ShellCommandTool) RequiresApproval(args map[string]any) bool {
	commandVal, ok := args["command"]
	if !ok {
		return true // Require approval if no command specified
	}

	command := fmt.Sprintf("%v", commandVal)
	return isDestructiveCommand(command)
}

// isDestructiveCommand checks if a command is destructive
func isDestructiveCommand(command string) bool {
	commandLower := strings.ToLower(command)

	for _, destructive := range destructiveCommands {
		if strings.Contains(commandLower, destructive) {
			return true
		}
	}

	// Check for dangerous redirects
	if strings.Contains(command, ">") && strings.Contains(command, "/dev/") {
		return true
	}

	// Check for force flags
	if strings.Contains(command, "-f") || strings.Contains(command, "--force") {
		return true
	}

	return false
}
