# ⚒ Anvil

A terminal-based, agentic coding CLI where you bring your own API keys and use language models to reason about, plan, and write software.

## Features

- **Local-First**: All data stays on your machine, you control your API keys
- **Model-Agnostic**: Support for Anthropic Claude, OpenAI, Google Gemini, and local models
- **Safety-First**: Explicit approval gates for file modifications and destructive commands
- **Transparent**: See the agent's reasoning and plan before execution
- **TUI Interface**: Beautiful terminal UI with multiple panels for conversation, diffs, and plans
- **Git-Aware**: Intelligent integration with your git workflow

## Status

🚧 **Under Active Development** - Currently implementing Phase 0 (Foundation)

Anvil is in early development. Core functionality is being built. See the [implementation plan](.claude/plans/) for details.

## Installation

### Prerequisites

- Go 1.21 or later
- macOS, Linux, or Windows

### From Source

```bash
git clone https://github.com/siddharth-bhatnagar/anvil.git
cd anvil
go build -o anvil ./cmd/anvil
./anvil
```

## Quick Start

1. Run Anvil:
   ```bash
   ./anvil
   ```

2. Configure your API key (stored securely in OS keychain):
   ```bash
   # This will be handled in the TUI in future versions
   ```

3. Start collaborating with AI agents on your code

## Philosophy

Anvil is built on three core principles:

- **Transparency > cleverness** - You see what's happening, always
- **Safety > speed** - No silent or irreversible changes
- **Clarity > completeness** - Clear, focused solutions over over-engineering

See [.claude/CLAUDE.md](.claude/CLAUDE.md) for the complete behavioral contract.

## Development

See [docs/developer.md](docs/developer.md) for development setup and guidelines.

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

Inspired by tools like Cursor, Claude Code, and the broader agentic coding movement.