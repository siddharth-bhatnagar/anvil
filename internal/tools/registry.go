package tools

import (
	"context"
	"fmt"
	"sync"

	"github.com/siddharth-bhatnagar/anvil/pkg/schema"
)

// Registry manages available tools
type Registry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register registers a tool
func (r *Registry) Register(tool Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := tool.Name()
	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool %s is already registered", name)
	}

	r.tools[name] = tool
	return nil
}

// Get retrieves a tool by name
func (r *Registry) Get(name string) (Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool %s not found", name)
	}

	return tool, nil
}

// List returns all registered tools
func (r *Registry) List() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}

	return tools
}

// ListDefinitions returns tool definitions for all registered tools
func (r *Registry) ListDefinitions() []schema.ToolDefinition {
	tools := r.List()
	defs := make([]schema.ToolDefinition, len(tools))

	for i, tool := range tools {
		defs[i] = tool.Definition()
	}

	return defs
}

// Execute executes a tool by name with the given arguments
func (r *Registry) Execute(ctx context.Context, toolCall schema.ToolCall) (*schema.ToolResult, error) {
	tool, err := r.Get(toolCall.Name)
	if err != nil {
		return &schema.ToolResult{
			ToolCallID: toolCall.ID,
			Success:    false,
			Error:      err.Error(),
		}, err
	}

	// Check if approval is required
	if tool.RequiresApproval(toolCall.Arguments) {
		// Return result with approval request
		// The caller will handle the approval flow
		return &schema.ToolResult{
			ToolCallID: toolCall.ID,
			Success:    false,
			Output:     "Approval required",
			Approval: &schema.ApprovalRequest{
				Action:      fmt.Sprintf("Execute %s", tool.Name()),
				Reason:      "This operation requires user approval",
				Destructive: true,
			},
		}, nil
	}

	// Execute the tool
	result, err := tool.Execute(ctx, toolCall.Arguments)
	if err != nil {
		return &schema.ToolResult{
			ToolCallID: toolCall.ID,
			Success:    false,
			Error:      err.Error(),
		}, err
	}

	result.ToolCallID = toolCall.ID
	return result, nil
}

// DefaultRegistry returns a registry with all default tools registered
func DefaultRegistry() (*Registry, error) {
	registry := NewRegistry()

	// Register file system tools
	if err := registry.Register(NewReadFileTool()); err != nil {
		return nil, err
	}

	if err := registry.Register(NewWriteFileTool()); err != nil {
		return nil, err
	}

	if err := registry.Register(NewSearchFilesTool()); err != nil {
		return nil, err
	}

	if err := registry.Register(NewGrepFilesTool()); err != nil {
		return nil, err
	}

	if err := registry.Register(NewListDirectoryTool()); err != nil {
		return nil, err
	}

	// Register git tools
	if err := registry.Register(NewGitStatusTool()); err != nil {
		return nil, err
	}

	if err := registry.Register(NewGitDiffTool()); err != nil {
		return nil, err
	}

	if err := registry.Register(NewGitLogTool()); err != nil {
		return nil, err
	}

	// Register shell tool
	if err := registry.Register(NewShellCommandTool()); err != nil {
		return nil, err
	}

	return registry, nil
}
