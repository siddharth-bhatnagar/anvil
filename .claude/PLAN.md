Anvil Implementation Plan

Understanding

Anvil is a terminal-based, agentic coding CLI where users provide their own API keys to use language models for software
development tasks. The project is currently in a greenfield state with only foundational files:

- Comprehensive behavioral contract (.claude/CLAUDE.md) defining agent philosophy
- Core principles: Transparency > cleverness, Safety > speed, Clarity > completeness
- Interaction model: Understand → Plan → Act → Verify
- Safety-first design: Explicit user control, no silent modifications, real consequences

The goal is to build a production-ready TUI application that enables developers to collaborate with AI agents on real codebases,
with strong safety guarantees and transparent operation.

---

Technology Stack

Language: Go (chosen for fast compilation, single-binary distribution, excellent TUI ecosystem, strong concurrency,
cross-platform support)

Core Dependencies:

- TUI: Bubbletea (Elm architecture), Lipgloss (styling), Bubbles (components)
- LLM: Multi-provider support (Anthropic Claude, OpenAI, Gemini, local models)
- Development: tree-sitter-go (parsing), go-git (git ops), viper (config), cobra (CLI)
- Security: golang.org/x/crypto, OS-native keychain integration

Rationale: Go provides the optimal balance of developer velocity, ease of deployment, and community accessibility. The Charm
ecosystem (Bubbletea/Lipgloss) is mature and well-documented for TUI development.

---

Architecture

┌─────────────────────────────────────────┐
│ Anvil CLI Binary │
├─────────────────────────────────────────┤
│ TUI Layer ←→ Agent Engine │
│ (Bubbletea) (Lifecycle) │
│ ↕ ↕ │
│ LLM Client ←→ Tool System │
│ (Multi-provider) (FS, Git, Shell) │
└─────────────────────────────────────────┘

Core Components:

1.  TUI Layer (/internal/tui): Multi-panel interface with conversation view, diff viewer, plan panel, file browser
2.  Agent Engine (/internal/agent): Orchestrates Understand→Plan→Act→Verify lifecycle with safety checks
3.  Tool System (/internal/tools): File operations, git commands, shell proposals with approval gates
4.  LLM Client (/internal/llm): Provider-agnostic streaming integration with token management

---

Implementation Phases

Phase 0: Foundation (2 weeks)

Goal: Establish project infrastructure

Tasks:

- Initialize Go module with directory structure
- Set up CI/CD (GitHub Actions for testing, linting, releases)
- Implement configuration system (file-based + OS keychain for secrets)
- Create basic TUI scaffold with Bubbletea
- Set up file-based structured logging
- Write development documentation

Key Files:

- go.mod, go.sum
- .github/workflows/ci.yml, .github/workflows/release.yml
- internal/config/manager.go, internal/config/keys.go
- internal/tui/app.go
- docs/developer.md

Success Criteria: Runnable binary with basic TUI, configuration loading works, tests run in CI

---

Phase 1: TUI Core (2 weeks)

Goal: Build functional multi-panel terminal interface

Tasks:

- Implement panel system (conversation, files, diff, plan)
- Create input handling with keyboard shortcuts
- Build conversation view with markdown rendering
- Implement streaming text display
- Add status bar showing mode/model/tokens
- Create theme system with Lipgloss
- Add mouse support for panel resizing

Key Files:

- internal/tui/app.go (main Bubbletea model)
- internal/tui/panels/conversation.go
- internal/tui/panels/diff.go
- internal/tui/panels/plan.go
- internal/tui/panels/files.go
- internal/tui/styles.go

Success Criteria: Full TUI with panel navigation, static content rendering, responsive layout across terminal sizes

---

Phase 2: LLM Integration (2 weeks)

Goal: Connect to LLM providers with streaming support

Tasks:

- Implement LLM client abstraction layer
- Add Anthropic Claude API integration (SSE streaming)
- Add OpenAI API integration
- Implement token counting and budget tracking
- Build retry logic with exponential backoff
- Test streaming responses in TUI conversation panel

Key Files:

- internal/llm/client.go (provider interface)
- internal/llm/anthropic.go
- internal/llm/openai.go
- internal/llm/streaming.go (SSE parser)
- internal/llm/token.go

Success Criteria: Working LLM integration with streaming responses, token usage displayed, graceful error handling

---

Phase 3: Tool System (3 weeks)

Goal: Implement core tools for file and git operations

Tasks:

- Design tool schema and registry system
- Implement file reading with permission checks
- Implement file writing with diff preview
- Add file search (glob patterns, content search)
- Implement git operations (status, diff, log, commit)
- Create shell command approval workflow
- Build safety checks for destructive operations
- Render tool results in TUI panels

Key Files:

- internal/tools/registry.go
- internal/tools/filesystem.go
- internal/tools/shell.go
- internal/tools/git.go
- internal/agent/safety.go
- pkg/schema/tool.go

