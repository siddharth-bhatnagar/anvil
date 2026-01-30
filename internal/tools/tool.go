package tools

import (
	"context"

	"github.com/siddharth-bhatnagar/anvil/pkg/schema"
)

// Tool represents a tool that can be executed
type Tool interface {
	// Name returns the tool name
	Name() string

	// Description returns the tool description
	Description() string

	// Definition returns the tool schema definition
	Definition() schema.ToolDefinition

	// Execute runs the tool with the given arguments
	Execute(ctx context.Context, args map[string]any) (*schema.ToolResult, error)

	// RequiresApproval returns true if this tool requires user approval
	RequiresApproval(args map[string]any) bool
}

// BaseTool provides common functionality for tools
type BaseTool struct {
	name        string
	description string
	parameters  []schema.ToolParameter
}

// NewBaseTool creates a new base tool
func NewBaseTool(name, description string, parameters []schema.ToolParameter) BaseTool {
	return BaseTool{
		name:        name,
		description: description,
		parameters:  parameters,
	}
}

// Name returns the tool name
func (t BaseTool) Name() string {
	return t.name
}

// Description returns the tool description
func (t BaseTool) Description() string {
	return t.description
}

// Definition returns the tool schema definition
func (t BaseTool) Definition() schema.ToolDefinition {
	return schema.ToolDefinition{
		Name:        t.name,
		Description: t.description,
		Parameters:  t.parameters,
	}
}
