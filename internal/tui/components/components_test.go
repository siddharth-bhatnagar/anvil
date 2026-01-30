package components

import (
	"strings"
	"testing"
	"time"
)

// Spinner tests
func TestNewSpinner(t *testing.T) {
	tests := []struct {
		name  string
		style SpinnerStyle
	}{
		{"dots", SpinnerDots},
		{"line", SpinnerLine},
		{"circle", SpinnerCircle},
		{"bounce", SpinnerBounce},
		{"pulse", SpinnerPulse},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSpinner(tt.style)
			if s == nil {
				t.Error("NewSpinner returned nil")
			}
			if len(s.frames) == 0 {
				t.Error("Spinner has no frames")
			}
			if s.active {
				t.Error("Spinner should not be active by default")
			}
		})
	}
}

func TestNewSpinnerWithMessage(t *testing.T) {
	msg := "Loading..."
	s := NewSpinnerWithMessage(SpinnerDots, msg)

	if s.message != msg {
		t.Errorf("Expected message %q, got %q", msg, s.message)
	}
}

func TestSpinnerStartStop(t *testing.T) {
	s := NewSpinner(SpinnerDots)

	// Start
	cmd := s.Start()
	if !s.IsActive() {
		t.Error("Spinner should be active after Start")
	}
	if cmd == nil {
		t.Error("Start should return a command")
	}

	// Stop
	s.Stop()
	if s.IsActive() {
		t.Error("Spinner should not be active after Stop")
	}
}

func TestSpinnerView(t *testing.T) {
	s := NewSpinner(SpinnerDots)

	// Not active - should return empty
	view := s.View()
	if view != "" {
		t.Error("Inactive spinner should return empty view")
	}

	// Active - should return frame
	s.Start()
	view = s.View()
	if view == "" {
		t.Error("Active spinner should return non-empty view")
	}

	// With message
	s.SetMessage("Loading")
	view = s.View()
	if !strings.Contains(view, "Loading") {
		t.Error("Spinner view should contain message")
	}
}

func TestSpinnerUpdate(t *testing.T) {
	s := NewSpinner(SpinnerDots)
	s.Start()

	initialIndex := s.index

	// Send tick message
	cmd := s.Update(SpinnerTickMsg(time.Now()))

	if s.index == initialIndex {
		t.Error("Spinner index should change after tick")
	}
	if cmd == nil {
		t.Error("Update should return a command when active")
	}

	// When not active
	s.Stop()
	cmd = s.Update(SpinnerTickMsg(time.Now()))
	if cmd != nil {
		t.Error("Update should return nil when not active")
	}
}

func TestSpinnerFrames(t *testing.T) {
	tests := []struct {
		style    SpinnerStyle
		expected int // minimum number of frames
	}{
		{SpinnerDots, 10},
		{SpinnerLine, 4},
		{SpinnerCircle, 4},
		{SpinnerBounce, 4},
		{SpinnerPulse, 6},
	}

	for _, tt := range tests {
		frames := SpinnerFrames(tt.style)
		if len(frames) < tt.expected {
			t.Errorf("SpinnerStyle %d: expected at least %d frames, got %d", tt.style, tt.expected, len(frames))
		}
	}
}

// ProgressBar tests
func TestNewProgressBar(t *testing.T) {
	pb := NewProgressBar(50)
	if pb == nil {
		t.Error("NewProgressBar returned nil")
	}
	if pb.width != 50 {
		t.Errorf("Expected width 50, got %d", pb.width)
	}
	if pb.progress != 0 {
		t.Error("Progress should start at 0")
	}
}

func TestProgressBarSetProgress(t *testing.T) {
	pb := NewProgressBar(50)

	tests := []struct {
		input    float64
		expected float64
	}{
		{0.5, 0.5},
		{0, 0},
		{1, 1},
		{-0.5, 0},   // Clamped to 0
		{1.5, 1},    // Clamped to 1
	}

	for _, tt := range tests {
		pb.SetProgress(tt.input)
		if pb.progress != tt.expected {
			t.Errorf("SetProgress(%v): expected %v, got %v", tt.input, tt.expected, pb.progress)
		}
	}
}

