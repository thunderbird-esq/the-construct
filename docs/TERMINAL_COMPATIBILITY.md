# Terminal Compatibility Design Document

## Goal
Allow ANY telnet-capable device from 1984 to 2025 to play Matrix MUD with the best possible experience for that device.

## Terminal Capability Levels

### Level 0: DUMB (Plain ASCII)
- No escape codes at all
- Pure 7-bit ASCII text
- 80 columns assumed (but graceful wrap)
- For: Truly ancient terminals, teletypes, accessibility screen readers

Example output:
```
The Construct - Loading Program
-------------------------------
A stark white room. Grid lines stretch to infinity.
Exits: north, east
Items: Nokia Phone
NPCs: The Merchant
> 
```

### Level 1: VT100 (Basic ANSI)
- Cursor positioning (ESC[H, ESC[row;colH)
- Clear screen (ESC[2J)
- Bold (ESC[1m), Reverse (ESC[7m), Reset (ESC[0m)
- NO colors
- For: Early 80s terminals, some vintage terminal emulators

Example output:
```
[BOLD]The Construct - Loading Program[RESET]

A stark white room. Grid lines stretch to infinity.

Exits: north, east
Items: [BOLD]Nokia Phone[RESET]
NPCs: [BOLD]The Merchant[RESET]
> 
```

### Level 2: ANSI Color (8 colors)
- Everything from VT100
- Basic 8 foreground colors (30-37)
- Basic 8 background colors (40-47)
- For: Late 80s terminals, Apple IIGS with ANSITerm, Mac Plus with color terminal

Example output:
```
[GREEN]The Construct - Loading Program[RESET]

A stark white room. Grid lines stretch to infinity.

Exits: [CYAN]north[RESET], [CYAN]east[RESET]
Items: [WHITE]Nokia Phone[RESET]
NPCs: [YELLOW]The Merchant[RESET]
> 
```

### Level 3: ANSI 256 / Full (Current implementation)
- Everything from ANSI Color
- Bright colors (90-97)
- 256 color support where available
- For: Modern terminals, xterm.js web client

## Detection Strategy

### Method 1: Telnet TTYPE Negotiation (RFC 1091)
```
Server: IAC DO TERMINAL-TYPE
Client: IAC WILL TERMINAL-TYPE
Server: IAC SB TERMINAL-TYPE SEND IAC SE
Client: IAC SB TERMINAL-TYPE IS "VT100" IAC SE
```

Known terminal types and their levels:
- "DUMB", "UNKNOWN" → Level 0
- "VT100", "VT102", "VT220" → Level 1
- "ANSI", "ANSI-BBS", "PC-ANSI" → Level 2
- "XTERM", "XTERM-256COLOR", "VT340" → Level 3

### Method 2: Terminal Response Query
Send: ESC[c (Device Attributes request)
If no response within 2 seconds → Level 0 or Level 1
If response → Parse for capabilities

### Method 3: User Selection (Fallback)
If negotiation fails, prompt:
```
Terminal type? [A]NSI color, [V]T100, [P]lain text, [D]etect: 
```

## Implementation Plan

1. Add TerminalLevel to Client struct
2. Implement TTYPE negotiation in handleConnection
3. Create color wrapper functions that respect level
4. Test with various terminal emulators

## Testing Without Vintage Hardware

### Option 1: Terminal Emulator Settings
Most modern terminal emulators can be set to VT100 mode:
- iTerm2: Profiles → Terminal → Terminal Type → VT100
- PuTTY: Terminal-type string: vt100

### Option 2: telnet with explicit TERM
```bash
TERM=vt100 telnet localhost 2323
TERM=dumb telnet localhost 2323
```

### Option 3: Create test mode
Add command-line flag: `--term-level=0|1|2|3`
Or in-game command: `/termtest 0`

### Option 4: Web-based VT100 emulator
- cool-retro-term: https://github.com/Swordfish90/cool-retro-term
- Online: https://www.masswerk.at/jsuix/ (Unix terminal in browser)

## ASCII Art Considerations

Current automap uses box-drawing characters which may not display on Level 0/1.
Need ASCII-safe alternatives:

Level 2-3 (Unicode box drawing):
```
┌───┬───┐
│ . │ @ │
├───┼───┤
│ . │ . │
└───┴───┘
```

Level 0-1 (Pure ASCII):
```
+---+---+
| . | @ |
+---+---+
| . | . |
+---+---+
```

## Network Considerations for Vintage Hardware

### Apple IIGS
- Uthernet II card with Marinetti TCP/IP stack
- Or: Serial connection to a Raspberry Pi running ser2net
- Terminal: ProTerm, Spectrum, ANSITerm

### Macintosh Plus
- MacTCP or Open Transport (System 7)
- Ethernet via SCSI-to-Ethernet adapter
- Or: Serial PPP to modern gateway
- Terminal: ZTerm, Microphone, NCSA Telnet

### Serial Gateway Option
For machines without TCP/IP:
```
[Vintage Mac] --serial--> [Raspberry Pi] --TCP/IP--> [MUD Server]
                          (running ser2net)
```

This is actually very common in the retro computing community.
