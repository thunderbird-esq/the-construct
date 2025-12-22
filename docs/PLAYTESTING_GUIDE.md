# Matrix MUD v1.90.0 - Comprehensive Playtesting Guide

A step-by-step guide to testing all systems implemented in Options A-D.

---

## Table of Contents

1. [Connection Methods](#1-connection-methods)
2. [Character Creation & Core Gameplay](#2-character-creation--core-gameplay)
3. [Option A: Content Expansion Testing](#3-option-a-content-expansion-testing)
4. [Option B: Multiplayer Enhancement Testing](#4-option-b-multiplayer-enhancement-testing)
5. [Option C: Technical Infrastructure Testing](#5-option-c-technical-infrastructure-testing)
6. [Option D: Player Experience Testing](#6-option-d-player-experience-testing)
7. [Admin Panel Access & Testing](#7-admin-panel-access--testing)
8. [Testing Checklist](#8-testing-checklist)

---

## 1. Connection Methods

### Option A: Web Interface (Recommended for Testing)

Open your browser and navigate to:
```
https://matrix-mud.fly.dev/
```

This provides a full terminal experience with:
- CRT monitor visual effect
- Matrix "digital rain" animation
- Full keyboard support
- Mobile-friendly touch controls

### Option B: Telnet (Classic MUD Experience)

```bash
# macOS/Linux
telnet matrix-mud.fly.dev 2323

# Or using netcat
nc matrix-mud.fly.dev 2323

# Windows (PowerShell)
telnet matrix-mud.fly.dev 2323
```

### Option C: Local Development

```bash
cd /Users/edsaga/claudes-world/the-construct-new
go run .

# Then in another terminal:
telnet localhost 2323
# Or open http://localhost:8080
```

---

## 2. Character Creation & Core Gameplay

### Step 2.1: Initial Connection

When you connect, you'll see:
```
Wake up, Neo...
The Matrix has you...

Enter your name (or 'new' for a new character):
```

### Step 2.2: Create a New Character

1. Type `new` and press Enter
2. Enter a username (e.g., `TestPlayer`)
3. Enter a password (twice for confirmation)
4. Choose a class:
   - **Hacker** - Digital warrior, high damage
   - **Runner** - Speed specialist, balanced
   - **Operator** - Support class, utility

### Step 2.3: Test Core Commands

Once logged in, test these basic commands:

```
look                    # View your surroundings
l                       # Shortcut for look
score                   # View your stats (HP, XP, level, money)
sc                      # Shortcut for score
inventory               # View your items
i                       # Shortcut
who                     # See online players
help                    # View command list
```

### Step 2.4: Test Movement

```
north                   # Move north (or n)
south                   # Move south (or s)
east                    # Move east (or e)
west                    # Move west (or w)
up                      # Move up (or u)
down                    # Move down (or d)
recall                  # Teleport to Dojo (safe room)
```

**Suggested Exploration Route:**
1. Start at Dojo
2. Go `east` to Hallway
3. Go `north` to Loading Program
4. Explore the world, noting room descriptions and exits

### Step 2.5: Test Item Interaction

```
look                    # See items in the room
get <item>              # Pick up an item (e.g., get phone)
drop <item>             # Drop an item
equip <item>            # Equip weapon/armor
remove <item>           # Unequip
use <item>              # Use consumable (e.g., use red_pill)
```

### Step 2.6: Test Combat

Find an NPC and engage:

```
look                    # Find NPCs listed in room
attack <npc>            # Start combat (e.g., attack agent)
kill <npc>              # Same as attack
cast <skill>            # Use class skill (e.g., cast glitch)
flee                    # Attempt to escape combat
```

**Combat Skills by Class:**
- Hacker: `cast glitch`, `cast patch`, `cast decrypt`
- Runner: `cast strike`, `cast evade`, `cast sprint`
- Operator: `cast scan`, `cast boost`, `cast disrupt`

---

## 3. Option A: Content Expansion Testing

### 3.1: Test Dialogue System (pkg/dialogue)

```
# Find an NPC with dialogue
look                    # Locate NPCs
talk morpheus           # Start dialogue with Morpheus
talk oracle             # Talk to The Oracle
talk merovingian        # Talk to The Merovingian

# During dialogue:
1                       # Select option 1
2                       # Select option 2
bye                     # End dialogue
```

**Expected Behavior:**
- Dialogue nodes display with numbered choices
- Selecting numbers advances conversation
- NPCs have unique personality in responses
- Some dialogue branches based on quest progress

### 3.2: Test Instance/Dungeon System (pkg/instance)

```
instance list           # View available dungeons
instance create gov_raid    # Create Government Building instance
instance enter <id>     # Enter the instance
instance info           # View instance status
instance leave          # Exit the instance

# Inside instance:
look                    # See instance-specific rooms
attack <enemy>          # Fight instance mobs
# Clear all enemies to unlock doors
# Defeat boss to complete
```

**Available Instances:**
1. `gov_raid` - Government Building Raid (Difficulty 1, Level 3+)
2. `club_depths` - Club Hel Depths (Difficulty 2, Level 5+)
3. `agent_stronghold` - Agent Stronghold (Difficulty 3, Level 8+)

### 3.3: Test Expanded Crafting

```
recipes                 # View all known recipes
recipes weapons         # View weapon recipes
recipes armor           # View armor recipes
recipes consumables     # View consumable recipes

craft <recipe>          # Craft an item (e.g., craft health_vial)
```

**Test Recipes:**
- `health_vial` - Basic healing item
- `focus_serum` - Damage buff
- `overclocked_pistol` - Upgraded weapon
- `reinforced_coat` - Upgraded armor

---

## 4. Option B: Multiplayer Enhancement Testing

> **Note:** Many multiplayer features require 2+ players. Use multiple terminal windows or have a friend connect.

### 4.1: Test Chat System (pkg/chat)

```
# Global chat
gossip Hello everyone!  # Send to all players
chat Hello!             # Same as gossip
; Hello!                # Shortcut (semicolon prefix)

# Private messaging
tell <player> <msg>     # Private message
whisper <player> <msg>  # Same as tell
t <player> <msg>        # Shortcut

# Room chat
say Hello!              # Say to room only
' Hello!                # Shortcut (apostrophe prefix)

# Channel commands (if implemented in game)
/channels               # List channels
/join trade             # Join trade channel
/leave trade            # Leave channel
/g <message>            # Global channel
/t <message>            # Trade channel
```

### 4.2: Test Party System

```
# Creating parties
party create            # Create a new party
party                   # View party status
invite <player>         # Invite player to party
accept                  # Accept invitation
decline                 # Decline invitation

# Party management
party kick <player>     # Remove from party
party promote <player>  # Make party leader
party leave             # Leave party
party disband           # Disband (leader only)

# Party communication
party say <message>     # Talk to party only
p <message>             # Shortcut
```

### 4.3: Test PvP Arena (pkg/pvp)

```
arena                   # View arena status
arena queue             # Join matchmaking queue
arena queue duel        # Queue for 1v1
arena queue team        # Queue for team battle
arena leave             # Leave queue
arena status            # Check queue status

# During arena match:
attack <player>         # Attack opponent
cast <skill>            # Use abilities
# Match ends when one side is defeated

# Rankings
pvp rank                # View your PvP rating
pvp stats               # View win/loss record
leaderboard pvp         # View PvP leaderboard
```

**Arena Types:**
- Duel (1v1)
- Team 2v2
- Team 3v3
- Free-for-All (4 players)

### 4.4: Test Trading System (pkg/trade)

```
# Player-to-player trading
trade <player>          # Initiate trade
trade accept            # Accept trade request
trade decline           # Decline trade

# During trade:
trade add <item>        # Add item to offer
trade addmoney <amount> # Add money to offer
trade remove <item>     # Remove item from offer
trade confirm           # Confirm your side
trade cancel            # Cancel trade
trade status            # View current trade

# Auction house (at The Archive)
auction list            # View all listings
auction search <term>   # Search items
auction sell <item> <price> [buyout]  # List item
auction bid <id> <amount>  # Place bid
auction buyout <id>     # Buy immediately
auction mylistings      # View your listings
auction mybids          # View your bids
```

---

## 5. Option C: Technical Infrastructure Testing

### 5.1: Test REST API (pkg/api)

The API requires an API key. For local testing:

```bash
# Health check (no auth required)
curl https://matrix-mud.fly.dev/api/health

# Status (no auth required)
curl https://matrix-mud.fly.dev/api/status

# Authenticated endpoints require X-API-Key header
# Note: API keys must be configured in the server
```

**API Endpoints:**
```
GET  /api/health              # Health check
GET  /api/status              # Server status
GET  /api/players             # List online players
GET  /api/players/:name       # Player details
GET  /api/players/:name/stats # Player statistics
GET  /api/world/rooms         # List rooms
GET  /api/world/rooms/:id     # Room details
GET  /api/world/npcs          # List NPCs
GET  /api/world/items         # Item templates
GET  /api/leaderboards        # Leaderboard categories
GET  /api/leaderboards/:cat   # Specific leaderboard
POST /api/messages            # Send message to player
```

### 5.2: Test Event System (pkg/events)

Events are internal but can be verified by:
1. Observing Discord webhook notifications (if configured)
2. Checking server logs for event emissions
3. Actions triggering achievement unlocks

**Events Emitted:**
- Player join/leave
- Level up
- NPC kill
- Quest complete
- Achievement unlock
- PvP results

### 5.3: Test Database (pkg/db)

The database layer is implemented but may not be fully integrated with the live save system yet. Verify by:

1. Create a character
2. Make progress (level up, get items)
3. Quit with `quit` command
4. Reconnect and verify progress saved

---

## 6. Option D: Player Experience Testing

### 6.1: Test Tutorial System (pkg/tutorial)

```
tutorial                # View tutorial status
tutorial list           # List available tutorials
tutorial start new_player  # Start new player tutorial
tutorial hint           # Get hint for current step
tutorial skip           # Skip current step
```

**New Player Tutorial Steps:**
1. "Wake Up, Neo" - Type `look`
2. "Learn to Move" - Move in any direction
3. "Check Your Inventory" - Type `inventory`
4. "Pick Up Items" - Pick up any item
5. "Equip Your Gear" - Equip a weapon or armor
6. "Enter Combat" - Attack any NPC
7. "Talk to NPCs" - Talk to any NPC
8. "Getting Help" - Type `help`

**Advanced Tutorials (after completing new_player):**
- `combat_mastery` - Combat skills
- `crafting_101` - Crafting system
- `faction_guide` - Faction system
- `party_play` - Party mechanics

### 6.2: Test Enhanced Help System (pkg/help)

```
# Basic help
help                    # Command list
help <command>          # Specific command help
help look               # Example

# Search functionality
help search combat      # Search for combat commands
help search move        # Search for movement commands

# Manual topics
help topic basics       # Basic guide
help topic combat       # Combat guide
help topic crafting     # Crafting guide
help topic quests       # Quest guide
help topic factions     # Faction guide
help topic classes      # Class comparison
help topic shortcuts    # Command shortcuts

# Typo correction (test with intentional typos)
hlep                    # Should suggest "help"
lok                     # Should suggest "look"
atack                   # Should suggest "attack"
```

### 6.3: Test Accessibility Features (pkg/accessibility)

```
# Toggle features (these are in-game settings)
accessibility           # View current settings
accessibility screenreader on   # Enable screen reader mode
accessibility highcontrast on   # Enable high contrast
accessibility largetext on      # Enable large text
accessibility reducedmotion on  # Disable animations
accessibility colorblind protanopia  # Colorblind mode

# Other colorblind modes:
# - deuteranopia (green-blind)
# - tritanopia (blue-blind)
# - none (default)

# Font scale
accessibility fontscale 1.5     # 150% font size
```

**Screen Reader Mode Tests:**
1. Enable screen reader mode
2. Verify ANSI codes are stripped
3. Box-drawing characters replaced with text
4. Semantic annotations added (e.g., [EXITS], [STATUS])

---

## 7. Admin Panel Access & Testing

### 7.1: Accessing the Admin Panel

**On Production (Fly.io):**

The admin panel is bound to `127.0.0.1:9090` for security, meaning it's only accessible from within the container. To access it:

```bash
# SSH into the Fly.io machine
flyctl ssh console -a matrix-mud

# From inside the container, use curl
curl http://127.0.0.1:9090/
```

**For Local Development:**

```bash
# Start the server locally
cd /Users/edsaga/claudes-world/the-construct-new
go run .

# Access admin panel in browser
open http://127.0.0.1:9090/
```

### 7.2: Admin Credentials

**Default Credentials:**
- Username: `admin`
- Password: Auto-generated on startup (check server logs)

**Setting Custom Credentials:**
```bash
export ADMIN_USER="your_username"
export ADMIN_PASS="your_secure_password"
```

### 7.3: Admin Panel Features

**Dashboard (`/`):**
- View all connected players
- See player names, current rooms, HP status
- Real-time connection monitoring

**Kick Player (`/kick?name=<playername>`):**
- Forcibly disconnect a player
- Use for moderation

### 7.4: Testing Admin Functions

1. **Connect as a player** in one terminal/browser
2. **Access admin panel** in another
3. **Verify player appears** in dashboard
4. **Test kick function** - click EJECT button
5. **Verify player disconnected** with message

---

## 8. Testing Checklist

### Core Systems
- [ ] Web connection works
- [ ] Telnet connection works
- [ ] Character creation succeeds
- [ ] Login with existing character works
- [ ] Movement in all directions
- [ ] Look command shows room details
- [ ] Inventory management (get, drop, equip)
- [ ] Combat initiates and resolves
- [ ] Class skills work
- [ ] Death and respawn work
- [ ] Save/load persists data

### Option A: Content
- [ ] Dialogue with Morpheus works
- [ ] Dialogue with Oracle works
- [ ] Multiple dialogue choices function
- [ ] Instance creation works
- [ ] Instance rooms are separate
- [ ] Instance enemies spawn
- [ ] Instance completion rewards
- [ ] Crafting recipes display
- [ ] Crafting with materials works
- [ ] Crafted items function correctly

### Option B: Multiplayer
- [ ] Global chat reaches all players
- [ ] Private tell works
- [ ] Room say is local only
- [ ] Party creation works
- [ ] Party invites send/receive
- [ ] Party chat is private
- [ ] PvP queue functions
- [ ] Arena matches complete
- [ ] Trade initiation works
- [ ] Trade completion exchanges items
- [ ] Auction listing works
- [ ] Auction bidding works

### Option C: Infrastructure
- [ ] API health endpoint responds
- [ ] API status shows server info
- [ ] Events emit on actions (check logs)
- [ ] Player data persists across sessions

### Option D: Experience
- [ ] Tutorial auto-starts for new players
- [ ] Tutorial steps complete correctly
- [ ] Tutorial rewards granted
- [ ] Help search finds commands
- [ ] Help suggestions for typos work
- [ ] Manual topics display
- [ ] Screen reader mode strips formatting
- [ ] High contrast changes colors
- [ ] Colorblind modes apply

### Admin Panel
- [ ] Dashboard loads
- [ ] Connected players displayed
- [ ] Kick function works
- [ ] Basic auth required

---

## Quick Reference Card

```
MOVEMENT        COMBAT          ITEMS           SOCIAL
─────────────   ──────────────  ──────────────  ──────────────
n/s/e/w/u/d     attack <npc>    get <item>      gossip <msg>
look            kill <npc>      drop <item>     tell <p> <msg>
recall          flee            equip <item>    say <msg>
                cast <skill>    use <item>      party create

INFO            QUESTS          TUTORIAL        HELP
─────────────   ──────────────  ──────────────  ──────────────
score           quest           tutorial list   help
inventory       quest log       tutorial hint   help <cmd>
who             quest hint      tutorial skip   help search <t>
achievements                                    help topic <t>
```

---

## Reporting Issues

If you find bugs during testing:

1. **Note the exact steps** to reproduce
2. **Copy any error messages** displayed
3. **Check server logs** if accessible
4. **Document expected vs actual** behavior

Good luck testing! 🎮
