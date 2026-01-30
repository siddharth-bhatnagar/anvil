# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability in Anvil, please report it responsibly.

### How to Report

1. **DO NOT** open a public GitHub issue for security vulnerabilities
2. Email security concerns to: [security@example.com] (or use GitHub's private vulnerability reporting)
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Any suggested fixes

### What to Expect

- **Acknowledgment**: Within 48 hours
- **Initial Assessment**: Within 7 days
- **Resolution Timeline**: Depends on severity
  - Critical: 24-48 hours
  - High: 7 days
  - Medium: 30 days
  - Low: 90 days

## Security Measures

### API Key Protection

Anvil takes API key security seriously:

1. **OS Keychain Storage**
   - API keys are stored in your operating system's secure keychain
   - macOS: Keychain Access
   - Linux: libsecret/GNOME Keyring
   - Windows: Credential Manager

2. **Never Logged**
   - API keys are never written to log files
   - Keys are masked in any error messages
   - Debug mode does not expose keys

3. **Memory Handling**
   - Keys are cleared from memory after use
   - No plaintext storage in config files

### File System Security

1. **Sandboxing**
   - File operations are restricted to the project directory
   - Path traversal attacks are prevented
   - Symbolic links are handled safely

2. **Approval Gates**
   - All file modifications require explicit user approval
   - Destructive operations show warnings
   - Changes can be previewed before applying

3. **Audit Logging**
   - All file operations are logged
   - Logs include timestamps and operation details
   - Logs do not contain file contents or secrets

### Shell Command Security

1. **Approval Required**
   - All shell commands require user approval
   - Destructive commands are flagged
   - Command preview shows full command

2. **Dangerous Command Detection**
   - Commands like `rm -rf`, `sudo`, etc. are flagged
   - Users must explicitly confirm dangerous operations
   - Option to run in dry-run mode

### Network Security

1. **HTTPS Only**
   - All API communications use HTTPS
   - TLS 1.2+ required
   - Certificate validation enforced

2. **No Telemetry**
   - Anvil does not collect usage data
   - No phone-home functionality
   - All data stays local

## Best Practices for Users

### API Key Management

```bash
# Good: Use environment variables
export ANVIL_ANTHROPIC_API_KEY="sk-..."

# Good: Use a secrets manager
anvil --api-key-cmd "pass show anthropic/api-key"

# Bad: Don't put keys in config files
# Bad: Don't commit keys to git
```

### Project Security

1. **Review Changes**
   - Always review diffs before approving
   - Pay attention to file permissions
   - Check for unintended modifications

2. **Git Integration**
   - Use `.gitignore` to exclude sensitive files
   - Review commits before pushing
   - Don't commit generated credentials

3. **Environment**
   - Run Anvil with minimal privileges
   - Use project-specific API keys when possible
   - Regularly rotate API keys

## Security Configuration

### Log Security

```yaml
# ~/.anvil/config.yaml
log_level: info  # Avoid 'debug' in production
log_dir: ~/.anvil/logs  # Ensure directory is private
```

Ensure log directory permissions:
```bash
chmod 700 ~/.anvil/logs
```

### File Permissions

Anvil creates files with these permissions:
- Config files: `0600` (user read/write only)
- Log files: `0600` (user read/write only)
- Session files: `0644` (user read/write, others read)

## Known Limitations

1. **Clipboard Access**
   - Anvil may read/write clipboard for copy/paste
   - Be cautious with sensitive data in clipboard

2. **Terminal History**
   - Commands may appear in shell history
   - Consider using `HISTCONTROL=ignorespace`

3. **Memory**
   - Conversation history is kept in memory
   - Large contexts may contain sensitive code

## Dependency Security

Anvil's dependencies are regularly audited:

```bash
# Check for vulnerabilities
go list -m all | nancy sleuth

# Update dependencies
go get -u ./...
go mod tidy
```

## Incident Response

If you believe your API key or system has been compromised:

1. **Immediately** revoke the affected API key
2. Generate a new API key from your provider
3. Update the key in Anvil
4. Review recent activity in your provider's dashboard
5. Check Anvil's logs for suspicious activity

## Security Updates

Security updates are released as soon as possible after a vulnerability is confirmed.

- Subscribe to releases on GitHub to receive notifications
- Check the CHANGELOG for security-related updates
- Keep Anvil updated to the latest version

## Contact

For security-related inquiries:
- GitHub Security Advisories: [Report a vulnerability](https://github.com/siddharth-bhatnagar/anvil/security/advisories/new)
- Email: [Create an issue requesting private contact]

Thank you for helping keep Anvil secure!
