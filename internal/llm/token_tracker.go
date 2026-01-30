package llm

import (
	"sync"
	"time"
)

// TokenTracker tracks token usage across requests
type TokenTracker struct {
	mu                sync.RWMutex
	totalPromptTokens     int
	totalCompletionTokens int
	totalTokens           int
	requestCount          int
	sessionStart          time.Time
}

// NewTokenTracker creates a new token tracker
func NewTokenTracker() *TokenTracker {
	return &TokenTracker{
		sessionStart: time.Now(),
	}
}

// AddUsage adds token usage from a response
func (t *TokenTracker) AddUsage(usage Usage) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.totalPromptTokens += usage.PromptTokens
	t.totalCompletionTokens += usage.CompletionTokens
	t.totalTokens += usage.TotalTokens
	t.requestCount++
}

// GetStats returns current usage statistics
func (t *TokenTracker) GetStats() TokenStats {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return TokenStats{
		TotalPromptTokens:     t.totalPromptTokens,
		TotalCompletionTokens: t.totalCompletionTokens,
		TotalTokens:           t.totalTokens,
		RequestCount:          t.requestCount,
		SessionDuration:       time.Since(t.sessionStart),
	}
}

// Reset resets all counters
func (t *TokenTracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.totalPromptTokens = 0
	t.totalCompletionTokens = 0
	t.totalTokens = 0
	t.requestCount = 0
	t.sessionStart = time.Now()
}

// TokenStats holds token usage statistics
type TokenStats struct {
	TotalPromptTokens     int
	TotalCompletionTokens int
	TotalTokens           int
	RequestCount          int
	SessionDuration       time.Duration
}

// EstimatedCost estimates the cost based on model pricing
// Returns cost in USD (approximate)
func (s TokenStats) EstimatedCost(model string) float64 {
	// Approximate pricing (as of 2026)
	// These would ideally come from a configuration file
	var inputCostPer1M, outputCostPer1M float64

	switch model {
	case "claude-sonnet-4-5", "claude-sonnet-4", "claude-sonnet-3-5":
		inputCostPer1M = 3.00
		outputCostPer1M = 15.00
	case "claude-opus-4-5", "claude-opus-4":
		inputCostPer1M = 15.00
		outputCostPer1M = 75.00
	case "claude-haiku-4", "claude-haiku-3-5":
		inputCostPer1M = 0.25
		outputCostPer1M = 1.25
	case "gpt-4-turbo", "gpt-4":
		inputCostPer1M = 10.00
		outputCostPer1M = 30.00
	case "gpt-3.5-turbo":
		inputCostPer1M = 0.50
		outputCostPer1M = 1.50
	default:
		// Unknown model, return 0
		return 0
	}

	inputCost := (float64(s.TotalPromptTokens) / 1_000_000) * inputCostPer1M
	outputCost := (float64(s.TotalCompletionTokens) / 1_000_000) * outputCostPer1M

	return inputCost + outputCost
}

// FormatStats returns a human-readable string of the stats
func (s TokenStats) FormatStats() string {
	return FormatTokenCount(s.TotalTokens)
}

// FormatTokenCount formats a token count with K/M suffixes
func FormatTokenCount(tokens int) string {
	if tokens < 1000 {
		return string(rune(tokens))
	} else if tokens < 1_000_000 {
		return string(rune(tokens/1000)) + "K"
	} else {
		return string(rune(tokens/1_000_000)) + "M"
	}
}
