# Contributing to go-cache

Thank you for your interest in contributing to go-cache! This document provides guidelines and instructions for contributing.

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers and help them learn
- Focus on constructive feedback
- Respect different viewpoints and experiences

## How to Contribute

### Reporting Bugs

If you find a bug, please open an issue with:
- Clear description of the bug
- Steps to reproduce
- Expected behavior
- Actual behavior
- Environment details (Go version, OS, Redis version, etc.)
- Minimal code example if possible

### Suggesting Features

Feature suggestions are welcome! Please open an issue with:
- Clear description of the feature
- Use case and motivation
- Proposed API design (if applicable)
- Examples of how it would be used

### Pull Requests

1. **Fork the repository**
2. **Create a feature branch** from `main`
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. **Make your changes**
   - Follow the coding standards (see below)
   - Add tests for new functionality
   - Update documentation as needed
4. **Run tests and checks**
   ```bash
   make test
   make test-cover
   ```
5. **Commit your changes**
   - Use clear, descriptive commit messages
   - Follow conventional commit format when possible
6. **Push and create a Pull Request**
   - Provide a clear description of changes
   - Reference any related issues

## Coding Standards

### Go Style Guide

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofmt` for formatting
- Follow the existing code style in the project

### Code Quality Requirements

- **Functions**: Keep functions under 25 lines when possible (max 30)
- **Parameters**: Maximum 4 parameters per function (5 requires refactoring)
- **Error Handling**: All errors must be handled explicitly
- **Comments**: Write self-documenting code; comments only when necessary
- **Tests**: Write tests for all exported functions

### Error Handling

- Always handle errors explicitly
- Use `fmt.Errorf` with `%w` verb for error wrapping
- Return sentinel errors for expected error conditions
- Document all error conditions in function comments

### Testing

- Write tests for all new functionality
- Maintain or improve test coverage
- Use table-driven tests for multiple test cases
- Test both success and error paths
- Clean up resources in tests (use `defer` for shutdowns)
- Run integration tests (requires Redis):
  ```bash
  go test ./test/integration/...
  ```
- Run end-to-end tests:
  ```bash
  go test ./test/e2e/...
  ```

### Documentation

- Add GoDoc comments to all exported functions
- Include examples in documentation when helpful
- Update README.md for user-facing changes
- Update API documentation for interface changes

## Project Structure

```text
go-cache/
├── internal/          # Internal packages (not exported)
│   ├── adapter/      # Tracer, Logger, Semaphore adapters
│   ├── cache/        # Core cache implementation
│   ├── errs/         # Error wrapping utilities
│   ├── hash/         # Hash generation for cache keys
│   ├── key/          # Key generation logic
│   ├── lock/         # Distributed locking
│   ├── serialize/    # Serialization utilities
│   └── version/      # Version-based invalidation
├── test/             # Test utilities and integration tests
│   ├── e2e/         # End-to-end tests
│   ├── integration/ # Integration tests
│   └── util/        # Test utilities
├── *.go              # Public API
├── *_test.go         # Unit tests
├── README.md         # User documentation
├── CONTRIBUTING.md   # This file
├── LICENSE           # MIT License
└── Makefile          # Build and test commands
```

## Development Workflow

1. **Set up your environment**
   ```bash
   git clone https://github.com/adityakw90/go-cache.git
   cd go-cache
   
   # Start Redis for integration tests (optional)
   docker-compose up -d
   ```

2. **Make your changes**
   - Create a feature branch
   - Make your code changes
   - Add tests
   - Update documentation

3. **Test your changes**
   ```bash
   # Run all tests
   make test
   
   # Run tests with coverage
   make test-cover
   
   # Run tests with verbose output
   make test verbose
   
   # Run integration tests (requires Redis)
   go test ./test/integration/...
   
   # Run end-to-end tests
   go test ./test/e2e/...
   ```

4. **Submit your changes**
   - Push to your fork
   - Create a Pull Request
   - Respond to feedback

## Review Process

- All PRs require at least one approval
- Maintainers will review code quality, tests, and documentation
- Address review comments promptly
- Keep PRs focused and reasonably sized

## Questions?

If you have questions about contributing, please:
- Open an issue with the `question` label
- Check existing issues and discussions
- Review the code and documentation

Thank you for contributing to go-cache! 🎉

