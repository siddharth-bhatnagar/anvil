package agent

import (
	"fmt"
	"strings"
)

// TeachingMode represents the level of explanation detail
type TeachingMode int

const (
	// TeachingOff provides concise responses (default)
	TeachingOff TeachingMode = iota
	// TeachingBasic provides brief explanations
	TeachingBasic
	// TeachingDetailed provides comprehensive explanations
	TeachingDetailed
	// TeachingExpert provides deep technical explanations
	TeachingExpert
)

// String returns the string representation of a teaching mode
func (t TeachingMode) String() string {
	switch t {
	case TeachingOff:
		return "off"
	case TeachingBasic:
		return "basic"
	case TeachingDetailed:
		return "detailed"
	case TeachingExpert:
		return "expert"
	default:
		return "unknown"
	}
}

// Description returns a description of the teaching mode
func (t TeachingMode) Description() string {
	switch t {
	case TeachingOff:
		return "Concise responses, minimal explanations"
	case TeachingBasic:
		return "Brief explanations of key concepts"
	case TeachingDetailed:
		return "Comprehensive explanations with examples"
	case TeachingExpert:
		return "Deep technical explanations for advanced users"
	default:
		return ""
	}
}

// ParseTeachingMode parses a string to TeachingMode
func ParseTeachingMode(s string) TeachingMode {
	switch strings.ToLower(s) {
	case "off", "none", "0":
		return TeachingOff
	case "basic", "brief", "1":
		return TeachingBasic
	case "detailed", "full", "2":
		return TeachingDetailed
	case "expert", "advanced", "3":
		return TeachingExpert
	default:
		return TeachingOff
	}
}

// TeachingConfig holds teaching mode configuration
type TeachingConfig struct {
	Mode             TeachingMode
	ExplainReasoning bool // Explain why decisions were made
	ShowAlternatives bool // Show alternative approaches
	ProvideExamples  bool // Provide code examples
	LinkDocs         bool // Link to relevant documentation
}

// DefaultTeachingConfig returns the default teaching configuration
func DefaultTeachingConfig() TeachingConfig {
	return TeachingConfig{
		Mode:             TeachingOff,
		ExplainReasoning: false,
		ShowAlternatives: false,
		ProvideExamples:  false,
		LinkDocs:         false,
	}
}

// TeachingConfigForMode returns a teaching config for a given mode
func TeachingConfigForMode(mode TeachingMode) TeachingConfig {
	switch mode {
	case TeachingOff:
		return DefaultTeachingConfig()
	case TeachingBasic:
		return TeachingConfig{
			Mode:             TeachingBasic,
			ExplainReasoning: true,
			ShowAlternatives: false,
			ProvideExamples:  false,
			LinkDocs:         false,
		}
	case TeachingDetailed:
		return TeachingConfig{
			Mode:             TeachingDetailed,
			ExplainReasoning: true,
			ShowAlternatives: true,
			ProvideExamples:  true,
			LinkDocs:         false,
		}
	case TeachingExpert:
		return TeachingConfig{
			Mode:             TeachingExpert,
			ExplainReasoning: true,
			ShowAlternatives: true,
			ProvideExamples:  true,
			LinkDocs:         true,
		}
	default:
		return DefaultTeachingConfig()
	}
}

// GetTeachingPromptAddition returns additional prompt text for teaching mode
func GetTeachingPromptAddition(config TeachingConfig) string {
	if config.Mode == TeachingOff {
		return ""
	}

	var parts []string

	parts = append(parts, "\n## Teaching Mode Active\n")

	switch config.Mode {
	case TeachingBasic:
		parts = append(parts, "You are in teaching mode. Please provide brief explanations of your actions and key concepts.")
	case TeachingDetailed:
		parts = append(parts, "You are in detailed teaching mode. Please provide comprehensive explanations with examples.")
	case TeachingExpert:
		parts = append(parts, "You are in expert teaching mode. Please provide deep technical explanations suitable for advanced developers.")
	}

	parts = append(parts, "\nWhen responding:")

	if config.ExplainReasoning {
		parts = append(parts, "- Explain your reasoning and decision-making process")
		parts = append(parts, "- Describe why you chose a particular approach")
	}

	if config.ShowAlternatives {
		parts = append(parts, "- Mention alternative approaches when relevant")
		parts = append(parts, "- Discuss trade-offs between different solutions")
	}

	if config.ProvideExamples {
		parts = append(parts, "- Provide concrete code examples to illustrate concepts")
		parts = append(parts, "- Show before/after comparisons when making changes")
	}

	if config.LinkDocs {
		parts = append(parts, "- Reference relevant documentation or resources")
		parts = append(parts, "- Suggest further reading for deeper understanding")
	}

	return strings.Join(parts, "\n")
}

