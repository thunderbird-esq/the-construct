# Matrix MUD - API Reference

Complete protocol documentation and command reference for Matrix MUD.

---

## Table of Contents

1. [Telnet Protocol](#telnet-protocol)
2. [Command Reference](#command-reference)
3. [HTTP API](#http-api)
4. [Admin Protocol](#admin-protocol)
5. [Data Formats](#data-formats)

---

## Telnet Protocol

### Connection

**Host**: `localhost` (or server IP)
**Port**: `2323`
**Protocol**: Raw TCP
**Encoding**: UTF-8

**Connect**:
```bash
telnet localhost 2323
```

### Authentication Flow

```
Server: Wake up...
Server: Identify yourself:
Client: alice
Server: Password:
Client: secret123
Server: [Authentication success or failure]
```

**New User**:
```
Server: New identity detected. Set a password:
Client: mysecurepassword
Server: Identity created.
```

### Command Format

```
command [argument]
```

**Examples**:
```
look
north
get phone
kill cop
say hello world
```

---

## Command Reference

### Movement Commands

#### `north`, `n`
Move north

**Syntax**: `north` or `n`
**Response**: "You move to {room_id}." or "No exit."

#### `south`, `s`
Move south

**Syntax**: `south` or `s`

#### `east`, `e`
Move east

#### `west`, `w`
Move west

#### `up`, `u`
Move up

#### `down`
Move down

### Observation Commands

#### `look [target]`, `l [target]`
Look around or examine target

**Syntax**:
- `look` - View current room
- `look cop` - Examine NPC
- `look phone` - Examine item

**Response**:
```
   @
  . .
 . . .

*** room_id ***
Description of the room.
Exits: [north] [south] [east] [west]
Visible Items: Nokia Phone Cyberdeck
Who is here: Riot Cop [HOSTILE] Merchant [MERCHANT]
Players: Bob Alice
```

### Inventory Commands

#### `get [item]`, `g [item]`
Pick up an item

**Syntax**: `get phone`
**Response**: "Got Nokia Phone." or "Not here."

#### `drop [item]`, `d [item]`
Drop an item

**Syntax**: `drop phone`
**Response**: "Dropped Nokia Phone." or "Don't have."

#### `inv`, `i`
Show inventory and equipment

**Response**:
```
HP: 20/20  MP: 10/10  STR: 12  AC: 12
[ EQUIPMENT ]
  Hand: Cyberdeck
  Body: <empty>
  Head: Pilot Shades
[ BACKPACK ]
  - Nokia Phone
  - Red Pill
```

### Equipment Commands

#### `wear [item]`, `wield [item]`, `equip [item]`
Equip an item

**Syntax**: `wear coat`
**Response**: "Equipped Leather Trenchcoat."

#### `remove [slot]`, `unequip [slot]`
Unequip item from slot

**Syntax**: `remove body`
**Slots**: `hand`, `body`, `head`

### Combat Commands

#### `kill [target]`, `k [target]`, `attack [target]`, `a [target]`
Initiate combat

**Syntax**: `kill cop`
**Response**: "Engaging Riot Cop!"

**Combat Round Output**:
```
You hit Riot Cop with fists for 3 damage!
Riot Cop hits you for 4 damage!
```

#### `cast [skill] [target]`, `c [skill] [target]`
Use class skill

**Skills by Class**:
- **Hacker**: `glitch` - Tech damage (5-15 dmg, 5 MP)
- **Rebel**: `smash` - Physical strike (8+STR dmg, 5 MP)
- **Operator**: `patch` - Heal self (10 HP, 5 MP)

**Syntax**:
- `cast glitch cop`
- `cast patch` (self-target)

#### `flee`, `stop`
Disengage from combat

**Syntax**: `flee`
**Response**: "Stopped."

### Item Usage

#### `use [item]`, `eat [item]`, `take [item]`
Use a consumable item

**Syntax**: `use red pill`
**Response**: "Swallowed Red Pill. Str +1!"

### Trading Commands

#### `list`, `vendor`
View merchant inventory

**Syntax**: `list`
**Response**:
```
Merchant offers:
 - Nokia Phone        : 10 Fragments
 - Leather Trenchcoat : 100 Fragments
 - Training Katana    : 50 Fragments
```

#### `buy [item]`
Purchase from merchant

**Syntax**: `buy phone`
**Response**: "Bought Nokia Phone." or "Not enough Fragments."

#### `sell [item]`
Sell to merchant

**Syntax**: `sell phone`
**Response**: "Sold Nokia Phone for 5." (50% of price)

#### `give [item] [target]`
Give item to NPC (quests)

**Syntax**: `give phone morpheus`
**Response**: "You give Nokia Phone to Morpheus.\n[Quest reward message]\n(Gained 100 XP)"

### Banking Commands

#### `deposit [item]`
Store item in The Archive

**Syntax**: `deposit katana`
**Requirements**: Must be in `construct_archive` room
**Response**: "You upload Training Katana to the Archive."

#### `withdraw [item]`
Retrieve item from The Archive

**Syntax**: `withdraw katana`
**Requirements**: Must be in `construct_archive` room

#### `storage`, `bank`
View stored items

**Response**:
```
[ ARCHIVE STORAGE ]
 - Training Katana
 - Leather Trenchcoat
```

### Social Commands

#### `say [message]`
Speak in current room

**Syntax**: `say hello everyone`
**Output** (to others): "Alice says: \"hello everyone\""

**NPC Interaction**: NPCs may respond to keywords

#### `gossip [message]`, `chat [message]`
Global chat

**Syntax**: `gossip looking for group`
**Output** (to all): "[GLOBAL] Alice: looking for group"

#### `tell [player] [message]`, `whisper [player] [message]`, `t [player] [message]`
Private message

**Syntax**: `tell bob meet me at nexus`
**Output** (to Bob): "Alice tells you: meet me at nexus"

#### `who`
List online players

**Response**:
```
Connected Signals:
- Alice [construct_nexus]
- Bob [loading_program]
```

### Information Commands

#### `score`, `sc`, `balance`, `bal`
View character stats

**Response**:
```
=== Alice ===
Class: Hacker
Level: 3
XP:    1500 / 3000
Fragments: 250
HP:    20 / 30
MP:    15 / 15
STR:   11
```

#### `help`
Show command list

### Builder Commands

**Note**: These commands allow world modification.

#### `generate city [rows] [cols]`
Generate procedural city grid

**Syntax**: `generate city 5 5`
**Response**: "Generated 5x5 City Grid."

**Generated**:
- Grid of connected rooms
- Random NPCs (10% hostile)
- Random items (30% trash)

#### `dig [direction] [room_name]`
Create new connected room

**Syntax**: `dig north Secret Hideout`
**Response**: "Created room 'Secret Hideout'."

#### `create [item|npc] [id]`
Spawn entity from template

**Syntax**:
- `create item phone`
- `create npc cop`

#### `delete [target]`, `del [target]`
Remove entity from room

**Syntax**: `delete cop`

#### `edit desc [text]`
Edit room description

**Syntax**: `edit desc A dark alley filled with fog.`

#### `save world`
Save world to disk

**Syntax**: `save world`
**Response**: "World saved to disk."

### Utility Commands

#### `teleport [room_id]`
Teleport to room

**Syntax**: `teleport construct_nexus`

#### `quit`
Disconnect and save

---

## HTTP API

### Endpoints

#### `GET /`
View world map (HTML)

**Response**: HTML page with visual map

#### `GET /map`
Get world data (JSON)

**Response**:
```json
{
  "rooms": {
    "construct_nexus": {
      "id": "construct_nexus",
      "description": "The Construct. A white void.",
      "exits": {"north": "city_street"},
      "symbol": "@",
      "color": "white"
    }
  },
  "players": [
    {
      "name": "Alice",
      "room_id": "construct_nexus",
      "level": 3,
      "hp": 20,
      "max_hp": 30
    }
  ]
}
```

---

## Admin Protocol

### Connection

**Port**: `9090`
**Protocol**: Raw TCP

**Connect**:
```bash
telnet localhost 9090
```

### Admin Commands

(To be implemented - currently placeholder)

Planned commands:
- `stats` - Server statistics
- `players` - List all players with details
- `kick [player]` - Disconnect player
- `broadcast [message]` - Server-wide message
- `shutdown` - Graceful server shutdown

---

## Data Formats

### Player Save Format

**File**: `data/players/{name}.json`

```json
{
  "Name": "Alice",
  "RoomID": "construct_nexus",
  "Inventory": [
    {
      "ID": "phone",
      "Name": "Nokia Phone",
      "Description": "An old school slider phone.",
      "Damage": 1,
      "Slot": "hand",
      "Price": 10
    }
  ],
  "Equipment": {
    "hand": {
      "ID": "deck",
      "Name": "Cyberdeck",
      "Damage": 2
    }
  },
  "Bank": [],
  "HP": 20,
  "MaxHP": 30,
  "MP": 15,
  "MaxMP": 15,
  "Strength": 11,
  "BaseAC": 10,
  "XP": 1500,
  "Level": 3,
  "Class": "Hacker",
  "Money": 250
}
```

### World Format

**File**: `data/world.json`

```json
{
  "rooms": {
    "construct_nexus": {
      "ID": "construct_nexus",
      "Description": "The Construct. A white void.",
      "Exits": {
        "north": "city_street"
      },
      "Symbol": "@",
      "Color": "white",
      "Items": [],
      "NPCs": [
        {
          "ID": "morpheus",
          "Name": "Morpheus",
          "Description": "The guide.",
          "HP": 100,
          "MaxHP": 100,
          "State": "IDLE",
          "Quest": {
            "wanted_item": "red_pill",
            "reward_xp": 500,
            "reward_msg": "You have chosen wisely."
          }
        }
      ]
    }
  }
}
```

### Item Rarity System

```
Rarity Levels:
0 = Common     (White text)
1 = Uncommon   (Bright Green) - +1 stats, 2x price
2 = Rare       (Cyan)         - +2 stats, 5x price
3 = Legendary  (Magenta)      - +4 stats, 10x price

Drop Rates:
Common:    75%
Uncommon:  15%
Rare:      9%
Legendary: 1%
```

---

## Error Codes

### Authentication Errors

```
"Identification required."     - No username provided
"Access Denied."              - Wrong password
"Password too short."         - Password < 3 characters
```

### Command Errors

```
"Unknown."                    - Invalid command
"Not here."                   - Target not found
"You don't have that."        - Item not in inventory
"Can't use."                  - Item not consumable
"Not enough MP."              - Insufficient mana
"No merchant."                - No vendor in room
"Protected."                  - Can't attack vendor
```

---

**Last Updated**: 2025-11-24
**API Version**: v1.0
**Protocol**: Telnet/HTTP
