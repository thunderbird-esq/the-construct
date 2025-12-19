// Package help provides the in-game help system for Matrix MUD.
// Includes context-sensitive help, command suggestions, and searchable manual.
package help

import (
	"sort"
	"strings"
)

// Entry represents a single help entry
type Entry struct {
	Command     string
	Aliases     []string
	Description string
	Usage       string
	Examples    []string
	Category    string
	Related     []string // Related commands
}

// Categories of commands
const (
	CatMovement    = "Movement"
	CatCombat      = "Combat"
	CatItems       = "Items"
	CatSocial      = "Social"
	CatInfo        = "Information"
	CatEconomy     = "Economy"
	CatBuilder     = "Builder"
	CatSystem      = "System"
	CatParty       = "Party"
	CatQuest       = "Quest"
	CatFaction     = "Faction"
	CatTraining    = "Training"
	CatAchievement = "Achievement"
	CatChat        = "Chat"
	CatPvP         = "PvP"
	CatTrade       = "Trade"
	CatTutorial    = "Tutorial"
)

// Manual topics for detailed guides
var ManualTopics = map[string]string{
	"basics": `
MATRIX MUD - BASIC GUIDE
=========================

Welcome to the Matrix! You have been freed from the simulation and must
now learn to survive in this digital world.

MOVING AROUND
  Use directional commands: north (n), south (s), east (e), west (w),
  up (u), down (d). The room description shows available exits.

LOOKING
  Type 'look' to see your surroundings. 'look <target>' examines
  something specific like an NPC or item.

INTERACTING
  'get <item>' picks up items. 'drop <item>' drops them.
  'inv' shows your inventory. 'score' shows your stats.

COMBAT
  'attack <target>' starts combat. 'flee' attempts to escape.
  'cast <skill>' uses special abilities in combat.

HELP
  'help <command>' for details on any command.
  'help search <term>' to search for commands.
`,

	"combat": `
COMBAT GUIDE
============

STARTING COMBAT
  Use 'attack <target>' or 'kill <target>' to engage an enemy.
  Combat is turn-based with automatic attacks each round.

COMBAT ACTIONS
  attack/kill - Start combat or continue attacking
  flee/stop - Attempt to escape (may fail)
  cast <skill> - Use a class ability
  use <item> - Use a consumable item

SKILLS
  Each class has unique skills:
  - Hacker: glitch (damage), patch (heal), decrypt (buff)
  - Runner: strike (damage), evade (defense), sprint (escape)
  - Operator: scan (info), boost (buff), disrupt (debuff)

DEATH & RESPAWN
  If you die, you respawn at the Dojo with reduced HP/MP.
  Your items remain with you.

TIPS
  - Check enemy levels before engaging
  - Keep healing items ready
  - Learn your class skills early
  - Party up for tough enemies
`,

	"crafting": `
CRAFTING GUIDE
==============

OVERVIEW
  Crafting lets you create items from components. You can find
  components on defeated enemies or scattered in the world.

COMMANDS
  recipes - List all known recipes
  recipe <name> - Show recipe details
  craft <name> - Craft an item if you have materials

CRAFTING STATIONS
  Some recipes require specific locations:
  - Workbench: basic items, tools
  - Forge: weapons, armor
  - Lab: consumables, tech items

TIPS
  - Gather components as you explore
  - Check vendor stocks for rare materials
  - Higher-tier recipes unlock at higher levels
  - Some recipes are discovered through quests
`,

	"quests": `
QUEST GUIDE
===========

OVERVIEW
  Quests provide structure to your adventure, offering rewards
  and advancing the story.

COMMANDS
  quest - View active quests
  quest log - Detailed quest journal
  quest hint - Get a hint for current objective
  quest abandon <name> - Give up on a quest

QUEST TYPES
  - Main Story: Follow the Awakening Protocol
  - Side Quests: Optional adventures for rewards
  - Daily: Repeatable challenges
  - Faction: Improve faction standing

TIPS
  - Read quest text for clues
  - Talk to NPCs for guidance
  - Some quests have multiple outcomes
  - Quest rewards scale with level
`,

	"factions": `
FACTION GUIDE
=============

THREE FACTIONS
  Zion (Resistance): Fight to free humanity
  Machines: Maintain order in the Matrix
  Exiles: Seek power and independence

REPUTATION
  Actions affect faction standing:
  - Killing faction NPCs: Major reputation loss
  - Helping faction members: Reputation gain
  - Quest choices: Faction alignment

BENEFITS
  High reputation unlocks:
  - Faction-specific vendors
  - Special quests
  - Unique titles
  - Safe zones

COMMANDS
  faction - View current allegiance
  faction list - See all factions
  faction join <name> - Join a faction
  rep - Check reputation levels
`,

	"classes": `
CLASS GUIDE
===========

HACKER
  The digital warrior. Manipulates code to damage enemies and
  support allies.
  - High damage, medium defense
  - Skills: Glitch, Patch, Decrypt, Virus

RUNNER
  Speed and agility specialist. Excels at combat and escape.
  - Balanced offense and defense
  - Skills: Strike, Evade, Sprint, Flurry

OPERATOR
  Support and intelligence. Buffs allies and debuffs enemies.
  - Low damage, high utility
  - Skills: Scan, Boost, Disrupt, Override

TIPS
  - Each class has a unique playstyle
  - Skills improve as you level up
  - Classes can party together for synergy
`,

	"shortcuts": `
COMMAND SHORTCUTS
=================

MOVEMENT
  n = north    s = south    e = east    w = west
  u = up       d = down

COMMON
  l = look     i = inv      k = kill    a = attack
  g = get      sc = score   t = tell

COMBAT
  c = cast     fl = flee

SOCIAL
  ' = say      ; = gossip

CHAT (if enabled)
  /g = global  /t = trade   /h = help channel

TOGGLE
  Use '!' before any command to repeat it
  Example: !n moves north repeatedly
`,
}

