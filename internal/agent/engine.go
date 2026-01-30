package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/siddharth-bhatnagar/anvil/internal/llm"
	"github.com/siddharth-bhatnagar/anvil/internal/tools"
	"github.com/siddharth-bhatnagar/anvil/pkg/schema"
)

// Agent orchestrates LLM interactions and tool execution
type Agent struct {
	llmClient      llm.Client
	toolRegistry   *tools.Registry
	systemPrompt   string
	context        *Context
	lifecycle      *Lifecycle
	teachingConfig TeachingConfig
}

// Config holds agent configuration
type Config struct {
	SystemPrompt   string
	Model          string
	Temperature    float64
	MaxTokens      int
	TeachingMode   TeachingMode
}

// NewAgent creates a new agent with the given configuration
func NewAgent(llmClient llm.Client, toolRegistry *tools.Registry, config Config) *Agent {
	return &Agent{
		llmClient:      llmClient,
		toolRegistry:   toolRegistry,
		systemPrompt:   config.SystemPrompt,
		context:        NewContext(),
		lifecycle:      NewLifecycle(),
		teachingConfig: TeachingConfigForMode(config.TeachingMode),
	}
}

// SetTeachingMode sets the teaching mode
func (a *Agent) SetTeachingMode(mode TeachingMode) {
	a.teachingConfig = TeachingConfigForMode(mode)
}

// GetTeachingMode returns the current teaching mode
func (a *Agent) GetTeachingMode() TeachingMode {
	return a.teachingConfig.Mode
}

// SetTeachingConfig sets the teaching configuration
func (a *Agent) SetTeachingConfig(config TeachingConfig) {
	a.teachingConfig = config
}

// GetTeachingConfig returns the current teaching configuration
func (a *Agent) GetTeachingConfig() TeachingConfig {
	return a.teachingConfig
}

// AskWhy asks for an explanation of a change
func (a *Agent) AskWhy(ctx context.Context, change, changeContext string) (*Response, error) {
	question := WhyQuestion(change, changeContext)
	return a.ProcessRequest(ctx, question)
}

// ExplainConcept asks for an explanation of a concept
func (a *Agent) ExplainConcept(ctx context.Context, concept, relatedCode string) (*Response, error) {
	prompt := ConceptExplanation(concept, relatedCode)
	return a.ProcessRequest(ctx, prompt)
}

// ReviewCode asks for a code review with explanation
func (a *Agent) ReviewCode(ctx context.Context, code string, concerns []string) (*Response, error) {
	prompt := CodeReviewExplanation(code, concerns)
	return a.ProcessRequest(ctx, prompt)
}

// ProcessRequest handles a user request through the agent loop
// Returns the assistant's response and any tool results that require approval
func (a *Agent) ProcessRequest(ctx context.Context, userMessage string) (*Response, error) {
	// Start in Understand phase
	a.lifecycle.SetPhase(PhaseUnderstand)

	// Add user message to context
	a.context.AddMessage(llm.Message{
		Role:    llm.RoleUser,
		Content: userMessage,
	})

	// Run through the lifecycle phases
	return a.runLifecycle(ctx)
}

