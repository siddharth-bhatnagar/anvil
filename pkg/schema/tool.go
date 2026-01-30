package schema

// ToolParameter represents a parameter for a tool
type ToolParameter struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"` // "string", "number", "boolean", "array", "object"
	Required    bool   `json:"required"`
	Default     any    `json:"default,omitempty"`
}

// ToolDefinition represents the schema of a tool
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  []ToolParameter `json:"parameters"`
}

// ToolCall represents a call to a tool with specific arguments
type ToolCall struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	ToolCallID string          `json:"tool_call_id"`
	Success    bool            `json:"success"`
	Output     string          `json:"output"`
	Error      string          `json:"error,omitempty"`
	Data       map[string]any  `json:"data,omitempty"` // Structured data
	Approval   *ApprovalRequest `json:"approval,omitempty"`
}

// ApprovalRequest represents a request for user approval
type ApprovalRequest struct {
	Action      string `json:"action"`       // Description of the action
	Reason      string `json:"reason"`       // Why approval is needed
	Destructive bool   `json:"destructive"`  // Whether this is a destructive operation
	Preview     string `json:"preview,omitempty"` // Preview of changes (e.g., diff)
}
