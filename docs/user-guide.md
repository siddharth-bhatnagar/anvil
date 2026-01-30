# Anvil User Guide

Complete documentation for using Anvil, the terminal-based agentic coding CLI.

## Table of Contents

1. [Overview](#overview)
2. [Installation](#installation)
3. [Configuration](#configuration)
4. [User Interface](#user-interface)
5. [Keyboard Shortcuts](#keyboard-shortcuts)
6. [Working with the Agent](#working-with-the-agent)
7. [File Operations](#file-operations)
8. [Git Integration](#git-integration)
9. [Session Management](#session-management)
10. [Teaching Mode](#teaching-mode)
11. [Advanced Features](#advanced-features)
12. [Troubleshooting](#troubleshooting)

---

## Overview

Anvil is a local-first, API-key driven coding assistant that runs in your terminal. It provides:

- **Multi-panel TUI**: Conversation, files, diffs, and plans all visible at once
- **Safety-first design**: All changes require explicit approval
- **Transparent operation**: See the agent's reasoning and plan before execution
- **Model-agnostic**: Support for Claude, GPT, Gemini, and local models
- **Git-aware**: Intelligent integration with your git workflow

### Core Principles

1. **Transparency > Cleverness**: You see what's happening, always
2. **Safety > Speed**: No silent or irreversible changes
3. **Clarity > Completeness**: Clear, focused solutions over over-engineering

---

## Installation

### Requirements

- Go 1.21+
- macOS, Linux, or Windows
- Terminal with 256-color support (recommended)

### Building from Source

```bash
git clone https://github.com/siddharth-bhatnagar/anvil.git
cd anvil
go build -o anvil ./cmd/anvil
```

### Verifying Installation

```bash
./anvil --version
./anvil --help
```

---

## Configuration

### Configuration File

Location: `~/.anvil/config.yaml`

```yaml
# LLM Settings
model: claude-sonnet-4        # Model to use
provider: anthropic           # Provider: anthropic, openai, gemini
temperature: 0.7              # Response creativity (0.0-1.0)
max_tokens: 4096              # Maximum response length

# Logging
log_level: info               # debug, info, warn, error
log_dir: ~/.anvil/logs        # Log file location
```

### API Keys

API keys are stored securely in your OS keychain:

- **macOS**: Keychain Access
- **Linux**: libsecret/GNOME Keyring
- **Windows**: Credential Manager

Set via environment variables for automation:

```bash
export ANVIL_ANTHROPIC_API_KEY="sk-ant-..."
export ANVIL_OPENAI_API_KEY="sk-..."
export ANVIL_GEMINI_API_KEY="..."
```

### Supported Providers

| Provider | Models | Environment Variable |
|----------|--------|---------------------|
| Anthropic | claude-sonnet-4, claude-opus-4, claude-haiku-4 | `ANVIL_ANTHROPIC_API_KEY` |
| OpenAI | gpt-4-turbo, gpt-4, gpt-3.5-turbo | `ANVIL_OPENAI_API_KEY` |
| Google | gemini-pro, gemini-ultra | `ANVIL_GEMINI_API_KEY` |

---

## User Interface

Anvil provides a multi-panel terminal interface:

```
┌─────────────────────────────────────────────────────────────────┐
│                        Conversation                              │
│                                                                  │
│  You: Add error handling to the API                             │
│                                                                  │
│  Assistant: I'll analyze your code and add proper error         │
│  handling. Let me first read the current implementation...       │
│                                                                  │
├─────────────────────────────┬───────────────────────────────────┤
│          Files              │              Diff                  │
│                             │                                    │
│  📁 src/                    │  --- a/src/api.go                 │
│    📄 api.go [M]            │  +++ b/src/api.go                 │
│    📄 handlers.go           │  @@ -10,6 +10,12 @@               │
│    📄 models.go             │  +  if err != nil {               │
│  📁 tests/                  │  +    return fmt.Errorf(...)       │
│                             │  +  }                              │
├─────────────────────────────┴───────────────────────────────────┤
│                           Plan                                   │
│  ● 1. Read current API implementation                           │
│  ◐ 2. Add error handling to endpoints                           │
│  ○ 3. Update tests                                              │
├─────────────────────────────────────────────────────────────────┤
│ Mode: Normal │ Panel: Conversation │ Tokens: 1.2K │ ? Help  q Quit │
└─────────────────────────────────────────────────────────────────┘
```

### Panels

| Panel | Description |
|-------|-------------|
| **Conversation** | Main chat interface with the agent |
| **Files** | Project file browser with git status |
| **Diff** | Preview of proposed changes |
| **Plan** | Current plan and progress tracking |

### Status Bar

The status bar shows:
- Current mode (Normal, Insert, etc.)
- Active panel
- Token usage
- Available shortcuts

---

## Keyboard Shortcuts

### Global

| Key | Action |
|-----|--------|
| `Tab` | Cycle through panels |
| `Shift+Tab` | Cycle panels (reverse) |
| `?` | Show help |
| `q` | Quit Anvil |
| `Ctrl+C` | Cancel current operation |

### Conversation Panel

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `Ctrl+Enter` | New line in message |
| `↑/↓` | Scroll history |
| `PgUp/PgDn` | Page scroll |

### Files Panel

| Key | Action |
|-----|--------|
| `j/↓` | Move down |
| `k/↑` | Move up |
| `Enter` | Open file / Expand directory |
| `g` | Go to top |
| `G` | Go to bottom |
| `r` | Refresh file list |

### Diff Panel

| Key | Action |
|-----|--------|
| `j/↓` | Scroll down |
| `k/↑` | Scroll up |
| `n/Tab` | Next file (multi-file diff) |
| `p/Shift+Tab` | Previous file |
| `s` | Toggle statistics |
| `y` | Approve changes |
| `N` | Reject changes |

### Plan Panel

| Key | Action |
|-----|--------|
| `j/↓` | Scroll down |
| `k/↑` | Scroll up |

---

## Working with the Agent

### Interaction Model

Anvil follows a structured interaction model:

1. **Understand**: Agent restates your goal and asks clarifying questions
2. **Plan**: Agent proposes a plan before making changes
3. **Act**: Agent executes the plan with your approval
4. **Verify**: Agent explains what changed and suggests testing

### Example Interactions

**Feature Request:**
```
You: Add rate limiting to the API endpoints

Agent: I'll add rate limiting to your API. A few questions:
- What rate limit do you want? (e.g., 100 requests/minute)
- Should it be per-user or global?
- Where should rate limit errors be logged?
```

**Bug Fix:**
```
You: Users are getting 500 errors on the /profile endpoint

Agent: I'll investigate. Let me read the profile handler...
[Reads files]

I found the issue: the code doesn't handle the case when the
user's profile data is null. Here's my proposed fix:
[Shows diff]
```

**Code Explanation:**
```
You: Explain how the authentication middleware works

Agent: The authentication middleware in src/middleware/auth.go
works as follows:

1. Extracts JWT token from Authorization header
2. Validates token signature using the secret key
3. Decodes claims and attaches user info to context
4. Passes request to next handler

[Provides detailed breakdown with code references]
```

### Best Practices

1. **Provide context**: Include relevant file names and error messages
2. **Be specific**: "Add validation to email field" is better than "fix form"
3. **Review diffs carefully**: Always read proposed changes before approving
4. **Iterate**: Ask for modifications if the first proposal isn't right

---

## File Operations

### Reading Files

The agent can read any file in your project. It will:
- Respect `.gitignore` patterns
- Skip binary files
- Handle large files with pagination

### Writing Files

All file modifications require explicit approval:

1. Agent proposes changes with a diff preview
2. You review the changes in the Diff panel
3. Press `y` to approve or `n` to reject

### Supported Operations

| Operation | Approval Required |
|-----------|------------------|
| Read file | No |
| Create file | Yes |
| Modify file | Yes |
| Delete file | Yes (with confirmation) |
| Rename file | Yes |

---

## Git Integration

### Git Status

The Files panel shows git status indicators:

| Indicator | Meaning |
|-----------|---------|
| `[M]` | Modified |
| `[A]` | Added (staged) |
| `[D]` | Deleted |
| `[?]` | Untracked |
| `[C]` | Conflicted |

### Git Operations

The agent can help with:
- Viewing diffs (`git diff`)
- Checking status (`git status`)
- Creating commits (proposed, requires approval)
- Viewing history (`git log`)

### Commit Workflow

```
You: Create a commit for the changes we just made

Agent: I'll prepare a commit. Here's my proposed message:

  feat: add rate limiting to API endpoints

  - Added redis-based rate limiter
  - Configured 100 req/min per user
  - Added rate limit headers to responses

Shall I create this commit? (y/n)
```

---

## Session Management

### Saving Sessions

Sessions are automatically saved to `~/.anvil/sessions/`.

Each session includes:
- Conversation history
- Model and provider used
- Token usage statistics
- Working directory

### Resuming Sessions

```bash
# List recent sessions
anvil sessions list

# Resume a specific session
anvil sessions resume <session-id>
```

### Session Files

Sessions are stored as JSON files:
```
~/.anvil/sessions/
├── 20240115-143022-123456.json
├── 20240115-150055-789012.json
└── ...
```

---

## Teaching Mode

Teaching mode provides detailed explanations for learning.

### Enabling Teaching Mode

In the conversation:
```
You: /teach on
```

Or via config:
```yaml
teaching_mode: detailed  # off, basic, detailed, expert
```

### Teaching Levels

| Level | Description |
|-------|-------------|
| `off` | Concise responses (default) |
| `basic` | Brief explanations of key concepts |
| `detailed` | Comprehensive explanations with examples |
| `expert` | Deep technical explanations |

### Example

```
You: /teach detailed
You: Add a binary search function

Agent: I'll create a binary search function. Let me explain
the approach:

## Binary Search Algorithm

Binary search is an efficient algorithm for finding items in
a sorted array. It works by repeatedly dividing the search
space in half.

**Time Complexity**: O(log n)
**Space Complexity**: O(1) for iterative, O(log n) for recursive

### How it works:

1. Start with the middle element
2. If target equals middle, we're done
3. If target < middle, search left half
4. If target > middle, search right half
5. Repeat until found or space exhausted

[Shows implementation with inline comments]
```

---

## Advanced Features

### Syntax Highlighting

Code in diffs and responses is syntax-highlighted for:
- Go, Python, JavaScript, TypeScript
- Rust, Java, C, C++
- HTML, CSS, JSON, YAML
- And many more...

### Code Symbol Navigation

The agent can navigate to specific symbols:
- Functions and methods
- Types and structs
- Constants and variables

### Token Tracking

Monitor your API usage:
- Tokens used per session
- Estimated cost
- Usage history

View in status bar or with:
```
You: /tokens
```

### Multi-File Changes

For changes spanning multiple files:
1. All changes shown in Diff panel
2. Navigate between files with `n`/`p`
3. Approve all or reject all

---

## Troubleshooting

### Common Issues

**"API key not found"**
```bash
# Set via environment
export ANVIL_ANTHROPIC_API_KEY="your-key"

# Or check config
cat ~/.anvil/config.yaml
```

**"Failed to connect"**
- Check internet connection
- Verify API key is valid
- Check provider status page

**"Context too long"**
- Start a new session
- Reduce file size in context
- Use `/clear` to reset conversation

**"Permission denied"**
- Check file permissions
- Ensure write access to project

### Debug Mode

Enable verbose logging:
```yaml
# ~/.anvil/config.yaml
log_level: debug
```

View logs:
```bash
tail -f ~/.anvil/logs/anvil-*.log
```

### Getting Help

- In-app: Press `?` for help
- GitHub: [Issues](https://github.com/siddharth-bhatnagar/anvil/issues)
- Docs: [Architecture](architecture.md)

---

## Appendix

### Environment Variables

| Variable | Description |
|----------|-------------|
| `ANVIL_CONFIG_DIR` | Config directory (default: `~/.anvil`) |
| `ANVIL_ANTHROPIC_API_KEY` | Anthropic API key |
| `ANVIL_OPENAI_API_KEY` | OpenAI API key |
| `ANVIL_GEMINI_API_KEY` | Google Gemini API key |
| `ANVIL_LOG_LEVEL` | Override log level |

### Command Line Flags

```bash
anvil [flags]

Flags:
  -c, --config string    Config file path
  -d, --dir string       Working directory
  -m, --model string     Model to use
  -p, --provider string  LLM provider
  -v, --version          Show version
  -h, --help             Show help
```