// Entries contains all help entries indexed by command name
var Entries = map[string]*Entry{
	"look": {
		Command:     "look",
		Aliases:     []string{"l"},
		Description: "Look at your surroundings, an item, or an NPC.",
		Usage:       "look [target]",
		Examples:    []string{"look", "look morpheus", "look katana"},
		Category:    CatInfo,
		Related:     []string{"examine", "inventory"},
	},
	"north": {
		Command:     "north",
		Aliases:     []string{"n"},
		Description: "Move north.",
		Usage:       "north",
		Examples:    []string{"north", "n"},
		Category:    CatMovement,
		Related:     []string{"south", "east", "west"},
	},
	"south": {
		Command:     "south",
		Aliases:     []string{"s"},
		Description: "Move south.",
		Usage:       "south",
		Examples:    []string{"south", "s"},
		Category:    CatMovement,
		Related:     []string{"north", "east", "west"},
	},
	"east": {
		Command:     "east",
		Aliases:     []string{"e"},
		Description: "Move east.",
		Usage:       "east",
		Examples:    []string{"east", "e"},
		Category:    CatMovement,
		Related:     []string{"north", "south", "west"},
	},
	"west": {
		Command:     "west",
		Aliases:     []string{"w"},
		Description: "Move west.",
		Usage:       "west",
		Examples:    []string{"west", "w"},
		Category:    CatMovement,
		Related:     []string{"north", "south", "east"},
	},
	"up": {
		Command:     "up",
		Aliases:     []string{"u"},
		Description: "Move up.",
		Usage:       "up",
		Examples:    []string{"up", "u"},
		Category:    CatMovement,
		Related:     []string{"down"},
	},
	"down": {
		Command:     "down",
		Aliases:     []string{"dn"},
		Description: "Move down.",
		Usage:       "down",
		Examples:    []string{"down", "dn"},
		Category:    CatMovement,
		Related:     []string{"up"},
	},
	"get": {
		Command:     "get",
		Aliases:     []string{"g", "take", "pick"},
		Description: "Pick up an item from the room.",
		Usage:       "get <item>",
		Examples:    []string{"get phone", "get katana", "get all"},
		Category:    CatItems,
		Related:     []string{"drop", "inventory"},
	},
	"drop": {
		Command:     "drop",
		Aliases:     []string{"d"},
		Description: "Drop an item from your inventory.",
		Usage:       "drop <item>",
		Examples:    []string{"drop phone", "drop trash"},
		Category:    CatItems,
		Related:     []string{"get", "inventory"},
	},
	"inv": {
		Command:     "inv",
		Aliases:     []string{"i", "inventory"},
		Description: "Show your inventory and equipped items.",
		Usage:       "inv",
		Examples:    []string{"inv", "i"},
		Category:    CatInfo,
		Related:     []string{"get", "drop", "equip"},
	},
	"score": {
		Command:     "score",
		Aliases:     []string{"sc", "balance", "bal", "stats"},
		Description: "Show your character stats, XP, level, and money.",
		Usage:       "score",
		Examples:    []string{"score", "sc", "bal"},
		Category:    CatInfo,
		Related:     []string{"skills", "achievements"},
	},
	"kill": {
		Command:     "kill",
		Aliases:     []string{"k", "attack", "a"},
		Description: "Attack an NPC to start combat.",
		Usage:       "kill <target>",
		Examples:    []string{"kill agent", "attack cop"},
		Category:    CatCombat,
		Related:     []string{"flee", "cast"},
	},
	"flee": {
		Command:     "flee",
		Aliases:     []string{"stop", "escape"},
		Description: "Attempt to flee from combat.",
		Usage:       "flee",
		Examples:    []string{"flee", "stop"},
		Category:    CatCombat,
		Related:     []string{"kill", "cast"},
	},
	"cast": {
		Command:     "cast",
		Aliases:     []string{"c", "skill", "use"},
		Description: "Cast a skill. Skills depend on your class.",
		Usage:       "cast <skill> [target]",
		Examples:    []string{"cast glitch agent", "cast patch", "cast smash cop"},
		Category:    CatCombat,
		Related:     []string{"skills", "kill"},
	},
	"wear": {
		Command:     "wear",
		Aliases:     []string{"wield", "equip"},
		Description: "Equip an item from your inventory.",
		Usage:       "wear <item>",
		Examples:    []string{"wear katana", "equip coat"},
		Category:    CatItems,
		Related:     []string{"remove", "inventory"},
	},
	"remove": {
		Command:     "remove",
		Aliases:     []string{"unequip"},
		Description: "Unequip an item and put it in your inventory.",
		Usage:       "remove <item>",
		Examples:    []string{"remove katana", "unequip coat"},
		Category:    CatItems,
		Related:     []string{"wear", "inventory"},
	},
	"say": {
		Command:     "say",
		Aliases:     []string{"'"},
		Description: "Say something to everyone in the room.",
		Usage:       "say <message>",
		Examples:    []string{"say Hello everyone!", "say I need help"},
		Category:    CatSocial,
		Related:     []string{"tell", "gossip"},
	},
	"gossip": {
		Command:     "gossip",
		Aliases:     []string{"chat", ";"},
		Description: "Send a message to all players in the game.",
		Usage:       "gossip <message>",
		Examples:    []string{"gossip Anyone want to group?", "chat Hello world"},
		Category:    CatSocial,
		Related:     []string{"say", "tell"},
	},
	"tell": {
		Command:     "tell",
		Aliases:     []string{"whisper", "t"},
		Description: "Send a private message to another player.",
		Usage:       "tell <player> <message>",
		Examples:    []string{"tell neo Meet me at the dojo", "t trinity Help!"},
		Category:    CatSocial,
		Related:     []string{"say", "gossip"},
	},
	"party": {
		Command:     "party",
		Aliases:     []string{"p"},
		Description: "Manage your party. View status, create, leave, kick, promote, or disband.",
		Usage:       "party [create|leave|kick <player>|promote <player>|disband]",
		Examples:    []string{"party", "party create", "party leave", "party kick neo"},
		Category:    CatParty,
		Related:     []string{"invite", "accept"},
	},
	"invite": {
		Command:     "invite",
		Description: "Invite a player to your party. Creates a party if you don't have one.",
		Usage:       "invite <player>",
		Examples:    []string{"invite neo", "invite trinity"},
		Category:    CatParty,
		Related:     []string{"party", "accept"},
	},
	"accept": {
		Command:     "accept",
		Description: "Accept a party invitation.",
		Usage:       "accept [leader_name]",
		Examples:    []string{"accept", "accept morpheus"},
		Category:    CatParty,
		Related:     []string{"party", "decline"},
	},
	"decline": {
		Command:     "decline",
		Description: "Decline a party invitation.",
		Usage:       "decline [leader_name]",
		Examples:    []string{"decline", "decline morpheus"},
		Category:    CatParty,
		Related:     []string{"party", "accept"},
	},
	"who": {
		Command:     "who",
		Aliases:     []string{"players"},
		Description: "List all players currently online.",
		Usage:       "who",
		Examples:    []string{"who"},
		Category:    CatInfo,
		Related:     []string{"tell", "party"},
	},
	"list": {
		Command:     "list",
		Aliases:     []string{"vendor", "shop"},
		Description: "List items for sale at a vendor.",
		Usage:       "list",
		Examples:    []string{"list", "vendor"},
		Category:    CatEconomy,
		Related:     []string{"buy", "sell"},
	},
	"buy": {
		Command:     "buy",
		Description: "Buy an item from a vendor.",
		Usage:       "buy <item>",
		Examples:    []string{"buy katana", "buy coat"},
		Category:    CatEconomy,
		Related:     []string{"list", "sell"},
	},
	"sell": {
		Command:     "sell",
		Description: "Sell an item to a vendor.",
		Usage:       "sell <item>",
		Examples:    []string{"sell trash", "sell baton"},
		Category:    CatEconomy,
		Related:     []string{"list", "buy"},
	},
	"deposit": {
		Command:     "deposit",
		Description: "Deposit an item into your bank storage (at The Archive).",
		Usage:       "deposit <item>",
		Examples:    []string{"deposit katana", "deposit red_pill"},
		Category:    CatEconomy,
		Related:     []string{"withdraw", "storage"},
	},
	"withdraw": {
		Command:     "withdraw",
		Description: "Withdraw an item from your bank storage (at The Archive).",
		Usage:       "withdraw <item>",
		Examples:    []string{"withdraw katana", "withdraw coat"},
		Category:    CatEconomy,
		Related:     []string{"deposit", "storage"},
	},
	"storage": {
		Command:     "storage",
		Aliases:     []string{"bank"},
		Description: "View items in your bank storage (at The Archive).",
		Usage:       "storage",
		Examples:    []string{"storage", "bank"},
		Category:    CatEconomy,
		Related:     []string{"deposit", "withdraw"},
	},
	"help": {
		Command:     "help",
		Aliases:     []string{"?", "commands"},
		Description: "Show help for commands. Use 'help <command>' for details.",
		Usage:       "help [command|search <term>|topic <name>]",
		Examples:    []string{"help", "help kill", "help search combat", "help topic basics"},
		Category:    CatSystem,
	},
	"recall": {
		Command:     "recall",
		Description: "Teleport back to the dojo (safe room). Useful if stuck.",
		Usage:       "recall",
		Examples:    []string{"recall"},
		Category:    CatSystem,
	},
	"quit": {
		Command:     "quit",
		Aliases:     []string{"exit", "logout"},
		Description: "Save and disconnect from the game.",
		Usage:       "quit",
		Examples:    []string{"quit"},
		Category:    CatSystem,
	},
	"brief": {
		Command:     "brief",
		Description: "Toggle brief mode for shorter room descriptions.",
		Usage:       "brief",
		Examples:    []string{"brief"},
		Category:    CatSystem,
	},
	"theme": {
		Command:     "theme",
		Description: "Change your terminal color theme.",
		Usage:       "theme [green|amber|white|none]",
		Examples:    []string{"theme", "theme amber", "theme none"},
		Category:    CatSystem,
	},
	"faction": {
		Command:     "faction",
		Aliases:     []string{"factions"},
		Description: "Manage your faction alignment. Join Zion, Machines, or Exiles.",
		Usage:       "faction [join|leave|list] [faction]",
		Examples:    []string{"faction", "faction list", "faction join zion", "faction leave"},
		Category:    CatFaction,
		Related:     []string{"reputation"},
	},
	"reputation": {
		Command:     "reputation",
		Aliases:     []string{"rep"},
		Description: "View your reputation with all factions.",
		Usage:       "reputation",
		Examples:    []string{"rep"},
		Category:    CatFaction,
		Related:     []string{"faction"},
	},
	"achievements": {
		Command:     "achievements",
		Aliases:     []string{"ach"},
		Description: "View your achievements and progress.",
		Usage:       "achievements [category]",
		Examples:    []string{"achievements", "ach combat", "ach exploration"},
		Category:    CatAchievement,
		Related:     []string{"title", "stats"},
	},
	"title": {
		Command:     "title",
		Aliases:     []string{"titles"},
		Description: "View or set your display title.",
		Usage:       "title [title_name|clear]",
		Examples:    []string{"title", "title Agent Slayer", "title clear"},
		Category:    CatAchievement,
		Related:     []string{"achievements"},
	},
	"rankings": {
		Command:     "rankings",
		Aliases:     []string{"leaderboard", "top"},
		Description: "View server leaderboards.",
		Usage:       "rankings [category]",
		Examples:    []string{"rankings", "top kills", "leaderboard pvp"},
		Category:    CatInfo,
		Related:     []string{"stats", "achievements"},
	},
	"quest": {
		Command:     "quest",
		Aliases:     []string{"quests", "journal"},
		Description: "View your quest log and active quests.",
		Usage:       "quest [log|hint|abandon <name>]",
		Examples:    []string{"quest", "quest log", "quest hint"},
		Category:    CatQuest,
	},
	"talk": {
		Command:     "talk",
		Aliases:     []string{"speak", "converse"},
		Description: "Talk to an NPC to start a dialogue.",
		Usage:       "talk <npc>",
		Examples:    []string{"talk morpheus", "talk oracle"},
		Category:    CatSocial,
		Related:     []string{"give", "quest"},
	},
	"skills": {
		Command:     "skills",
		Aliases:     []string{"abilities"},
		Description: "View your available skills and abilities.",
		Usage:       "skills",
		Examples:    []string{"skills"},
		Category:    CatCombat,
		Related:     []string{"cast", "score"},
	},
	"recipes": {
		Command:     "recipes",
		Description: "View available crafting recipes.",
		Usage:       "recipes [category]",
		Examples:    []string{"recipes", "recipes weapons"},
		Category:    CatItems,
		Related:     []string{"craft"},
	},
	"craft": {
		Command:     "craft",
		Description: "Craft an item from components.",
		Usage:       "craft <recipe>",
		Examples:    []string{"craft health_vial", "craft katana"},
		Category:    CatItems,
		Related:     []string{"recipes"},
	},
	"tutorial": {
		Command:     "tutorial",
		Aliases:     []string{"tut"},
		Description: "Manage tutorials and view progress.",
		Usage:       "tutorial [list|start <name>|skip|hint]",
		Examples:    []string{"tutorial", "tutorial list", "tutorial skip"},
		Category:    CatTutorial,
	},
	"hint": {
		Command:     "hint",
		Description: "Get a hint for your current objective.",
		Usage:       "hint",
		Examples:    []string{"hint"},
		Category:    CatTutorial,
		Related:     []string{"tutorial", "quest"},
	},
	// Chat commands
	"channel": {
		Command:     "channel",
		Aliases:     []string{"channels", "/channels"},
		Description: "List or manage chat channels.",
		Usage:       "channel [list|join <name>|leave <name>]",
		Examples:    []string{"channel list", "channel join trade"},
		Category:    CatChat,
	},
	// PvP commands
	"arena": {
		Command:     "arena",
		Aliases:     []string{"pvp"},
		Description: "Enter the PvP arena.",
		Usage:       "arena [queue|leave|status]",
		Examples:    []string{"arena queue", "arena status"},
		Category:    CatPvP,
	},
	"duel": {
		Command:     "duel",
		Description: "Challenge a player to a duel.",
		Usage:       "duel <player>",
		Examples:    []string{"duel neo"},
		Category:    CatPvP,
	},
	// Trade commands
	"trade": {
		Command:     "trade",
		Description: "Trade items with another player.",
		Usage:       "trade <player>",
		Examples:    []string{"trade neo", "trade accept", "trade cancel"},
		Category:    CatTrade,
	},
	"auction": {
		Command:     "auction",
		Aliases:     []string{"ah"},
		Description: "Use the auction house.",
		Usage:       "auction [list|sell <item> <price>|buy <id>]",
		Examples:    []string{"auction list", "auction sell katana 500"},
		Category:    CatTrade,
	},
}

