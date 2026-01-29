# Contributing to Anvil

Thank you for your interest in contributing to Anvil! This document provides guidelines for contributing to the project.

## Code of Conduct

Be respectful, constructive, and professional in all interactions.

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally
3. Create a feature branch from `main`
4. Make your changes
5. Push to your fork and submit a pull request

See [docs/developer.md](docs/developer.md) for detailed development setup.

## Pull Request Process

1. **Before Starting**:
   - Check existing issues and PRs to avoid duplication
   - For major changes, open an issue first to discuss the approach
   - Make sure you understand the project philosophy (see [.claude/CLAUDE.md](.claude/CLAUDE.md))

2. **During Development**:
   - Follow Go idioms and existing code style
   - Write tests for new functionality
   - Keep changes focused (one feature/fix per PR)
   - Update documentation as needed
   - Ensure all tests pass: `go test ./...`
   - Run linter: `golangci-lint run`

3. **Commit Messages**:
   - Use [Conventional Commits](https://www.conventionalcommits.org/) format
   - Be descriptive but concise
   - Examples:
     - `feat: add syntax highlighting to diff viewer`
     - `fix: prevent API key leakage in logs`
     - `docs: update installation instructions`
     - `test: add unit tests for key manager`

4. **Pull Request**:
   - Provide a clear description of the changes
   - Reference any related issues
   - Include screenshots for UI changes
   - Ensure CI passes (tests, linting, build)

5. **Review Process**:
   - Address reviewer feedback promptly
   - Be open to suggestions and iterate
   - Maintainers may request changes or clarifications

## What to Contribute

### Good First Issues

Look for issues labeled `good first issue` - these are suitable for newcomers.

### Areas We Need Help

- **Core Features**: Implementing items from the [implementation plan](.claude/plans/)
- **Testing**: Writing unit and integration tests
- **Documentation**: Improving guides, tutorials, examples
- **Bug Fixes**: Addressing reported issues
- **Performance**: Optimizing slow operations
- **Platform Support**: Testing and fixing issues on different OS

### What We're Not Looking For

- Major architectural changes without prior discussion
- Features that conflict with the project philosophy
- Dependencies that significantly increase binary size
- Changes that break backwards compatibility without strong justification

## Code Style

- Run `gofmt` before committing (enforced by CI)
- Follow Go best practices and idioms
- Keep functions small and focused
- Add comments for exported types and functions
- Use meaningful variable names
- Avoid premature optimization

## Testing

- Write unit tests for business logic
- Use table-driven tests for multiple cases
- Test error paths, not just happy paths
- Mock external dependencies
- Target >80% code coverage

Example test structure:

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid input", "test", "result", false},
        {"invalid input", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("unexpected error: %v", err)
            }
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Documentation

- Update relevant documentation for your changes
- Add comments for complex logic
- Include examples where helpful
- Update README.md if adding user-facing features

## Security

- Never commit API keys, secrets, or credentials
- Report security vulnerabilities privately (don't open public issues)
- Follow secure coding practices
- Review [internal/config/keys.go](internal/config/keys.go) for secrets handling

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Questions?

- Open an issue for questions about contributing
- Check [docs/developer.md](docs/developer.md) for development guides
- Review existing PRs to see the contribution process in action

## Thank You!

Your contributions make Anvil better for everyone. We appreciate your time and effort!
