package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	anthropicAPIURL = "https://api.anthropic.com/v1/messages"
	anthropicVersion = "2023-06-01"
)

// AnthropicClient implements the Client interface for Anthropic's Claude API
type AnthropicClient struct {
	config     ClientConfig
	httpClient *http.Client
}

// NewAnthropicClient creates a new Anthropic client
func NewAnthropicClient(config ClientConfig) (*AnthropicClient, error) {
	if config.APIKey == "" {
		return nil, &LLMError{
			Type:    ErrorTypeAuth,
			Message: "API key is required for Anthropic",
		}
	}

	if config.BaseURL == "" {
		config.BaseURL = anthropicAPIURL
	}

	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}

	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}

	return &AnthropicClient{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}, nil
}

// anthropicRequest represents the Anthropic API request format
type anthropicRequest struct {
	Model       string              `json:"model"`
	Messages    []anthropicMessage  `json:"messages"`
	MaxTokens   int                 `json:"max_tokens"`
	Temperature float64             `json:"temperature,omitempty"`
	TopP        float64             `json:"top_p,omitempty"`
	Stream      bool                `json:"stream"`
	System      string              `json:"system,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// anthropicResponse represents the Anthropic API response format
type anthropicResponse struct {
	ID           string                  `json:"id"`
	Type         string                  `json:"type"`
	Role         string                  `json:"role"`
	Content      []anthropicContent      `json:"content"`
	Model        string                  `json:"model"`
	StopReason   string                  `json:"stop_reason"`
	Usage        anthropicUsage          `json:"usage"`
	Error        *anthropicError         `json:"error,omitempty"`
}

type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type anthropicError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// anthropicStreamEvent represents a server-sent event from Anthropic
type anthropicStreamEvent struct {
	Type         string                  `json:"type"`
	Index        int                     `json:"index,omitempty"`
	Delta        *anthropicDelta         `json:"delta,omitempty"`
	ContentBlock *anthropicContent       `json:"content_block,omitempty"`
	Message      *anthropicResponse      `json:"message,omitempty"`
	Usage        *anthropicUsage         `json:"usage,omitempty"`
	Error        *anthropicError         `json:"error,omitempty"`
}

type anthropicDelta struct {
	Type       string `json:"type"`
	Text       string `json:"text"`
	StopReason string `json:"stop_reason,omitempty"`
}

// Complete sends a non-streaming request
func (c *AnthropicClient) Complete(ctx context.Context, req Request) (*Response, error) {
	apiReq := c.buildRequest(req)
	apiReq.Stream = false

	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, &LLMError{
			Type:    ErrorTypeInvalidRequest,
			Message: "failed to marshal request: " + err.Error(),
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, &LLMError{
			Type:    ErrorTypeNetwork,
			Message: "failed to create request: " + err.Error(),
		}
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, &LLMError{
			Type:    ErrorTypeNetwork,
			Message: "request failed: " + err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleError(resp)
	}

	var apiResp anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, &LLMError{
			Type:    ErrorTypeServer,
			Message: "failed to decode response: " + err.Error(),
		}
	}

	return c.convertResponse(&apiResp), nil
}

// Stream sends a streaming request
func (c *AnthropicClient) Stream(ctx context.Context, req Request, callback StreamCallback) error {
	apiReq := c.buildRequest(req)
	apiReq.Stream = true

	body, err := json.Marshal(apiReq)
	if err != nil {
		return &LLMError{
			Type:    ErrorTypeInvalidRequest,
			Message: "failed to marshal request: " + err.Error(),
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL, bytes.NewReader(body))
	if err != nil {
		return &LLMError{
			Type:    ErrorTypeNetwork,
			Message: "failed to create request: " + err.Error(),
		}
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return &LLMError{
			Type:    ErrorTypeNetwork,
			Message: "request failed: " + err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.handleError(resp)
	}

	return c.processStream(resp.Body, callback)
}

// processStream processes the SSE stream
func (c *AnthropicClient) processStream(body io.Reader, callback StreamCallback) error {
	scanner := bufio.NewScanner(body)
	var usage *Usage

	for scanner.Scan() {
		line := scanner.Text()

		// SSE format: "data: {json}" or "event: {type}"
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			callback(StreamEvent{
				Done:      true,
				Usage:     usage,
				Timestamp: time.Now(),
			})
			return nil
		}

		var event anthropicStreamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue // Skip malformed events
		}

		// Handle different event types
		switch event.Type {
		case "content_block_delta":
			if event.Delta != nil && event.Delta.Text != "" {
				callback(StreamEvent{
					Delta:     event.Delta.Text,
					Done:      false,
					Timestamp: time.Now(),
				})
			}

		case "message_delta":
			if event.Delta != nil && event.Delta.StopReason != "" {
				callback(StreamEvent{
					Done:         false,
					FinishReason: event.Delta.StopReason,
					Timestamp:    time.Now(),
				})
			}
			if event.Usage != nil {
				usage = &Usage{
					PromptTokens:     event.Usage.InputTokens,
					CompletionTokens: event.Usage.OutputTokens,
					TotalTokens:      event.Usage.InputTokens + event.Usage.OutputTokens,
				}
			}

		case "message_stop":
			callback(StreamEvent{
				Done:      true,
				Usage:     usage,
				Timestamp: time.Now(),
			})
			return nil

		case "error":
			if event.Error != nil {
				return &LLMError{
					Type:    ErrorTypeServer,
					Message: event.Error.Message,
					Details: event.Error.Type,
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return &LLMError{
			Type:    ErrorTypeNetwork,
			Message: "stream reading error: " + err.Error(),
		}
	}

	return nil
}

// buildRequest converts a generic Request to Anthropic format
func (c *AnthropicClient) buildRequest(req Request) anthropicRequest {
	apiReq := anthropicRequest{
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stream:      req.Stream,
		System:      req.SystemPrompt,
		Messages:    make([]anthropicMessage, len(req.Messages)),
	}

	// Use config defaults if not specified
	if apiReq.Model == "" {
		apiReq.Model = c.config.Model
	}
	if apiReq.MaxTokens == 0 {
		apiReq.MaxTokens = c.config.MaxTokens
	}
	if apiReq.Temperature == 0 {
		apiReq.Temperature = c.config.Temperature
	}

	// Convert messages
	for i, msg := range req.Messages {
		apiReq.Messages[i] = anthropicMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}
	}

	return apiReq
}

// convertResponse converts Anthropic response to generic Response
func (c *AnthropicClient) convertResponse(resp *anthropicResponse) *Response {
	var content string
	if len(resp.Content) > 0 {
		content = resp.Content[0].Text
	}

	return &Response{
		Content:      content,
		Role:         Role(resp.Role),
		FinishReason: resp.StopReason,
		Model:        resp.Model,
		Usage: Usage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
	}
}

// setHeaders sets the required headers for Anthropic API
func (c *AnthropicClient) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.config.APIKey)
	req.Header.Set("anthropic-version", anthropicVersion)
}

// handleError processes error responses
func (c *AnthropicClient) handleError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	var apiErr anthropicResponse
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Error != nil {
		return &LLMError{
			Type:    c.mapErrorType(resp.StatusCode),
			Message: apiErr.Error.Message,
			Code:    resp.StatusCode,
			Details: apiErr.Error.Type,
		}
	}

	return &LLMError{
		Type:    c.mapErrorType(resp.StatusCode),
		Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)),
		Code:    resp.StatusCode,
	}
}

// mapErrorType maps HTTP status codes to error types
func (c *AnthropicClient) mapErrorType(statusCode int) ErrorType {
	switch statusCode {
	case 401, 403:
		return ErrorTypeAuth
	case 429:
		return ErrorTypeRateLimit
	case 400, 422:
		return ErrorTypeInvalidRequest
	case 500, 502, 503, 504:
		return ErrorTypeServer
	default:
		return ErrorTypeUnknown
	}
}

// Provider returns the provider type
func (c *AnthropicClient) Provider() ProviderType {
	return ProviderAnthropic
}

// Model returns the configured model
func (c *AnthropicClient) Model() string {
	return c.config.Model
}

// CountTokens estimates token count (rough approximation)
func (c *AnthropicClient) CountTokens(text string) int {
	// Rough approximation: ~4 characters per token
	return len(text) / 4
}
