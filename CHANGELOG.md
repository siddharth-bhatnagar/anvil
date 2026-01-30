# Changelog

All notable changes to Anvil will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2026-01-30

### Added

#### Core Features
- Multi-panel terminal UI with Bubbletea
  - Conversation panel for chat interface
  - Files panel with git status integration
  - Diff panel with syntax highlighting
  - Plan panel for tracking agent progress
- Agent engine with Understand-Plan-Act-Verify lifecycle
- Multi-provider LLM support (Anthropic Claude, OpenAI, Google Gemini)
- Streaming responses with real-time display

#### Tool System
- File operations (read, write, search)
- Git integration (status, diff, log)
- Shell command execution with approval gates
- Code analysis and symbol navigation

#### Safety Features
- Explicit approval required for all file modifications
- Destructive command detection and warnings
- API keys stored securely in OS keychain
- Change rollback support

#### User Experience
- Syntax highlighting for code display
- Markdown rendering in terminal
- Session save/restore functionality
- Teaching mode for learning
- Token usage tracking with cost estimation
- Configurable themes and styles

#### Documentation
- Quickstart guide
- Comprehensive user guide
- Architecture documentation
- Developer documentation

### Security
- API keys never logged or stored in plaintext
- OS keychain integration for credential storage
- File operations sandboxed to project directory
- Audit logging for all operations

## [0.1.0] - 2026-01-15

### Added
- Initial project structure
- Basic TUI scaffold
- Configuration system
- Logging infrastructure

---

## Version History

| Version | Date | Highlights |
|---------|------|------------|
| 1.0.0 | 2026-01-30 | First stable release |
| 0.1.0 | 2026-01-15 | Initial development |

[Unreleased]: https://github.com/siddharth-bhatnagar/anvil/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/siddharth-bhatnagar/anvil/compare/v0.1.0...v1.0.0
[0.1.0]: https://github.com/siddharth-bhatnagar/anvil/releases/tag/v0.1.0
