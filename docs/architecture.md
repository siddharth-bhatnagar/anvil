# Anvil Architecture

Technical documentation of Anvil's architecture, components, and design decisions.

## Table of Contents

1. [Overview](#overview)
2. [Directory Structure](#directory-structure)
3. [Core Components](#core-components)
4. [Data Flow](#data-flow)
5. [Component Details](#component-details)
6. [Design Decisions](#design-decisions)
7. [Extension Points](#extension-points)

---

## Overview

Anvil is built with Go, following a modular architecture that separates concerns:

```
┌─────────────────────────────────────────────────────────────────┐
│                        Anvil CLI Binary                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   ┌─────────────────┐         ┌─────────────────┐              │
│   │    TUI Layer    │◄───────►│   Agent Engine   │              │
│   │   (Bubbletea)   │         │   (Lifecycle)    │              │
│   └────────┬────────┘         └────────┬────────┘              │
│            │                           │                        │
│            ▼                           ▼                        │
│   ┌─────────────────┐         ┌─────────────────┐              │
│   │   LLM Client    │◄───────►│   Tool System    │              │
│   │ (Multi-provider)│         │ (FS, Git, Shell) │              │
│   └─────────────────┘         └─────────────────┘              │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Key Technologies

- **Language**: Go 1.21+
- **TUI Framework**: Bubbletea (Elm architecture)
- **Styling**: Lipgloss
- **Components**: Bubbles (viewport, textarea)
- **Config**: Viper
- **Logging**: Zerolog

---

## Directory Structure

```
anvil/
├── cmd/anvil/
│   └── main.go              # Entry point
│
├── internal/
│   ├── agent/               # Agent engine & lifecycle
│   │   ├── engine.go        # Core agent loop
│   │   ├── lifecycle.go     # Understand→Plan→Act→Verify
│   │   ├── context.go       # Conversation context
│   │   ├── approval.go      # Approval gates
│   │   ├── session.go       # Session persistence
│   │   ├── teaching.go      # Teaching mode
│   │   └── changes.go       # Multi-file coordination
│   │
│   ├── tui/                 # Terminal UI
│   │   ├── app.go           # Main Bubbletea model
│   │   ├── styles.go        # Lipgloss themes
│   │   ├── messages.go      # TUI messages
│   │   ├── panel.go         # Panel interface
│   │   ├── panel_manager.go # Panel orchestration
│   │   ├── panels/          # Panel implementations
│   │   │   ├── conversation.go
│   │   │   ├── diff.go
│   │   │   ├── plan.go
│   │   │   └── files.go
│   │   └── components/      # Reusable components
│   │       ├── spinner.go
│   │       ├── statusbar.go
│   │       ├── errors.go
│   │       └── markdown.go
│   │
│   ├── llm/                 # LLM providers
│   │   ├── client.go        # Provider interface
│   │   ├── anthropic.go     # Anthropic Claude
│   │   ├── openai.go        # OpenAI GPT
│   │   ├── types.go         # Common types
│   │   ├── retry.go         # Retry logic
│   │   └── token_tracker.go # Usage tracking
│   │
│   ├── tools/               # Tool system
│   │   ├── tool.go          # Tool interface
│   │   ├── registry.go      # Tool registry
│   │   ├── filesystem.go    # File operations
│   │   ├── git.go           # Git operations
│   │   ├── shell.go         # Shell commands
│   │   └── analysis.go      # Code analysis
│   │
│   ├── analysis/            # Code analysis
│   │   ├── syntax.go        # Syntax highlighting
│   │   └── symbols.go       # Symbol extraction
│   │
│   ├── config/              # Configuration
│   │   ├── manager.go       # Config management
│   │   ├── keys.go          # Keychain integration
│   │   └── defaults.go      # Default values
│   │
│   └── util/                # Utilities
│       ├── logger.go        # Logging setup
│       └── diff.go          # Diff generation
│
├── pkg/schema/              # Public schemas
│   ├── tool.go              # Tool definitions
│   └── message.go           # Message types
│
├── docs/                    # Documentation
│   ├── quickstart.md
│   ├── user-guide.md
│   └── architecture.md
│
└── .claude/                 # Agent configuration
    └── CLAUDE.md            # Behavioral contract
```

---

## Core Components

### 1. TUI Layer (`internal/tui/`)

The TUI uses the Bubbletea framework with the Elm architecture:

```go
type Model struct {
    panels       *PanelManager
    statusBar    *StatusBar
    agent        *Agent
    width, height int
    // ...
}

func (m Model) Init() tea.Cmd
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m Model) View() string
```

**Responsibilities:**
- Render multi-panel interface
- Handle keyboard/mouse input
- Display streaming responses
- Manage panel focus and layout

### 2. Agent Engine (`internal/agent/`)

The agent orchestrates the interaction lifecycle:

```go
type Agent struct {
    client        llm.Client
    tools         *tools.Registry
    context       *Context
    lifecycle     *Lifecycle
    changeManager *ChangeManager
}

func (a *Agent) Process(input string) (*Response, error)
func (a *Agent) ExecuteTool(call ToolCall) (*ToolResult, error)
```

**Lifecycle Phases:**
1. **Understand**: Parse request, identify ambiguities
2. **Plan**: Generate execution plan
3. **Act**: Execute tools with approval
4. **Verify**: Confirm results, suggest testing

### 3. LLM Client (`internal/llm/`)

Provider-agnostic LLM integration:

```go
type Client interface {
    Chat(ctx context.Context, req *Request) (*Response, error)
    Stream(ctx context.Context, req *Request) (<-chan StreamEvent, error)
    CountTokens(messages []Message) (int, error)
}
```

**Features:**
- Streaming responses (SSE)
- Automatic retry with backoff
- Token counting and tracking
- Multi-provider support

### 4. Tool System (`internal/tools/`)

Extensible tool framework:

```go
type Tool interface {
    Name() string
    Description() string
    Schema() ToolSchema
    Execute(ctx context.Context, args map[string]any) (*Result, error)
    RequiresApproval() bool
}
```

**Built-in Tools:**
- `read_file`: Read file contents
- `write_file`: Create/modify files
- `list_files`: List directory contents
- `search_files`: Search file contents
- `git_status`: Get git status
- `git_diff`: Get git diff
- `run_command`: Execute shell commands

---

## Data Flow

### Request Flow

```
User Input
    │
    ▼
┌─────────┐
│   TUI   │ ──► Parse input
└────┬────┘
     │
     ▼
┌─────────┐
│  Agent  │ ──► Determine phase
└────┬────┘
     │
     ▼
┌─────────┐
│   LLM   │ ──► Generate response
└────┬────┘
     │
     ▼
┌─────────┐
│  Tools  │ ──► Execute if needed
└────┬────┘
     │
     ▼
┌─────────┐
│Approval │ ──► User confirms
└────┬────┘
     │
     ▼
  Response
```

### Streaming Flow

```
LLM API
    │
    ▼ (SSE events)
┌─────────────┐
│ LLM Client  │ ──► Parse stream
└──────┬──────┘
       │
       ▼ (StreamEvent)
┌─────────────┐
│    Agent    │ ──► Process content
└──────┬──────┘
       │
       ▼ (tea.Msg)
┌─────────────┐
│     TUI     │ ──► Update display
└─────────────┘
```

---

## Component Details

### Panel System

Each panel implements the Panel interface:

```go
type Panel interface {
    Init() tea.Cmd
    Update(msg tea.Msg) tea.Cmd
    View() string
    SetSize(width, height int)
    Focus()
    Blur()
    IsFocused() bool
    Type() PanelType
    Title() string
}
```

**Panel Manager** handles:
- Layout calculation
- Focus management
- Panel switching
- Resize events

### Context Management

The Context maintains conversation state:

```go
type Context struct {
    messages     []llm.Message
    systemPrompt string
    maxTokens    int
    tools        []llm.Tool
}
```

**Features:**
- Automatic pruning when context exceeds limits
- Message summarization
- Tool result integration

### Approval System

Changes require explicit approval:

```go
type ApprovalRequest struct {
    Type        ApprovalType
    Description string
    Changes     []*FileChange
    Command     string
}

type ApprovalGate struct {
    pending chan *ApprovalRequest
    results chan ApprovalResult
}
```

### Session Persistence

Sessions are saved as JSON:

```go
type Session struct {
    ID        string
    Name      string
    CreatedAt time.Time
    UpdatedAt time.Time
    Messages  []llm.Message
    Metadata  SessionMeta
}
```

---

## Design Decisions

### 1. Go over Rust

**Decision**: Use Go for implementation

**Rationale**:
- Faster development velocity
- Excellent TUI ecosystem (Charm)
- Strong standard library
- Easy cross-platform compilation
- Better community accessibility

### 2. Elm Architecture (Bubbletea)

**Decision**: Use Bubbletea's Elm architecture

**Rationale**:
- Predictable state management
- Easy testing
- Clear separation of concerns
- Well-documented patterns

### 3. Propose-Then-Execute

**Decision**: All modifications require approval

**Rationale**:
- Aligns with "Safety > Speed" philosophy
- Prevents accidental damage
- Builds user trust
- Enables learning from diffs

### 4. Provider Abstraction

**Decision**: Abstract LLM providers from day one

**Rationale**:
- Avoid vendor lock-in
- Enable model switching
- Support local models
- Future-proof design

### 5. File-Based Logging

**Decision**: Log to files, not stdout

**Rationale**:
- TUI controls stdout
- Persistent debugging
- Structured logging (JSON)
- Easy log rotation

### 6. OS Keychain for Secrets

**Decision**: Store API keys in OS keychain

**Rationale**:
- Maximum security
- Platform-native integration
- No plaintext storage
- Follows security best practices

---

## Extension Points

### Adding a New LLM Provider

1. Implement the `Client` interface:

```go
type MyProvider struct {
    apiKey string
    // ...
}

func (p *MyProvider) Chat(ctx context.Context, req *Request) (*Response, error) {
    // Implementation
}
```

2. Register in factory:

```go
func NewClient(provider, apiKey string) (Client, error) {
    switch provider {
    case "myprovider":
        return NewMyProvider(apiKey), nil
    // ...
    }
}
```

### Adding a New Tool

1. Implement the `Tool` interface:

```go
type MyTool struct{}

func (t *MyTool) Name() string { return "my_tool" }
func (t *MyTool) Execute(ctx context.Context, args map[string]any) (*Result, error) {
    // Implementation
}
```

2. Register in registry:

```go
registry.Register(&MyTool{})
```

### Adding a New Panel

1. Implement the `Panel` interface
2. Add to PanelManager
3. Define keyboard shortcuts
4. Update layout calculation

---

## Testing Strategy

### Unit Tests

- Test individual components in isolation
- Mock external dependencies
- Target 80%+ coverage

### Integration Tests

- Test component interactions
- Use test fixtures for file operations
- Mock LLM responses

### E2E Tests

- Full workflow testing
- Automated TUI interaction
- Real API calls (optional, with test keys)

---

## Performance Considerations

### Streaming

- Process tokens as they arrive
- Update UI incrementally
- Avoid blocking main thread

### Large Files

- Lazy loading
- Pagination for diffs
- Memory-efficient parsing

### Context Window

- Smart pruning strategies
- Summarization for long conversations
- User-visible context usage

---

## Security

### API Keys

- Never logged or displayed
- Stored in OS keychain
- Memory cleared after use

### File Operations

- Sandboxed to project directory
- Explicit approval required
- Audit logging

### Shell Commands

- Whitelist approach for safe commands
- Destructive commands flagged
- Dry-run option available

---

## Future Considerations

### Planned Enhancements

1. Plugin system for custom tools
2. Local model support (Ollama, llama.cpp)
3. Collaborative features
4. IDE integrations
5. Custom themes

### Scalability

- Designed for single-user, single-project
- Could extend to project workspaces
- Potential for daemon mode
