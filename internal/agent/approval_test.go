package agent

import (
	"testing"

	"github.com/siddharth-bhatnagar/anvil/pkg/schema"
)

func TestNewApprovalManager(t *testing.T) {
	am := NewApprovalManager()

	if am == nil {
		t.Fatal("NewApprovalManager returned nil")
	}

	if am.HasPending() {
		t.Error("New manager should have no pending approvals")
	}
}

func TestApprovalManagerAdd(t *testing.T) {
	am := NewApprovalManager()

	toolCall := schema.ToolCall{
		ID:   "test-1",
		Name: "write_file",
		Arguments: map[string]any{
			"path": "/test/file.txt",
		},
	}

	request := schema.ApprovalRequest{
		Action:      "Write file",
		Reason:      "File modification requires approval",
		Destructive: false,
	}

	item := am.Add(toolCall, request)

	if item == nil {
		t.Fatal("Add returned nil")
	}

	if item.Status != ApprovalPending {
		t.Errorf("Expected status Pending, got %v", item.Status)
	}

	if !am.HasPending() {
		t.Error("Manager should have pending approvals")
	}

	if am.PendingCount() != 1 {
		t.Errorf("Expected 1 pending, got %d", am.PendingCount())
	}
}

func TestApprovalManagerApprove(t *testing.T) {
	am := NewApprovalManager()

	item := am.Add(
		schema.ToolCall{ID: "test-1", Name: "shell_command"},
		schema.ApprovalRequest{Action: "Execute command"},
	)

	err := am.Approve(item.ID)
	if err != nil {
		t.Fatalf("Approve failed: %v", err)
	}

	if item.Status != ApprovalApproved {
		t.Errorf("Expected status Approved, got %v", item.Status)
	}

	if am.HasPending() {
		t.Error("Should have no pending approvals after approval")
	}
}

func TestApprovalManagerReject(t *testing.T) {
	am := NewApprovalManager()

	item := am.Add(
		schema.ToolCall{ID: "test-1", Name: "shell_command"},
		schema.ApprovalRequest{Action: "Execute command"},
	)

	err := am.Reject(item.ID, "Too dangerous")
	if err != nil {
		t.Fatalf("Reject failed: %v", err)
	}

	if item.Status != ApprovalRejected {
		t.Errorf("Expected status Rejected, got %v", item.Status)
	}

	if item.Reason != "Too dangerous" {
		t.Errorf("Expected reason 'Too dangerous', got '%s'", item.Reason)
	}
}

func TestApprovalManagerGet(t *testing.T) {
	am := NewApprovalManager()

	item := am.Add(
		schema.ToolCall{ID: "test-1", Name: "write_file"},
		schema.ApprovalRequest{Action: "Write"},
	)

	got := am.Get(item.ID)
	if got == nil {
		t.Fatal("Get returned nil")
	}

	if got.ID != item.ID {
		t.Errorf("Got wrong item")
	}

	// Get non-existent
	notFound := am.Get("non-existent")
	if notFound != nil {
		t.Error("Should return nil for non-existent ID")
	}
}

func TestApprovalManagerGetPending(t *testing.T) {
	am := NewApprovalManager()

	am.Add(schema.ToolCall{ID: "1", Name: "cmd1"}, schema.ApprovalRequest{})
	am.Add(schema.ToolCall{ID: "2", Name: "cmd2"}, schema.ApprovalRequest{})
	am.Add(schema.ToolCall{ID: "3", Name: "cmd3"}, schema.ApprovalRequest{})

	pending := am.GetPending()
	if len(pending) != 3 {
		t.Errorf("Expected 3 pending, got %d", len(pending))
	}

	// Approve one
	am.Approve(pending[0].ID)

	pending = am.GetPending()
	if len(pending) != 2 {
		t.Errorf("Expected 2 pending after approval, got %d", len(pending))
	}
}

