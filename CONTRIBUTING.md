# Contributing to AsyncAPI Generator

Thank you for your interest in contributing to AsyncAPI Generator! This document provides guidelines for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Submitting Changes](#submitting-changes)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Documentation](#documentation)

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for everyone.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/asyncapi-generator.git
   cd asyncapi-generator
   ```
3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/fedanant/asyncapi-generator.git
   ```

## Development Setup

### Prerequisites

- Go 1.22 or higher
- Make
- Git

### Install Dependencies

```bash
# Install Go dependencies
go mod download

# Install development tools
make lint-install
```

### Build

```bash
# Build the binary
make build

# Run tests
make test

# Run linter
make lint

# Format code
make fmt
```

## Making Changes

### Create a Branch

Always create a new branch for your changes:

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

### Branch Naming Convention

- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `refactor/` - Code refactoring
- `test/` - Adding or updating tests
- `chore/` - Maintenance tasks

### Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation changes
- `style` - Code style changes (formatting, etc.)
- `refactor` - Code refactoring
- `test` - Adding or updating tests
- `chore` - Maintenance tasks

**Examples:**
```
feat(parser): add support for MQTT protocol
fix(generator): handle nil pointer in type extraction
docs(readme): update installation instructions
test(parser): add tests for parameterized channels
```

## Submitting Changes

### Before Submitting

1. **Run tests**: `make test`
2. **Run linter**: `make lint`
3. **Format code**: `make fmt`
4. **Update documentation** if needed
5. **Add tests** for new features

### Pull Request Process

1. **Update your branch** with the latest upstream changes:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Push your changes**:
   ```bash
   git push origin your-branch-name
   ```

3. **Create a Pull Request** on GitHub:
   - Provide a clear title and description
   - Reference any related issues (e.g., "Fixes #123")
   - Describe what changes you made and why
   - Add screenshots or examples if applicable

4. **Respond to feedback**:
   - Address review comments
   - Make requested changes
   - Keep the conversation respectful and constructive

### Pull Request Checklist

- [ ] Tests pass locally
- [ ] Code follows project style (linter passes)
- [ ] Documentation updated (if needed)
- [ ] Commit messages follow conventions
- [ ] No breaking changes (or clearly documented)
- [ ] Added tests for new features
- [ ] Updated CHANGELOG.md (for significant changes)

## Coding Standards

### Go Style Guide

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for formatting
- Use `golangci-lint` for linting

### Code Guidelines

1. **Keep it simple** - Prefer clarity over cleverness
2. **Write tests** - Aim for good test coverage
3. **Document exports** - All exported functions, types, and constants should have comments
4. **Handle errors** - Always check and handle errors appropriately
5. **Use meaningful names** - Variables, functions, and types should have descriptive names
6. **Keep functions small** - Functions should do one thing well
7. **Avoid global state** - Prefer dependency injection

### Code Review Guidelines

When reviewing code:
- Be constructive and respectful
- Focus on the code, not the person
- Explain your reasoning
- Suggest alternatives when possible
- Approve when ready, request changes when needed

## Testing

### Writing Tests

- Place tests in `*_test.go` files
- Use table-driven tests when appropriate
- Test both success and error cases
- Mock external dependencies

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test
go test -v ./internal/asyncapi -run TestParseFolder
```

### Test Coverage

We aim for:
- Minimum 70% overall coverage
- New code should have tests
- Critical paths should have high coverage

## Documentation

### Code Documentation

- All exported functions must have comments
- Comments should explain "why", not just "what"
- Use examples in documentation when helpful

### README Updates

When adding features:
- Update README.md with usage examples
- Add to the annotations reference if applicable
- Update the Quick Start if needed

### Example Updates

- Keep examples up to date with changes
- Add new examples for significant features
- Test examples to ensure they work

## Release Process

Releases are automated through GitHub Actions:

1. Maintainers create a new tag:
   ```bash
   git tag -a v1.2.3 -m "Release v1.2.3"
   git push origin v1.2.3
   ```

2. GitHub Actions automatically:
   - Builds binaries for all platforms
   - Creates a GitHub release
   - Generates changelog

## Questions?

- Open an issue for bugs or feature requests
- Start a discussion for questions or ideas
- Check existing issues before creating new ones

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to AsyncAPI Generator! 🎉
