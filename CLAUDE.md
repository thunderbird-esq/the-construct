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

## Remaining Implementation Phases

This section documents the detailed steps to complete the Matrix MUD project setup. These phases were planned during the initial project initialization and should be completed in order.

### Phase 4: Development Workflow Enhancements

#### 4.1 Git Hooks Enhancement

**Status**: ✅ Completed

**Steps**:

1. **Verify Pre-Commit Hook**:
   ```bash
   # Check if hook exists
   ls -la .git/hooks/pre-commit

   # Test it
   git add -A
   git commit -m "test" --dry-run
   ```

2. **Create Post-Commit Hook**:
   Create `.git/hooks/post-commit`:
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

   Make executable:
   ```bash
   chmod +x .git/hooks/post-commit
   ```

3. **Create Commit-Msg Hook**:
   Create `.git/hooks/commit-msg`:
   ```bash
   #!/bin/sh

   # Check commit message format
   MESSAGE=$(cat "$1")

   # Conventional commits check
   if ! echo "$MESSAGE" | grep -qE '^(feat|fix|docs|style|refactor|test|chore|perf)(\(.+\))?: .+'; then
       echo "❌ Commit message must follow conventional commits format:"
       echo "   type(scope): description"
       echo ""
       echo "   Types: feat, fix, docs, style, refactor, test, chore, perf"
       exit 1
   fi
   ```

   Make executable:
   ```bash
   chmod +x .git/hooks/commit-msg
   ```

#### 4.2 MCP Server Implementation

**Status**: Not Started

**Critical Note**: These servers enhance development but are optional. Skip if time-constrained.

**Steps**:

1. **Create MCP Servers Directory**:
   ```bash
   mkdir -p mcp-servers/{inspector,generator,test-orchestrator}
   ```

2. **Implement matrix-mud-inspector**:

   Create `mcp-servers/inspector/package.json`:
   ```json
   {
     "name": "matrix-mud-inspector",
     "version": "1.0.0",
     "type": "module",
     "dependencies": {
       "@modelcontextprotocol/sdk": "^0.5.0"
     }
   }
   ```

   Create `mcp-servers/inspector/index.js`:
   ```javascript
   import { Server } from '@modelcontextprotocol/sdk/server/index.js';
   import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
   import fs from 'fs/promises';
   import path from 'path';

   const DATA_DIR = process.env.DATA_DIR || './data';

   const server = new Server({
     name: 'matrix-mud-inspector',
     version: '1.0.0'
   }, {
     capabilities: {
       tools: {}
     }
   });

   // Tool: list_players
   server.setRequestHandler('tools/list', async () => ({
     tools: [
       {
         name: 'list_players',
         description: 'List all player files',
         inputSchema: {
           type: 'object',
           properties: {}
         }
       },
       {
         name: 'get_player',
         description: 'Get player details by name',
         inputSchema: {
           type: 'object',
           properties: {
             name: { type: 'string', description: 'Player name' }
           },
           required: ['name']
         }
       },
       {
         name: 'inspect_world',
         description: 'Get world state',
         inputSchema: {
           type: 'object',
           properties: {}
         }
       }
     ]
   }));

   server.setRequestHandler('tools/call', async (request) => {
     const { name, arguments: args } = request.params;

     switch (name) {
       case 'list_players':
         const files = await fs.readdir(path.join(DATA_DIR, 'players'));
         return {
           content: [{
             type: 'text',
             text: JSON.stringify(files.filter(f => f.endsWith('.json')), null, 2)
           }]
         };

       case 'get_player':
         const playerData = await fs.readFile(
           path.join(DATA_DIR, 'players', `${args.name}.json`),
           'utf-8'
         );
         return {
           content: [{
             type: 'text',
             text: playerData
           }]
         };

       case 'inspect_world':
         const worldData = await fs.readFile(
           path.join(DATA_DIR, 'world.json'),
           'utf-8'
         );
         return {
           content: [{
             type: 'text',
             text: worldData
           }]
         };

       default:
         throw new Error(`Unknown tool: ${name}`);
     }
   });

   const transport = new StdioServerTransport();
   await server.connect(transport);
   ```

3. **Create .claude/mcp.json**:
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

4. **Install MCP Dependencies**:
   ```bash
   cd mcp-servers/inspector
   npm install
   cd ../..
   ```

5. **Test MCP Server**:
   ```bash
   # Test manually
   node ./mcp-servers/inspector/index.js

   # Use in Claude Code
   # "Claude, use matrix-mud-inspector to list all players"
   ```

**Note**: world-generator and test-orchestrator servers are similar. Implement only if needed.

#### 4.3 Hot Reload Configuration

**Status**: ✅ Completed

**Steps**:

1. **Install Air**:
   ```bash
   go install github.com/cosmtrek/air@latest
   ```

