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
	openaiAPIURL = "https://api.openai.com/v1/chat/completions"
)

// OpenAIClient implements the Client interface for OpenAI-compatible APIs
type OpenAIClient struct {
	config     ClientConfig
	httpClient *http.Client
}

// NewOpenAIClient creates a new OpenAI-compatible client
func NewOpenAIClient(config ClientConfig) (*OpenAIClient, error) {
	if config.APIKey == "" && config.Provider != ProviderLocal {
		return nil, &LLMError{
			Type:    ErrorTypeAuth,
			Message: "API key is required for OpenAI",
		}
	}

	if config.BaseURL == "" {
		config.BaseURL = openaiAPIURL
	}

	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}

	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}

	return &OpenAIClient{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}, nil
}

// openaiRequest represents the OpenAI API request format
type openaiRequest struct {
	Model       string          `json:"model"`
	Messages    []openaiMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	TopP        float64         `json:"top_p,omitempty"`
	Stream      bool            `json:"stream"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openaiResponse represents the OpenAI API response format
type openaiResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []openaiChoice `json:"choices"`
	Usage   openaiUsage    `json:"usage"`
	Error   *openaiError   `json:"error,omitempty"`
}

type openaiChoice struct {
	Index        int           `json:"index"`
	Message      openaiMessage `json:"message"`
	Delta        *openaiDelta  `json:"delta,omitempty"`
	FinishReason string        `json:"finish_reason"`
}

type openaiDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

type openaiUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type openaiError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// Complete sends a non-streaming request
func (c *OpenAIClient) Complete(ctx context.Context, req Request) (*Response, error) {
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

	var apiResp openaiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, &LLMError{
			Type:    ErrorTypeServer,
			Message: "failed to decode response: " + err.Error(),
		}
	}

	if apiResp.Error != nil {
		return nil, &LLMError{
			Type:    ErrorTypeServer,
			Message: apiResp.Error.Message,
			Details: apiResp.Error.Type,
		}
	}

	return c.convertResponse(&apiResp), nil
}

// Stream sends a streaming request
func (c *OpenAIClient) Stream(ctx context.Context, req Request, callback StreamCallback) error {
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
func (c *OpenAIClient) processStream(body io.Reader, callback StreamCallback) error {
	scanner := bufio.NewScanner(body)
	var usage *Usage

	for scanner.Scan() {
		line := scanner.Text()

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

		var chunk openaiResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if len(chunk.Choices) > 0 {
			choice := chunk.Choices[0]

			if choice.Delta != nil && choice.Delta.Content != "" {
				callback(StreamEvent{
					Delta:     choice.Delta.Content,
					Done:      false,
					Timestamp: time.Now(),
				})
			}

			if choice.FinishReason != "" {
				callback(StreamEvent{
					Done:         false,
					FinishReason: choice.FinishReason,
					Timestamp:    time.Now(),
				})
			}
		}

		if chunk.Usage.TotalTokens > 0 {
			usage = &Usage{
				PromptTokens:     chunk.Usage.PromptTokens,
				CompletionTokens: chunk.Usage.CompletionTokens,
				TotalTokens:      chunk.Usage.TotalTokens,
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return &LLMError{
			Type:    ErrorTypeNetwork,
			Message: "stream reading error: " + err.Error(),
		}
	}

	callback(StreamEvent{
		Done:      true,
		Usage:     usage,
		Timestamp: time.Now(),
	})

	return nil
}

// buildRequest converts a generic Request to OpenAI format
func (c *OpenAIClient) buildRequest(req Request) openaiRequest {
	apiReq := openaiRequest{
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stream:      req.Stream,
		Messages:    make([]openaiMessage, 0, len(req.Messages)+1),
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

	// Add system prompt if present
	if req.SystemPrompt != "" {
		apiReq.Messages = append(apiReq.Messages, openaiMessage{
			Role:    "system",
			Content: req.SystemPrompt,
		})
	}

	// Convert messages
	for _, msg := range req.Messages {
		apiReq.Messages = append(apiReq.Messages, openaiMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}

	return apiReq
}

// convertResponse converts OpenAI response to generic Response
func (c *OpenAIClient) convertResponse(resp *openaiResponse) *Response {
	if len(resp.Choices) == 0 {
		return &Response{
			Model: resp.Model,
		}
	}

	choice := resp.Choices[0]

	return &Response{
		Content:      choice.Message.Content,
		Role:         Role(choice.Message.Role),
		FinishReason: choice.FinishReason,
		Model:        resp.Model,
		Usage: Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}
}

// setHeaders sets the required headers for OpenAI API
func (c *OpenAIClient) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}
}

// handleError processes error responses
func (c *OpenAIClient) handleError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	var apiResp openaiResponse
	if err := json.Unmarshal(body, &apiResp); err == nil && apiResp.Error != nil {
		return &LLMError{
			Type:    c.mapErrorType(resp.StatusCode),
			Message: apiResp.Error.Message,
			Code:    resp.StatusCode,
			Details: apiResp.Error.Type,
		}
	}

	return &LLMError{
		Type:    c.mapErrorType(resp.StatusCode),
		Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)),
		Code:    resp.StatusCode,
	}
}

// mapErrorType maps HTTP status codes to error types
func (c *OpenAIClient) mapErrorType(statusCode int) ErrorType {
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
func (c *OpenAIClient) Provider() ProviderType {
	return c.config.Provider
}

// Model returns the configured model
func (c *OpenAIClient) Model() string {
	return c.config.Model
}

// CountTokens estimates token count (rough approximation)
func (c *OpenAIClient) CountTokens(text string) int {
	// Rough approximation: ~4 characters per token
	return len(text) / 4
}
