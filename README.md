# Matrix MUD

A Matrix-themed Multi-User Dungeon (MUD) game written in Go with real-time multiplayer support.

## Features

- **Multi-user gameplay** - Multiple players can connect and interact simultaneously
- **Class system** - Choose from Hacker, Rebel, or Operator classes
- **Real-time combat** - Engage in combat with NPCs and other players
- **Procedural generation** - Generate city grids and explore dynamic environments
- **Inventory & Equipment** - Collect items, weapons, and armor with rarity tiers
- **Quest system** - Complete quests for NPCs to gain XP and rewards
- **Banking system** - Store items in The Archive for safekeeping
- **Web interface** - Monitor the game world via HTTP
- **Admin console** - Manage the game server via dedicated admin port

## Installation

### Prerequisites

- Go 1.21 or higher
- Git

### Clone and Build

```bash
git clone https://github.com/yourusername/matrix-mud.git
cd matrix-mud
make install
make build
```

## Quick Start

### Start the server

```bash
make run
```

The server will start on:
- Telnet: `localhost:2323`
- Web interface: `localhost:8080`
- Admin console: `localhost:9090`

### Connect with a client

```bash
telnet localhost 2323
```

Or use any MUD client that supports telnet.

## Documentation

Comprehensive documentation is available to help you understand, develop, and extend Matrix MUD:

### Core Documentation

- **[CHANGELOG.md](CHANGELOG.md)** - Version history and notable changes
- **[DEVLOG.md](DEVLOG.md)** - Development journal with technical decisions and insights
- **[CONTRIBUTING.md](CONTRIBUTING.md)** - Development workflow and contribution guidelines

### Technical Documentation

- **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** - System design, components, and concurrency model
- **[docs/API.md](docs/API.md)** - Complete protocol reference and command documentation
- **[docs/DEVELOPMENT.md](docs/DEVELOPMENT.md)** - Development environment setup and debugging

### Claude Code Integration

- **[CLAUDE.md](CLAUDE.md)** - Claude Code integration guide with custom commands and MCP servers
- **[AGENTS.md](AGENTS.md)** - Multi-agent development patterns and workflows

### Quick Links

| Document | Purpose |
|----------|---------|
| [ARCHITECTURE.md](docs/ARCHITECTURE.md) | Understand system design |
| [API.md](docs/API.md) | Command reference |
| [DEVELOPMENT.md](docs/DEVELOPMENT.md) | Set up dev environment |
| [CLAUDE.md](CLAUDE.md) | Use Claude Code effectively |
| [AGENTS.md](AGENTS.md) | Multi-agent workflows |

## Project Structure

```
matrix-mud/
├── cmd/
│   └── matrix-mud/      # Main application entry point
├── pkg/
│   ├── game/           # Game logic and entities
│   ├── world/          # World management
│   ├── network/        # Network handling
│   └── auth/           # Authentication
├── internal/
│   └── config/         # Configuration management
├── data/
│   ├── world.json      # World data
│   ├── dialogue.json   # NPC dialogue
│   └── players/        # Player save files
├── tests/
│   ├── unit/           # Unit tests
│   └── integration/    # Integration tests
└── docs/               # Documentation

```

## Development

### Running tests

```bash
make test
```

### Linting

```bash
make lint
```

### Build for production

```bash
make build
```

### Run with Docker

```bash
docker build -t matrix-mud .
docker run -p 2323:2323 -p 8080:8080 matrix-mud
```

## Game Commands

### Movement
- `north`, `n` - Move north
- `south`, `s` - Move south
- `east`, `e` - Move east
- `west`, `w` - Move west
- `up`, `u` - Move up
- `down` - Move down

### Interaction
- `look [target]`, `l` - Look around or at a target
- `get [item]`, `g` - Pick up an item
- `drop [item]`, `d` - Drop an item
- `inv`, `i` - Show inventory
- `score`, `balance` - Show character stats

### Combat
- `kill [target]`, `attack [target]` - Attack an NPC
- `cast [skill] [target]` - Cast a skill
- `flee`, `stop` - Stop combat

### Items
- `wear/equip [item]` - Equip an item
- `remove [slot]` - Unequip an item
- `use/eat [item]` - Use a consumable

### Social
- `say [message]` - Say something in the room
- `gossip [message]` - Send a global message
- `tell [player] [message]` - Send a private message
- `who` - List online players

### Trading
- `list` - View merchant inventory
- `buy [item]` - Buy from merchant
- `sell [item]` - Sell to merchant
- `give [item] [target]` - Give item to NPC

### Banking
- `deposit [item]` - Store item in The Archive
- `withdraw [item]` - Retrieve item from The Archive
- `storage` - View stored items

### Builder Commands
- `generate city [rows] [cols]` - Generate a city grid
- `dig [direction] [name]` - Create a new room
- `create [item|npc] [id]` - Spawn an entity
- `delete [target]` - Remove an entity
- `edit desc [text]` - Edit room description
- `save world` - Save world to disk

## Classes

### Hacker
- **HP**: 15
- **Strength**: 10
- **Starting Item**: Cyberdeck
- **Special Skill**: Glitch (MP-based tech attack)

### Rebel
- **HP**: 30
- **Strength**: 14
- **Starting Item**: Combat Boots
- **Special Skill**: Smash (Physical power attack)

### Operator
- **HP**: 20
- **Strength**: 12
- **Starting Item**: Pilot Shades
- **Special Skill**: Patch (Healing ability)

## API Endpoints

### Web Interface
- `GET /` - View the world map
- `GET /map` - Get JSON map data

### Admin Console
- Accessible via telnet on port 9090
- Real-time server statistics and management

## Configuration

Configuration can be managed through environment variables:

```bash
TELNET_PORT=2323
WEB_PORT=8080
ADMIN_PORT=9090
DATA_DIR=./data
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines.

## License

MIT License - See LICENSE file for details

## Acknowledgments

Inspired by the Matrix franchise and classic MUD games.