// GetHelp returns the help entry for a command, checking aliases
func GetHelp(cmd string) *Entry {
	cmd = strings.ToLower(strings.TrimSpace(cmd))
	
	// Direct lookup
	if entry, ok := Entries[cmd]; ok {
		return entry
	}

	// Check aliases
	for _, entry := range Entries {
		for _, alias := range entry.Aliases {
			if strings.EqualFold(alias, cmd) {
				return entry
			}
		}
	}

	return nil
}

// GetAllByCategory returns all help entries grouped by category
func GetAllByCategory() map[string][]*Entry {
	result := make(map[string][]*Entry)

	for _, entry := range Entries {
		result[entry.Category] = append(result[entry.Category], entry)
	}

	return result
}

// GetCategories returns all category names
func GetCategories() []string {
	return []string{
		CatMovement,
		CatCombat,
		CatItems,
		CatSocial,
		CatInfo,
		CatEconomy,
		CatParty,
		CatQuest,
		CatFaction,
		CatAchievement,
		CatChat,
		CatPvP,
		CatTrade,
		CatTutorial,
		CatSystem,
	}
}

// Search finds help entries matching a search term
func Search(term string) []*Entry {
	term = strings.ToLower(term)
	var results []*Entry
	seen := make(map[string]bool)

	for _, entry := range Entries {
		if seen[entry.Command] {
			continue
		}

		// Check command name
		if strings.Contains(strings.ToLower(entry.Command), term) {
			results = append(results, entry)
			seen[entry.Command] = true
			continue
		}

		// Check aliases
		for _, alias := range entry.Aliases {
			if strings.Contains(strings.ToLower(alias), term) {
				results = append(results, entry)
				seen[entry.Command] = true
				break
			}
		}

		if seen[entry.Command] {
			continue
		}

		// Check description
		if strings.Contains(strings.ToLower(entry.Description), term) {
			results = append(results, entry)
			seen[entry.Command] = true
			continue
		}

		// Check category
		if strings.Contains(strings.ToLower(entry.Category), term) {
			results = append(results, entry)
			seen[entry.Command] = true
		}
	}

	// Sort by relevance (command name match first)
	sort.Slice(results, func(i, j int) bool {
		iName := strings.Contains(strings.ToLower(results[i].Command), term)
		jName := strings.Contains(strings.ToLower(results[j].Command), term)
		if iName != jName {
			return iName
		}
		return results[i].Command < results[j].Command
	})

	return results
}

