package llm

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestNewClient tests client creation for different providers
func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		config      ClientConfig
		wantErr     bool
		wantErrType ErrorType
	}{
		{
			name: "Anthropic client creation",
			config: ClientConfig{
				Provider: ProviderAnthropic,
				APIKey:   "test-key",
				Model:    "claude-sonnet-4",
			},
			wantErr: false,
		},
		{
			name: "OpenAI client creation",
			config: ClientConfig{
				Provider: ProviderOpenAI,
				APIKey:   "test-key",
				Model:    "gpt-4",
			},
			wantErr: false,
		},
		{
			name: "Unsupported provider",
			config: ClientConfig{
				Provider: "unsupported",
				APIKey:   "test-key",
			},
			wantErr:     true,
			wantErrType: ErrorTypeInvalidRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if llmErr, ok := err.(*LLMError); ok {
					if llmErr.Type != tt.wantErrType {
						t.Errorf("expected error type %s, got %s", tt.wantErrType, llmErr.Type)
					}
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if client == nil {
					t.Fatal("expected client, got nil")
				}
			}
		})
	}
}

// TestAnthropicClientCreation tests Anthropic client initialization
func TestAnthropicClientCreation(t *testing.T) {
	tests := []struct {
		name        string
		config      ClientConfig
		wantErr     bool
		wantErrType ErrorType
	}{
		{
			name: "valid configuration",
			config: ClientConfig{
				APIKey:  "test-key",
				Model:   "claude-sonnet-4",
				BaseURL: "https://api.anthropic.com/v1/messages",
			},
			wantErr: false,
		},
		{
			name: "missing API key",
			config: ClientConfig{
				Model: "claude-sonnet-4",
			},
			wantErr:     true,
			wantErrType: ErrorTypeAuth,
		},
		{
			name: "defaults applied",
			config: ClientConfig{
				APIKey: "test-key",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewAnthropicClient(tt.config)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if llmErr, ok := err.(*LLMError); ok {
					if llmErr.Type != tt.wantErrType {
						t.Errorf("expected error type %s, got %s", tt.wantErrType, llmErr.Type)
					}
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if client == nil {
					t.Fatal("expected client, got nil")
				}
				// Check defaults were applied
				if client.config.BaseURL == "" {
					t.Error("expected BaseURL to be set")
				}
				if client.config.Timeout == 0 {
					t.Error("expected Timeout to be set")
				}
				if client.config.MaxRetries == 0 {
					t.Error("expected MaxRetries to be set")
				}
			}
		})
	}
}

// TestAnthropicClientMethods tests basic Anthropic client methods
func TestAnthropicClientMethods(t *testing.T) {
	client := &AnthropicClient{
		config: ClientConfig{
			Provider: ProviderAnthropic,
			Model:    "claude-sonnet-4",
		},
	}

	if client.Provider() != ProviderAnthropic {
		t.Errorf("expected provider %s, got %s", ProviderAnthropic, client.Provider())
	}

	if client.Model() != "claude-sonnet-4" {
		t.Errorf("expected model claude-sonnet-4, got %s", client.Model())
	}

	// Test token counting
	text := "This is a test message"
	tokens := client.CountTokens(text)
	if tokens <= 0 {
		t.Error("expected positive token count")
	}
}

// TestAnthropicErrorMapping tests HTTP status code to error type mapping
func TestAnthropicErrorMapping(t *testing.T) {
	client := &AnthropicClient{}

	tests := []struct {
		statusCode int
		wantType   ErrorType
	}{
		{401, ErrorTypeAuth},
		{403, ErrorTypeAuth},
		{429, ErrorTypeRateLimit},
		{400, ErrorTypeInvalidRequest},
		{422, ErrorTypeInvalidRequest},
		{500, ErrorTypeServer},
		{502, ErrorTypeServer},
		{503, ErrorTypeServer},
		{504, ErrorTypeServer},
		{418, ErrorTypeUnknown}, // I'm a teapot
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.statusCode)), func(t *testing.T) {
			gotType := client.mapErrorType(tt.statusCode)
			if gotType != tt.wantType {
				t.Errorf("status %d: expected error type %s, got %s", tt.statusCode, tt.wantType, gotType)
			}
		})
	}
}

