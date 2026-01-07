# Contributing to File-Meta

Thank you for your interest in contributing to File-Meta! This document provides guidelines and instructions for contributing.

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers and help them get started
- Give and receive constructive feedback gracefully
- Focus on what's best for the community

## Getting Started

1. **Fork the repository**
2. **Clone your fork**
   ```bash
   git clone https://github.com/your-username/file-meta.git
   cd file-meta
   ```
3. **Install dependencies**
   ```bash
   make deps
   ```
4. **Create a branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Development Workflow

### 1. Make Your Changes

- Write clear, idiomatic Go code
- Follow existing code style and patterns
- Keep changes focused and atomic

### 2. Write Tests

- Add unit tests for new functionality
- Ensure existing tests still pass
- Aim for high test coverage

```bash
make test
```

### 3. Format Code

```bash
make fmt
```

### 4. Run Linters

```bash
make lint
```

### 5. Test Your Changes

```bash
# Run all tests
make test

# Check coverage
make test-coverage

# Run the application locally
make run
```

## Coding Standards

### Go Style Guide

- Follow the [Effective Go](https://golang.org/doc/effective_go) guidelines
- Use `gofmt` for formatting
- Run `golangci-lint` before committing

### Naming Conventions

- Use descriptive names for variables and functions
- Follow Go naming conventions (e.g., `MixedCaps` for exported names)
- Prefer short, clear names for local variables

### Error Handling

- Always check and handle errors
- Provide context in error messages
- Use custom error types when appropriate

### Comments

- Write clear, concise comments
- Document all exported functions, types, and packages
- Use `//` for line comments, `/* */` for package docs

## Testing Guidelines

### Unit Tests

- Test files should end with `_test.go`
- Use table-driven tests where appropriate
- Mock external dependencies
- Test edge cases and error conditions

Example:
```go
func TestMetadataExtraction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *Result
        wantErr bool
    }{
        // test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

### Integration Tests

- Place in `*_integration_test.go` files
- Use build tags: `// +build integration`
- Run with: `go test -tags=integration`

## Pull Request Process

1. **Update Documentation**
   - Update README.md if adding features
   - Add/update API documentation
   - Include code comments

2. **Commit Messages**
   - Use clear, descriptive commit messages
   - Follow conventional commits format:
     ```
     type(scope): subject
     
     body (optional)
     
     footer (optional)
     ```
   - Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`
   - Example: `feat(handlers): add batch file processing endpoint`

3. **Create Pull Request**
   - Provide a clear description of changes
   - Reference related issues
   - Include screenshots for UI changes
   - Ensure CI/CD checks pass

4. **Code Review**
   - Address reviewer feedback promptly
   - Keep discussions focused and professional
   - Update PR based on feedback

## Issue Reporting

### Bug Reports

Include:
- Clear description of the issue
- Steps to reproduce
- Expected vs actual behavior
- Environment details (OS, Go version)
- Relevant logs or error messages

### Feature Requests

Include:
- Clear description of the feature
- Use cases and benefits
- Possible implementation approach
- Any related issues or PRs

## Project Structure

```
file-meta/
├── config/          # Configuration management
├── handlers/        # HTTP handlers
├── internal/        # Internal packages
│   ├── logger/      # Logging utilities
│   ├── metadata/    # Core metadata logic
│   └── models/      # Data models
├── middleware/      # HTTP middleware
├── testdata/        # Test fixtures
└── main.go          # Entry point
```

## Questions?

- Open an issue for questions
- Check existing issues and PRs
- Review documentation

Thank you for contributing! 🎉