// runLifecycle executes the Understand → Plan → Act → Verify cycle
func (a *Agent) runLifecycle(ctx context.Context) (*Response, error) {
	var response Response

	for {
		phase := a.lifecycle.CurrentPhase()
		response.Phase = phase

		switch phase {
		case PhaseUnderstand:
			// Gather information about the request
			resp, err := a.understand(ctx)
			if err != nil {
				response.Error = err
				return &response, err
			}
			response.Message = resp.Message
			response.ToolCalls = append(response.ToolCalls, resp.ToolCalls...)
			response.ToolResults = append(response.ToolResults, resp.ToolResults...)

			if resp.RequiresApproval {
				response.RequiresApproval = true
				response.PendingApprovals = resp.PendingApprovals
				return &response, nil
			}

			// Check if we have a plan in the response
			if a.detectPlan(resp.Message) {
				a.lifecycle.NextPhase() // Move to Plan
			} else if a.detectAction(resp.Message) {
				a.lifecycle.SetPhase(PhaseAct) // Skip to Act if direct action
			} else {
				// Simple response, no planning needed
				response.Done = true
				a.lifecycle.SetPhase(PhaseVerify)
			}

		case PhasePlan:
			// Extract and set the plan steps
			steps := a.extractPlanSteps(response.Message)
			if len(steps) > 0 {
				a.lifecycle.SetPlan(steps)
				response.PlanSteps = a.lifecycle.GetPlan()
			}
			a.lifecycle.NextPhase() // Move to Act

		case PhaseAct:
			// Execute the plan steps
			resp, err := a.act(ctx)
			if err != nil {
				response.Error = err
				return &response, err
			}
			response.Message = resp.Message
			response.ToolCalls = append(response.ToolCalls, resp.ToolCalls...)
			response.ToolResults = append(response.ToolResults, resp.ToolResults...)
			response.PlanSteps = a.lifecycle.GetPlan()

			if resp.RequiresApproval {
				response.RequiresApproval = true
				response.PendingApprovals = resp.PendingApprovals
				return &response, nil
			}

			if resp.Done || a.lifecycle.AllStepsCompleted() {
				a.lifecycle.NextPhase() // Move to Verify
			} else {
				// Continue acting
				continue
			}

		case PhaseVerify:
			// Verify the results
			resp, err := a.verify(ctx)
			if err != nil {
				response.Error = err
				return &response, err
			}
			response.Message = resp.Message
			response.Done = true
			return &response, nil
		}
	}
}

// understand gathers information about the request
func (a *Agent) understand(ctx context.Context) (*Response, error) {
	return a.loop(ctx)
}

// act executes the planned actions
func (a *Agent) act(ctx context.Context) (*Response, error) {
	// Start the next step if available
	step := a.lifecycle.StartNextStep()
	if step != nil {
		// Add step context to the conversation
		a.context.AddMessage(llm.Message{
			Role:    llm.RoleUser,
			Content: fmt.Sprintf("Execute step %d: %s", step.ID+1, step.Description),
		})
	}

	resp, err := a.loop(ctx)
	if err != nil {
		if step != nil {
			a.lifecycle.FailCurrentStep(err)
		}
		return resp, err
	}

	if step != nil {
		a.lifecycle.CompleteCurrentStep(resp.Message)
	}

	return resp, nil
}

// verify confirms the results of the actions
func (a *Agent) verify(ctx context.Context) (*Response, error) {
	// Add verification prompt
	a.context.AddMessage(llm.Message{
		Role:    llm.RoleUser,
		Content: "Please verify the changes made and confirm everything is working correctly.",
	})

	return a.loop(ctx)
}

// detectPlan checks if the response contains a plan
func (a *Agent) detectPlan(content string) bool {
	planIndicators := []string{
		"here's my plan",
		"i will:",
		"steps:",
		"plan:",
		"1.",
		"step 1:",
		"first,",
	}

	contentLower := strings.ToLower(content)
	for _, indicator := range planIndicators {
		if strings.Contains(contentLower, indicator) {
			return true
		}
	}
	return false
}

// detectAction checks if the response indicates direct action
func (a *Agent) detectAction(content string) bool {
	actionIndicators := []string{
		"<tool_use>",
		"let me",
		"i'll read",
		"i'll check",
		"reading",
		"checking",
	}

	contentLower := strings.ToLower(content)
	for _, indicator := range actionIndicators {
		if strings.Contains(contentLower, indicator) {
			return true
		}
	}
	return false
}

// extractPlanSteps parses plan steps from the message
func (a *Agent) extractPlanSteps(content string) []string {
	var steps []string
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for numbered steps (1., 2., etc.)
		if len(line) > 2 && line[0] >= '1' && line[0] <= '9' && line[1] == '.' {
			step := strings.TrimSpace(line[2:])
			if step != "" {
				steps = append(steps, step)
			}
		}

		// Look for bullet points with "Step N:"
		if strings.HasPrefix(strings.ToLower(line), "step ") {
			colonIdx := strings.Index(line, ":")
			if colonIdx > 0 && colonIdx < len(line)-1 {
				step := strings.TrimSpace(line[colonIdx+1:])
				if step != "" {
					steps = append(steps, step)
				}
			}
		}
	}

	return steps
}