// TestAnthropicComplete tests non-streaming requests
func TestAnthropicComplete(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("x-api-key") != "test-key" {
			t.Error("missing or incorrect API key header")
		}
		if r.Header.Get("anthropic-version") == "" {
			t.Error("missing anthropic-version header")
		}

		// Return mock response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "msg_123",
			"type": "message",
			"role": "assistant",
			"content": [{"type": "text", "text": "Hello!"}],
			"model": "claude-sonnet-4",
			"stop_reason": "end_turn",
			"usage": {"input_tokens": 10, "output_tokens": 5}
		}`))
	}))
	defer server.Close()

	client, err := NewAnthropicClient(ClientConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "claude-sonnet-4",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()
	req := Request{
		Model: "claude-sonnet-4",
		Messages: []Message{
			{Role: RoleUser, Content: "Hi"},
		},
		MaxTokens: 100,
	}

	resp, err := client.Complete(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}

	if resp.Content != "Hello!" {
		t.Errorf("expected content 'Hello!', got '%s'", resp.Content)
	}

	if resp.Usage.PromptTokens != 10 {
		t.Errorf("expected 10 prompt tokens, got %d", resp.Usage.PromptTokens)
	}

	if resp.Usage.CompletionTokens != 5 {
		t.Errorf("expected 5 completion tokens, got %d", resp.Usage.CompletionTokens)
	}

	if resp.Usage.TotalTokens != 15 {
		t.Errorf("expected 15 total tokens, got %d", resp.Usage.TotalTokens)
	}
}

// TestAnthropicCompleteError tests error handling in Complete
func TestAnthropicCompleteError(t *testing.T) {
	// Mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{
			"error": {
				"type": "authentication_error",
				"message": "Invalid API key"
			}
		}`))
	}))
	defer server.Close()

	client, _ := NewAnthropicClient(ClientConfig{
		APIKey:  "invalid-key",
		BaseURL: server.URL,
	})

	ctx := context.Background()
	req := Request{
		Messages: []Message{{Role: RoleUser, Content: "Hi"}},
	}

	_, err := client.Complete(ctx, req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	llmErr, ok := err.(*LLMError)
	if !ok {
		t.Fatal("expected LLMError")
	}

	if llmErr.Type != ErrorTypeAuth {
		t.Errorf("expected auth error, got %s", llmErr.Type)
	}

	if !strings.Contains(llmErr.Message, "Invalid API key") {
		t.Errorf("expected error message to contain 'Invalid API key', got: %s", llmErr.Message)
	}
}