func TestProgressBarView(t *testing.T) {
	pb := NewProgressBar(20)

	// Empty
	pb.SetProgress(0)
	view := pb.View()
	if view == "" {
		t.Error("View should not be empty")
	}

	// Half
	pb.SetProgress(0.5)
	view = pb.View()
	if !strings.Contains(view, "█") {
		t.Error("View should contain filled blocks")
	}

	// Full
	pb.SetProgress(1)
	view = pb.View()
	if strings.Contains(view, "░") {
		t.Error("Full progress should not have empty blocks")
	}

	// Without percentage
	pb.ShowPercentage(false)
	view = pb.View()
	if strings.Contains(view, "%") {
		t.Error("View should not contain percentage when disabled")
	}
}

// StatusIndicator tests
func TestNewStatusIndicator(t *testing.T) {
	si := NewStatusIndicator()
	if si == nil {
		t.Error("NewStatusIndicator returned nil")
	}
	if si.status != StatusIdle {
		t.Error("Initial status should be StatusIdle")
	}
}

func TestStatusIndicatorSetStatus(t *testing.T) {
	si := NewStatusIndicator()

	statuses := []Status{StatusIdle, StatusLoading, StatusSuccess, StatusWarning, StatusError}

	for _, status := range statuses {
		si.SetStatus(status, "test message")
		if si.status != status {
			t.Errorf("Expected status %d, got %d", status, si.status)
		}
		if si.message != "test message" {
			t.Error("Message not set correctly")
		}
	}
}

func TestStatusIndicatorView(t *testing.T) {
	si := NewStatusIndicator()

	si.SetStatus(StatusSuccess, "Done")
	view := si.View()

	if view == "" {
		t.Error("View should not be empty")
	}
	if !strings.Contains(view, "Done") {
		t.Error("View should contain message")
	}
}

// Toast tests
func TestNewToast(t *testing.T) {
	toast := NewToast("Test message", 3*time.Second)

	if toast == nil {
		t.Error("NewToast returned nil")
	}
	if toast.message != "Test message" {
		t.Error("Message not set correctly")
	}
	if toast.duration != 3*time.Second {
		t.Error("Duration not set correctly")
	}
	if !toast.visible {
		t.Error("Toast should be visible by default")
	}
}

func TestToastVariants(t *testing.T) {
	success := NewSuccessToast("Success!")
	if success == nil || success.message != "Success!" {
		t.Error("NewSuccessToast failed")
	}

	error := NewErrorToast("Error!")
	if error == nil || error.message != "Error!" {
		t.Error("NewErrorToast failed")
	}

	warning := NewWarningToast("Warning!")
	if warning == nil || warning.message != "Warning!" {
		t.Error("NewWarningToast failed")
	}
}

func TestToastShowHide(t *testing.T) {
	toast := NewToast("Test", time.Second)

	// Show
	cmd := toast.Show()
	if cmd == nil {
		t.Error("Show should return a command")
	}
	if !toast.IsVisible() {
		t.Error("Toast should be visible after Show")
	}

	// Hide
	toast.Hide()
	if toast.IsVisible() {
		t.Error("Toast should not be visible after Hide")
	}
}

func TestToastView(t *testing.T) {
	toast := NewToast("Hello", time.Second)

	// Visible
	view := toast.View()
	if !strings.Contains(view, "Hello") {
		t.Error("View should contain message")
	}

	// Hidden
	toast.Hide()
	view = toast.View()
	if view != "" {
		t.Error("Hidden toast should return empty view")
	}
}

// StatusBar tests
func TestNewStatusBar(t *testing.T) {
	sb := NewStatusBar(100)

	if sb == nil {
		t.Error("NewStatusBar returned nil")
	}
	if sb.width != 100 {
		t.Errorf("Expected width 100, got %d", sb.width)
	}
}

func TestStatusBarItems(t *testing.T) {
	sb := NewStatusBar(100)

	sb.AddLeftItem(StatusBarItem{Label: "Mode", Value: "Normal"})
	sb.AddRightItem(StatusBarItem{Label: "Help", Value: "?"})

	if len(sb.leftItems) != 1 {
		t.Error("Left item not added")
	}
	if len(sb.rightItems) != 1 {
		t.Error("Right item not added")
	}
}

func TestStatusBarSetItem(t *testing.T) {
	sb := NewStatusBar(100)

	sb.AddLeftItem(StatusBarItem{Label: "Mode", Value: "Normal"})
	sb.SetLeftItem("Mode", "Insert")

	if sb.leftItems[0].Value != "Insert" {
		t.Error("Left item not updated")
	}

	// Set non-existent item (should add)
	sb.SetLeftItem("New", "Value")
	if len(sb.leftItems) != 2 {
		t.Error("New item should be added")
	}
}