2. **Create .air.toml**:
   ```toml
   root = "."
   testdata_dir = "testdata"
   tmp_dir = "tmp"

   [build]
     args_bin = []
     bin = "./tmp/main"
     cmd = "go build -o ./tmp/main ."
     delay = 1000
     exclude_dir = ["assets", "tmp", "vendor", "testdata", "data"]
     exclude_file = []
     exclude_regex = ["_test.go"]
     exclude_unchanged = false
     follow_symlink = false
     full_bin = ""
     include_dir = []
     include_ext = ["go", "tpl", "tmpl", "html"]
     include_file = []
     kill_delay = "0s"
     log = "build-errors.log"
     poll = false
     poll_interval = 0
     rerun = false
     rerun_delay = 500
     send_interrupt = false
     stop_on_error = false

   [color]
     app = ""
     build = "yellow"
     main = "magenta"
     runner = "green"
     watcher = "cyan"

   [log]
     main_only = false
     time = false

   [misc]
     clean_on_exit = false

   [screen]
     clear_on_rebuild = false
     keep_scroll = true
   ```

3. **Add to Makefile**:
   ```makefile
   .PHONY: dev
   dev: ## Run with hot reload
   	@air
   ```

4. **Test Hot Reload**:
   ```bash
   make dev
   # Edit a .go file and watch it rebuild
   ```

#### 4.4 Development Scripts

**Status**: ✅ Completed

**Steps**:

1. **Create scripts directory**:
   ```bash
   mkdir -p scripts
   ```

2. **Create scripts/dev.sh**:
   ```bash
   #!/bin/bash
   # Development startup script

   set -e

   echo "🚀 Starting Matrix MUD Development Environment..."

   # Check dependencies
   if ! command -v go &> /dev/null; then
       echo "❌ Go not found. Install Go 1.21+"
       exit 1
   fi

   # Build
   echo "📦 Building..."
   make build

   # Start services in background
   echo "🌐 Starting web server..."
   ./bin/matrix-mud &
   MUD_PID=$!

   # Wait for startup
   sleep 2

   echo "✅ Matrix MUD is running!"
   echo "   Telnet: localhost:2323"
   echo "   Web:    localhost:8080"
   echo "   Admin:  localhost:9090"
   echo ""
   echo "Press Ctrl+C to stop"

   # Cleanup on exit
   trap "kill $MUD_PID" EXIT

   wait
   ```

   Make executable:
   ```bash
   chmod +x scripts/dev.sh
   ```

3. **Create scripts/load-test.sh**:
   ```bash
   #!/bin/bash
   # Simple load testing script

   NUM_PLAYERS=${1:-10}

   echo "🔥 Running load test with $NUM_PLAYERS concurrent players..."

   for i in $(seq 1 $NUM_PLAYERS); do
       (
           echo "player$i"
           echo "pass$i"
           sleep 1
           echo "look"
           sleep 1
           echo "north"
           sleep 1
           echo "quit"
       ) | telnet localhost 2323 > /dev/null 2>&1 &
   done

   wait
   echo "✅ Load test complete"
   ```

   Make executable:
   ```bash
   chmod +x scripts/load-test.sh
   ```

4. **Create scripts/backup-data.sh**:
   ```bash
   #!/bin/bash
   # Backup game data

   BACKUP_DIR="backups/$(date +%Y%m%d_%H%M%S)"

   echo "💾 Creating backup in $BACKUP_DIR..."

   mkdir -p "$BACKUP_DIR"
   cp -r data/* "$BACKUP_DIR/"

   echo "✅ Backup complete: $BACKUP_DIR"
   ```

   Make executable:
   ```bash
   chmod +x scripts/backup-data.sh
   ```

### Phase 5: Code Quality Improvements

#### 5.1 Complete Godoc Comments

**Status**: ✅ Completed (world.go, web.go, admin.go documented)

**Remaining Files**:

