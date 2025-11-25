# Development Guide

Complete guide to setting up your development environment and working with Matrix MUD.

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [Development Environment](#development-environment)
3. [Build & Run](#build--run)
4. [Testing](#testing)
5. [Debugging](#debugging)
6. [IDE Setup](#ide-setup)
7. [Common Issues](#common-issues)

---

## Quick Start

### Prerequisites

```bash
# Check Go version (need 1.21+)
go version

# Check Git
git --version

# Check Make (optional but recommended)
make --version
```

### Clone & Setup

```bash
# Clone repository
git clone <repository-url>
cd matrix-mud

# Install dependencies
make install

# Verify setup
make check
```

### Run Server

```bash
# Build and run
make run

# Or manually
go build -o bin/matrix-mud .
./bin/matrix-mud
```

### Connect as Client

```bash
# Terminal 1: Server running
./bin/matrix-mud

# Terminal 2: Connect
telnet localhost 2323
```

---

## Development Environment

### Required Tools

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.21+ | Language runtime |
| Git | Any | Version control |
| Make | Any | Build automation |

### Recommended Tools

| Tool | Purpose |
|------|---------|
| golangci-lint | Linting |
| air | Hot reload |
| delve | Debugging |
| pprof | Profiling |

### Install Recommended Tools

```bash
# golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# air (hot reload)
go install github.com/cosmtrek/air@latest

# delve (debugger)
go install github.com/go-delve/delve/cmd/dlv@latest
```

### Directory Structure

```
matrix-mud/
├── cmd/
│   └── matrix-mud/          # Main application (future)
├── pkg/
│   ├── world/              # World management (future)
│   ├── game/               # Game logic (future)
│   ├── auth/               # Authentication (future)
│   └── network/            # Networking (future)
├── internal/
│   └── config/             # Configuration (future)
├── tests/
│   ├── unit/               # Unit tests
│   └── integration/        # Integration tests
├── data/
│   ├── world.json          # World data
│   ├── dialogue.json       # NPC dialogue
│   ├── users.json          # User accounts
│   └── players/            # Player saves
├── docs/                   # Documentation
├── .github/                # GitHub Actions
└── Current files:
    ├── main.go             # Entry point & connection handling
    ├── world.go            # World state & game logic
    ├── terminal.go         # Terminal/ANSI utilities
    ├── web.go              # Web server
    └── admin.go            # Admin server
```

---

## Build & Run

### Using Make

```bash
# Show all commands
make help

# Build
make build

# Run
make run

# Hot reload (requires air)
make dev

# Clean
make clean
```

### Manual Build

```bash
# Basic build
go build -o bin/matrix-mud .

# Optimized build (smaller binary)
go build -ldflags="-s -w" -o bin/matrix-mud .

# Cross-compile for Linux
GOOS=linux GOARCH=amd64 go build -o bin/matrix-mud-linux .
```

### Development Mode

```bash
# With hot reload
air

# With race detector
go run -race .

# With verbose output
go run -v .
```

---

## Testing

### Run Tests

```bash
# All tests
make test

# Unit tests only
make test-unit

# Integration tests only
make test-integration

# With coverage
make test-coverage
# Opens coverage.html in browser
```

### Write Tests

**Unit Test Example**:

```go
// tests/unit/world_test.go
package unit

import (
    "testing"
)

func TestPlayerMovement(t *testing.T) {
    tests := []struct {
        name      string
        direction string
        expected  string
    }{
        {"Move North", "north", "new_room"},
        {"Invalid Direction", "invalid", "No exit."},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Test Coverage Goals

- **Target**: 80%+ coverage
- **Priority**: Core game logic, combat system, inventory
- **Current**: Infrastructure ready, tests pending

---

## Debugging

### Using Delve

```bash
# Start debugger
dlv debug

# Set breakpoint
(dlv) break main.handleConnection
(dlv) continue

# Inspect variables
(dlv) print player
(dlv) locals
```

### Printf Debugging

```go
// Add logging
fmt.Printf("DEBUG: Player %s in room %s\n", player.Name, player.RoomID)
```

### Race Detector

```bash
# Run with race detector
go run -race .

# Build with race detector
go build -race -o bin/matrix-mud-race .
./bin/matrix-mud-race
```

### Profiling

```bash
# Add pprof HTTP endpoint to main.go
import _ "net/http/pprof"
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

# Run server, then profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Memory profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Interactive analysis
(pprof) top
(pprof) list functionName
```

---

## IDE Setup

### VSCode

**Extensions**:
- Go (official)
- Go Test Explorer
- Error Lens
- GitLens

**settings.json**:
```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.formatTool": "goimports",
  "[go]": {
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
      "source.organizeImports": true
    }
  },
  "go.testFlags": ["-v", "-race"],
  "go.coverOnSave": true
}
```

**launch.json** (debugging):
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch Server",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}",
      "args": []
    }
  ]
}
```

### GoLand

**Configuration**:
1. File → Settings → Go → GOROOT: Set to Go SDK
2. File → Settings → Tools → File Watchers: Add gofmt
3. Run → Edit Configurations → Add Go Build

**Run Configuration**:
- Run kind: Package
- Package path: github.com/yourusername/matrix-mud
- Working directory: $PROJECT_DIR$

---

## Common Issues

### Issue: Port Already in Use

**Error**: `bind: address already in use`

**Solution**:
```bash
# Find process using port
lsof -i :2323

# Kill process
kill -9 <PID>

# Or use different port (edit main.go)
listener, err := net.Listen("tcp", ":2324")
```

### Issue: Permission Denied (data files)

**Error**: `permission denied: data/world.json`

**Solution**:
```bash
# Fix permissions
chmod 644 data/*.json
chmod 755 data/players/

# Or run with sudo (not recommended)
sudo ./bin/matrix-mud
```

### Issue: Module Not Found

**Error**: `cannot find module`

**Solution**:
```bash
# Re-download modules
go mod download

# Tidy dependencies
go mod tidy

# Verify modules
go mod verify
```

### Issue: Build Fails

**Error**: Various compilation errors

**Solution**:
```bash
# Clean and rebuild
make clean
make build

# Check Go version
go version  # Should be 1.21+

# Update dependencies
go get -u ./...
go mod tidy
```

### Issue: Tests Fail

**Error**: Tests failing after changes

**Solution**:
```bash
# Run specific test
go test -v -run TestName ./tests/unit/

# Check for race conditions
go test -race ./...

# Verbose output
go test -v ./...
```

### Issue: Slow Performance

**Symptoms**: High latency, slow responses

**Solutions**:
```bash
# Profile CPU
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Profile memory
go tool pprof http://localhost:6060/debug/pprof/heap

# Check goroutines
go tool pprof http://localhost:6060/debug/pprof/goroutine

# Enable race detector
go run -race .
```

---

## Development Workflow

### Feature Development

```bash
# 1. Create branch
git checkout -b feature/new-combat-system

# 2. Make changes
vim world.go

# 3. Test locally
make test

# 4. Lint
make lint

# 5. Format
make fmt

# 6. Commit
git add .
git commit -m "Add new combat system"

# 7. Push
git push origin feature/new-combat-system

# 8. Create PR on GitHub
```

### Code Review Checklist

- [ ] Tests pass (`make test`)
- [ ] Linter passes (`make lint`)
- [ ] Code formatted (`make fmt`)
- [ ] Documentation updated
- [ ] No race conditions (`go test -race`)
- [ ] Performance acceptable
- [ ] Security considerations addressed

---

## Performance Tips

### Optimization Strategies

1. **Use RLock for Reads**:
   ```go
   // Instead of:
   world.mutex.Lock()

   // Use:
   world.mutex.RLock()  // For read-only operations
   ```

2. **Preallocate Slices**:
   ```go
   // Instead of:
   items := []Item{}

   // Use:
   items := make([]Item, 0, expectedSize)
   ```

3. **Reuse Buffers**:
   ```go
   var buf bytes.Buffer
   buf.Reset()  // Reuse instead of allocate
   ```

4. **Profile Before Optimizing**:
   ```bash
   go test -bench=. -cpuprofile=cpu.prof
   go tool pprof cpu.prof
   ```

### Memory Optimization

```bash
# Check memory usage
go build -o bin/matrix-mud .
./bin/matrix-mud &
ps aux | grep matrix-mud

# Profile memory
go tool pprof http://localhost:6060/debug/pprof/heap
(pprof) top
```

---

## Resources

### Documentation
- [Go Documentation](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

### Project Docs
- [README.md](../README.md)
- [ARCHITECTURE.md](ARCHITECTURE.md)
- [API.md](API.md)
- [CONTRIBUTING.md](../CONTRIBUTING.md)

### Tools
- [golangci-lint](https://golangci-lint.run/)
- [delve](https://github.com/go-delve/delve)
- [air](https://github.com/cosmtrek/air)

---

**Last Updated**: 2025-11-24
**Go Version**: 1.21+
