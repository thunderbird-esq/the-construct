# Claude Code Integration Guide

Complete guide for using Claude Code with Matrix MUD, including custom commands, MCP servers, hooks, and multi-agent development patterns.

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [Custom Slash Commands](#custom-slash-commands)
3. [MCP Servers](#mcp-servers)
4. [Git Hooks](#git-hooks)
5. [Multi-Agent Workflows](#multi-agent-workflows)
6. [Prompt Templates](#prompt-templates)
7. [Best Practices](#best-practices)
8. [Troubleshooting](#troubleshooting)

---

## Quick Start

### Prerequisites

- Claude Code CLI installed
- Go 1.21+ installed
- Git configured

### Initial Setup

```bash
# Clone the repository
git clone <repository-url>
cd matrix-mud

# Install dependencies
make install

# Set up git hooks
make setup-hooks

# Verify everything works
make check
```

### First Steps with Claude

```bash
# Start Claude Code in the project directory
claude

# Try these commands:
/help              # List all available commands
make help          # List all Makefile commands
```

---

## Custom Slash Commands

Custom slash commands are defined in `.claude/commands/` (create this directory if needed).

### Available Commands

#### `/run-mud` - Start Development Server

**Location**: `.claude/commands/run-mud.md`

```markdown
Start the Matrix MUD development server with auto-reload.

Steps:
1. Build the latest version: `make build`
2. Start the server: `./bin/matrix-mud`
3. Inform the user that the server is running on:
   - Telnet: localhost:2323
   - Web: localhost:8080
   - Admin: localhost:9090
4. Suggest connecting with: `telnet localhost 2323`
```

**Usage**: `/run-mud`

#### `/test-combat` - Test Combat System

**Location**: `.claude/commands/test-combat.md`

```markdown
Run comprehensive combat system tests.

Steps:
1. Run unit tests for combat: `go test -v -run TestCombat ./tests/unit/`
2. Check for race conditions: `go test -race -run TestCombat ./tests/unit/`
3. Run integration tests: `go test -v ./tests/integration/`
4. Report test results and any failures
5. If failures, analyze and suggest fixes
```

**Usage**: `/test-combat`

#### `/gen-world` - Generate World Sections

**Location**: `.claude/commands/gen-world.md`

```markdown
Generate new world sections with rooms, NPCs, and items.

Ask the user:
1. What type of area? (city, dungeon, wasteland, etc.)
2. How many rooms? (default: 10)
3. Difficulty level? (1-10)

Then:
1. Generate JSON for world.json
2. Create appropriate NPCs with loot tables
3. Add items and encounters
4. Suggest testing the new area
```

**Usage**: `/gen-world`

#### `/refactor` - Code Refactoring with Agents

**Location**: `.claude/commands/refactor.md`

```markdown
Refactor code using specialized agents.

Ask the user what to refactor:
1. Specific file or package
2. Type of refactoring (structure, performance, style)

Then:
1. Use golang-pro agent for Go best practices
2. Use code-reviewer agent for quality checks
3. Show proposed changes
4. Run tests to verify
5. Update relevant documentation
```

**Usage**: `/refactor`

#### `/changelog-add` - Add Changelog Entry

**Location**: `.claude/commands/changelog-add.md`

```markdown
Add an entry to CHANGELOG.md.

Parameters:
- type: added|changed|deprecated|removed|fixed|security
- message: Description of the change

Steps:
1. Read current CHANGELOG.md
2. Add entry under [Unreleased] section
3. Use proper formatting
4. Save and confirm
```

**Usage**: `/changelog-add [type] [message]`

#### `/devlog-entry` - Create Development Log Entry

**Location**: `.claude/commands/devlog-entry.md`

```markdown
Create a new DEVLOG.md entry.

Ask the user:
1. Entry title
2. Key points to document

Then:
1. Read current DEVLOG.md
2. Create new dated entry
3. Include sections: What, Why, How, Results
4. Add technical details
5. Save and commit
```

**Usage**: `/devlog-entry`

#### `/security-audit` - Run Security Analysis

**Location**: `.claude/commands/security-audit.md`

```markdown
Perform comprehensive security audit.

Use the security-engineer agent to:
1. Scan for common vulnerabilities
2. Check authentication security
3. Review input validation
4. Check for SQL injection risks
5. Verify error handling doesn't leak info
6. Check for OWASP Top 10 issues
7. Generate security report
8. Provide remediation steps
```

**Usage**: `/security-audit`

#### `/performance-profile` - Profile and Optimize

**Location**: `.claude/commands/performance-profile.md`

```markdown
Profile the application and suggest optimizations.

Use the performance-engineer agent to:
1. Build with profiling: `go build -o bin/matrix-mud-prof`
2. Add pprof HTTP endpoint if needed
3. Run load tests
4. Analyze CPU and memory profiles
5. Identify hot paths
6. Suggest optimizations
7. Benchmark before/after
```

**Usage**: `/performance-profile`

### Creating Custom Commands

Create a file in `.claude/commands/your-command.md`:

```markdown
# Command description

Your command instructions here.

Steps:
1. First step
2. Second step
3. etc.
```

Then use with: `/your-command`

---

## MCP Servers

MCP (Model Context Protocol) servers provide specialized tools and context to Claude.

### matrix-mud-inspector

**Purpose**: Inspect game state, players, and world data.

**Configuration**: Create `.claude/mcp.json`:

```json
{
  "mcpServers": {
    "matrix-mud-inspector": {
      "command": "node",
      "args": ["./mcp-servers/inspector/index.js"],
      "env": {
        "DATA_DIR": "./data"
      }
    }
  }
}
```

**Tools Provided**:
- `list_players` - List all active players
- `get_player` - Get player details
- `list_rooms` - List all rooms
- `get_room` - Get room details with NPCs/items
- `inspect_npc` - Get NPC information
- `query_world` - Query world state

**Usage Example**:
```
Claude, use the matrix-mud-inspector to show me all players in the loading_program room.
```

### world-generator

**Purpose**: Generate world content (rooms, NPCs, items, quests).

**Configuration**:

```json
{
  "mcpServers": {
    "world-generator": {
      "command": "node",
      "args": ["./mcp-servers/generator/index.js"],
      "env": {
        "WORLD_FILE": "./data/world.json",
        "TEMPLATES_DIR": "./data/templates"
      }
    }
  }
}
```

**Tools Provided**:
- `generate_room` - Create new room with description
- `generate_npc` - Create NPC with stats and dialogue
- `generate_item` - Create item with properties
- `generate_quest` - Create quest chain
- `generate_area` - Generate connected area (multiple rooms)

### test-orchestrator

**Purpose**: Coordinate parallel test execution.

**Configuration**:

```json
{
  "mcpServers": {
    "test-orchestrator": {
      "command": "node",
      "args": ["./mcp-servers/test-orchestrator/index.js"]
    }
  }
}
```

**Tools Provided**:
- `run_tests` - Run specific test suites
- `parallel_test` - Run tests in parallel
- `coverage_report` - Generate coverage report
- `test_watch` - Watch mode for tests

---

## Git Hooks

### Pre-Commit Hook

Automatically installed with `make setup-hooks`.

**Location**: `.git/hooks/pre-commit`

```bash
#!/bin/sh

# Format code
make fmt

# Run linter
make lint || exit 1

# Run tests
make test || exit 1

exit 0
```

**What it does**:
1. Formats all Go code
2. Runs golangci-lint
3. Runs all tests
4. Blocks commit if tests fail

### Post-Commit Hook

**Location**: `.git/hooks/post-commit` (create manually)

```bash
#!/bin/sh

# Get commit message
MESSAGE=$(git log -1 --pretty=%B)

# If commit affects code, suggest devlog entry
if git diff-tree --no-commit-id --name-only -r HEAD | grep -E '\.(go|md)$' > /dev/null; then
    echo ""
    echo "💡 Consider adding a DEVLOG.md entry for this change:"
    echo "   /devlog-entry"
    echo ""
fi
```

---

## Multi-Agent Workflows

### Pattern 1: Parallel Documentation

Use multiple agents for independent documentation tasks.

**Command**:
```
Create comprehensive documentation using parallel agents:

1. Agent 1 (technical-writer): Create ARCHITECTURE.md
2. Agent 2 (documentation-expert): Create API.md
3. Agent 3 (technical-writer): Create DEVELOPMENT.md

Run these in parallel and coordinate results.
```

**Memory Budget**: ~6GB (3 agents × ~2GB each)

### Pattern 2: Code Review Chain

Sequential agents for thorough code review.

**Command**:
```
Review the code with agent chain:

1. golang-pro: Review Go best practices
2. security-engineer: Check for security issues
3. performance-engineer: Identify optimization opportunities
4. code-reviewer: Final quality check

Run sequentially, each building on previous findings.
```

**Memory Budget**: ~2GB (sequential execution)

### Pattern 3: Test Generation

Parallel test creation for different components.

**Command**:
```
Generate comprehensive tests using parallel agents:

1. test-engineer (Agent 1): Unit tests for pkg/world/
2. test-engineer (Agent 2): Unit tests for pkg/game/
3. test-automator (Agent 3): Integration tests for server
4. test-engineer (Agent 4): Benchmark tests

Max 3 agents in parallel, then Agent 4.
```

**Memory Budget**: ~6GB (3 agents) + 2GB (Agent 4)

### Pattern 4: Feature Implementation

Multi-agent feature development workflow.

**Command**:
```
Implement REST API feature:

Phase 1 (Planning):
- backend-architect: Design API structure

Phase 2 (Parallel Implementation):
- Agent 1 (golang-pro): Implement handlers
- Agent 2 (test-engineer): Write API tests
- Agent 3 (api-documenter): Create OpenAPI spec

Phase 3 (Review):
- code-reviewer: Review all changes
- security-engineer: Security audit

Phase 4 (Documentation):
- technical-writer: Update documentation
```

**Memory Budget**: Phases are sequential, max 6GB in Phase 2

---

## Prompt Templates

### Code Review Template

```
Please review the following code following Go best practices:

File: [filename]

Focus areas:
- Error handling
- Concurrency safety
- Resource management
- Code organization
- Performance

Use the golang-pro agent for specialized review.
```

### Bug Fix Template

```
Debug and fix the following issue:

Issue: [description]
File: [filename]
Error: [error message]

Steps:
1. Analyze the root cause
2. Propose fix
3. Write test to prevent regression
4. Verify fix works
```

### Feature Implementation Template

```
Implement the following feature:

Feature: [name]
Description: [details]
Requirements:
- [requirement 1]
- [requirement 2]

Use appropriate agents:
- Planning: backend-architect
- Implementation: golang-pro
- Testing: test-engineer
- Documentation: technical-writer
```

### Refactoring Template

```
Refactor the code in [file/package] to:

Goals:
- [goal 1]
- [goal 2]

Constraints:
- Maintain backward compatibility
- Don't break existing tests
- Follow Go best practices

Use golang-pro agent with code-reviewer for validation.
```

---

## Best Practices

### 1. Use the Right Agent for the Job

| Task | Recommended Agent |
|------|------------------|
| Go code review | golang-pro |
| API design | backend-architect |
| Testing | test-engineer, test-automator |
| Documentation | technical-writer, documentation-expert |
| Security | security-engineer |
| Performance | performance-engineer |
| Refactoring | golang-pro + code-reviewer |

### 2. Memory Management

**Monitor memory usage**:
```bash
# Check memory before spawning agents
top -l 1 | grep PhysMem

# If using >6GB, use sequential execution
```

**Agent memory estimates**:
- Simple tasks: ~1GB
- Code generation: ~2GB
- Complex analysis: ~3GB

**Stay under 8GB total**.

### 3. Parallel vs Sequential

**Use Parallel When**:
- Tasks are independent
- Different files/packages
- Read-only operations
- Memory budget allows (<8GB)

**Use Sequential When**:
- Tasks depend on each other
- Same file modifications
- Memory constrained
- Complex reasoning required

### 4. Context Management

**Keep context focused**:
```
# Good
Review the authentication code in pkg/auth/

# Less optimal
Review the entire codebase
```

**Use file patterns**:
```
# Review all test files
Review tests in tests/**/*_test.go

# Review specific package
Review pkg/world/*.go
```

### 5. Iterative Development

**Start small, iterate**:
```
1. Implement core feature
2. Add tests
3. Add error handling
4. Optimize
5. Document
```

**Don't try to do everything at once.**

---

## Troubleshooting

### Agent Memory Issues

**Symptom**: System slows down, swap usage increases

**Solution**:
```bash
# Kill running agents
pkill -f "claude.*agent"

# Use sequential execution
# Instead of parallel agents, run one at a time
```

### Build Failures

**Symptom**: Code doesn't compile after agent changes

**Solution**:
```bash
# Revert recent changes
git diff HEAD

# Run tests
make test

# Check linter
make lint

# Ask Claude to fix:
"The build is failing with error: [error]. Please fix."
```

### Test Failures

**Symptom**: Tests fail after changes

**Solution**:
```bash
# Run specific test
go test -v -run TestName ./path/to/test

# Check for race conditions
go test -race ./...

# Ask for debugging help:
"/test-combat and analyze failures"
```

### MCP Server Not Responding

**Symptom**: MCP tools unavailable

**Solution**:
```bash
# Check MCP configuration
cat .claude/mcp.json

# Test MCP server manually
node ./mcp-servers/inspector/index.js

# Restart Claude Code
```

### Slow Performance

**Symptom**: Claude responses are slow

**Solution**:
- Use more specific prompts
- Limit context scope
- Use simpler agents for simple tasks
- Close unused terminal sessions

---

## Advanced Usage

### Custom Agent Workflows

Create complex multi-agent workflows:

```javascript
// .claude/workflows/comprehensive-review.json
{
  "name": "Comprehensive Code Review",
  "steps": [
    {
      "agent": "golang-pro",
      "task": "Review Go best practices"
    },
    {
      "agent": "security-engineer",
      "task": "Security audit",
      "depends_on": ["golang-pro"]
    },
    {
      "agent": "test-engineer",
      "task": "Verify test coverage",
      "parallel": true
    },
    {
      "agent": "code-reviewer",
      "task": "Final review",
      "depends_on": ["security-engineer", "test-engineer"]
    }
  ]
}
```

### Integration with CI/CD

Use Claude Code in GitHub Actions:

```yaml
# .github/workflows/claude-review.yml
name: Claude Code Review

on: [pull_request]

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Claude Code Review
        run: |
          claude "/refactor --ci-mode"
```

### Metrics and Monitoring

Track agent usage:

```bash
# Log agent invocations
echo "$(date): Agent $AGENT_TYPE - Task: $TASK" >> .claude/agent-log.txt

# Analyze usage
grep "Agent" .claude/agent-log.txt | sort | uniq -c
```

---

## Resources

### Documentation
- [Claude Code Documentation](https://docs.anthropic.com/claude/docs/claude-code)
- [MCP Protocol](https://modelcontextprotocol.io/)
- [Agent SDK](https://github.com/anthropics/agent-sdk)

### Community
- [Claude Code GitHub Discussions](https://github.com/anthropics/claude-code/discussions)
- [MCP Servers Repository](https://github.com/modelcontextprotocol/servers)

### Project Specific
- [README.md](README.md) - Project overview
- [CONTRIBUTING.md](CONTRIBUTING.md) - Development guidelines
- [AGENTS.md](AGENTS.md) - Multi-agent patterns
- [DEVLOG.md](DEVLOG.md) - Development journal

---

**Last Updated**: 2025-11-24
**Claude Code Version**: Latest
**Project Version**: v1.0.0