1. **world.go - Core Types**:

   Add to top of file:
   ```go
   // Package main implements the Matrix MUD game world simulation.
   // This file contains the core game state, player management, and world mechanics.
   ```

   Document World type:
   ```go
   // World represents the entire game state including all rooms, players, NPCs, and items.
   // It uses sync.RWMutex for concurrent access from multiple player goroutines.
   // The game loop calls World.Update() every 500ms to handle combat, NPC AI, and events.
   type World struct {
       Rooms   map[string]*Room         // All rooms indexed by ID
       Players map[*Client]*Player      // Active players indexed by connection
       mutex   sync.RWMutex            // Protects concurrent access to world state
   }
   ```

   Document Player type:
   ```go
   // Player represents a connected player character with stats, inventory, and position.
   // Player data is persisted to data/players/<name>.json on disconnect.
   type Player struct {
       Name      string    // Unique player name (case-insensitive)
       Class     string    // Character class (Hacker, Rebel, Operator)
       HP        int       // Current hit points
       MaxHP     int       // Maximum hit points
       Strength  int       // Attack damage modifier
       BaseAC    int       // Base armor class
       Level     int       // Character level
       XP        int       // Experience points
       Credits   int       // In-game currency
       RoomID    string    // Current room location
       Inventory []*Item   // Carried items
       Equipment map[string]*Item // Equipped items by slot
       Storage   []*Item   // Bank storage items
       Conn      *Client   // Network connection
   }
   ```

   Document Room type:
   ```go
   // Room represents a location in the game world with connections to other rooms.
   // Rooms contain NPCs, items, and can have special vendors or storage facilities.
   type Room struct {
       ID          string            // Unique room identifier
       Name        string            // Display name
       Description string            // Room description text
       Exits       map[string]string // Available exits (direction -> room ID)
       Items       []*Item           // Items on the ground
       NPCs        []*NPC            // Non-player characters in this room
       HasVendor   bool              // Whether room has a shop
       HasStorage  bool              // Whether room has bank storage
   }
   ```

   Document Item type:
   ```go
   // Item represents an object that can be picked up, equipped, or used.
   // Items have different slots (hand, body, head) and can provide AC bonuses or damage.
   type Item struct {
       ID          string // Unique item identifier
       Name        string // Display name
       Description string // Item description
       Slot        string // Equipment slot (hand, body, head, consumable)
       Damage      int    // Weapon damage bonus
       AC          int    // Armor class bonus
       Healing     int    // HP restoration amount (consumables)
       Value       int    // Shop price in credits
       Rarity      string // common, uncommon, rare, epic
   }
   ```

   Document NPC type:
   ```go
   // NPC represents a non-player character that can engage in combat or dialogue.
   // NPCs have AI behaviors, loot tables, and can respond to player speech.
   type NPC struct {
       ID          string   // Unique NPC identifier
       Name        string   // Display name
       Description string   // NPC description
       HP          int      // Current hit points
       MaxHP       int      // Maximum hit points
       Damage      int      // Attack damage
       AC          int      // Armor class
       XP          int      // XP reward on death
       Credits     int      // Credits dropped on death
       LootTable   []string // Possible item drops (item IDs)
       Hostile     bool     // Auto-attacks players
       Dialogue    map[string]string // Keyword -> response mapping
   }
   ```

2. **world.go - Key Functions**:

   ```go
   // NewWorld creates and initializes a new game world.
   // It loads world data from data/world.json or creates default rooms if missing.
   func NewWorld() *World { ... }

   // Update is called every game tick (500ms) to process combat, NPC AI, and events.
   // This runs in its own goroutine and uses the world mutex for safe concurrent access.
   func (w *World) Update() { ... }

   // LoadPlayer retrieves or creates a player from persistent storage.
   // Player data is loaded from data/players/<name>.json if it exists.
   func (w *World) LoadPlayer(name string, client *Client) *Player { ... }

   // SavePlayer persists player data to data/players/<name>.json.
   // This is called on disconnect and periodically during gameplay.
   func (w *World) SavePlayer(p *Player) error { ... }

   // MovePlayer attempts to move a player in the specified direction.
   // Returns a message describing the result (success, blocked, invalid exit).
   func (w *World) MovePlayer(p *Player, direction string) string { ... }

   // StartCombat initiates combat between a player and an NPC.
   // Combat runs automatically in the Update() loop until one side dies or flees.
   func (w *World) StartCombat(p *Player, target string) string { ... }

   // ResolveCombatRound processes one round of combat for all active combatants.
   // Calculates attack rolls, applies damage, checks for death, and awards XP.
   func (w *World) ResolveCombatRound() { ... }
   ```

3. **web.go**:

   Add to top:
   ```go
   // Package main implements the HTTP web server for Matrix MUD.
   // This file provides REST API endpoints and a web-based client interface.
   ```

   Document startWebServer:
   ```go
   // startWebServer initializes the HTTP server on port 8080.
   // Provides endpoints:
   //   GET  /          - Web client interface
   //   GET  /api/world - World state JSON
   //   GET  /api/stats - Server statistics
   //   POST /api/cmd   - Execute game command
   func startWebServer(world *World) { ... }
   ```

4. **admin.go**:

   Add to top:
   ```go
   // Package main implements the administrative server for Matrix MUD.
   // This file provides monitoring, management, and debugging endpoints.
   ```

   Document startAdminServer:
   ```go
   // startAdminServer initializes the admin HTTP server on port 9090.
   // Provides endpoints:
   //   GET /metrics     - Prometheus metrics
   //   GET /debug/pprof - Go profiling endpoints
   //   GET /players     - Active player list
   //   POST /broadcast  - Send message to all players
   func startAdminServer(world *World) { ... }
   ```

