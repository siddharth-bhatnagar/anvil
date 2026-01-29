# Developer Guide

This document provides guidance for developers contributing to Anvil.

## Prerequisites

- Go 1.21 or later
- Git
- A terminal with Unicode and color support

## Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/siddharth-bhatnagar/anvil.git
   cd anvil
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build the project:
   ```bash
   go build -o anvil ./cmd/anvil
   ```

4. Run the binary:
   ```bash
   ./anvil
   ```

## Project Structure

```
anvil/
├── cmd/anvil/           # Entry point
├── internal/            # Internal packages
│   ├── agent/          # Agent engine and lifecycle
│   ├── tui/            # Terminal UI (Bubbletea)
│   ├── llm/            # LLM provider integrations
│   ├── tools/          # Tool system (file ops, git, shell)
│   ├── config/         # Configuration management
│   └── util/           # Utilities (logging, diff, etc.)
├── pkg/                # Public packages
│   └── schema/         # Data schemas
├── docs/               # Documentation
├── .claude/            # Agent behavioral contract
└── .github/            # CI/CD workflows
```

## Architecture

Anvil follows a clean architecture with four main components:

1. **TUI Layer** (`internal/tui`): Handles all UI rendering and user input using Bubbletea
2. **Agent Engine** (`internal/agent`): Orchestrates the Understand→Plan→Act→Verify lifecycle
3. **LLM Client** (`internal/llm`): Provider-agnostic interface for LLM APIs
4. **Tool System** (`internal/tools`): Implements tools the agent can invoke (file ops, git, shell)

## Development Workflow

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run with race detector
go test -race ./...
```

### Linting

We use `golangci-lint` for code quality:

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run

# Auto-fix issues
golangci-lint run --fix
```

### Building

```bash
# Standard build
go build -o anvil ./cmd/anvil

# Build with version info
go build -ldflags="-X main.version=v0.1.0" -o anvil ./cmd/anvil

# Cross-compile for different platforms
GOOS=linux GOARCH=amd64 go build -o anvil-linux ./cmd/anvil
GOOS=darwin GOARCH=arm64 go build -o anvil-macos ./cmd/anvil
GOOS=windows GOARCH=amd64 go build -o anvil.exe ./cmd/anvil
```

## Code Style

- Follow Go idioms and best practices
- Use `gofmt` for formatting (enforced by CI)
- Add comments for exported types and functions
- Keep functions small and focused
- Avoid premature abstractions

## Safety & Security

Anvil handles sensitive data (API keys, user code). Follow these guidelines:

1. **Never log API keys** - Use the logger's redaction features
2. **Store secrets in OS keychain** - Never in config files or environment variables
3. **Validate all user input** - Especially file paths and shell commands
4. **Check file permissions** - Respect user's file system boundaries
5. **Audit destructive operations** - Flag and require explicit approval

## Testing

- Write unit tests for all business logic
- Use table-driven tests for multiple cases
- Mock external dependencies (LLM APIs, file system)
- Test error paths, not just happy paths
- Aim for >80% code coverage

## Debugging

### Logs

Anvil logs to `~/.anvil/logs/anvil-YYYY-MM-DD.log`:

```bash
# Watch logs in real-time
tail -f ~/.anvil/logs/anvil-$(date +%Y-%m-%d).log

# Search logs
grep "ERROR" ~/.anvil/logs/anvil-*.log
```

### TUI Debugging

Since stdout is used by the TUI, use file-based logging:

```go
import "github.com/siddharth-bhatnagar/anvil/internal/util"

util.Logger.Debug().
    Str("key", "value").
    Msg("Debug message")
```

## Adding New Features

1. Create a feature branch: `git checkout -b feature/my-feature`
2. Implement the feature following the architecture
3. Add tests (unit + integration where applicable)
4. Update documentation
5. Run tests and linting: `go test ./... && golangci-lint run`
6. Commit with clear messages following [Conventional Commits](https://www.conventionalcommits.org/)
7. Push and create a pull request

## Bubbletea Development

Anvil uses [Bubbletea](https://github.com/charmbracelet/bubbletea) for the TUI.

### Model-Update-View Pattern

```go
type Model struct {
    // state
}

func (m Model) Init() tea.Cmd {
    // initialization
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // handle messages, update state
}

func (m Model) View() string {
    // render UI
}
```

### Useful Resources

- [Bubbletea Tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
- [Bubbles Components](https://github.com/charmbracelet/bubbles)
- [Lipgloss Styling](https://github.com/charmbracelet/lipgloss)

## Release Process

Releases are automated via GitHub Actions and GoReleaser:

1. Update version in relevant files
2. Create a git tag: `git tag -a v0.1.0 -m "Release v0.1.0"`
3. Push the tag: `git push origin v0.1.0`
4. GitHub Actions will build and release automatically

## Getting Help

- Open an issue on GitHub for bugs or feature requests
- Check existing issues for known problems
- Review the [implementation plan](.claude/plans/) for roadmap

## Resources

- [Go Documentation](https://go.dev/doc/)
- [Bubbletea Examples](https://github.com/charmbracelet/bubbletea/tree/master/examples)
- [Anvil Behavioral Contract](.claude/CLAUDE.md)
- [Implementation Plan](.claude/plans/)
