package llm

import (
	"context"
	"math"
	"time"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Multiplier     float64
}

// DefaultRetryConfig returns sensible default retry settings
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		Multiplier:     2.0,
	}
}

// RetryableClient wraps a Client with retry logic
type RetryableClient struct {
	client Client
	config RetryConfig
}

// NewRetryableClient creates a client with retry logic
func NewRetryableClient(client Client, config RetryConfig) *RetryableClient {
	return &RetryableClient{
		client: client,
		config: config,
	}
}

// Complete sends a request with retry logic
func (r *RetryableClient) Complete(ctx context.Context, req Request) (*Response, error) {
	var lastErr error

	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		resp, err := r.client.Complete(ctx, req)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Check if error is retryable
		if !r.isRetryable(err) {
			return nil, err
		}

		// Don't sleep after the last attempt
		if attempt < r.config.MaxRetries {
			backoff := r.calculateBackoff(attempt)
			select {
			case <-time.After(backoff):
				// Continue to next attempt
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	return nil, lastErr
}

// Stream sends a streaming request with retry logic
func (r *RetryableClient) Stream(ctx context.Context, req Request, callback StreamCallback) error {
	var lastErr error

	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		err := r.client.Stream(ctx, req, callback)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !r.isRetryable(err) {
			return err
		}

		// Don't sleep after the last attempt
		if attempt < r.config.MaxRetries {
			backoff := r.calculateBackoff(attempt)
			select {
			case <-time.After(backoff):
				// Continue to next attempt
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return lastErr
}

// isRetryable determines if an error should trigger a retry
func (r *RetryableClient) isRetryable(err error) bool {
	if err == nil {
		return false
	}

	llmErr, ok := err.(*LLMError)
	if !ok {
		// Unknown error type, don't retry
		return false
	}

	switch llmErr.Type {
	case ErrorTypeRateLimit, ErrorTypeTimeout, ErrorTypeNetwork, ErrorTypeServer:
		// These errors are retryable
		return true
	case ErrorTypeAuth, ErrorTypeInvalidRequest:
		// These errors are not retryable
		return false
	default:
		return false
	}
}

// calculateBackoff calculates the backoff duration for a given attempt
func (r *RetryableClient) calculateBackoff(attempt int) time.Duration {
	backoff := float64(r.config.InitialBackoff) * math.Pow(r.config.Multiplier, float64(attempt))
	if backoff > float64(r.config.MaxBackoff) {
		backoff = float64(r.config.MaxBackoff)
	}
	return time.Duration(backoff)
}

// Provider returns the provider type
func (r *RetryableClient) Provider() ProviderType {
	return r.client.Provider()
}

// Model returns the configured model
func (r *RetryableClient) Model() string {
	return r.client.Model()
}

// CountTokens estimates token count
func (r *RetryableClient) CountTokens(text string) int {
	return r.client.CountTokens(text)
}
