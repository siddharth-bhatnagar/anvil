package agent

import (
	"fmt"
	"strings"
	"sync"

	"github.com/siddharth-bhatnagar/anvil/internal/llm"
)

// ContextConfig holds configuration for context management
type ContextConfig struct {
	MaxMessages   int // Maximum number of messages (0 = unlimited)
	MaxTokens     int // Maximum estimated tokens (0 = unlimited)
	CharsPerToken int // Characters per token estimate (default: 4)
}

// DefaultContextConfig returns the default context configuration
func DefaultContextConfig() ContextConfig {
	return ContextConfig{
		MaxMessages:   100,
		MaxTokens:     100000, // ~100k tokens
		CharsPerToken: 4,
	}
}

// Context manages the conversation history and context window
type Context struct {
	mu            sync.RWMutex
	messages      []llm.Message
	config        ContextConfig
	prunedSummary string       // Summary of pruned content
	prunedCount   int          // Number of messages pruned
	onPrune       PruneCallback // Called when messages are pruned
}

// PruneCallback is called when messages are pruned from context
type PruneCallback func(pruned []llm.Message, summary string)

// NewContext creates a new context manager
func NewContext() *Context {
	return &Context{
		messages: make([]llm.Message, 0),
		config:   DefaultContextConfig(),
	}
}

// NewContextWithConfig creates a new context manager with custom config
func NewContextWithConfig(config ContextConfig) *Context {
	if config.CharsPerToken <= 0 {
		config.CharsPerToken = 4
	}
	return &Context{
		messages: make([]llm.Message, 0),
		config:   config,
	}
}

// SetPruneCallback sets the callback for when messages are pruned
func (c *Context) SetPruneCallback(cb PruneCallback) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onPrune = cb
}

// AddMessage adds a message to the context
func (c *Context) AddMessage(msg llm.Message) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.messages = append(c.messages, msg)

	// Check if we need to prune
	c.checkAndPrune()
}

// checkAndPrune checks if pruning is needed and performs it
// Must be called with lock held
func (c *Context) checkAndPrune() {
	needsPrune := false

	// Check message count limit
	if c.config.MaxMessages > 0 && len(c.messages) > c.config.MaxMessages {
		needsPrune = true
	}

	// Check token limit
	if c.config.MaxTokens > 0 && c.estimateTokensLocked() > c.config.MaxTokens {
		needsPrune = true
	}

	if needsPrune {
		c.prune()
	}
}

// estimateTokensLocked estimates tokens without acquiring lock
// Must be called with lock held
func (c *Context) estimateTokensLocked() int {
	totalChars := 0
	for _, msg := range c.messages {
		totalChars += len(msg.Content)
	}
	return totalChars / c.config.CharsPerToken
}

// GetMessages returns all messages in the context
func (c *Context) GetMessages() []llm.Message {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to prevent external modification
	messages := make([]llm.Message, len(c.messages))
	copy(messages, c.messages)
	return messages
}

// GetRecentMessages returns the N most recent messages
func (c *Context) GetRecentMessages(n int) []llm.Message {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if n >= len(c.messages) {
		messages := make([]llm.Message, len(c.messages))
		copy(messages, c.messages)
		return messages
	}

	start := len(c.messages) - n
	messages := make([]llm.Message, n)
	copy(messages, c.messages[start:])
	return messages
}

// Clear removes all messages from the context
func (c *Context) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.messages = make([]llm.Message, 0)
}

// Size returns the number of messages in the context
func (c *Context) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.messages)
}

// SetMaxSize sets the maximum number of messages to keep
func (c *Context) SetMaxSize(size int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.config.MaxMessages = size
	c.checkAndPrune()
}

// SetMaxTokens sets the maximum number of tokens to keep
func (c *Context) SetMaxTokens(tokens int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.config.MaxTokens = tokens
	c.checkAndPrune()
}

// GetConfig returns the current context configuration
func (c *Context) GetConfig() ContextConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.config
}

// SetConfig sets the context configuration
func (c *Context) SetConfig(config ContextConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if config.CharsPerToken <= 0 {
		config.CharsPerToken = 4
	}
	c.config = config
	c.checkAndPrune()
}

