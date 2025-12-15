# Matrix MUD - System Architecture

Comprehensive architectural documentation for the Matrix MUD game server.

---

## Table of Contents

1. [Overview](#overview)
2. [System Components](#system-components)
3. [Data Architecture](#data-architecture)
4. [Concurrency Model](#concurrency-model)
5. [Network Architecture](#network-architecture)
6. [Game Loop Design](#game-loop-design)
7. [Scalability Considerations](#scalability-considerations)
8. [Future Architecture](#future-architecture)

---

## Overview

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Client Layer                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ Telnet   │  │ Web      │  │ Admin    │  │ Future   │   │
│  │ Client   │  │ Browser  │  │ Console  │  │ WebSocket│   │
│  └─────┬────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘   │
└────────┼────────────┼─────────────┼──────────────┼─────────┘
         │            │             │              │
         │ :2323      │ :8080       │ :9090        │ :WS Port
         │ TCP        │ HTTP        │ TCP          │ WebSocket
┌────────▼────────────▼─────────────▼──────────────▼─────────┐
│                    Server Layer                             │
│  ┌───────────────────────────────────────────────────────┐ │
│  │               Main Server Process                     │ │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐           │ │
│  │  │ Telnet   │  │ Web      │  │ Admin    │           │ │
│  │  │ Listener │  │ Server   │  │ Server   │           │ │
│  │  └────┬─────┘  └────┬─────┘  └────┬─────┘           │ │
│  │       │             │             │                  │ │
│  │       └─────────────┼─────────────┘                  │ │
│  │                     │                                │ │
│  │            ┌────────▼────────┐                       │ │
│  │            │  World Manager  │                       │ │
│  │            │   (Shared State)│                       │ │
│  │            └────────┬────────┘                       │ │
│  │                     │                                │ │
│  │       ┌─────────────┼─────────────┐                 │ │
│  │       │             │             │                 │ │
│  │  ┌────▼────┐  ┌────▼────┐  ┌────▼────┐            │ │
│  │  │ Game    │  │ Combat  │  │ Player  │            │ │
│  │  │ Loop    │  │ System  │  │ Manager │            │ │
│  │  │ (500ms) │  │         │  │         │            │ │
│  │  └─────────┘  └─────────┘  └─────────┘            │ │
│  └───────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
         │
         │ File I/O
┌────────▼────────────────────────────────────────────────────┐
│                    Data Layer                               │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ world    │  │ players/ │  │ users    │  │ dialogue │   │
│  │ .json    │  │ *.json   │  │ .json    │  │ .json    │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### Design Principles

1. **Simplicity First**: Start simple, add complexity as needed
2. **Concurrency by Default**: Leverage Go's goroutines for natural concurrency
3. **Shared State with Locks**: Central world state protected by RWMutex
4. **Event-Driven Updates**: 500ms tick for world updates and combat
5. **Stateless Protocol**: Commands are stateless (except player session)

---

## System Components

### Core Components

#### 1. Main Server (`main.go`)

**Responsibilities**:
- Initialize world state
- Start network listeners (Telnet, HTTP, Admin)
- Launch game loop
- Handle graceful shutdown

**Key Functions**:
```go
func main()
  - listener := net.Listen("tcp", ":2323")     // Telnet
  - go startWebServer(world)                    // HTTP :8080
  - go startAdminServer(world)                  // Admin :9090
  - go gameLoop(world)                          // 500ms tick
  - for { handleConnection(conn, world) }      // Accept loop
```

**Lifecycle**:
```
[Startup]
   │
   ├──> Load World Data (world.json)
   ├──> Load Dialogue (dialogue.json)
   ├──> Initialize Item Templates
   │
   ├──> Start Web Server (goroutine)
   ├──> Start Admin Server (goroutine)
   ├──> Start Game Loop (goroutine)
   │
   └──> Listen for Connections (main thread)
        │
        └──> For each connection:
             └──> spawn handleConnection(goroutine)
```

#### 2. World Manager (`world.go`)

**Data Structure**:
```go
type World struct {
    Rooms         map[string]*Room            // All rooms indexed by ID
    Players       map[*Client]*Player         // Active players
    Dialogue      map[string]map[string]string // NPC dialogue
    DeadNPCs      []*NPC                      // Respawn queue
    ItemTemplates map[string]*Item            // Item templates
    mutex         sync.RWMutex                // Concurrency control
}
```

**Responsibilities**:
- Maintain global game state
- Coordinate player actions
- Handle world updates (game loop)
- Manage NPC respawns
- Persist player data

**Concurrency Pattern**:
```go
// Write operations (modify state)
world.mutex.Lock()
defer world.mutex.Unlock()
// ... modify world state ...

// Read operations (query state)
world.mutex.RLock()
defer world.mutex.RUnlock()
// ... read world state ...
```

#### 3. Connection Handler (`main.go:handleConnection`)

**Per-Client Goroutine**:

```
[New Connection]
   │
   ├──> Authenticate User
   │    ├──> Check users.json
   │    ├──> Verify password (bcrypt hash comparison)
   │    └──> Create account if new
   │
   ├──> Load Player Data
   │    ├──> Read data/players/{name}.json
   │    └──> Initialize new player if needed
   │
   ├──> Choose Class (if new)
   │    └──> Hacker / Rebel / Operator
   │
   ├──> Add to World.Players
   │
   └──> Command Loop
        ├──> Read input
        ├──> Parse command
        ├──> Execute action
        ├──> Send response
        └──> Repeat until disconnect
                │
                └──> On Disconnect:
                     ├──> Save player data
                     └──> Remove from World.Players
```

**Command Processing**:
```go
input := "kill cop"
parts := strings.Fields(input)
cmd := parts[0]           // "kill"
arg := parts[1:]          // "cop"

switch cmd {
    case "kill":
        response = world.StartCombat(player, arg)
    // ... other commands ...
}
```

#### 4. Game Loop

**Update Cycle** (500ms):

```go
go func() {
    ticker := time.NewTicker(500 * time.Millisecond)
    for range ticker.C {
        world.Update()
    }
}()
```

**Update Operations**:
```
Every 500ms:
   │
   ├──> NPC Respawn Check
   │    └──> If 30s elapsed: respawn NPC to original room
   │
   ├──> For each Player:
   │    │
   │    ├──> Aggro Check (if IDLE)
   │    │    └──> If aggro NPC present: start combat
   │    │
   │    ├──> MP Regeneration (16.6% chance)
   │    │    └──> If MP < MaxMP: MP++
   │    │
   │    └──> Combat Resolution (if in COMBAT)
   │         └──> If 1.5s elapsed: resolve combat round
   │
   └──> Next tick
```

#### 5. Combat System

**State Machine**:

```
Player States:
   IDLE ──┬──> COMBAT ──> IDLE
          │       ▲         │
          │       │         │
          └───────┴─────────┘
           (aggro NPC)  (flee/victory/death)
```

**Combat Resolution** (per round):

```
[Combat Round]
   │
   ├──> Player Attack
   │    ├──> Roll d20 + modifiers
   │    ├──> Check vs NPC AC
   │    └──> Apply damage if hit
   │
   ├──> Check NPC Death
   │    └──> If HP <= 0:
   │         ├──> Award XP
   │         ├──> Drop loot (GenerateLoot)
   │         ├──> Check level up
   │         └──> End combat
   │
   ├──> NPC Attack
   │    ├──> Roll d20 + modifiers
   │    ├──> Check vs Player AC
   │    └──> Apply damage if hit
   │
   └──> Check Player Death
        └──> If HP <= 0:
             ├──> Respawn at loading_program
             ├──> Restore HP
             └──> End combat
```

**Damage Calculation**:
```
Player Damage = 1 + (Strength - 10) / 2 + Weapon.Damage
NPC Damage = random(1, NPC.Damage)
Hit Check = d20 >= Target.AC
```

---

## Data Architecture

### Data Model

```
World
 ├── Rooms (map[string]*Room)
 │    ├── Room
 │    │    ├── ID: string
 │    │    ├── Description: string
 │    │    ├── Exits: map[string]string
 │    │    ├── Items: []*Item
 │    │    ├── NPCs: []*NPC
 │    │    ├── Symbol: string (map display)
 │    │    └── Color: string (map display)
 │    └── ...
 │
 ├── Players (map[*Client]*Player)
 │    ├── Player
 │    │    ├── Name: string
 │    │    ├── RoomID: string
 │    │    ├── Conn: *Client
 │    │    ├── Inventory: []*Item
 │    │    ├── Equipment: map[string]*Item
 │    │    ├── Bank: []*Item
 │    │    ├── HP, MaxHP: int
 │    │    ├── MP, MaxMP: int
 │    │    ├── Strength, BaseAC: int
 │    │    ├── State: string (IDLE/COMBAT)
 │    │    ├── Target: string (NPC ID in combat)
 │    │    ├── XP, Level, Money: int
 │    │    └── Class: string
 │    └── ...
 │
 ├── Item Templates (map[string]*Item)
 │    └── Item
 │         ├── ID: string
 │         ├── Name: string
 │         ├── Description: string
 │         ├── Damage, AC: int
 │         ├── Slot: string (hand/body/head)
 │         ├── Type: string (consumable/equipment)
 │         ├── Effect: string
 │         ├── Value, Price: int
 │         └── Rarity: int (0-3)
 │
 └── NPC Templates
      └── NPC
           ├── ID: string
           ├── Name: string
           ├── HP, MaxHP: int
           ├── Damage, AC: int
           ├── State: string (IDLE/COMBAT)
           ├── Loot: []string (item template IDs)
           ├── XP, DropMoney: int
           ├── Vendor: bool
           ├── Inventory: []string (for vendors)
           ├── Aggro: bool
           └── Quest: Quest
```

### Persistence

**Files**:
```
data/
├── world.json          # Room definitions, NPCs, items
├── dialogue.json       # NPC dialogue trees
├── users.json          # Authentication (username -> password)
└── players/
    ├── alice.json      # Individual player saves
    ├── bob.json
    └── ...
```

**Save Strategy**:
- **World**: Manual save with "save world" command
- **Players**: Automatic save on disconnect + periodic backup
- **Users**: Save immediately on account creation

**Limitations**:
- No ACID guarantees
- Concurrent write risks
- File corruption possible
- No transaction support

**Migration Path**: PostgreSQL with proper schema

---

## Concurrency Model

### Goroutine Architecture

```
Main Goroutine
   ├──> Game Loop Goroutine (1)
   │    └──> Updates world every 500ms
   │
   ├──> Web Server Goroutine (1)
   │    └──> Handles HTTP requests
   │
   ├──> Admin Server Goroutine (1)
   │    └──> Handles admin connections
   │
   └──> Client Goroutines (N)
        └──> One per connected player
```

**Total Goroutines**: 3 + N (N = number of connected players)

### Synchronization

**Shared Resource**: `World` struct

**Protection Mechanism**: `sync.RWMutex`

**Lock Hierarchy**:
```
World.mutex (global)
  └──> No nested locks (avoid deadlock)
```

**Lock Usage Pattern**:
```go
// Many readers allowed concurrently
func (w *World) Look(p *Player) string {
    w.mutex.RLock()
    defer w.mutex.RUnlock()
    // Read world state...
}

// Only one writer at a time
func (w *World) MovePlayer(p *Player, dir string) {
    w.mutex.Lock()
    defer w.mutex.Unlock()
    // Modify world state...
}
```

### Potential Race Conditions

**1. Player State Updates**

**Issue**: Player struct modified from multiple goroutines
- Game loop modifies HP (regen)
- Client goroutine modifies inventory (commands)
- Combat system modifies HP/state

**Current**: Protected by World.mutex (all access serialized)

**Future**: Consider per-player mutex for finer granularity

**2. NPC Respawn**

**Issue**: NPC added to room while player interacting

**Current**: Protected by World.mutex

**Race Example**:
```go
// Goroutine 1 (Game Loop)
npc.HP = npc.MaxHP
room.NPCMap[npc.ID] = npc  // Respawn

// Goroutine 2 (Player)
targetNPC := room.NPCMap[targetID]  // May be nil or just respawned
```

**Mitigation**: All access under world.mutex

**3. File I/O**

**Issue**: Multiple players disconnecting simultaneously

**Current**: No file locking ⚠️

**Risk**: Corrupted player saves

**TODO**: Implement file locking or channel-based serialization

### Performance Considerations

**Bottlenecks**:
1. **Global mutex**: All world operations serialized
2. **File I/O**: Blocks during saves
3. **JSON parsing**: Slow for large datasets

**Optimization Strategies**:
1. **Read/Write separation**: Use RLock for reads (allows parallel)
2. **Async saves**: Queue save requests, process in background
3. **Caching**: Cache frequently accessed data
4. **Lock-free structures**: Consider for hot paths

---

## Network Architecture

### Protocol Stack

```
Application Layer:  [Game Commands]
                           │
Presentation Layer: [Text/ANSI Colors]
                           │
Session Layer:      [Player Session]
                           │
Transport Layer:    [TCP / HTTP / WebSocket]
                           │
Network Layer:      [IP]
```

### Telnet Protocol

**Port**: 2323
**Transport**: Raw TCP
**Encoding**: UTF-8 text + ANSI escape codes

**Connection Flow**:
```
Client                          Server
  │                               │
  ├──── TCP Connect ─────────────>│
  │<──── "Wake up..." ────────────┤
  │                               │
  ├──── "Alice\n" ───────────────>│ (username)
  │<──── "Password:" ─────────────┤
  │                               │
  ├──── "secret\n" ──────────────>│ (password)
  │<──── (auth success) ──────────┤
  │                               │
  ├──── "look\n" ────────────────>│ (command)
  │<──── (room description) ──────┤
  │                               │
  └──── Commands... ──────────────>
```

**ANSI Color Codes**:
```
Reset: \033[0m
Green: \033[32m  (default text)
White: \033[97m  (highlights)
Red:   \033[31m  (danger)
Yellow:\033[33m  (merchants)
```

### HTTP Protocol

**Port**: 8080
**Endpoints**:
- `GET /`: World map visualization (HTML)
- `GET /map`: JSON map data

**Response Format** (JSON):
```json
{
  "rooms": [...],
  "players": [...]
}
```

### Admin Protocol

**Port**: 9090
**Transport**: Raw TCP
**Purpose**: Server administration and monitoring

---

## Game Loop Design

### Tick System

**Fixed Timestep**: 500ms (2 updates per second)

**Advantages**:
- Predictable behavior
- Easy debugging
- Low CPU usage

**Disadvantages**:
- Limited responsiveness
- Not suitable for action games

**Implementation**:
```go
ticker := time.NewTicker(500 * time.Millisecond)
for range ticker.C {
    world.Update()
}
```

### Update Phases

```
Tick N (time = t)
   │
   ├──> Phase 1: NPC Respawn (O(dead NPCs))
   │    └──> Check each dead NPC
   │         └──> If 30s elapsed: respawn
   │
   ├──> Phase 2: Player Updates (O(players))
   │    └──> For each player:
   │         ├──> Aggro check
   │         ├──> MP regen
   │         └──> Combat resolution
   │
   └──> Complete (t + delta)

Tick N+1 (time = t + 500ms)
   └──> Repeat...
```

### Time Complexity

Per tick:
- NPC respawn: O(D) where D = dead NPCs
- Player updates: O(P) where P = players
- Combat resolution: O(C) where C = players in combat
- **Total**: O(D + P + C) ≈ O(P) typically

**Performance**:
- 100 players: ~0.1ms per tick
- 1000 players: ~1ms per tick
- 10000 players: ~10ms per tick (20% of tick budget)

---

## Scalability Considerations

### Current Limits

**Estimated Capacity** (single server):
- Concurrent players: 1,000 - 5,000
- World size: 1,000 - 10,000 rooms
- Active NPCs: 500 - 1,000

**Bottlenecks**:
1. Global mutex (lock contention)
2. File I/O (player saves)
3. Memory (all players in RAM)
4. Network (TCP connections)

### Horizontal Scaling Strategy

**Approach**: Shard by zone

```
         ┌─────────────┐
         │   Gateway   │
         │   (Router)  │
         └──────┬──────┘
                │
       ┌────────┼────────┐
       │        │        │
  ┌────▼───┐ ┌─▼────┐ ┌─▼────┐
  │ Zone 1 │ │Zone 2│ │Zone 3│
  │ Server │ │Server│ │Server│
  └────────┘ └──────┘ └──────┘
    City      Wasteland  Dungeon
```

**Characteristics**:
- Players in different zones on different servers
- Zone transitions require server migration
- Shared database for persistence
- Message passing for cross-zone chat

### Vertical Scaling

**Optimizations**:
1. **Database**: PostgreSQL instead of JSON files
2. **Caching**: Redis for session data
3. **Connection pooling**: Reuse DB connections
4. **Async I/O**: Non-blocking file operations
5. **Profiling**: Identify and optimize hot paths

---

## Future Architecture

### Planned Improvements

#### 1. Database Backend

```
┌─────────┐     ┌──────────┐     ┌─────────┐
│  Game   │────>│PostgreSQL│<────│  Web    │
│ Server  │     │ Database │     │  API    │
└─────────┘     └──────────┘     └─────────┘
```

**Schema**:
- `players`: Player data
- `rooms`: World structure
- `items`: Item instances
- `npcs`: NPC instances

#### 2. Microservices

```
┌──────────┐  ┌──────────┐  ┌──────────┐
│  Game    │  │  Auth    │  │  Chat    │
│ Service  │  │ Service  │  │ Service  │
└────┬─────┘  └────┬─────┘  └────┬─────┘
     │             │              │
     └─────────────┼──────────────┘
                   │
            ┌──────▼──────┐
            │   Message   │
            │    Queue    │
            │  (RabbitMQ) │
            └─────────────┘
```

#### 3. Event-Driven Architecture

**Event Bus**:
```go
type Event struct {
    Type    string
    PlayerID string
    Data    interface{}
}

// Publish
bus.Publish(Event{Type: "player.move", PlayerID: "alice", Data: newRoom})

// Subscribe
bus.Subscribe("player.move", func(e Event) {
    // Handle move event...
})
```

---

**Last Updated**: 2025-11-24
**Architecture Version**: v1.0
**Status**: Production-Ready (with noted limitations)
