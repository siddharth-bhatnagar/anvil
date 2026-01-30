# Anvil Quickstart Guide

Get up and running with Anvil in minutes.

## Prerequisites

- Go 1.21 or later
- An API key from one of: Anthropic (Claude), OpenAI, or Google (Gemini)
- macOS, Linux, or Windows

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/siddharth-bhatnagar/anvil.git
cd anvil

# Build the binary
go build -o anvil ./cmd/anvil

# Optional: Move to your PATH
sudo mv anvil /usr/local/bin/
```

### Verify Installation

```bash
anvil --version
```

## Configuration

### Setting Your API Key

Anvil stores API keys securely in your OS keychain. On first run, you'll be prompted to enter your API key.

Alternatively, set it via environment variable:

```bash
# For Anthropic Claude
export ANVIL_ANTHROPIC_API_KEY="your-api-key"

# For OpenAI
export ANVIL_OPENAI_API_KEY="your-api-key"

# For Google Gemini
export ANVIL_GEMINI_API_KEY="your-api-key"
```

### Configuration File

Anvil creates a config file at `~/.anvil/config.yaml`:

```yaml
model: claude-sonnet-4
provider: anthropic
temperature: 0.7
max_tokens: 4096
log_level: info
```

## Your First Session

1. Navigate to your project directory:
   ```bash
   cd /path/to/your/project
   ```

2. Start Anvil:
   ```bash
   anvil
   ```

3. You'll see the TUI with multiple panels:
   - **Conversation**: Main chat interface
   - **Files**: Project file browser
   - **Diff**: View proposed changes
   - **Plan**: See the agent's plan

## Basic Commands

### Navigation

| Key | Action |
|-----|--------|
| `Tab` | Switch between panels |
| `j/k` or arrows | Navigate within panels |
| `Enter` | Select/confirm |
| `Esc` | Cancel/go back |
| `q` | Quit Anvil |
| `?` | Show help |

### Interaction

1. **Ask a question**: Just type your question and press Enter
2. **Request changes**: Describe what you want changed
3. **Review diffs**: Changes appear in the Diff panel
4. **Approve changes**: Press `y` to approve, `n` to reject

## Example Workflow

### Adding a Feature

```
You: Add input validation to the login form

Anvil: I'll analyze your codebase and propose changes.

Plan:
1. Read current login form implementation
2. Add email format validation
3. Add password strength check
4. Update error messages
5. Add tests

[Anvil shows diff in Diff panel]

Do you want to apply these changes? (y/n)
```

### Fixing a Bug

```
You: The API returns 500 when user is null

Anvil: I'll investigate the issue.

[Reads relevant files]

I found the problem in src/api/users.go:42.
The code doesn't check for null user before accessing properties.

Proposed fix:
[Shows diff]
```

### Understanding Code

```
You: Explain how the authentication middleware works

Anvil: Let me analyze the authentication flow...

[Provides detailed explanation with code references]
```

## Tips

1. **Be specific**: The more context you provide, the better the results
2. **Review diffs carefully**: Always review proposed changes before accepting
3. **Use the plan panel**: Watch the agent's progress and reasoning
4. **Iterate**: Don't hesitate to ask for modifications to proposals

## Troubleshooting

### API Key Issues

```
Error: API key not found
```

Solution: Set your API key via environment variable or the config file.

### Connection Issues

```
Error: Failed to connect to API
```

Solution: Check your internet connection and verify your API key is valid.

### Permission Issues

```
Error: Permission denied
```

Solution: Ensure you have read/write access to your project directory.

## Next Steps

- Read the [User Guide](user-guide.md) for comprehensive documentation
- Check [Architecture](architecture.md) for technical details
- Visit the [GitHub repository](https://github.com/siddharth-bhatnagar/anvil) for updates

## Getting Help

- Use `?` within Anvil to see available commands
- Check the [FAQ](https://github.com/siddharth-bhatnagar/anvil/wiki/FAQ)
- Open an issue on [GitHub](https://github.com/siddharth-bhatnagar/anvil/issues)