func TestStatusBarShowHide(t *testing.T) {
	sb := NewStatusBar(100)

	sb.AddLeftItem(StatusBarItem{Label: "Test", Value: "Value"})

	sb.HideItem("Test")
	if sb.leftItems[0].Visible {
		t.Error("Item should be hidden")
	}

	sb.ShowItem("Test")
	if !sb.leftItems[0].Visible {
		t.Error("Item should be visible")
	}
}

func TestStatusBarView(t *testing.T) {
	sb := NewStatusBar(80)

	sb.AddLeftItem(StatusBarItem{Label: "Mode", Value: "Normal", Visible: true})
	sb.AddRightItem(StatusBarItem{Value: "Help", Visible: true})

	view := sb.View()

	if view == "" {
		t.Error("View should not be empty")
	}
	if !strings.Contains(view, "Normal") {
		t.Error("View should contain left item value")
	}
}

func TestDefaultStatusBar(t *testing.T) {
	sb := DefaultStatusBar(100)

	if sb == nil {
		t.Error("DefaultStatusBar returned nil")
	}
	if len(sb.leftItems) == 0 {
		t.Error("Default status bar should have left items")
	}
	if len(sb.rightItems) == 0 {
		t.Error("Default status bar should have right items")
	}
}

func TestFormatKeyHints(t *testing.T) {
	hints := []KeyHint{
		{Key: "q", Description: "Quit"},
		{Key: "?", Description: "Help"},
	}

	result := FormatKeyHints(hints)

	if !strings.Contains(result, "q") {
		t.Error("Result should contain key")
	}
	if !strings.Contains(result, "Quit") {
		t.Error("Result should contain description")
	}
}

func TestModeIndicator(t *testing.T) {
	mi := NewModeIndicator("Normal")

	if mi.mode != "Normal" {
		t.Error("Mode not set correctly")
	}

	mi.SetMode("Insert")
	if mi.mode != "Insert" {
		t.Error("SetMode failed")
	}

	mi.SetModeStyle("Dangerous", true)
	view := mi.View()
	if !strings.Contains(view, "Dangerous") {
		t.Error("View should contain mode")
	}
}

// ErrorDisplay tests
func TestNewErrorDisplay(t *testing.T) {
	ed := NewErrorDisplay(SeverityError, "Test Error", "Something went wrong")

	if ed == nil {
		t.Error("NewErrorDisplay returned nil")
	}
	if ed.title != "Test Error" {
		t.Error("Title not set correctly")
	}
	if ed.message != "Something went wrong" {
		t.Error("Message not set correctly")
	}
}

func TestErrorDisplaySuggestions(t *testing.T) {
	ed := NewErrorDisplay(SeverityError, "Error", "Message")

	ed.AddSuggestion("Try this")
	ed.AddSuggestion("Or that")

	if len(ed.suggestions) != 2 {
		t.Error("Suggestions not added correctly")
	}
}

func TestErrorDisplayView(t *testing.T) {
	ed := NewErrorDisplay(SeverityWarning, "Warning", "Be careful")
	ed.SetDetails("More info here")
	ed.AddSuggestion("Do this instead")

	view := ed.View()

	if !strings.Contains(view, "Warning") {
		t.Error("View should contain title")
	}
	if !strings.Contains(view, "Be careful") {
		t.Error("View should contain message")
	}
	if !strings.Contains(view, "More info") {
		t.Error("View should contain details")
	}
	if !strings.Contains(view, "Do this") {
		t.Error("View should contain suggestion")
	}
}

func TestQuickHelpers(t *testing.T) {
	error := QuickError("Error message")
	if !strings.Contains(error, "Error message") {
		t.Error("QuickError failed")
	}

	warning := QuickWarning("Warning message")
	if !strings.Contains(warning, "Warning message") {
		t.Error("QuickWarning failed")
	}

	info := QuickInfo("Info message")
	if !strings.Contains(info, "Info message") {
		t.Error("QuickInfo failed")
	}
}

func TestAPIError(t *testing.T) {
	ed := APIError("anthropic", 401, "Unauthorized")

	if ed.severity != SeverityError {
		t.Error("Severity should be Error")
	}
	if len(ed.suggestions) == 0 {
		t.Error("APIError should add suggestions for 401")
	}
}