**Use golang-pro agent for help**:
```
Use golang-pro agent to add comprehensive godoc comments to world.go following Go documentation standards. Include package comment, type documentation, and function descriptions.
```

#### 5.2 Implement Proper Error Handling

**Status**: Not Started

**Critical Issue**: Many errors are currently ignored with `_, _ := ...`

**Steps**:

1. **Audit Ignored Errors**:
   ```bash
   grep -n "_, _" *.go
   ```

   Common patterns to fix:
   ```go
   // Bad
   file, _ := os.ReadFile("data/users.json")

   // Good
   file, err := os.ReadFile("data/users.json")
   if err != nil {
       log.Printf("Failed to read users: %v", err)
       return fmt.Errorf("loading users: %w", err)
   }
   ```

2. **Fix Authentication Errors** (main.go:22-23):
   ```go
   // Current (BAD)
   file, _ := os.ReadFile("data/users.json")
   json.Unmarshal(file, &users)

   // Fixed
   file, err := os.ReadFile("data/users.json")
   if err != nil {
       if os.IsNotExist(err) {
           // First run, no users yet
           return false
       }
       log.Printf("Error reading users.json: %v", err)
       c.Write(Red + "Authentication system error.\\r\\n" + Reset)
       return false
   }

   if err := json.Unmarshal(file, &users); err != nil {
       log.Printf("Error parsing users.json: %v", err)
       c.Write(Red + "Authentication system error.\\r\\n" + Reset)
       return false
   }
   ```

3. **Fix Player Class Selection** (main.go:59):
   ```go
   // Current (BAD)
   choice, _ := c.reader.ReadString('\n')

   // Fixed
   choice, err := c.reader.ReadString('\n')
   if err != nil {
       log.Printf("Error reading class choice: %v", err)
       return // Connection lost
   }
   ```

4. **Fix Connection Handler** (main.go:141, 173):
   ```go
   // Current (BAD)
   name, _ := client.reader.ReadString('\n')
   input, err := client.reader.ReadString('\n')

   // Fixed
   name, err := client.reader.ReadString('\n')
   if err != nil {
       log.Printf("Error reading name: %v", err)
       return
   }

   input, err := client.reader.ReadString('\n')
   if err != nil {
       if err != io.EOF {
           log.Printf("Read error for player %s: %v", player.Name, err)
       }
       break
   }
   ```

5. **Fix World Loading**:
   ```go
   // In NewWorld(), add error handling
   data, err := os.ReadFile("data/world.json")
   if err != nil {
       if os.IsNotExist(err) {
           log.Println("No world.json found, creating default world")
           return createDefaultWorld()
       }
       log.Fatalf("Failed to load world: %v", err)
   }

   if err := json.Unmarshal(data, &world); err != nil {
       log.Fatalf("Failed to parse world.json: %v", err)
   }
   ```

6. **Create Custom Error Types**:

   Add to world.go:
   ```go
   // Common errors
   var (
       ErrPlayerNotFound = errors.New("player not found")
       ErrItemNotFound   = errors.New("item not found")
       ErrNPCNotFound    = errors.New("NPC not found")
       ErrInvalidExit    = errors.New("no exit in that direction")
       ErrNotInCombat    = errors.New("not in combat")
       ErrAlreadyEquipped = errors.New("slot already occupied")
   )
   ```

**Use golang-pro agent**:
```
Use golang-pro agent to audit and fix all error handling in the codebase. Replace all ignored errors with proper error handling, add logging, and create custom error types where appropriate.
```

#### 5.3 Add Structured Logging

**Status**: Not Started

**Current Issue**: Using fmt.Printf and log.Printf inconsistently

**Steps**:

1. **Choose Logging Library**:

   Recommended: **zerolog** (zero-allocation, fast, structured)

   ```bash
   go get github.com/rs/zerolog/log
   ```

2. **Create Logger Package**:

   Create `pkg/logger/logger.go`:
   ```go
   package logger

   import (
       "os"
       "github.com/rs/zerolog"
       "github.com/rs/zerolog/log"
   )

   // Init initializes the global logger
   func Init(debug bool) {
       zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

       if debug {
           zerolog.SetGlobalLevel(zerolog.DebugLevel)
           log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
       } else {
           zerolog.SetGlobalLevel(zerolog.InfoLevel)
       }
   }

   // Player returns a logger with player context
   func Player(name string) zerolog.Logger {
       return log.With().Str("player", name).Logger()
   }

   // Combat returns a logger with combat context
   func Combat(attacker, defender string) zerolog.Logger {
       return log.With().
           Str("attacker", attacker).
           Str("defender", defender).
           Logger()
   }
   ```