// Response represents the agent's response to a request
type Response struct {
	Message          string                // The assistant's text response
	ToolCalls        []schema.ToolCall     // Tools that were called
	ToolResults      []schema.ToolResult   // Results from tool execution
	RequiresApproval bool                  // Whether any tools require approval
	PendingApprovals []PendingApproval     // Tools waiting for approval
	Done             bool                  // Whether processing is complete
	Error            error                 // Any error that occurred
	Phase            Phase                 // Current lifecycle phase
	PlanSteps        []PlanStep            // Plan steps if in planning phase
}

// PendingApproval represents a tool call waiting for user approval
type PendingApproval struct {
	ToolCall schema.ToolCall
	Request  schema.ApprovalRequest
}

// loop implements the core agent request/response loop
func (a *Agent) loop(ctx context.Context) (*Response, error) {
	var response Response
	maxIterations := 10 // Prevent infinite loops

	for iteration := 0; iteration < maxIterations; iteration++ {
		// Prepare LLM request with conversation history
		llmReq := a.prepareLLMRequest()

		// Call LLM
		llmResp, err := a.llmClient.Complete(ctx, llmReq)
		if err != nil {
			response.Error = fmt.Errorf("LLM request failed: %w", err)
			return &response, err
		}

		// Add assistant message to context
		a.context.AddMessage(llm.Message{
			Role:    llm.RoleAssistant,
			Content: llmResp.Content,
		})

		// Update response message
		response.Message = llmResp.Content

		// Check if there are tool calls in the response
		toolCalls := a.extractToolCalls(llmResp.Content)
		if len(toolCalls) == 0 {
			// No tool calls, we're done
			response.Done = true
			return &response, nil
		}

		// Execute tools
		var pendingApprovals []PendingApproval
		var executedResults []schema.ToolResult

		for _, toolCall := range toolCalls {
			result, err := a.toolRegistry.Execute(ctx, toolCall)
			if err != nil {
				// Tool execution failed, add error to context and continue
				a.context.AddMessage(llm.Message{
					Role:    llm.RoleUser,
					Content: fmt.Sprintf("Tool %s failed: %v", toolCall.Name, err),
				})
				continue
			}

			// Check if approval is required
			if result.Approval != nil {
				pendingApprovals = append(pendingApprovals, PendingApproval{
					ToolCall: toolCall,
					Request:  *result.Approval,
				})
			} else {
				// Tool executed successfully
				executedResults = append(executedResults, *result)

				// Add tool result to context
				a.context.AddMessage(llm.Message{
					Role:    llm.RoleUser,
					Content: fmt.Sprintf("Tool %s result:\n%s", toolCall.Name, result.Output),
				})
			}
		}

		response.ToolCalls = toolCalls
		response.ToolResults = executedResults

		// If there are pending approvals, return and wait for user
		if len(pendingApprovals) > 0 {
			response.RequiresApproval = true
			response.PendingApprovals = pendingApprovals
			response.Done = false
			return &response, nil
		}

		// If all tools executed, continue the loop to let LLM process results
		if len(executedResults) > 0 {
			continue
		}

		// No tools executed and no pending approvals, we're done
		response.Done = true
		return &response, nil
	}

	// Hit max iterations
	response.Error = fmt.Errorf("agent loop exceeded maximum iterations")
	response.Done = true
	return &response, response.Error
}