// SuggestCommand suggests corrections for typos
func SuggestCommand(input string) []string {
	input = strings.ToLower(input)
	suggestions := make(map[string]int) // command -> edit distance

	for cmd := range Entries {
		dist := levenshteinDistance(input, cmd)
		if dist <= 2 { // Within 2 edits
			suggestions[cmd] = dist
		}
		
		// Check aliases too
		for _, alias := range Entries[cmd].Aliases {
			dist := levenshteinDistance(input, strings.ToLower(alias))
			if dist <= 2 {
				if existing, ok := suggestions[cmd]; !ok || dist < existing {
					suggestions[cmd] = dist
				}
			}
		}
	}

	// Sort by distance
	var result []string
	for cmd := range suggestions {
		result = append(result, cmd)
	}
	sort.Slice(result, func(i, j int) bool {
		if suggestions[result[i]] != suggestions[result[j]] {
			return suggestions[result[i]] < suggestions[result[j]]
		}
		return result[i] < result[j]
	})

	if len(result) > 3 {
		result = result[:3]
	}

	return result
}

// levenshteinDistance calculates edit distance between two strings
func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i
	}
	for j := 0; j <= len(b); j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(a)][len(b)]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// GetAutocompleteSuggestions returns command suggestions for partial input
func GetAutocompleteSuggestions(prefix string) []string {
	prefix = strings.ToLower(prefix)
	var matches []string

	for cmd := range Entries {
		if strings.HasPrefix(strings.ToLower(cmd), prefix) {
			matches = append(matches, cmd)
		}
	}

	sort.Strings(matches)
	return matches
}