3. **Update main.go**:
   ```go
   import (
       "github.com/rs/zerolog/log"
       "github.com/yourusername/matrix-mud/pkg/logger"
   )

   func main() {
       logger.Init(true) // debug mode

       log.Info().
           Int("port", 2323).
           Str("version", "v1.28").
           Msg("Matrix MUD server starting")

       // ... rest of main
   }
   ```

4. **Update Authentication**:
   ```go
   func authenticate(c *Client, name string) bool {
       logger.Player(name).Debug().Msg("Authentication attempt")

       // ... auth logic ...

       if pass == storedPass {
           logger.Player(name).Info().Msg("Authentication successful")
           return true
       }

       logger.Player(name).Warn().Msg("Authentication failed")
       return false
   }
   ```

5. **Update World Methods**:
   ```go
   func (w *World) MovePlayer(p *Player, direction string) string {
       log.Debug().
           Str("player", p.Name).
           Str("direction", direction).
           Str("from_room", p.RoomID).
           Msg("Player movement attempt")

       // ... movement logic ...
   }

   func (w *World) StartCombat(p *Player, target string) string {
       logger.Combat(p.Name, target).Info().Msg("Combat initiated")

       // ... combat logic ...
   }
   ```

6. **Add Metrics Logging**:
   ```go
   func (w *World) Update() {
       combats := 0
       for _, p := range w.Players {
           if p.InCombat {
               combats++
           }
       }

       if combats > 0 {
           log.Debug().
               Int("active_combats", combats).
               Int("total_players", len(w.Players)).
               Msg("World update tick")
       }
   }
   ```

**Use golang-pro agent**:
```
Use golang-pro agent to implement structured logging with zerolog throughout the codebase. Replace all fmt.Printf and log.Printf calls with structured logging including relevant context fields.
```

#### 5.4 Create Unit Tests

**Status**: Not Started

**Target**: 80%+ code coverage

**Steps**:

1. **Create Test Directory Structure**:
   ```bash
   mkdir -p tests/unit
   ```

2. **Create tests/unit/world_test.go**:
   ```go
   package unit

   import (
       "testing"

       // Import your packages once refactored
   )

   func TestNewWorld(t *testing.T) {
       world := NewWorld()

       if world == nil {
           t.Fatal("NewWorld returned nil")
       }

       if len(world.Rooms) == 0 {
           t.Error("World should have default rooms")
       }

       // Check for loading_program room
       if _, exists := world.Rooms["loading_program"]; !exists {
           t.Error("World missing loading_program room")
       }
   }

   func TestPlayerCreation(t *testing.T) {
       p := &Player{
           Name:  "TestPlayer",
           Class: "Hacker",
           HP:    15,
           MaxHP: 15,
       }

       if p.HP != p.MaxHP {
           t.Error("New player should start at max HP")
       }
   }

   func TestMovePlayer(t *testing.T) {
       world := NewWorld()
       player := &Player{
           Name:   "TestPlayer",
           RoomID: "loading_program",
       }

       // Test valid movement
       result := world.MovePlayer(player, "north")
       if player.RoomID == "loading_program" {
           t.Error("Player should have moved")
       }

       // Test invalid movement
       oldRoom := player.RoomID
       world.MovePlayer(player, "invalid")
       if player.RoomID != oldRoom {
           t.Error("Player should not move on invalid direction")
       }
   }
   ```

3. **Create tests/unit/combat_test.go**:
   ```go
   package unit

   import (
       "testing"
   )

   func TestCombatInitiation(t *testing.T) {
       world := NewWorld()
       player := &Player{
           Name:   "TestPlayer",
           HP:     20,
           Strength: 12,
       }

       // Add NPC to room
       room := world.Rooms["loading_program"]
       npc := &NPC{
           Name: "Agent",
           HP:   30,
           Damage: 5,
       }
       room.NPCs = append(room.NPCs, npc)

       // Start combat
       result := world.StartCombat(player, "agent")

       if !player.InCombat {
           t.Error("Player should be in combat")
       }

       if player.Target == nil {
           t.Error("Player should have a target")
       }
   }

   func TestCombatDamage(t *testing.T) {
       attacker := &Player{
           Name: "Attacker",
           Strength: 10,
       }

       defender := &NPC{
           Name: "Defender",
           HP:   50,
           AC:   10,
       }

       initialHP := defender.HP

       // Simulate attack (implement calculateDamage function)
       damage := calculateDamage(attacker.Strength, defender.AC)

       if damage < 0 {
           t.Error("Damage cannot be negative")
       }

       defender.HP -= damage

       if defender.HP >= initialHP {
           t.Error("Defender should take damage")
       }
   }
   ```

