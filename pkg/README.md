# Matrix MUD Packages

This directory contains all reusable packages for the Matrix MUD game server.

## Package Overview

| Package | Coverage | Purpose |
|---------|----------|---------|
| `achievements` | 95%+ | Achievement and title system |
| `admin` | 92.9% | Admin dashboard and management endpoints |
| `analytics` | 96.0% | Player behavior and game analytics tracking |
| `cooldown` | 88.9% | Ability and spell cooldown management |
| `crafting` | 92.2% | Item crafting system with recipes |
| `errors` | 100% | Custom error types and sentinel errors |
| `faction` | 95%+ | Faction system (Zion, Machines, Exiles) |
| `game` | 90.7% | Core game logic (combat, inventory, intro) |
| `help` | 100% | In-game help system with command documentation |
| `leaderboard` | 95%+ | Player rankings and statistics |
| `logging` | - | Structured logging wrapper (zerolog) |
| `metrics` | 100% | Prometheus-compatible metrics collection |
| `party` | 90.1% | Player party/group system |
| `quest` | 90%+ | Multi-stage quest system |
| `ratelimit` | - | Request rate limiting |
| `readline` | 90%+ | Terminal line editing with history |
| `session` | 89.5% | Persistent player session management |
| `training` | 95%+ | Training programs and PvP arenas |
| `validation` | - | Input validation utilities |
| `world` | 79.1% | World simulation (day/night cycle) |

## Package Details

### achievements
Tracks player milestones and unlockable titles. Categories include combat, exploration, social, and progression. Hidden achievements for special discoveries.

### admin
Admin dashboard for monitoring connected players, server stats, and management operations. HTTP Basic Auth protected.

### analytics
Tracks player events, session duration, commands used, and generates insights about game usage patterns.

### cooldown
Per-player, per-ability cooldown tracking with configurable durations for all class skills.

### crafting
Recipe-based crafting system. Loads recipes from `data/recipes.json`. Supports skill requirements and XP rewards.

### errors
Sentinel errors (ErrNotFound, ErrPermissionDenied, etc.) and GameError wrapper with operation context.

### faction
Three-faction system: Zion (resistance), Machines (AI), and Exiles (rogue programs). Reputation tracking with standing levels from Hated to Exalted. Faction choice affects gameplay and NPC interactions.

### game
Core game modules:
- `combat.go` - Combat calculations and damage
- `inventory.go` - Inventory management
- `intro.go` - Matrix rain intro animation
- `types.go` - Shared type definitions

### help
Comprehensive help system with 40+ commands documented. Supports aliases, categories, usage examples.

### leaderboard
Server-wide rankings for XP, kills, deaths, quests completed, money, PvP wins, and achievements. Supports top-N queries and individual rank lookups.

### logging
Structured logging wrapper using zerolog. Context-aware logging with connection and player info.

### metrics
Prometheus-compatible metrics for monitoring. Tracks commands, connections, errors, and performance.

### party
Group system allowing up to 6 players. Features include invites, kick, promote, disband. XP sharing with party bonuses.

### quest
Multi-stage quest system with objectives (kill, collect, deliver, visit, talk). Prerequisites, rewards, and repeatable quests supported.

### readline
Terminal line editing with command history support. Handles escape sequences for arrow keys, backspace, and line navigation.

### session
30-minute reconnection window for disconnected players. Preserves state including inventory and location.

### training
Instanced training programs for combat practice and PvP. Types include combat, survival, PvP arena, and timed trials. No death penalty in training. Challenge leaderboards with records.

### validation
Input sanitization and validation. Username validation, command filtering, and security helpers.

### world
Day/night cycle simulation with 8 time periods affecting NPC activity and light levels.

## Adding New Packages

1. Create directory under `pkg/`
2. Add Go files with package declaration
3. Add `*_test.go` for unit tests
4. Update this README with package info
5. Run `go test ./pkg/...` to verify