func TestFileError(t *testing.T) {
	ed := FileError("read", "/path/to/file", &testError{"permission denied"})

	if ed.severity != SeverityError {
		t.Error("Severity should be Error")
	}
	if len(ed.suggestions) == 0 {
		t.Error("FileError should add suggestions")
	}
}

func TestSeverityMethods(t *testing.T) {
	severities := []ErrorSeverity{SeverityInfo, SeverityWarning, SeverityError, SeverityCritical}

	for _, s := range severities {
		if s.String() == "" {
			t.Errorf("Severity %d has empty string", s)
		}
		if s.Icon() == "" {
			t.Errorf("Severity %d has empty icon", s)
		}
		// Color returns a type, just check it doesn't panic
		_ = s.Color()
	}
}

// MarkdownRenderer tests
func TestNewMarkdownRenderer(t *testing.T) {
	mr := NewMarkdownRenderer(80)

	if mr == nil {
		t.Error("NewMarkdownRenderer returned nil")
	}
	if mr.width != 80 {
		t.Errorf("Expected width 80, got %d", mr.width)
	}
}

func TestMarkdownRenderHeaders(t *testing.T) {
	mr := NewMarkdownRenderer(80)

	tests := []struct {
		input    string
		contains string
	}{
		{"# Header 1", "Header 1"},
		{"## Header 2", "Header 2"},
		{"### Header 3", "Header 3"},
	}

	for _, tt := range tests {
		result := mr.Render(tt.input)
		if !strings.Contains(result, tt.contains) {
			t.Errorf("Render(%q) should contain %q", tt.input, tt.contains)
		}
	}
}

func TestMarkdownRenderLists(t *testing.T) {
	mr := NewMarkdownRenderer(80)

	bullet := mr.Render("- Item 1\n- Item 2")
	if !strings.Contains(bullet, "Item 1") || !strings.Contains(bullet, "Item 2") {
		t.Error("Bullet list not rendered correctly")
	}

	numbered := mr.Render("1. First\n2. Second")
	if !strings.Contains(numbered, "First") || !strings.Contains(numbered, "Second") {
		t.Error("Numbered list not rendered correctly")
	}
}

func TestMarkdownRenderCodeBlock(t *testing.T) {
	mr := NewMarkdownRenderer(80)

	code := mr.Render("```go\nfunc main() {}\n```")
	if !strings.Contains(code, "func main()") {
		t.Error("Code block not rendered correctly")
	}
}

func TestMarkdownRenderBlockquote(t *testing.T) {
	mr := NewMarkdownRenderer(80)

	quote := mr.Render("> This is a quote")
	if !strings.Contains(quote, "This is a quote") {
		t.Error("Blockquote not rendered correctly")
	}
}

func TestMarkdownRenderInline(t *testing.T) {
	mr := NewMarkdownRenderer(80)

	// Bold
	bold := mr.Render("This is **bold** text")
	if !strings.Contains(bold, "bold") {
		t.Error("Bold not rendered")
	}

	// Inline code
	code := mr.Render("Use `code` here")
	if !strings.Contains(code, "code") {
		t.Error("Inline code not rendered")
	}

	// Links
	link := mr.Render("[Click here](http://example.com)")
	if !strings.Contains(link, "Click here") {
		t.Error("Link not rendered")
	}
}

func TestMarkdownHelpers(t *testing.T) {
	code := CodeBlock("func main() {}", "go", 60)
	if !strings.Contains(code, "func main()") {
		t.Error("CodeBlock failed")
	}

	inline := InlineCode("code")
	if !strings.Contains(inline, "code") {
		t.Error("InlineCode failed")
	}

	bold := Bold("text")
	if !strings.Contains(bold, "text") {
		t.Error("Bold failed")
	}

	italic := Italic("text")
	if !strings.Contains(italic, "text") {
		t.Error("Italic failed")
	}

	h1 := Header("Title", 1)
	if !strings.Contains(h1, "Title") {
		t.Error("Header failed")
	}

	bullets := BulletList([]string{"a", "b"})
	if !strings.Contains(bullets, "a") || !strings.Contains(bullets, "b") {
		t.Error("BulletList failed")
	}

	numbers := NumberedList([]string{"a", "b"})
	if !strings.Contains(numbers, "a") || !strings.Contains(numbers, "b") {
		t.Error("NumberedList failed")
	}

	quote := Quote("quoted")
	if !strings.Contains(quote, "quoted") {
		t.Error("Quote failed")
	}
}

// Helper for tests
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