4. **Create tests/unit/inventory_test.go**:
   ```go
   package unit

   import (
       "testing"
   )

   func TestPickUpItem(t *testing.T) {
       world := NewWorld()
       player := &Player{
           Name:      "TestPlayer",
           RoomID:    "loading_program",
           Inventory: []*Item{},
       }

       // Add item to room
       room := world.Rooms[player.RoomID]
       item := &Item{
           ID:   "test_item",
           Name: "Test Item",
       }
       room.Items = append(room.Items, item)

       // Pick up item
       result := world.GetItem(player, "test item")

       if len(player.Inventory) == 0 {
           t.Error("Item should be in player inventory")
       }

       if len(room.Items) > 0 {
           t.Error("Item should be removed from room")
       }
   }

   func TestEquipItem(t *testing.T) {
       player := &Player{
           Name:      "TestPlayer",
           Equipment: make(map[string]*Item),
       }

       weapon := &Item{
           ID:     "sword",
           Name:   "Sword",
           Slot:   "hand",
           Damage: 10,
       }
       player.Inventory = append(player.Inventory, weapon)

       // Equip weapon
       world := NewWorld()
       world.WearItem(player, "sword")

       if player.Equipment["hand"] == nil {
           t.Error("Weapon should be equipped")
       }
   }
   ```

5. **Run Tests**:
   ```bash
   go test -v ./tests/unit/
   go test -race ./tests/unit/
   go test -cover ./tests/unit/
   ```

**Use test-engineer agent**:
```
Use test-engineer agent to create comprehensive unit tests for all game mechanics including world management, combat system, inventory management, and player actions. Target 80%+ code coverage.
```

#### 5.5 Create Integration Tests

**Status**: Not Started

**Steps**:

1. **Create tests/integration/server_test.go**:
   ```go
   package integration

   import (
       "bufio"
       "net"
       "strings"
       "testing"
       "time"
   )

   func TestServerConnection(t *testing.T) {
       // Start server in background
       go main() // Your actual main function

       time.Sleep(1 * time.Second) // Wait for startup

       // Connect
       conn, err := net.Dial("tcp", "localhost:2323")
       if err != nil {
           t.Fatalf("Failed to connect: %v", err)
       }
       defer conn.Close()

       reader := bufio.NewReader(conn)

       // Read welcome message
       line, _ := reader.ReadString('\n')
       if !strings.Contains(line, "Wake up") {
           t.Error("Expected welcome message")
       }
   }

   func TestPlayerAuthentication(t *testing.T) {
       conn, _ := net.Dial("tcp", "localhost:2323")
       defer conn.Close()

       reader := bufio.NewReader(conn)
       writer := bufio.NewWriter(conn)

       // Skip welcome
       reader.ReadString('\n')

       // Send username
       writer.WriteString("testplayer\n")
       writer.Flush()

       // Should ask for password (new user)
       line, _ := reader.ReadString('\n')
       if !strings.Contains(line, "password") {
           t.Error("Expected password prompt")
       }
   }

   func TestGameCommands(t *testing.T) {
       conn := connectAndAuth(t, "testuser", "testpass")
       defer conn.Close()

       reader := bufio.NewReader(conn)
       writer := bufio.NewWriter(conn)

       testCases := []struct {
           command  string
           expected string
       }{
           {"look", "loading_program"},
           {"inv", "Inventory"},
           {"score", "HP"},
       }

       for _, tc := range testCases {
           writer.WriteString(tc.command + "\n")
           writer.Flush()

           response, _ := reader.ReadString('>')
           if !strings.Contains(response, tc.expected) {
               t.Errorf("Command %s: expected %s in response", tc.command, tc.expected)
           }
       }
   }
   ```

2. **Create tests/integration/combat_integration_test.go**:
   ```go
   package integration

   import (
       "testing"
       "time"
   )

   func TestFullCombatFlow(t *testing.T) {
       conn := connectAndAuth(t, "combattest", "pass")
       defer conn.Close()

       reader := bufio.NewReader(conn)
       writer := bufio.NewWriter(conn)

       // Navigate to room with NPC
       sendCommand(writer, "north")

       // Initiate combat
       sendCommand(writer, "kill agent")
       time.Sleep(100 * time.Millisecond)

       // Check combat started
       response := readResponse(reader)
       if !strings.Contains(response, "combat") {
           t.Error("Combat should have started")
       }

       // Combat should resolve automatically
       time.Sleep(2 * time.Second)

       // Check combat ended
       sendCommand(writer, "look")
       response = readResponse(reader)
       // Either player or NPC should be dead
   }
   ```

3. **Run Integration Tests**:
   ```bash
   go test -v ./tests/integration/
   ```

**Use test-engineer agent**:
```
Use test-engineer agent to create integration tests that verify the full server workflow including connection, authentication, commands, and combat system.
```

#### 5.6 Security Hardening