// ExplanationRequest represents a request for explanation
type ExplanationRequest struct {
	Topic       string           // What to explain
	Context     string           // Additional context
	Depth       TeachingMode     // Depth of explanation
	CodeSnippet string           // Related code if any
	Question    string           // Specific question to answer
}

// FormatExplanationPrompt formats an explanation request into a prompt
func FormatExplanationPrompt(req ExplanationRequest) string {
	var parts []string

	parts = append(parts, "Please explain the following:\n")

	if req.Topic != "" {
		parts = append(parts, fmt.Sprintf("Topic: %s\n", req.Topic))
	}

	if req.Question != "" {
		parts = append(parts, fmt.Sprintf("Question: %s\n", req.Question))
	}

	if req.Context != "" {
		parts = append(parts, fmt.Sprintf("Context: %s\n", req.Context))
	}

	if req.CodeSnippet != "" {
		parts = append(parts, fmt.Sprintf("\nRelevant code:\n```\n%s\n```\n", req.CodeSnippet))
	}

	// Add depth-specific instructions
	switch req.Depth {
	case TeachingBasic:
		parts = append(parts, "\nPlease provide a brief, beginner-friendly explanation.")
	case TeachingDetailed:
		parts = append(parts, "\nPlease provide a detailed explanation with examples.")
	case TeachingExpert:
		parts = append(parts, "\nPlease provide an in-depth technical explanation suitable for experienced developers.")
	}

	return strings.Join(parts, "")
}

// WhyQuestion formats a "why" question about a change
func WhyQuestion(change, context string) string {
	return fmt.Sprintf(
		"I'd like to understand the reasoning behind this change:\n\n"+
			"Change: %s\n\n"+
			"Context: %s\n\n"+
			"Please explain:\n"+
			"1. Why was this change made?\n"+
			"2. What problem does it solve?\n"+
			"3. Were there alternative approaches considered?\n"+
			"4. What are the implications of this change?",
		change, context,
	)
}

// ConceptExplanation formats a concept explanation request
func ConceptExplanation(concept string, relatedCode string) string {
	prompt := fmt.Sprintf("Please explain the concept of '%s'", concept)

	if relatedCode != "" {
		prompt += fmt.Sprintf(" in the context of this code:\n\n```\n%s\n```", relatedCode)
	}

	prompt += "\n\nInclude:\n" +
		"- A clear definition\n" +
		"- Why it's important\n" +
		"- Common use cases\n" +
		"- Best practices"

	return prompt
}

// CodeReviewExplanation formats a code review explanation
func CodeReviewExplanation(code string, issues []string) string {
	var prompt strings.Builder

	prompt.WriteString("Please review and explain the following code:\n\n")
	prompt.WriteString(fmt.Sprintf("```\n%s\n```\n\n", code))

	if len(issues) > 0 {
		prompt.WriteString("Specific concerns:\n")
		for _, issue := range issues {
			prompt.WriteString(fmt.Sprintf("- %s\n", issue))
		}
		prompt.WriteString("\n")
	}

	prompt.WriteString("Please provide:\n")
	prompt.WriteString("1. An overview of what this code does\n")
	prompt.WriteString("2. Analysis of potential issues\n")
	prompt.WriteString("3. Suggestions for improvement\n")
	prompt.WriteString("4. Best practices that apply here")

	return prompt.String()
}