func TestApprovalManagerApproveAll(t *testing.T) {
	am := NewApprovalManager()

	am.Add(schema.ToolCall{ID: "1", Name: "cmd1"}, schema.ApprovalRequest{})
	am.Add(schema.ToolCall{ID: "2", Name: "cmd2"}, schema.ApprovalRequest{})

	approved := am.ApproveAll()

	if len(approved) != 2 {
		t.Errorf("Expected 2 approved, got %d", len(approved))
	}

	if am.HasPending() {
		t.Error("Should have no pending after ApproveAll")
	}
}

func TestApprovalManagerRejectAll(t *testing.T) {
	am := NewApprovalManager()

	am.Add(schema.ToolCall{ID: "1", Name: "cmd1"}, schema.ApprovalRequest{})
	am.Add(schema.ToolCall{ID: "2", Name: "cmd2"}, schema.ApprovalRequest{})

	rejected := am.RejectAll("Batch reject")

	if len(rejected) != 2 {
		t.Errorf("Expected 2 rejected, got %d", len(rejected))
	}

	all := am.GetAll()
	for _, item := range all {
		if item.Status != ApprovalRejected {
			t.Error("All items should be rejected")
		}
		if item.Reason != "Batch reject" {
			t.Errorf("Expected reason 'Batch reject', got '%s'", item.Reason)
		}
	}
}

func TestApprovalManagerClear(t *testing.T) {
	am := NewApprovalManager()

	am.Add(schema.ToolCall{ID: "1", Name: "cmd1"}, schema.ApprovalRequest{})
	am.Add(schema.ToolCall{ID: "2", Name: "cmd2"}, schema.ApprovalRequest{})

	am.Clear()

	if am.PendingCount() != 0 {
		t.Error("Should have no items after Clear")
	}
}

func TestApprovalManagerClearResolved(t *testing.T) {
	am := NewApprovalManager()

	item1 := am.Add(schema.ToolCall{ID: "1", Name: "cmd1"}, schema.ApprovalRequest{})
	am.Add(schema.ToolCall{ID: "2", Name: "cmd2"}, schema.ApprovalRequest{})

	am.Approve(item1.ID)

	am.ClearResolved()

	// Should only have the pending one left
	if am.PendingCount() != 1 {
		t.Errorf("Expected 1 pending after ClearResolved, got %d", am.PendingCount())
	}
}

func TestApprovalManagerCallback(t *testing.T) {
	am := NewApprovalManager()

	callbackCalled := false
	var callbackItem *ApprovalItem

	am.SetCallback(func(item *ApprovalItem) {
		callbackCalled = true
		callbackItem = item
	})

	item := am.Add(schema.ToolCall{ID: "1", Name: "cmd1"}, schema.ApprovalRequest{})
	am.Approve(item.ID)

	if !callbackCalled {
		t.Error("Callback should have been called")
	}

	if callbackItem == nil || callbackItem.ID != item.ID {
		t.Error("Callback received wrong item")
	}
}

func TestApprovalStatusString(t *testing.T) {
	tests := []struct {
		status   ApprovalStatus
		expected string
	}{
		{ApprovalPending, "Pending"},
		{ApprovalApproved, "Approved"},
		{ApprovalRejected, "Rejected"},
	}

	for _, tt := range tests {
		if tt.status.String() != tt.expected {
			t.Errorf("Status %d: Expected '%s', got '%s'", tt.status, tt.expected, tt.status.String())
		}
	}
}

func TestFormatApprovalRequest(t *testing.T) {
	item := &ApprovalItem{
		ID: "test-1",
		ToolCall: schema.ToolCall{
			Name: "write_file",
			Arguments: map[string]any{
				"path":    "/test/file.txt",
				"content": "hello",
			},
		},
		Request: schema.ApprovalRequest{
			Action:      "Write file",
			Reason:      "File modification",
			Destructive: true,
			Preview:     "--- old\n+++ new",
		},
	}

	formatted := FormatApprovalRequest(item)

	if formatted == "" {
		t.Error("FormatApprovalRequest returned empty string")
	}

	// Check it contains key information
	if !contains(formatted, "Write file") {
		t.Error("Should contain action")
	}
	if !contains(formatted, "destructive") {
		t.Error("Should indicate destructive operation")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
