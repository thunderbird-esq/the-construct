# Contributing to Matrix MUD

Thank you for your interest in contributing to Matrix MUD! This document provides guidelines and instructions for contributing to the project.

## Code of Conduct

Be respectful, inclusive, and professional in all interactions. We're here to build a fun game together.

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Git
- Make (optional, but recommended)

### Setting Up Development Environment

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/matrix-mud.git
   cd matrix-mud
   ```

3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/yourusername/matrix-mud.git
   ```

4. Install dependencies:
   ```bash
   make install
   ```

5. Set up git hooks (optional but recommended):
   ```bash
   make setup-hooks
   ```

## Development Workflow

### 1. Create a Branch

Always create a new branch for your work:

```bash
git checkout -b feature/your-feature-name
```

Branch naming conventions:
- `feature/` - New features
- `bugfix/` - Bug fixes
- `hotfix/` - Critical production fixes
- `refactor/` - Code refactoring
- `docs/` - Documentation updates
- `test/` - Test additions or updates

### 2. Make Your Changes

- Write clean, readable code
- Follow Go best practices and idioms
- Add tests for new functionality
- Update documentation as needed
- Keep commits focused and atomic

### 3. Test Your Changes

Before submitting, ensure all tests pass:

```bash
# Run all tests
make test

# Run linting
make lint

# Run formatting
make fmt

# Run all checks
make check
```

### 4. Commit Your Changes

Write clear, descriptive commit messages:

```bash
git add .
git commit -m "Add feature: brief description

Detailed explanation of what changed and why.
Reference any related issues."
```

Commit message guidelines:
- Use present tense ("Add feature" not "Added feature")
- First line should be 50 characters or less
- Add detailed explanation after a blank line if needed
- Reference issues and PRs where relevant

### 5. Push to Your Fork

```bash
git push origin feature/your-feature-name
```

### 6. Submit a Pull Request

1. Go to the original repository on GitHub
2. Click "New Pull Request"
3. Select your fork and branch
4. Fill out the PR template with:
   - Clear description of changes
   - Related issues
   - Testing performed
   - Screenshots (if applicable)

## Code Standards

### Go Style Guide

- Follow the [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` for formatting (run `make fmt`)
- Run `go vet` to catch common mistakes (run `make vet`)
- Pass `golangci-lint` checks (run `make lint`)

### Code Organization

```
matrix-mud/
├── cmd/           # Application entry points
├── pkg/           # Public library code
├── internal/      # Private application code
├── tests/         # Test files
│   ├── unit/      # Unit tests
│   └── integration/ # Integration tests
└── data/          # Game data files
```

### Testing

- Write unit tests for all new functions
- Aim for >80% test coverage
- Use table-driven tests where appropriate
- Mock external dependencies

Example test:

```go
func TestPlayerMovement(t *testing.T) {
    tests := []struct {
        name      string
        direction string
        expected  string
    }{
        {"Move North", "north", "new_room_north"},
        {"Move South", "south", "new_room_south"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Documentation

- Add godoc comments for exported functions and types
- Update README.md for user-facing changes
- Add inline comments for complex logic
- Keep documentation up-to-date with code changes

Example godoc comment:

```go
// MovePlayer moves a player in the specified direction.
// It returns an error if the direction is invalid or blocked.
func MovePlayer(p *Player, direction string) error {
    // Implementation
}
```

## What to Contribute

### Good First Issues

Look for issues labeled `good-first-issue` - these are great for newcomers!

### Areas Needing Help

- Bug fixes
- New game features
- Performance improvements
- Documentation improvements
- Test coverage
- Code refactoring

### Feature Requests

Before starting work on a major feature:
1. Open an issue to discuss the idea
2. Wait for maintainer feedback
3. Get approval before investing significant time

## Pull Request Process

1. **Update Documentation** - Ensure README and relevant docs are updated
2. **Add Tests** - New code should include tests
3. **Pass CI Checks** - All automated checks must pass
4. **Code Review** - Address reviewer feedback promptly
5. **Squash Commits** - Clean up commit history if requested
6. **Merge** - Maintainers will merge once approved

### PR Checklist

- [ ] Code follows project style guidelines
- [ ] Tests added/updated and passing
- [ ] Documentation updated
- [ ] Commit messages are clear
- [ ] No merge conflicts
- [ ] CI checks passing
- [ ] Reviewed and approved

## Reporting Bugs

### Before Reporting

1. Check existing issues to avoid duplicates
2. Test on the latest version
3. Gather relevant information

### Bug Report Template

```markdown
**Describe the bug**
A clear and concise description.

**To Reproduce**
Steps to reproduce the behavior:
1. Go to '...'
2. Click on '...'
3. See error

**Expected behavior**
What you expected to happen.

**Environment:**
- OS: [e.g., Ubuntu 22.04]
- Go Version: [e.g., 1.22]
- Version: [e.g., v1.0.0]

**Additional context**
Any other relevant information.
```

## Feature Requests

### Feature Request Template

```markdown
**Is your feature request related to a problem?**
A clear description of the problem.

**Describe the solution you'd like**
What you want to happen.

**Describe alternatives you've considered**
Other solutions you've thought about.

**Additional context**
Any other relevant information.
```

## Community

- Be patient and respectful
- Help others when you can
- Give constructive feedback
- Celebrate contributions from all skill levels

## Questions?

- Open a GitHub issue with the `question` label
- Check existing documentation first
- Be specific and provide context

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Recognition

Contributors will be recognized in:
- The project README
- Release notes for significant contributions
- Special thanks for first-time contributors

Thank you for contributing to Matrix MUD!