**Status**: ✅ Completed - **CRITICAL SECURITY FIXES IMPLEMENTED**

**Completed**: Bcrypt password hashing, input validation, rate limiting, secure file permissions

**Steps**:

1. **Install bcrypt**:
   ```bash
   go get golang.org/x/crypto/bcrypt
   ```

2. **Update Authentication (main.go)**:
   ```go
   import "golang.org/x/crypto/bcrypt"

   func authenticate(c *Client, name string) bool {
       userMutex.Lock()
       defer userMutex.Unlock()

       users := make(map[string]string) // Now stores password hashes
       file, err := os.ReadFile("data/users.json")
       if err != nil && !os.IsNotExist(err) {
           log.Printf("Error reading users: %v", err)
           c.Write(Red + "Authentication error.\\r\\n" + Reset)
           return false
       }

       if file != nil {
           if err := json.Unmarshal(file, &users); err != nil {
               log.Printf("Error parsing users: %v", err)
               c.Write(Red + "Authentication error.\\r\\n" + Reset)
               return false
           }
       }

       cleanName := strings.ToLower(name)

       if storedHash, exists := users[cleanName]; exists {
           // Existing user - verify password
           c.Write("Password: ")
           pass, err := c.reader.ReadString('\n')
           if err != nil {
               return false
           }
           pass = strings.TrimSpace(pass)

           // Compare with bcrypt
           err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(pass))
           if err == nil {
               log.Printf("User %s authenticated successfully", cleanName)
               return true
           }

           c.Write(Red + "Access Denied.\\r\\n" + Reset)
           log.Printf("Failed auth attempt for user %s", cleanName)
           return false
       } else {
           // New user - create account
           c.Write("New identity detected. Set a password: ")
           pass, err := c.reader.ReadString('\n')
           if err != nil {
               return false
           }
           pass = strings.TrimSpace(pass)

           if len(pass) < 8 {
               c.Write("Password must be at least 8 characters.\\r\\n")
               return false
           }

           // Hash password with bcrypt
           hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
           if err != nil {
               log.Printf("Error hashing password: %v", err)
               c.Write(Red + "Error creating account.\\r\\n" + Reset)
               return false
           }

           users[cleanName] = string(hash)
           data, _ := json.MarshalIndent(users, "", "  ")
           if err := os.WriteFile("data/users.json", data, 0600); err != nil {
               log.Printf("Error saving user: %v", err)
               c.Write(Red + "Error creating account.\\r\\n" + Reset)
               return false
           }

           c.Write("Identity created.\\r\\n")
           log.Printf("New user created: %s", cleanName)
           return true
       }
   }
   ```

3. **Add Input Validation**:

   Create `pkg/validation/validation.go`:
   ```go
   package validation

   import (
       "regexp"
       "strings"
   )

   var (
       validUsername = regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)
       validCommand  = regexp.MustCompile(`^[a-z ]{1,100}$`)
   )

   // ValidateUsername checks if username is safe
   func ValidateUsername(name string) bool {
       return validUsername.MatchString(name)
   }

   // ValidateCommand checks if command is safe
   func ValidateCommand(cmd string) bool {
       return validCommand.MatchString(strings.ToLower(cmd))
   }

   // SanitizeInput removes potentially dangerous characters
   func SanitizeInput(input string) string {
       // Remove control characters except newline/tab
       cleaned := strings.Map(func(r rune) rune {
           if r < 32 && r != '\n' && r != '\t' {
               return -1
           }
           return r
       }, input)

       return strings.TrimSpace(cleaned)
   }
   ```

   Use in main.go:
   ```go
   import "github.com/yourusername/matrix-mud/pkg/validation"

   func handleConnection(conn net.Conn, world *World) {
       // ... existing code ...

       name, _ := client.reader.ReadString('\n')
       name = validation.SanitizeInput(name)

       if !validation.ValidateUsername(name) {
           client.Write("Invalid username. Use 3-20 alphanumeric characters.\\r\\n")
           return
       }

       // ... rest of handler ...
   }
   ```

4. **Add Rate Limiting**:

   Create `pkg/ratelimit/ratelimit.go`:
   ```go
   package ratelimit

   import (
       "sync"
       "time"
   )

   type RateLimiter struct {
       requests map[string][]time.Time
       mutex    sync.Mutex
       limit    int
       window   time.Duration
   }

   func New(limit int, window time.Duration) *RateLimiter {
       return &RateLimiter{
           requests: make(map[string][]time.Time),
           limit:    limit,
           window:   window,
       }
   }

   func (rl *RateLimiter) Allow(key string) bool {
       rl.mutex.Lock()
       defer rl.mutex.Unlock()

       now := time.Now()
       cutoff := now.Add(-rl.window)

       // Clean old requests
       requests := rl.requests[key]
       var recent []time.Time
       for _, t := range requests {
           if t.After(cutoff) {
               recent = append(recent, t)
           }
       }

       if len(recent) >= rl.limit {
           return false
       }

       rl.requests[key] = append(recent, now)
       return true
   }
   ```

   Use in main.go:
   ```go
   var authLimiter = ratelimit.New(5, 1*time.Minute) // 5 attempts per minute

   func authenticate(c *Client, name string) bool {
       if !authLimiter.Allow(name) {
           c.Write(Red + "Too many attempts. Try again later.\\r\\n" + Reset)
           time.Sleep(3 * time.Second)
           return false
       }

       // ... rest of auth ...
   }
   ```