// prune removes old messages when the context exceeds limits
// This is called with the lock already held
func (c *Context) prune() {
	// Separate system messages (always keep) from others
	var systemMessages []llm.Message
	var otherMessages []llm.Message

	for _, msg := range c.messages {
		if msg.Role == llm.RoleSystem {
			systemMessages = append(systemMessages, msg)
		} else {
			otherMessages = append(otherMessages, msg)
		}
	}

	// Calculate current state
	systemTokens := 0
	for _, msg := range systemMessages {
		systemTokens += len(msg.Content) / c.config.CharsPerToken
	}

	var prunedMessages []llm.Message
	var keptMessages []llm.Message

	// Prune based on message count first
	if c.config.MaxMessages > 0 {
		maxOther := c.config.MaxMessages - len(systemMessages)
		if maxOther < 0 {
			maxOther = 0
		}
		if len(otherMessages) > maxOther {
			prunedMessages = otherMessages[:len(otherMessages)-maxOther]
			otherMessages = otherMessages[len(otherMessages)-maxOther:]
		}
	}

	// Then prune based on token count
	if c.config.MaxTokens > 0 {
		availableTokens := c.config.MaxTokens - systemTokens
		currentTokens := 0

		// Count tokens from newest to oldest
		for i := len(otherMessages) - 1; i >= 0; i-- {
			msgTokens := len(otherMessages[i].Content) / c.config.CharsPerToken
			if currentTokens+msgTokens <= availableTokens {
				currentTokens += msgTokens
				keptMessages = append([]llm.Message{otherMessages[i]}, keptMessages...)
			} else {
				prunedMessages = append(prunedMessages, otherMessages[i])
			}
		}
		otherMessages = keptMessages
	}

	// Update pruned count
	c.prunedCount += len(prunedMessages)

	// Generate summary of pruned content
	if len(prunedMessages) > 0 {
		c.prunedSummary = c.generatePrunedSummary(prunedMessages)

		// Call callback if set
		callback := c.onPrune
		if callback != nil {
			// Make copies for callback
			prunedCopy := make([]llm.Message, len(prunedMessages))
			copy(prunedCopy, prunedMessages)
			go callback(prunedCopy, c.prunedSummary)
		}
	}

	// Combine back together
	c.messages = append(systemMessages, otherMessages...)
}

// generatePrunedSummary creates a brief summary of pruned messages
func (c *Context) generatePrunedSummary(pruned []llm.Message) string {
	if len(pruned) == 0 {
		return ""
	}

	var userCount, assistantCount int
	for _, msg := range pruned {
		switch msg.Role {
		case llm.RoleUser:
			userCount++
		case llm.RoleAssistant:
			assistantCount++
		}
	}

	var parts []string
	if userCount > 0 {
		parts = append(parts, fmt.Sprintf("%d user message(s)", userCount))
	}
	if assistantCount > 0 {
		parts = append(parts, fmt.Sprintf("%d assistant message(s)", assistantCount))
	}

	return fmt.Sprintf("Pruned %s from context", strings.Join(parts, " and "))
}

// GetPrunedSummary returns the summary of pruned content
func (c *Context) GetPrunedSummary() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.prunedSummary
}

// GetPrunedCount returns the total number of pruned messages
func (c *Context) GetPrunedCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.prunedCount
}

// EstimateTokens estimates the total number of tokens in the context
// This is a rough estimate based on character count
func (c *Context) EstimateTokens() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.estimateTokensLocked()
}

// Stats returns statistics about the context
func (c *Context) Stats() ContextStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var userCount, assistantCount, systemCount int
	var totalChars int

	for _, msg := range c.messages {
		totalChars += len(msg.Content)
		switch msg.Role {
		case llm.RoleUser:
			userCount++
		case llm.RoleAssistant:
			assistantCount++
		case llm.RoleSystem:
			systemCount++
		}
	}

	return ContextStats{
		MessageCount:     len(c.messages),
		UserMessages:     userCount,
		AssistantMessages: assistantCount,
		SystemMessages:   systemCount,
		EstimatedTokens:  totalChars / c.config.CharsPerToken,
		PrunedCount:      c.prunedCount,
		MaxMessages:      c.config.MaxMessages,
		MaxTokens:        c.config.MaxTokens,
	}
}

// ContextStats holds statistics about the context
type ContextStats struct {
	MessageCount      int
	UserMessages      int
	AssistantMessages int
	SystemMessages    int
	EstimatedTokens   int
	PrunedCount       int
	MaxMessages       int
	MaxTokens         int
}

// RemoveLastMessage removes the most recent message
func (c *Context) RemoveLastMessage() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.messages) > 0 {
		c.messages = c.messages[:len(c.messages)-1]
	}
}

// RemoveLastN removes the N most recent messages
func (c *Context) RemoveLastN(n int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if n >= len(c.messages) {
		c.messages = make([]llm.Message, 0)
		return
	}

	c.messages = c.messages[:len(c.messages)-n]
}