// TestAnthropicStream tests streaming requests
func TestAnthropicStream(t *testing.T) {
	// Mock server with SSE stream
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		// Send SSE events
		events := []string{
			`data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"Hello"}}`,
			`data: {"type":"content_block_delta","delta":{"type":"text_delta","text":" world"}}`,
			`data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"input_tokens":10,"output_tokens":5}}`,
			`data: {"type":"message_stop"}`,
		}

		for _, event := range events {
			w.Write([]byte(event + "\n\n"))
		}
	}))
	defer server.Close()

	client, _ := NewAnthropicClient(ClientConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	ctx := context.Background()
	req := Request{
		Messages: []Message{{Role: RoleUser, Content: "Hi"}},
	}

	var receivedText string
	var done bool
	var finalUsage *Usage

	err := client.Stream(ctx, req, func(event StreamEvent) {
		if event.Delta != "" {
			receivedText += event.Delta
		}
		if event.Done {
			done = true
			finalUsage = event.Usage
		}
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !done {
		t.Error("expected stream to complete")
	}

	if receivedText != "Hello world" {
		t.Errorf("expected 'Hello world', got '%s'", receivedText)
	}

	if finalUsage == nil {
		t.Fatal("expected final usage stats")
	}

	if finalUsage.PromptTokens != 10 {
		t.Errorf("expected 10 prompt tokens, got %d", finalUsage.PromptTokens)
	}
}

// TestRetryableClient tests retry logic
func TestRetryableClient(t *testing.T) {
	// Create a mock client that fails twice then succeeds
	attempts := 0
	mockClient := &mockClient{
		completeFunc: func(ctx context.Context, req Request) (*Response, error) {
			attempts++
			if attempts < 3 {
				return nil, &LLMError{
					Type:    ErrorTypeRateLimit,
					Message: "rate limited",
				}
			}
			return &Response{Content: "success"}, nil
		},
	}

	config := DefaultRetryConfig()
	config.InitialBackoff = 10 * time.Millisecond // Speed up test
	retryClient := NewRetryableClient(mockClient, config)

	ctx := context.Background()
	resp, err := retryClient.Complete(ctx, Request{})

	if err != nil {
		t.Fatalf("unexpected error after retries: %v", err)
	}

	if resp.Content != "success" {
		t.Errorf("expected 'success', got '%s'", resp.Content)
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

// TestRetryableClientNonRetryableError tests that auth errors aren't retried
func TestRetryableClientNonRetryableError(t *testing.T) {
	attempts := 0
	mockClient := &mockClient{
		completeFunc: func(ctx context.Context, req Request) (*Response, error) {
			attempts++
			return nil, &LLMError{
				Type:    ErrorTypeAuth,
				Message: "unauthorized",
			}
		},
	}

	config := DefaultRetryConfig()
	retryClient := NewRetryableClient(mockClient, config)

	ctx := context.Background()
	_, err := retryClient.Complete(ctx, Request{})

	if err == nil {
		t.Fatal("expected error")
	}

	if attempts != 1 {
		t.Errorf("expected 1 attempt (no retry), got %d", attempts)
	}
}

// TestRetryBackoffCalculation tests exponential backoff calculation
func TestRetryBackoffCalculation(t *testing.T) {
	config := RetryConfig{
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		Multiplier:     2.0,
	}

	client := &RetryableClient{config: config}

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 1 * time.Second},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
		{10, 30 * time.Second}, // Should cap at MaxBackoff
	}

	for _, tt := range tests {
		got := client.calculateBackoff(tt.attempt)
		if got != tt.want {
			t.Errorf("attempt %d: expected backoff %v, got %v", tt.attempt, tt.want, got)
		}
	}
}

// TestTokenTracker tests token usage tracking
func TestTokenTracker(t *testing.T) {
	tracker := NewTokenTracker()

	// Add some usage
	tracker.AddUsage(Usage{
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
	})

	tracker.AddUsage(Usage{
		PromptTokens:     200,
		CompletionTokens: 75,
		TotalTokens:      275,
	})

	stats := tracker.GetStats()

	if stats.TotalPromptTokens != 300 {
		t.Errorf("expected 300 prompt tokens, got %d", stats.TotalPromptTokens)
	}

	if stats.TotalCompletionTokens != 125 {
		t.Errorf("expected 125 completion tokens, got %d", stats.TotalCompletionTokens)
	}

	if stats.TotalTokens != 425 {
		t.Errorf("expected 425 total tokens, got %d", stats.TotalTokens)
	}

	if stats.RequestCount != 2 {
		t.Errorf("expected 2 requests, got %d", stats.RequestCount)
	}

	// Test reset
	tracker.Reset()
	stats = tracker.GetStats()

	if stats.TotalTokens != 0 {
		t.Errorf("expected 0 tokens after reset, got %d", stats.TotalTokens)
	}

	if stats.RequestCount != 0 {
		t.Errorf("expected 0 requests after reset, got %d", stats.RequestCount)
	}
}

// TestTokenStatsEstimatedCost tests cost estimation
func TestTokenStatsEstimatedCost(t *testing.T) {
	stats := TokenStats{
		TotalPromptTokens:     1_000_000, // 1M input tokens
		TotalCompletionTokens: 500_000,   // 500K output tokens
	}

	// Test Claude Sonnet pricing
	cost := stats.EstimatedCost("claude-sonnet-4-5")
	expectedCost := (1.0 * 3.0) + (0.5 * 15.0) // (1M * $3/1M) + (0.5M * $15/1M)
	if cost != expectedCost {
		t.Errorf("expected cost $%.2f, got $%.2f", expectedCost, cost)
	}

	// Test unknown model
	cost = stats.EstimatedCost("unknown-model")
	if cost != 0 {
		t.Errorf("expected cost $0 for unknown model, got $%.2f", cost)
	}
}

// TestLLMError tests error type
func TestLLMError(t *testing.T) {
	err := &LLMError{
		Type:    ErrorTypeAuth,
		Message: "authentication failed",
		Code:    401,
		Details: "invalid_api_key",
	}

	if err.Error() != "authentication failed" {
		t.Errorf("expected error message 'authentication failed', got '%s'", err.Error())
	}
}

// Mock client for testing
type mockClient struct {
	completeFunc func(ctx context.Context, req Request) (*Response, error)
	streamFunc   func(ctx context.Context, req Request, callback StreamCallback) error
}

func (m *mockClient) Complete(ctx context.Context, req Request) (*Response, error) {
	if m.completeFunc != nil {
		return m.completeFunc(ctx, req)
	}
	return nil, errors.New("not implemented")
}

func (m *mockClient) Stream(ctx context.Context, req Request, callback StreamCallback) error {
	if m.streamFunc != nil {
		return m.streamFunc(ctx, req, callback)
	}
	return errors.New("not implemented")
}

func (m *mockClient) Provider() ProviderType {
	return ProviderAnthropic
}

func (m *mockClient) Model() string {
	return "test-model"
}

func (m *mockClient) CountTokens(text string) int {
	return len(text) / 4
}

// TestOpenAIClientCreation tests OpenAI client initialization
func TestOpenAIClientCreation(t *testing.T) {
	tests := []struct {
		name        string
		config      ClientConfig
		wantErr     bool
		wantErrType ErrorType
	}{
		{
			name: "valid OpenAI configuration",
			config: ClientConfig{
				Provider: ProviderOpenAI,
				APIKey:   "test-key",
				Model:    "gpt-4",
			},
			wantErr: false,
		},
		{
			name: "local provider without API key",
			config: ClientConfig{
				Provider: ProviderLocal,
				Model:    "llama2",
				BaseURL:  "http://localhost:11434",
			},
			wantErr: false,
		},
		{
			name: "OpenAI missing API key",
			config: ClientConfig{
				Provider: ProviderOpenAI,
				Model:    "gpt-4",
			},
			wantErr:     true,
			wantErrType: ErrorTypeAuth,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewOpenAIClient(tt.config)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if llmErr, ok := err.(*LLMError); ok {
					if llmErr.Type != tt.wantErrType {
						t.Errorf("expected error type %s, got %s", tt.wantErrType, llmErr.Type)
					}
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if client == nil {
					t.Fatal("expected client, got nil")
				}
			}
		})
	}
}

// TestOpenAIClientMethods tests basic OpenAI client methods
func TestOpenAIClientMethods(t *testing.T) {
	client := &OpenAIClient{
		config: ClientConfig{
			Provider: ProviderOpenAI,
			Model:    "gpt-4",
		},
	}

	if client.Provider() != ProviderOpenAI {
		t.Errorf("expected provider %s, got %s", ProviderOpenAI, client.Provider())
	}

	if client.Model() != "gpt-4" {
		t.Errorf("expected model gpt-4, got %s", client.Model())
	}

	// Test token counting
	text := "This is a test message"
	tokens := client.CountTokens(text)
	if tokens <= 0 {
		t.Error("expected positive token count")
	}
}

// TestOpenAIComplete tests non-streaming requests
func TestOpenAIComplete(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			t.Error("missing or incorrect Authorization header")
		}

		// Return mock response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "chatcmpl-123",
			"object": "chat.completion",
			"created": 1677652288,
			"model": "gpt-4",
			"choices": [{
				"index": 0,
				"message": {
					"role": "assistant",
					"content": "Hello! How can I help you?"
				},
				"finish_reason": "stop"
			}],
			"usage": {
				"prompt_tokens": 10,
				"completion_tokens": 8,
				"total_tokens": 18
			}
		}`))
	}))
	defer server.Close()

	client, err := NewOpenAIClient(ClientConfig{
		Provider: ProviderOpenAI,
		APIKey:   "test-key",
		BaseURL:  server.URL,
		Model:    "gpt-4",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()
	req := Request{
		Messages: []Message{
			{Role: RoleUser, Content: "Hi"},
		},
		MaxTokens: 100,
	}

	resp, err := client.Complete(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}

	if resp.Content != "Hello! How can I help you?" {
		t.Errorf("expected specific content, got '%s'", resp.Content)
	}

	if resp.Usage.PromptTokens != 10 {
		t.Errorf("expected 10 prompt tokens, got %d", resp.Usage.PromptTokens)
	}

	if resp.Usage.CompletionTokens != 8 {
		t.Errorf("expected 8 completion tokens, got %d", resp.Usage.CompletionTokens)
	}
}

// TestOpenAICompleteWithSystemPrompt tests system prompt handling
func TestOpenAICompleteWithSystemPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Decode request to verify system prompt was included
		var reqBody openaiRequest
		json.NewDecoder(r.Body).Decode(&reqBody)

		if len(reqBody.Messages) < 2 {
			t.Error("expected at least 2 messages (system + user)")
		}

		if reqBody.Messages[0].Role != "system" {
			t.Errorf("expected first message to be system, got %s", reqBody.Messages[0].Role)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "test",
			"model": "gpt-4",
			"choices": [{"message": {"role": "assistant", "content": "OK"}, "finish_reason": "stop"}],
			"usage": {"prompt_tokens": 10, "completion_tokens": 1, "total_tokens": 11}
		}`))
	}))
	defer server.Close()

	client, _ := NewOpenAIClient(ClientConfig{
		Provider: ProviderOpenAI,
		APIKey:   "test-key",
		BaseURL:  server.URL,
	})

	ctx := context.Background()
	req := Request{
		SystemPrompt: "You are a helpful assistant.",
		Messages:     []Message{{Role: RoleUser, Content: "Hi"}},
	}

	_, err := client.Complete(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestOpenAIStream tests streaming requests
func TestOpenAIStream(t *testing.T) {
	// Mock server with SSE stream
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		// Send SSE events
		events := []string{
			`data: {"choices":[{"delta":{"content":"Hello"},"finish_reason":""}]}`,
			`data: {"choices":[{"delta":{"content":" world"},"finish_reason":""}]}`,
			`data: {"choices":[{"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":10,"completion_tokens":5,"total_tokens":15}}`,
			`data: [DONE]`,
		}

		for _, event := range events {
			w.Write([]byte(event + "\n\n"))
		}
	}))
	defer server.Close()

	client, _ := NewOpenAIClient(ClientConfig{
		Provider: ProviderOpenAI,
		APIKey:   "test-key",
		BaseURL:  server.URL,
	})

	ctx := context.Background()
	req := Request{
		Messages: []Message{{Role: RoleUser, Content: "Hi"}},
	}

	var receivedText string
	var done bool
	var finalUsage *Usage

	err := client.Stream(ctx, req, func(event StreamEvent) {
		if event.Delta != "" {
			receivedText += event.Delta
		}
		if event.Done {
			done = true
			finalUsage = event.Usage
		}
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !done {
		t.Error("expected stream to complete")
	}

	if receivedText != "Hello world" {
		t.Errorf("expected 'Hello world', got '%s'", receivedText)
	}

	if finalUsage == nil {
		t.Fatal("expected final usage stats")
	}
}

// TestOpenAIErrorHandling tests error response handling
func TestOpenAIErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{
			"error": {
				"message": "Invalid API key provided",
				"type": "invalid_request_error",
				"code": "invalid_api_key"
			}
		}`))
	}))
	defer server.Close()

	client, _ := NewOpenAIClient(ClientConfig{
		Provider: ProviderOpenAI,
		APIKey:   "invalid-key",
		BaseURL:  server.URL,
	})

	ctx := context.Background()
	req := Request{
		Messages: []Message{{Role: RoleUser, Content: "Hi"}},
	}

	_, err := client.Complete(ctx, req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	llmErr, ok := err.(*LLMError)
	if !ok {
		t.Fatal("expected LLMError")
	}

	if llmErr.Type != ErrorTypeAuth {
		t.Errorf("expected auth error, got %s", llmErr.Type)
	}
}
