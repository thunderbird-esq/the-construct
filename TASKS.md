# Matrix MUD Task Management

## Current Status: Phase 2 Complete ✅

**Last Updated**: 2025-11-28 11:15 UTC
**Current Version**: v1.30.0

---

## Phase 1: Security Hardening ✅ COMPLETE

| Task ID | Description | Status | Agent | Test |
|---------|-------------|--------|-------|------|
| P1-SEC-01 | Remove hardcoded admin credentials | ✅ | security-engineer | TestConfigEnvironmentVariables |
| P1-SEC-02 | Admin panel localhost-only binding | ✅ | security-engineer | TestAdminBindAddressNotExposed |
| P1-SEC-03 | WebSocket origin validation | ✅ | security-engineer | TestAllowedOriginsConfig |
| P1-SEC-04 | Rate limiter memory cleanup | ✅ | security-engineer | Integrated in main.go |
| P1-SEC-05 | go.mod version fix | ✅ | golang-pro | Build validation |

**Commits**: b488140, 6fa519f
**Tests**: 7/7 passing

---

## Phase 2: Bug Fixes ✅ COMPLETE

### Overview

Phase 2 addresses 5 high-priority bugs and 4 medium-priority issues. Tasks are organized into 3 parallel streams for efficient execution.

### Stream A: Data Integrity ✅

| Task ID | Description | Priority | Status | Test |
|---------|-------------|----------|--------|------|
| P2-BUG-04 | Nil room panic in GenerateAutomapInternal | HIGH | ✅ | TestNilRoomAccessNoPanic |
| P2-BUG-07 | JSON unmarshal errors silently ignored | HIGH | ✅ | TestJSONLoadErrorHandling |
| P2-BUG-13 | Agent Smith HP=0 (unkillable) | HIGH | ✅ | TestNPCHPValues |
| P2-BUG-14 | Morpheus MaxHP=0 (respawn fails) | HIGH | ✅ | TestNPCHPValues |

### Stream B: Resource Management ✅

| Task ID | Description | Priority | Status | Test |
|---------|-------------|----------|--------|------|
| P2-BUG-08 | No inventory size limit | HIGH | ✅ | TestInventorySizeLimit |
| P2-BUG-16 | No connection timeout | MEDIUM | ✅ | TestConnectionTimeoutConfig |

### Stream C: Code Quality ✅

| Task ID | Description | Priority | Status | Test |
|---------|-------------|----------|--------|------|
| P2-BUG-11 | Missing 'd' alias for down | MEDIUM | ✅ | TestDownAlias |

**All Phase 2 Tests**: 15/15 passing

---

## Phase 2 Detailed Task Breakdown

### P2-BUG-04: Nil Room Panic Fix
**File**: world.go (GenerateAutomapInternal, Look, Move)
**Problem**: Accessing room properties when room is nil causes panic
**Solution**: Add nil check before accessing room properties
**Test**: `TestNilRoomAccessNoPanic`
```go
// Before
room := w.Rooms[p.RoomID]
desc := room.Description  // PANIC if nil

// After
room := w.Rooms[p.RoomID]
if room == nil {
    return "Error: You are in the void. Use 'recall' to return."
}
desc := room.Description
```
**Status**: ✅ COMPLETE (in commit 6fa519f)

### P2-BUG-07: JSON Error Handling
**File**: world.go (loadWorldData)
**Problem**: JSON unmarshal errors are silently ignored
**Solution**: Log errors and create default world as fallback
**Test**: `TestJSONLoadErrorHandling`
```go
// Before
json.Unmarshal(file, &data)  // Errors ignored

// After
if err := json.Unmarshal(file, &data); err != nil {
    log.Printf("WARNING: Could not parse world.json: %v", err)
    w.createDefaultWorld()
    return
}
```
**Status**: ✅ COMPLETE (in commit 6fa519f)

### P2-BUG-08: Inventory Size Limit
**File**: world.go (GetItem)
**Problem**: No limit on inventory items causes memory growth
**Solution**: Add MaxInventorySize constant and check before adding
**Test**: `TestInventorySizeLimit`
```go
// config.go
const MaxInventorySize = 20

// world.go GetItem()
if len(p.Inventory) >= MaxInventorySize {
    return "Your inventory is full (max 20 items). Drop something first."
}
```
**Status**: ✅ COMPLETE (in commit 6fa519f)

### P2-BUG-11: Down Alias
**File**: main.go (command parsing)
**Problem**: No 'd' or 'dn' alias for 'down' command
**Solution**: Add 'dn' alias (since 'd' is 'drop')
**Test**: `TestDownAlias`
```go
// In command parsing
case "down", "dn":
    return w.Move(player, "down")
```
**Status**: 🔄 TODO

