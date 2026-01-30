package panels

// PanelType represents the type of panel
type PanelType int

const (
	PanelConversation PanelType = iota
	PanelFiles
	PanelDiff
	PanelPlan
)

// String returns the string representation of the panel type
func (p PanelType) String() string {
	switch p {
	case PanelConversation:
		return "Conversation"
	case PanelFiles:
		return "Files"
	case PanelDiff:
		return "Diff"
	case PanelPlan:
		return "Plan"
	default:
		return "Unknown"
	}
}