// GetContextHelp returns contextual help based on player state
func GetContextHelp(state string, lastCommand string, errorMsg string) string {
	var hints []string

	// Based on error message
	if strings.Contains(errorMsg, "not found") {
		hints = append(hints, "Use 'look' to see what's in the room")
	}
	if strings.Contains(errorMsg, "inventory") {
		hints = append(hints, "Check your inventory with 'inv'")
	}
	if strings.Contains(errorMsg, "combat") {
		hints = append(hints, "You can 'flee' to escape combat")
	}

	// Based on state
	switch state {
	case "combat":
		hints = append(hints, "In combat: use 'attack', 'cast <skill>', or 'flee'")
	case "dialogue":
		hints = append(hints, "In dialogue: type a number to select an option, or 'bye' to leave")
	case "trading":
		hints = append(hints, "Trading: 'trade add <item>', 'trade confirm', or 'trade cancel'")
	case "dead":
		hints = append(hints, "You died! Type 'respawn' or wait to auto-respawn")
	}

	// Based on last command
	if lastCommand != "" {
		if entry := GetHelp(lastCommand); entry != nil && len(entry.Related) > 0 {
			hints = append(hints, "Related commands: "+strings.Join(entry.Related, ", "))
		}
	}

	if len(hints) == 0 {
		return "Type 'help' for a list of commands"
	}

	return strings.Join(hints, "\n")
}

// GetManualTopic returns a manual topic
func GetManualTopic(topic string) string {
	topic = strings.ToLower(topic)
	if content, ok := ManualTopics[topic]; ok {
		return content
	}
	return ""
}

// ListManualTopics returns available manual topics
func ListManualTopics() []string {
	var topics []string
	for topic := range ManualTopics {
		topics = append(topics, topic)
	}
	sort.Strings(topics)
	return topics
}