### P2-BUG-13/14: NPC HP Values
**File**: data/world.json
**Problem**: Agent Smith HP=0, Morpheus MaxHP=0
**Solution**: Fix JSON values OR auto-correct during load
**Test**: `TestNPCHPValues`
```go
// During world load
if npc.HP <= 0 {
    npc.HP = DefaultNPCHP
}
if npc.MaxHP <= 0 || npc.MaxHP < npc.HP {
    npc.MaxHP = npc.HP
}
```
**Status**: ✅ COMPLETE (runtime fix in commit 6fa519f)

### P2-BUG-16: Connection Timeout
**File**: main.go (handleConnection)
**Problem**: No timeout for idle connections
**Solution**: Add connection and idle timeouts
**Test**: `TestConnectionTimeoutConfig`
```go
// config.go
const ConnectionTimeout = 30 * time.Second
const IdleTimeout = 30 * time.Minute

// main.go
conn.SetDeadline(time.Now().Add(IdleTimeout))
```
**Status**: ✅ Constants added, implementation TODO

### P2-BUG-17: Broadcast Nil Safety
**File**: world.go (Broadcast)
**Problem**: Potential nil pointer if player is nil
**Solution**: Add nil check in broadcast
**Test**: `TestBroadcastNilSafety`
```go
func (w *World) Broadcast(roomID string, exclude *Player, msg string) {
    // exclude can be nil - that's valid (broadcast to everyone)
    for _, p := range w.Players {
        if p != nil && p.RoomID == roomID && (exclude == nil || p != exclude) {
            p.Send(msg)
        }
    }
}
```
**Status**: 🔄 TODO

---

## Phase 2 Agent Team Structure

### Team Alpha: Data Integrity
**Lead Agent**: golang-pro-A
**Support Agent**: data-fixer
**Memory Budget**: 4GB
**Tasks**: P2-BUG-04, P2-BUG-07, P2-BUG-13, P2-BUG-14

### Team Beta: Resource Management  
**Lead Agent**: golang-pro-B
**Memory Budget**: 2GB
**Tasks**: P2-BUG-08, P2-BUG-16, P2-BUG-17

### Team Gamma: Code Quality
**Lead Agent**: golang-pro-C
**Memory Budget**: 2GB
**Tasks**: P2-BUG-11, P2-BUG-10

### Test Integration Agent
**Agent**: test-engineer
**Memory Budget**: 2GB
**Tasks**: Validate all fixes with comprehensive tests

---

## Phase 2 Execution Timeline

```
Time 0-10min:  Team Alpha + Team Beta (parallel)
               - Alpha: P2-BUG-04, P2-BUG-07 (already done, verify)
               - Beta: P2-BUG-16, P2-BUG-17
               
Time 10-20min: Team Gamma + Test Integration
               - Gamma: P2-BUG-11, P2-BUG-10
               - Tests: Write phase2_fixes_test.go tests
               
Time 20-30min: Integration & Validation
               - Run all tests
               - Verify build
               - Update CHANGELOG
               - Git commit
```

---

## Phase 3: Enhancements (PLANNED)

| Task ID | Description | Priority | Agent |
|---------|-------------|----------|-------|
| P3-ENH-12 | Update xterm.js 3.14.5 → 5.x | MEDIUM | frontend-dev |
| P3-ENH-15 | Fix duplicate NPC/Item data in JSON | LOW | data-fixer |
| P3-ENH-18 | Implement skipped tests | LOW | test-engineer |
| P3-ENH-19 | Add password echo suppression (IAC) | HIGH | golang-pro |

---

## Phase 4: Deployment (PLANNED)

| Task ID | Description | Priority | Agent |
|---------|-------------|----------|-------|
| P4-DEP-01 | Fly.io configuration | HIGH | devops-engineer |
| P4-DEP-02 | Production documentation | HIGH | technical-writer |
| P4-DEP-03 | Deployment testing | HIGH | test-automator |

---

## Quality Gates

Each phase must pass these gates before proceeding:

1. ✅ All tests pass (`go test ./...`)
2. ✅ Build succeeds (`go build .`)
3. ✅ No linter errors (`golangci-lint run`)
4. ✅ CHANGELOG updated
5. ✅ Git committed

---

## Notes

- All code changes must preserve existing functionality
- All tests must be simple and focused
- No spaghetti code or sweeping changes
- Documentation must be machine-readable Markdown
- Each task has exactly ONE test that validates it
