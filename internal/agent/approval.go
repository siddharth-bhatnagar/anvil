package agent

import (
	"fmt"
	"sync"

	"github.com/siddharth-bhatnagar/anvil/pkg/schema"
)

// ApprovalStatus represents the status of an approval request
type ApprovalStatus int

const (
	ApprovalPending ApprovalStatus = iota
	ApprovalApproved
	ApprovalRejected
)

// String returns the string representation of an approval status
func (s ApprovalStatus) String() string {
	switch s {
	case ApprovalPending:
		return "Pending"
	case ApprovalApproved:
		return "Approved"
	case ApprovalRejected:
		return "Rejected"
	default:
		return "Unknown"
	}
}

// ApprovalItem represents a single approval request with its status
type ApprovalItem struct {
	ID        string
	ToolCall  schema.ToolCall
	Request   schema.ApprovalRequest
	Status    ApprovalStatus
	Reason    string // Reason for rejection if rejected
}

// ApprovalManager manages pending approval requests
type ApprovalManager struct {
	mu       sync.RWMutex
	items    map[string]*ApprovalItem
	order    []string // Maintains order of additions
	callback ApprovalCallback
}

// ApprovalCallback is called when an approval decision is made
type ApprovalCallback func(item *ApprovalItem)

// NewApprovalManager creates a new approval manager
func NewApprovalManager() *ApprovalManager {
	return &ApprovalManager{
		items: make(map[string]*ApprovalItem),
		order: make([]string, 0),
	}
}

// SetCallback sets the callback for approval decisions
func (am *ApprovalManager) SetCallback(cb ApprovalCallback) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.callback = cb
}

// Add adds a new approval request
func (am *ApprovalManager) Add(toolCall schema.ToolCall, request schema.ApprovalRequest) *ApprovalItem {
	am.mu.Lock()
	defer am.mu.Unlock()

	id := fmt.Sprintf("approval_%d", len(am.items))
	item := &ApprovalItem{
		ID:       id,
		ToolCall: toolCall,
		Request:  request,
		Status:   ApprovalPending,
	}

	am.items[id] = item
	am.order = append(am.order, id)

	return item
}

// AddPending adds multiple pending approvals from a Response
func (am *ApprovalManager) AddPending(approvals []PendingApproval) []*ApprovalItem {
	var items []*ApprovalItem
	for _, pa := range approvals {
		item := am.Add(pa.ToolCall, pa.Request)
		items = append(items, item)
	}
	return items
}

// Get retrieves an approval item by ID
func (am *ApprovalManager) Get(id string) *ApprovalItem {
	am.mu.RLock()
	defer am.mu.RUnlock()

	return am.items[id]
}

// GetPending returns all pending approval items in order
func (am *ApprovalManager) GetPending() []*ApprovalItem {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var pending []*ApprovalItem
	for _, id := range am.order {
		item := am.items[id]
		if item != nil && item.Status == ApprovalPending {
			pending = append(pending, item)
		}
	}

	return pending
}

// GetAll returns all approval items in order
func (am *ApprovalManager) GetAll() []*ApprovalItem {
	am.mu.RLock()
	defer am.mu.RUnlock()

	items := make([]*ApprovalItem, 0, len(am.order))
	for _, id := range am.order {
		if item := am.items[id]; item != nil {
			items = append(items, item)
		}
	}

	return items
}

// Approve marks an approval item as approved
func (am *ApprovalManager) Approve(id string) error {
	am.mu.Lock()
	item, exists := am.items[id]
	if !exists {
		am.mu.Unlock()
		return fmt.Errorf("approval item %s not found", id)
	}

	if item.Status != ApprovalPending {
		am.mu.Unlock()
		return fmt.Errorf("approval item %s is not pending", id)
	}

	item.Status = ApprovalApproved
	callback := am.callback
	am.mu.Unlock()

	// Call callback outside of lock
	if callback != nil {
		callback(item)
	}

	return nil
}

// Reject marks an approval item as rejected
func (am *ApprovalManager) Reject(id string, reason string) error {
	am.mu.Lock()
	item, exists := am.items[id]
	if !exists {
		am.mu.Unlock()
		return fmt.Errorf("approval item %s not found", id)
	}

	if item.Status != ApprovalPending {
		am.mu.Unlock()
		return fmt.Errorf("approval item %s is not pending", id)
	}

	item.Status = ApprovalRejected
	item.Reason = reason
	callback := am.callback
	am.mu.Unlock()

	// Call callback outside of lock
	if callback != nil {
		callback(item)
	}

	return nil
}

// ApproveAll approves all pending items
func (am *ApprovalManager) ApproveAll() []string {
	pending := am.GetPending()
	var approved []string

	for _, item := range pending {
		if err := am.Approve(item.ID); err == nil {
			approved = append(approved, item.ID)
		}
	}

	return approved
}

// RejectAll rejects all pending items
func (am *ApprovalManager) RejectAll(reason string) []string {
	pending := am.GetPending()
	var rejected []string

	for _, item := range pending {
		if err := am.Reject(item.ID, reason); err == nil {
			rejected = append(rejected, item.ID)
		}
	}

	return rejected
}

// HasPending returns true if there are pending approvals
func (am *ApprovalManager) HasPending() bool {
	am.mu.RLock()
	defer am.mu.RUnlock()

	for _, item := range am.items {
		if item.Status == ApprovalPending {
			return true
		}
	}

	return false
}

// PendingCount returns the number of pending approvals
func (am *ApprovalManager) PendingCount() int {
	am.mu.RLock()
	defer am.mu.RUnlock()

	count := 0
	for _, item := range am.items {
		if item.Status == ApprovalPending {
			count++
		}
	}

	return count
}

// Clear removes all approval items
func (am *ApprovalManager) Clear() {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.items = make(map[string]*ApprovalItem)
	am.order = make([]string, 0)
}

// ClearResolved removes all resolved (approved/rejected) items
func (am *ApprovalManager) ClearResolved() {
	am.mu.Lock()
	defer am.mu.Unlock()

	newItems := make(map[string]*ApprovalItem)
	var newOrder []string

	for _, id := range am.order {
		item := am.items[id]
		if item != nil && item.Status == ApprovalPending {
			newItems[id] = item
			newOrder = append(newOrder, id)
		}
	}

	am.items = newItems
	am.order = newOrder
}

// FormatApprovalRequest formats an approval request for display
func FormatApprovalRequest(item *ApprovalItem) string {
	var result string

	result += fmt.Sprintf("Action: %s\n", item.Request.Action)
	result += fmt.Sprintf("Reason: %s\n", item.Request.Reason)

	if item.Request.Destructive {
		result += "⚠️  This is a destructive operation\n"
	}

	if item.Request.Preview != "" {
		result += fmt.Sprintf("\nPreview:\n%s\n", item.Request.Preview)
	}

	result += fmt.Sprintf("\nTool: %s\n", item.ToolCall.Name)
	if len(item.ToolCall.Arguments) > 0 {
		result += "Arguments:\n"
		for key, val := range item.ToolCall.Arguments {
			result += fmt.Sprintf("  %s: %v\n", key, val)
		}
	}

	return result
}