5. **Add File Permission Checks**:
   ```go
   // In SavePlayer, SaveWorld, etc.
   if err := os.WriteFile(filename, data, 0600); err != nil { // Owner read/write only
       return err
   }
   ```

6. **Create Security Checklist**:

   Create `docs/SECURITY.md`:
   ```markdown
   # Security Guidelines

   ## Completed
   - [x] Bcrypt password hashing
   - [x] Input validation
   - [x] Rate limiting on authentication
   - [x] Secure file permissions (0600)

   ## TODO
   - [ ] Add TLS/SSL for telnet connections
   - [ ] Implement session tokens
   - [ ] Add CSRF protection for web interface
   - [ ] SQL injection protection (when DB added)
   - [ ] XSS protection in web client
   - [ ] Add admin authentication
   - [ ] Implement IP-based rate limiting
   - [ ] Add logging for security events
   - [ ] Regular dependency updates
   - [ ] Penetration testing
   ```

**Use security-engineer agent**:
```
Use security-engineer agent to implement comprehensive security hardening including bcrypt password hashing, input validation, rate limiting, and secure file permissions. Then perform a security audit to identify any remaining vulnerabilities.
```

### Phase 6: Final Polish

**Status**: In Progress

#### 6.1 Update CHANGELOG.md

**Status**: ✅ Completed

**Steps**:

1. **Review all changes made**:
   ```bash
   git log --oneline --since="2025-11-24"
   ```

2. **Update CHANGELOG.md**:
   ```markdown
   ## [1.1.0] - YYYY-MM-DD

   ### Added
   - Comprehensive godoc comments for all packages
   - Structured logging with zerolog
   - Unit test suite with 80%+ coverage
   - Integration tests for server and combat
   - Bcrypt password hashing for secure authentication
   - Input validation and sanitization
   - Rate limiting on authentication
   - Custom slash commands for Claude Code
   - MCP server for game inspection
   - Hot reload development environment
   - Development scripts (dev.sh, load-test.sh, backup-data.sh)

   ### Changed
   - Improved error handling throughout codebase
   - File permissions set to 0600 for sensitive data
   - Minimum password length increased to 8 characters

   ### Security
   - **CRITICAL**: Replaced plaintext password storage with bcrypt hashing
   - Added input validation to prevent injection attacks
   - Implemented rate limiting to prevent brute force
   - Secured file permissions on data files
   ```

#### 6.2 Update DEVLOG.md

**Steps**:

1. **Add comprehensive entry**:
   ```bash
   /devlog-entry
   ```

2. **Document**:
   - Security hardening implementation
   - Testing infrastructure creation
   - Performance improvements
   - Challenges overcome
   - Metrics (test coverage, lines of code, etc.)

#### 6.3 Verify All Documentation

**Checklist**:

```bash
# Check all doc links work
grep -r "](docs/" *.md

# Verify all slash commands exist
ls -la .claude/commands/

# Check MCP config
cat .claude/mcp.json

# Verify git hooks
ls -la .git/hooks/

# Test all Make targets
make help
```

#### 6.4 Final Commit

**Steps**:

1. **Review all changes**:
   ```bash
   git status
   git diff
   ```

2. **Create commit**:
   ```bash
   git add .
   git commit -m "feat: complete project setup with security, tests, and documentation

   - Add comprehensive godoc comments
   - Implement bcrypt password hashing (CRITICAL security fix)
   - Add structured logging with zerolog
   - Create unit and integration test suites (80%+ coverage)
   - Add input validation and rate limiting
   - Create custom Claude Code slash commands
   - Implement MCP inspector server
   - Add development scripts and hot reload
   - Update all documentation

   🤖 Generated with Claude Code

   Co-Authored-By: Claude <noreply@anthropic.com>"
   ```

3. **Optional: Create version tag**:
   ```bash
   git tag -a v1.1.0 -m "Version 1.1.0 - Security & Quality Release"
   git push origin v1.1.0
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

**Last Updated**: 2025-11-25
**Claude Code Version**: Latest
**Project Version**: v1.0.0 (v1.1.0 in development)
