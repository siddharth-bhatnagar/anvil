# Anvil

[![CI](https://github.com/siddharth-bhatnagar/anvil/actions/workflows/ci.yml/badge.svg)](https://github.com/siddharth-bhatnagar/anvil/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/siddharth-bhatnagar/anvil)](https://goreportcard.com/report/github.com/siddharth-bhatnagar/anvil)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/siddharth-bhatnagar/anvil)](https://github.com/siddharth-bhatnagar/anvil/releases)

A terminal-based, agentic coding CLI where you bring your own API keys and use language models to reason about, plan, and write software.

![Anvil Demo](docs/assets/demo.gif)

## Features

- **Local-First**: All data stays on your machine, you control your API keys
- **Model-Agnostic**: Support for Anthropic Claude, OpenAI, Google Gemini, and local models
- **Safety-First**: Explicit approval gates for file modifications and destructive commands
- **Transparent**: See the agent's reasoning and plan before execution
- **TUI Interface**: Beautiful terminal UI with multiple panels for conversation, diffs, and plans
- **Git-Aware**: Intelligent integration with your git workflow
- **Teaching Mode**: Get detailed explanations while learning

## Installation

### Homebrew (macOS/Linux)

```bash
brew install siddharth-bhatnagar/tap/anvil
```

### Download Binary

Download the latest release for your platform from the [releases page](https://github.com/siddharth-bhatnagar/anvil/releases).

### From Source

```bash
git clone https://github.com/siddharth-bhatnagar/anvil.git
cd anvil
go build -o anvil ./cmd/anvil
sudo mv anvil /usr/local/bin/  # Optional: add to PATH
```

## Quick Start

1. **Set your API key**:
   ```bash
   # Anthropic Claude (recommended)
   export ANVIL_ANTHROPIC_API_KEY="your-api-key"

   # Or OpenAI
   export ANVIL_OPENAI_API_KEY="your-api-key"
   ```

2. **Navigate to your project**:
   ```bash
   cd /path/to/your/project
   ```

3. **Start Anvil**:
   ```bash
   anvil
   ```

4. **Start coding**:
   ```
   You: Add input validation to the login form

   Anvil: I'll analyze your codebase and propose changes...
   ```

See the [Quickstart Guide](docs/quickstart.md) for more details.

## Usage

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Tab` | Switch panels |
| `Enter` | Send message / Confirm |
| `j/k` | Navigate up/down |
| `y/n` | Approve/reject changes |
| `?` | Show help |
| `q` | Quit |

### Example Workflows

**Adding a Feature:**
```
You: Add rate limiting to the API endpoints
Anvil: [Shows plan, then diff for review]
```

**Fixing a Bug:**
```
You: Users are getting 500 errors on /api/profile
Anvil: [Investigates, identifies issue, proposes fix]
```

**Understanding Code:**
```
You: Explain how the authentication middleware works
Anvil: [Provides detailed explanation with code references]
```

## Configuration

Configuration is stored in `~/.anvil/config.yaml`:

```yaml
model: claude-sonnet-4
provider: anthropic
temperature: 0.7
max_tokens: 4096
```

API keys are stored securely in your OS keychain.

## Philosophy

Anvil is built on three core principles:

- **Transparency > Cleverness**: You see what's happening, always
- **Safety > Speed**: No silent or irreversible changes
- **Clarity > Completeness**: Clear, focused solutions over over-engineering

## Documentation

- [Quickstart Guide](docs/quickstart.md) - Get started in minutes
- [User Guide](docs/user-guide.md) - Comprehensive documentation
- [Architecture](docs/architecture.md) - Technical deep-dive
- [Contributing](CONTRIBUTING.md) - How to contribute

## Supported Providers

| Provider | Models | Status |
|----------|--------|--------|
| Anthropic | Claude Opus, Sonnet, Haiku | Fully Supported |
| OpenAI | GPT-4, GPT-3.5 | Fully Supported |
| Google | Gemini Pro, Ultra | Supported |
| Local | Ollama, llama.cpp | Planned |

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

```bash
# Development setup
git clone https://github.com/siddharth-bhatnagar/anvil.git
cd anvil
go mod download
go test ./...
```

## Security

API keys are stored in your OS keychain and never logged. See [SECURITY.md](SECURITY.md) for our security policy.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

Inspired by tools like Cursor, Claude Code, and the broader agentic coding movement.

---

Built with Go and the [Charm](https://charm.sh) ecosystem.
