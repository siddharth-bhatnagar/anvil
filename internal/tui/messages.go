package tui

import (
	"github.com/siddharth-bhatnagar/anvil/internal/llm"
)

// StreamChunkMsg represents a streaming chunk from the LLM
type StreamChunkMsg struct {
	Delta string
	Done  bool
	Error error
	Usage *llm.Usage
}

// SendMessageMsg is sent when the user submits a message
type SendMessageMsg struct {
	Content string
}

// ErrorMsg represents an error message
type ErrorMsg struct {
	Error error
}