Success Criteria: Full file system operations working, git-aware functionality, shell commands require approval, safety policies
enforced

Risks: File permission edge cases, git repository states, false positives in destructive command detection

---

Phase 4: Agent Engine (3 weeks)

Goal: Implement agentic workflow and lifecycle management

Tasks:

- Build agent loop (request → tool calls → response)
- Implement Understand → Plan → Act → Verify phases from CLAUDE.md
- Create tool execution orchestration
- Add approval gates for file modifications
- Implement plan generation and step tracking
- Build context window management with pruning
- Add conversation history persistence
- Implement teaching mode for explanations

Key Files:

- internal/agent/engine.go (core agent loop)
- internal/agent/lifecycle.go (phase management)
- internal/agent/context.go (conversation context)
- internal/agent/stream.go (streaming handler)

Success Criteria: Full agent lifecycle working, plan panel tracks progress, approval workflow functional, conversations saved and
resumable

Risks: Context window overflow, tool call validation, complex approval flows

---

Phase 5: Advanced Features (3 weeks)

Goal: Add polish, code analysis, and production readiness

Tasks:

- Integrate tree-sitter for syntax highlighting
- Implement code symbol navigation
- Enhance teaching mode with detailed explanations
- Create syntax-highlighted diff viewer
- Implement multi-file change coordination
- Add session management (save/restore)
- Build usage analytics (token tracking)
- Polish UX (animations, loading states, error messages)

Key Files:

- internal/tools/analysis.go (tree-sitter integration)
- internal/tui/components/markdown.go
- internal/util/diff.go

Success Criteria: Syntax-highlighted diffs, teaching mode functional, session persistence working, production-quality UX

---

Phase 6: Testing & Documentation (2 weeks)

Goal: Comprehensive testing and user documentation

Tasks:

- Write unit tests (target 80% coverage)
- Create integration tests (E2E workflows)
- Build user manual and quickstart guide
- Write architectural documentation
- Create example workflows and tutorials
- Beta testing with real users
- Bug fixes based on feedback

Key Files:

- \*\_test.go throughout codebase
- docs/user-guide.md
- docs/quickstart.md
- docs/architecture.md

Success Criteria: Test coverage >80%, complete documentation, beta feedback incorporated

---

Phase 7: Release (Ongoing)

Goal: Public release and ecosystem growth

Tasks:

- Security audit (especially API key handling)
- Release v1.0.0 with binaries for macOS, Linux, Windows
- Publish to package managers (homebrew, apt, scoop)
- Create plugin system documentation
- Monitor issues and iterate
- Build community resources

Success Criteria: Public v1.0.0 available, multi-platform binaries, active community engagement

---

Critical Design Decisions

1.  Go vs Rust

Decision: Go
Rationale: Developer velocity and community accessibility outweigh raw performance benefits for a CLI tool

2.  Tool Execution Model

Decision: Propose-then-execute with approval gates
Rationale: Aligns with CLAUDE.md "Safety > speed" philosophy; prevents irreversible mistakes

3.  Multi-Provider LLM Support

Decision: Abstract provider interface from day one
Rationale: "Model-agnostic" requirement; avoids vendor lock-in; user freedom

4.  Configuration & Secrets

Decision: File-based config + OS keychain for API keys
Rationale: Maximum security; "Never log, echo, or store API keys" from CLAUDE.md

5.  Context Management

Decision: Intelligent pruning with user visibility
Rationale: Transparency over magic; users see what's included in context

6.  Streaming by Default

Decision: All LLM responses stream to UI
Rationale: Real-time feedback critical for terminal tool UX

---

Risk Mitigation
┌─────────────────────────┬──────────┬──────────────────────────────────────────────────────────────────┐
│ Risk │ Impact │ Mitigation │
├─────────────────────────┼──────────┼──────────────────────────────────────────────────────────────────┤
│ Context window overflow │ High │ Smart pruning, user controls, summarization │
├─────────────────────────┼──────────┼──────────────────────────────────────────────────────────────────┤
│ Shell command safety │ Critical │ Whitelist approach, explicit confirmation, dry-run mode │
├─────────────────────────┼──────────┼──────────────────────────────────────────────────────────────────┤
│ API rate limiting │ High │ Client-side throttling, queue management, clear user feedback │
├─────────────────────────┼──────────┼──────────────────────────────────────────────────────────────────┤
│ Cross-platform issues │ Medium │ Abstract file operations, extensive testing on all platforms │
├─────────────────────────┼──────────┼──────────────────────────────────────────────────────────────────┤
│ API key leakage │ Critical │ OS keychain, never log keys, pre-commit hooks for .env detection │
├─────────────────────────┼──────────┼──────────────────────────────────────────────────────────────────┤
│ TUI performance │ Medium │ Lazy loading, pagination, virtual scrolling │
└─────────────────────────┴──────────┴──────────────────────────────────────────────────────────────────┘