// ApproveToolCall executes a tool that was previously pending approval
func (a *Agent) ApproveToolCall(ctx context.Context, toolCall schema.ToolCall) (*schema.ToolResult, error) {
	// Get the tool
	tool, err := a.toolRegistry.Get(toolCall.Name)
	if err != nil {
		return nil, fmt.Errorf("tool not found: %w", err)
	}

	// Execute the tool directly (bypassing approval check)
	result, err := tool.Execute(ctx, toolCall.Arguments)
	if err != nil {
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	result.ToolCallID = toolCall.ID

	// Add tool result to context
	a.context.AddMessage(llm.Message{
		Role:    llm.RoleUser,
		Content: fmt.Sprintf("Tool %s approved and executed:\n%s", toolCall.Name, result.Output),
	})

	return result, nil
}

// RejectToolCall rejects a tool that was pending approval
func (a *Agent) RejectToolCall(toolCall schema.ToolCall, reason string) {
	// Add rejection to context
	a.context.AddMessage(llm.Message{
		Role:    llm.RoleUser,
		Content: fmt.Sprintf("Tool %s rejected: %s", toolCall.Name, reason),
	})
}

// ContinueAfterApproval continues the agent loop after approvals/rejections
func (a *Agent) ContinueAfterApproval(ctx context.Context) (*Response, error) {
	// Continue from the current lifecycle phase
	return a.runLifecycle(ctx)
}

// prepareLLMRequest creates an LLM request from the current context
func (a *Agent) prepareLLMRequest() llm.Request {
	messages := a.context.GetMessages()

	// Build system prompt with teaching additions
	systemPrompt := a.systemPrompt
	if teachingAddition := GetTeachingPromptAddition(a.teachingConfig); teachingAddition != "" {
		systemPrompt += teachingAddition
	}

	// Add system prompt if provided
	if systemPrompt != "" {
		messages = append([]llm.Message{
			{
				Role:    llm.RoleSystem,
				Content: systemPrompt,
			},
		}, messages...)
	}

	// Add tool definitions
	toolDefs := a.toolRegistry.ListDefinitions()

	return llm.Request{
		Messages:    messages,
		Tools:       convertToolDefinitions(toolDefs),
		MaxTokens:   4096,
		Temperature: 0.7,
	}
}

// extractToolCalls parses tool calls from the LLM response
// This is a simplified implementation - real tool calls would use structured output
func (a *Agent) extractToolCalls(content string) []schema.ToolCall {
	var toolCalls []schema.ToolCall

	// Look for tool call markers in the format:
	// <tool_use>
	// {
	//   "name": "tool_name",
	//   "arguments": {...}
	// }
	// </tool_use>

	// For now, we'll use a simple approach
	// In a production system, this would parse the LLM's structured tool call format

	start := 0
	for {
		// Find tool_use tag
		startTag := "<tool_use>"
		endTag := "</tool_use>"

		startIdx := strings.Index(content[start:], startTag)
		if startIdx == -1 {
			break
		}
		startIdx += start

		endIdx := strings.Index(content[startIdx:], endTag)
		if endIdx == -1 {
			break
		}
		endIdx += startIdx

		// Extract JSON between tags
		jsonStr := content[startIdx+len(startTag) : endIdx]
		jsonStr = strings.TrimSpace(jsonStr)

		// Parse the tool call
		var toolCall struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}

		if err := json.Unmarshal([]byte(jsonStr), &toolCall); err == nil {
			toolCalls = append(toolCalls, schema.ToolCall{
				ID:        fmt.Sprintf("call_%d", len(toolCalls)),
				Name:      toolCall.Name,
				Arguments: toolCall.Arguments,
			})
		}

		start = endIdx + len(endTag)
	}

	return toolCalls
}

// convertToolDefinitions converts schema.ToolDefinition to llm.Tool
func convertToolDefinitions(defs []schema.ToolDefinition) []llm.Tool {
	tools := make([]llm.Tool, len(defs))
	for i, def := range defs {
		tools[i] = llm.Tool{
			Name:        def.Name,
			Description: def.Description,
			InputSchema: convertParameters(def.Parameters),
		}
	}
	return tools
}

// convertParameters converts []schema.ToolParameter to map[string]any
func convertParameters(params []schema.ToolParameter) map[string]any {
	properties := make(map[string]any)
	var required []string

	for _, param := range params {
		prop := map[string]any{
			"type":        param.Type,
			"description": param.Description,
		}

		if param.Default != nil {
			prop["default"] = param.Default
		}

		properties[param.Name] = prop

		if param.Required {
			required = append(required, param.Name)
		}
	}

	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

// Reset clears the agent's context and state
func (a *Agent) Reset() {
	a.context = NewContext()
	a.lifecycle.Reset()
}

// GetContext returns the current conversation context
func (a *Agent) GetContext() *Context {
	return a.context
}

// GetLifecycle returns the lifecycle manager
func (a *Agent) GetLifecycle() *Lifecycle {
	return a.lifecycle
}
