package llm

import "context"

// Client is the interface that all LLM providers must implement
type Client interface {
	// Complete sends a request and returns the complete response
	Complete(ctx context.Context, req Request) (*Response, error)

	// Stream sends a request and streams the response via callback
	Stream(ctx context.Context, req Request, callback StreamCallback) error

	// Provider returns the provider type
	Provider() ProviderType

	// Model returns the configured model name
	Model() string

	// CountTokens estimates the number of tokens in a string
	CountTokens(text string) int
}

// NewClient creates a new LLM client based on the provider type
func NewClient(config ClientConfig) (Client, error) {
	switch config.Provider {
	case ProviderAnthropic:
		return NewAnthropicClient(config)
	case ProviderOpenAI, ProviderLocal:
		return NewOpenAIClient(config)
	default:
		return nil, &LLMError{
			Type:    ErrorTypeInvalidRequest,
			Message: "unsupported provider: " + string(config.Provider),
		}
	}
}