---

Directory Structure

anvil/
├── cmd/anvil/main.go # Entry point
├── internal/
│ ├── agent/ # Agent engine & lifecycle
│ │ ├── engine.go # Core loop
│ │ ├── lifecycle.go # Understand→Plan→Act→Verify
│ │ ├── context.go # Conversation context
│ │ ├── safety.go # Safety policies
│ │ └── stream.go # Streaming handler
│ ├── tui/ # Terminal UI
│ │ ├── app.go # Main Bubbletea model
│ │ ├── styles.go # Lipgloss themes
│ │ └── panels/ # UI panels
│ │ ├── conversation.go
│ │ ├── diff.go
│ │ ├── plan.go
│ │ └── files.go
│ ├── llm/ # LLM providers
│ │ ├── client.go # Provider interface
│ │ ├── anthropic.go
│ │ ├── openai.go
│ │ ├── streaming.go # SSE parser
│ │ └── token.go # Token counting
│ ├── tools/ # Tool system
│ │ ├── registry.go
│ │ ├── filesystem.go
│ │ ├── shell.go
│ │ ├── git.go
│ │ └── analysis.go
│ ├── config/ # Configuration
│ │ ├── manager.go
│ │ ├── keys.go # Keychain integration
│ │ └── defaults.go
│ └── util/ # Utilities
│ ├── diff.go
│ └── git.go
├── pkg/schema/ # Public schemas
│ ├── tool.go
│ └── message.go
├── docs/ # Documentation
│ ├── user-guide.md
│ ├── developer.md
│ └── architecture.md
├── .claude/
│ ├── CLAUDE.md # Agent behavioral contract
│ └── settings.local.json
├── .github/workflows/
│ ├── ci.yml
│ └── release.yml
├── go.mod
├── go.sum
├── README.md
└── LICENSE

---

Verification Plan

Phase-by-Phase Testing

Phase 0-1: Run anvil, verify TUI renders, navigate between empty panels with keyboard shortcuts

Phase 2: Configure API key, send test message, verify streaming response appears in conversation panel

Phase 3: Issue file read command, verify content appears; propose file write, verify diff preview and approval gate

Phase 4: Request multi-step task, verify plan appears in plan panel, steps execute in order, approval gates work

Phase 5: Request code change, verify syntax-highlighted diff, ask for explanation, verify teaching mode

Phase 6: Run full test suite (go test ./...), verify coverage >80%, test on macOS/Linux/Windows

Phase 7: Install from package manager, run in fresh environment, verify API key setup flow

End-to-End Workflow Test

1.  User launches anvil in a git repository
2.  User configures API key (stored in OS keychain)
3.  User requests: "Add error handling to server.go"
4.  Agent understands → reads server.go → proposes plan in panel
5.  User approves plan
6.  Agent proposes file modifications with diff preview
7.  User approves changes
8.  Agent verifies changes applied, suggests testing steps
9.  User requests explanation of approach (teaching mode)
10. Agent explains with references to code paths

Success: All steps complete without errors, files modified correctly, git status shows changes, no API keys leaked to logs

---

Critical Files for Implementation

- /Users/siddharth/Workspace/anvil/internal/tui/app.go - Main Bubbletea model (TUI state machine)
- /Users/siddharth/Workspace/anvil/internal/agent/engine.go - Agent orchestration (Understand→Plan→Act→Verify)
- /Users/siddharth/Workspace/anvil/internal/llm/client.go - LLM provider abstraction
- /Users/siddharth/Workspace/anvil/internal/tools/registry.go - Tool system with safety policies
- /Users/siddharth/Workspace/anvil/.claude/CLAUDE.md - Behavioral contract (passed as system prompt)

---

Open Questions

1.  Platform Priority: Build for macOS/Linux first, or Windows support in parallel?
2.  Default Model: Which model should be the recommended default for new users?
3.  Telemetry: Include opt-in crash reporting and usage analytics?
4.  Community: Set up Discord/Slack for support from the start?
5.  Plugin System: Include in v1.0 or defer to v2.0?

---

Philosophy Alignment

This plan directly implements the CLAUDE.md principles:

✓ Transparency > cleverness: All agent actions visible in plan panel, diffs shown before applying
✓ Safety > speed: Approval gates for modifications, destructive commands flagged, no silent changes
✓ Clarity > completeness: Phased approach, clear documentation, teaching mode for explanations
✓ Local-first: No cloud dependencies, user-provided API keys, works offline except LLM calls
✓ Model-agnostic: Multi-provider from day one, no vendor lock-in
✓ Real consequences: Explicit user control, no autonomous execution

---

Next Steps

After approval:

1.  Initialize Go module and directory structure
2.  Set up CI/CD pipeline
3.  Begin Phase 0 foundation work
4.  Weekly progress updates with working demos
