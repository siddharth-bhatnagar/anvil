package llm

import "time"

// Role represents the role of a message sender
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
)

// Message represents a conversation message
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// Request represents a request to an LLM
type Request struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
	Stream      bool      `json:"stream"`
	SystemPrompt string   `json:"-"` // Handled differently by providers
}

// Response represents a response from an LLM
type Response struct {
	Content      string
	Role         Role
	FinishReason string
	Usage        Usage
	Model        string
	Error        error
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// StreamEvent represents a streaming event
type StreamEvent struct {
	Delta        string    // The incremental text
	Done         bool      // Whether the stream is complete
	Error        error     // Error if any
	Usage        *Usage    // Final usage (only on last event)
	FinishReason string    // Reason for finishing (only on last event)
	Timestamp    time.Time // When this event occurred
}

// StreamCallback is called for each streaming event
type StreamCallback func(event StreamEvent)

// ProviderType represents the LLM provider
type ProviderType string

const (
	ProviderAnthropic ProviderType = "anthropic"
	ProviderOpenAI    ProviderType = "openai"
	ProviderLocal     ProviderType = "local" // Ollama, LM Studio, etc.
)

// ClientConfig holds configuration for an LLM client
type ClientConfig struct {
	Provider    ProviderType
	APIKey      string
	BaseURL     string // For custom endpoints
	Model       string // Default model
	MaxRetries  int    // Max retry attempts
	Timeout     time.Duration
	Temperature float64
	MaxTokens   int
}

// Error types
type ErrorType string

const (
	ErrorTypeAuth         ErrorType = "auth"          // Authentication failed
	ErrorTypeRateLimit    ErrorType = "rate_limit"    // Rate limit exceeded
	ErrorTypeInvalidRequest ErrorType = "invalid_request" // Invalid request
	ErrorTypeTimeout      ErrorType = "timeout"       // Request timeout
	ErrorTypeNetwork      ErrorType = "network"       // Network error
	ErrorTypeServer       ErrorType = "server"        // Server error
	ErrorTypeUnknown      ErrorType = "unknown"       // Unknown error
)

// LLMError represents an error from the LLM provider
type LLMError struct {
	Type    ErrorType
	Message string
	Code    int    // HTTP status code
	Details string // Additional details
}

func (e *LLMError) Error() string {
	return e.Message
}
