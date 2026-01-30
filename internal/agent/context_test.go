package agent

import (
	"testing"

	"github.com/siddharth-bhatnagar/anvil/internal/llm"
)

func TestNewContext(t *testing.T) {
	ctx := NewContext()

	if ctx == nil {
		t.Fatal("NewContext returned nil")
	}

	if ctx.Size() != 0 {
		t.Errorf("New context should be empty, got %d messages", ctx.Size())
	}
}

func TestContextAddMessage(t *testing.T) {
	ctx := NewContext()

	ctx.AddMessage(llm.Message{Role: llm.RoleUser, Content: "Hello"})

	if ctx.Size() != 1 {
		t.Errorf("Expected 1 message, got %d", ctx.Size())
	}

	messages := ctx.GetMessages()
	if messages[0].Content != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", messages[0].Content)
	}
}

func TestContextGetMessages(t *testing.T) {
	ctx := NewContext()

	ctx.AddMessage(llm.Message{Role: llm.RoleUser, Content: "First"})
	ctx.AddMessage(llm.Message{Role: llm.RoleAssistant, Content: "Second"})

	messages := ctx.GetMessages()

	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	// Verify it returns a copy
	messages[0].Content = "Modified"
	original := ctx.GetMessages()
	if original[0].Content == "Modified" {
		t.Error("GetMessages should return a copy, not the original")
	}
}

func TestContextGetRecentMessages(t *testing.T) {
	ctx := NewContext()

	for i := 0; i < 5; i++ {
		ctx.AddMessage(llm.Message{Role: llm.RoleUser, Content: string(rune('A' + i))})
	}

	recent := ctx.GetRecentMessages(3)

	if len(recent) != 3 {
		t.Fatalf("Expected 3 messages, got %d", len(recent))
	}

	// Should be the last 3: C, D, E
	if recent[0].Content != "C" || recent[1].Content != "D" || recent[2].Content != "E" {
		t.Error("GetRecentMessages returned wrong messages")
	}
}

func TestContextGetRecentMessagesMoreThanAvailable(t *testing.T) {
	ctx := NewContext()

	ctx.AddMessage(llm.Message{Role: llm.RoleUser, Content: "Only one"})

	recent := ctx.GetRecentMessages(10)

	if len(recent) != 1 {
		t.Errorf("Expected 1 message, got %d", len(recent))
	}
}

func TestContextClear(t *testing.T) {
	ctx := NewContext()

	ctx.AddMessage(llm.Message{Role: llm.RoleUser, Content: "Test"})
	ctx.Clear()

	if ctx.Size() != 0 {
		t.Errorf("Context should be empty after Clear, got %d messages", ctx.Size())
	}
}

func TestContextSetMaxSize(t *testing.T) {
	ctx := NewContext()

	// Add 10 messages
	for i := 0; i < 10; i++ {
		ctx.AddMessage(llm.Message{Role: llm.RoleUser, Content: string(rune('0' + i))})
	}

	// Set max size to 5
	ctx.SetMaxSize(5)

	if ctx.Size() > 5 {
		t.Errorf("Context should have at most 5 messages after SetMaxSize, got %d", ctx.Size())
	}
}

func TestContextPrunePreservesSystemMessages(t *testing.T) {
	ctx := NewContext()
	ctx.SetMaxSize(3)

	// Add a system message
	ctx.AddMessage(llm.Message{Role: llm.RoleSystem, Content: "System prompt"})

	// Add several user messages
	for i := 0; i < 5; i++ {
		ctx.AddMessage(llm.Message{Role: llm.RoleUser, Content: string(rune('A' + i))})
	}

	messages := ctx.GetMessages()

	// System message should still be present
	hasSystem := false
	for _, msg := range messages {
		if msg.Role == llm.RoleSystem {
			hasSystem = true
			break
		}
	}

	if !hasSystem {
		t.Error("System message should be preserved during pruning")
	}
}

func TestContextEstimateTokens(t *testing.T) {
	ctx := NewContext()

	// Add a message with known length
	ctx.AddMessage(llm.Message{Role: llm.RoleUser, Content: "12345678"}) // 8 chars

	tokens := ctx.EstimateTokens()

	// Default is 4 chars per token, so 8 chars = 2 tokens
	if tokens != 2 {
		t.Errorf("Expected 2 tokens, got %d", tokens)
	}
}

func TestContextRemoveLastMessage(t *testing.T) {
	ctx := NewContext()

	ctx.AddMessage(llm.Message{Role: llm.RoleUser, Content: "First"})
	ctx.AddMessage(llm.Message{Role: llm.RoleUser, Content: "Second"})

	ctx.RemoveLastMessage()

	if ctx.Size() != 1 {
		t.Errorf("Expected 1 message, got %d", ctx.Size())
	}

	messages := ctx.GetMessages()
	if messages[0].Content != "First" {
		t.Error("Wrong message removed")
	}
}

func TestContextRemoveLastN(t *testing.T) {
	ctx := NewContext()

	for i := 0; i < 5; i++ {
		ctx.AddMessage(llm.Message{Role: llm.RoleUser, Content: string(rune('A' + i))})
	}

	ctx.RemoveLastN(3)

	if ctx.Size() != 2 {
		t.Errorf("Expected 2 messages, got %d", ctx.Size())
	}
}

func TestContextStats(t *testing.T) {
	ctx := NewContext()

	ctx.AddMessage(llm.Message{Role: llm.RoleSystem, Content: "System"})
	ctx.AddMessage(llm.Message{Role: llm.RoleUser, Content: "User 1"})
	ctx.AddMessage(llm.Message{Role: llm.RoleUser, Content: "User 2"})
	ctx.AddMessage(llm.Message{Role: llm.RoleAssistant, Content: "Assistant"})

	stats := ctx.Stats()

	if stats.MessageCount != 4 {
		t.Errorf("Expected 4 messages, got %d", stats.MessageCount)
	}

	if stats.UserMessages != 2 {
		t.Errorf("Expected 2 user messages, got %d", stats.UserMessages)
	}

	if stats.AssistantMessages != 1 {
		t.Errorf("Expected 1 assistant message, got %d", stats.AssistantMessages)
	}

	if stats.SystemMessages != 1 {
		t.Errorf("Expected 1 system message, got %d", stats.SystemMessages)
	}
}

func TestContextWithConfig(t *testing.T) {
	config := ContextConfig{
		MaxMessages:   50,
		MaxTokens:     10000,
		CharsPerToken: 3,
	}

	ctx := NewContextWithConfig(config)

	gotConfig := ctx.GetConfig()

	if gotConfig.MaxMessages != 50 {
		t.Errorf("Expected MaxMessages 50, got %d", gotConfig.MaxMessages)
	}

	if gotConfig.MaxTokens != 10000 {
		t.Errorf("Expected MaxTokens 10000, got %d", gotConfig.MaxTokens)
	}

	if gotConfig.CharsPerToken != 3 {
		t.Errorf("Expected CharsPerToken 3, got %d", gotConfig.CharsPerToken)
	}
}

func TestContextSetMaxTokens(t *testing.T) {
	ctx := NewContext()

	// Add messages totaling ~100 chars = ~25 tokens at 4 chars/token
	for i := 0; i < 10; i++ {
		ctx.AddMessage(llm.Message{Role: llm.RoleUser, Content: "1234567890"}) // 10 chars each
	}

	initialTokens := ctx.EstimateTokens()

	// Set max tokens to a lower value
	ctx.SetMaxTokens(15)

	// Context should be pruned to have fewer tokens than before
	afterTokens := ctx.EstimateTokens()
	if afterTokens >= initialTokens {
		t.Errorf("Expected tokens to decrease after SetMaxTokens, before: %d, after: %d", initialTokens, afterTokens)
	}
}
