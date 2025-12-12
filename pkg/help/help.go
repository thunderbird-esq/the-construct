// Package help provides the in-game help system for Matrix MUD.
package help

// Entry represents a single help entry
type Entry struct {
	Command     string
	Aliases     []string
	Description string
	Usage       string
	Examples    []string
	Category    string
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
)

// Entries contains all help entries indexed by command name
var Entries = map[string]*Entry{
	"look": {
		Command:     "look",
		Aliases:     []string{"l"},
		Description: "Look at your surroundings, an item, or an NPC.",
		Usage:       "look [target]",
		Examples:    []string{"look", "look morpheus", "look katana"},
		Category:    CatInfo,
	},
	"north": {
		Command:     "north",
		Aliases:     []string{"n"},
		Description: "Move north.",
		Usage:       "north",
		Examples:    []string{"north", "n"},
		Category:    CatMovement,
	},
	"south": {
		Command:     "south",
		Aliases:     []string{"s"},
		Description: "Move south.",
		Usage:       "south",
		Examples:    []string{"south", "s"},
		Category:    CatMovement,
	},
	"east": {
		Command:     "east",
		Aliases:     []string{"e"},
		Description: "Move east.",
		Usage:       "east",
		Examples:    []string{"east", "e"},
		Category:    CatMovement,
	},
	"west": {
		Command:     "west",
		Aliases:     []string{"w"},
		Description: "Move west.",
		Usage:       "west",
		Examples:    []string{"west", "w"},
		Category:    CatMovement,
	},
	"up": {
		Command:     "up",
		Aliases:     []string{"u"},
		Description: "Move up.",
		Usage:       "up",
		Examples:    []string{"up", "u"},
		Category:    CatMovement,
	},
	"down": {
		Command:     "down",
		Aliases:     []string{"dn"},
		Description: "Move down.",
		Usage:       "down",
		Examples:    []string{"down", "dn"},
		Category:    CatMovement,
	},
	"get": {
		Command:     "get",
		Aliases:     []string{"g"},
		Description: "Pick up an item from the room.",
		Usage:       "get <item>",
		Examples:    []string{"get phone", "get katana"},
		Category:    CatItems,
	},
	"drop": {
		Command:     "drop",
		Aliases:     []string{"d"},
		Description: "Drop an item from your inventory.",
		Usage:       "drop <item>",
		Examples:    []string{"drop phone", "drop trash"},
		Category:    CatItems,
	},
	"inv": {
		Command:     "inv",
		Aliases:     []string{"i", "inventory"},
		Description: "Show your inventory and equipped items.",
		Usage:       "inv",
		Examples:    []string{"inv", "i"},
		Category:    CatInfo,
	},
	"score": {
		Command:     "score",
		Aliases:     []string{"sc", "balance", "bal"},
		Description: "Show your character stats, XP, level, and money.",
		Usage:       "score",
		Examples:    []string{"score", "sc", "bal"},
		Category:    CatInfo,
	},
	"kill": {
		Command:     "kill",
		Aliases:     []string{"k", "attack", "a"},
		Description: "Attack an NPC to start combat.",
		Usage:       "kill <target>",
		Examples:    []string{"kill agent", "attack cop"},
		Category:    CatCombat,
	},
	"flee": {
		Command:     "flee",
		Aliases:     []string{"stop"},
		Description: "Attempt to flee from combat.",
		Usage:       "flee",
		Examples:    []string{"flee", "stop"},
		Category:    CatCombat,
	},
	"cast": {
		Command:     "cast",
		Aliases:     []string{"c"},
		Description: "Cast a skill. Skills depend on your class.",
		Usage:       "cast <skill> [target]",
		Examples:    []string{"cast glitch agent", "cast patch", "cast smash cop"},
		Category:    CatCombat,
	},
	"wear": {
		Command:     "wear",
		Aliases:     []string{"wield", "equip"},
		Description: "Equip an item from your inventory.",
		Usage:       "wear <item>",
		Examples:    []string{"wear katana", "equip coat"},
		Category:    CatItems,
	},
	"remove": {
		Command:     "remove",
		Aliases:     []string{"unequip"},
		Description: "Unequip an item and put it in your inventory.",
		Usage:       "remove <item>",
		Examples:    []string{"remove katana", "unequip coat"},
		Category:    CatItems,
	},
	"use": {
		Command:     "use",
		Aliases:     []string{"eat", "take"},
		Description: "Use a consumable item (potions, pills, etc).",
		Usage:       "use <item>",
		Examples:    []string{"use red_pill", "eat health_vial"},
		Category:    CatItems,
	},
	"give": {
		Command:     "give",
		Aliases:     []string{},
		Description: "Give an item to another player or NPC.",
		Usage:       "give <item> <target>",
		Examples:    []string{"give sunglasses morpheus", "give phone neo"},
		Category:    CatItems,
	},
	"say": {
		Command:     "say",
		Aliases:     []string{},
		Description: "Say something to everyone in the room.",
		Usage:       "say <message>",
		Examples:    []string{"say Hello everyone!", "say I need help"},
		Category:    CatSocial,
	},
	"gossip": {
		Command:     "gossip",
		Aliases:     []string{"chat"},
		Description: "Send a message to all players in the game.",
		Usage:       "gossip <message>",
		Examples:    []string{"gossip Anyone want to group?", "chat Hello world"},
		Category:    CatSocial,
	},
	"tell": {
		Command:     "tell",
		Aliases:     []string{"whisper", "t"},
		Description: "Send a private message to another player.",
		Usage:       "tell <player> <message>",
		Examples:    []string{"tell neo Meet me at the dojo", "t trinity Help!"},
		Category:    CatSocial,
	},
	"party": {
		Command:     "party",
		Aliases:     []string{},
		Description: "Manage your party. View status, create, leave, kick, promote, or disband.",
		Usage:       "party [create|leave|kick <player>|promote <player>|disband]",
		Examples:    []string{"party", "party create", "party leave", "party kick neo"},
		Category:    CatSocial,
	},
	"invite": {
		Command:     "invite",
		Aliases:     []string{},
		Description: "Invite a player to your party. Creates a party if you don't have one.",
		Usage:       "invite <player>",
		Examples:    []string{"invite neo", "invite trinity"},
		Category:    CatSocial,
	},
	"accept": {
		Command:     "accept",
		Aliases:     []string{},
		Description: "Accept a party invitation.",
		Usage:       "accept [leader_name]",
		Examples:    []string{"accept", "accept morpheus"},
		Category:    CatSocial,
	},
	"decline": {
		Command:     "decline",
		Aliases:     []string{},
		Description: "Decline a party invitation.",
		Usage:       "decline [leader_name]",
		Examples:    []string{"decline", "decline morpheus"},
		Category:    CatSocial,
	},
	"who": {
		Command:     "who",
		Aliases:     []string{},
		Description: "List all players currently online.",
		Usage:       "who",
		Examples:    []string{"who"},
		Category:    CatInfo,
	},
	"list": {
		Command:     "list",
		Aliases:     []string{"vendor"},
		Description: "List items for sale at a vendor.",
		Usage:       "list",
		Examples:    []string{"list", "vendor"},
		Category:    CatEconomy,
	},
	"buy": {
		Command:     "buy",
		Aliases:     []string{},
		Description: "Buy an item from a vendor.",
		Usage:       "buy <item>",
		Examples:    []string{"buy katana", "buy coat"},
		Category:    CatEconomy,
	},
	"sell": {
		Command:     "sell",
		Aliases:     []string{},
		Description: "Sell an item to a vendor.",
		Usage:       "sell <item>",
		Examples:    []string{"sell trash", "sell baton"},
		Category:    CatEconomy,
	},
	"deposit": {
		Command:     "deposit",
		Aliases:     []string{},
		Description: "Deposit an item into your bank storage (at The Archive).",
		Usage:       "deposit <item>",
		Examples:    []string{"deposit katana", "deposit red_pill"},
		Category:    CatEconomy,
	},
	"withdraw": {
		Command:     "withdraw",
		Aliases:     []string{},
		Description: "Withdraw an item from your bank storage (at The Archive).",
		Usage:       "withdraw <item>",
		Examples:    []string{"withdraw katana", "withdraw coat"},
		Category:    CatEconomy,
	},
	"storage": {
		Command:     "storage",
		Aliases:     []string{"bank"},
		Description: "View items in your bank storage (at The Archive).",
		Usage:       "storage",
		Examples:    []string{"storage", "bank"},
		Category:    CatEconomy,
	},
	"teleport": {
		Command:     "teleport",
		Aliases:     []string{},
		Description: "Teleport to a specific room (if you know the room ID).",
		Usage:       "teleport <room_id>",
		Examples:    []string{"teleport dojo", "teleport loading_program"},
		Category:    CatMovement,
	},
	"help": {
		Command:     "help",
		Aliases:     []string{"?"},
		Description: "Show help for commands. Use 'help <command>' for details.",
		Usage:       "help [command]",
		Examples:    []string{"help", "help kill", "help cast"},
		Category:    CatSystem,
	},
	"quit": {
		Command:     "quit",
		Aliases:     []string{"exit"},
		Description: "Save and disconnect from the game.",
		Usage:       "quit",
		Examples:    []string{"quit"},
		Category:    CatSystem,
	},
	// --- FACTION COMMANDS ---
	"faction": {
		Command:     "faction",
		Aliases:     []string{"factions"},
		Description: "Manage your faction alignment. Join Zion, Machines, or Exiles.",
		Usage:       "faction [join|leave|list] [faction]",
		Examples:    []string{"faction", "faction list", "faction join zion", "faction leave"},
		Category:    CatSocial,
	},
	"reputation": {
		Command:     "reputation",
		Aliases:     []string{"rep"},
		Description: "View your reputation with all factions.",
		Usage:       "reputation",
		Examples:    []string{"rep"},
		Category:    CatInfo,
	},
	// --- ACHIEVEMENT COMMANDS ---
	"achievements": {
		Command:     "achievements",
		Aliases:     []string{"ach"},
		Description: "View your achievements and progress.",
		Usage:       "achievements [category]",
		Examples:    []string{"achievements", "ach combat", "ach exploration"},
		Category:    CatInfo,
	},
	"title": {
		Command:     "title",
		Aliases:     []string{"titles"},
		Description: "View or set your display title.",
		Usage:       "title [title_name|clear]",
		Examples:    []string{"title", "title Agent Slayer", "title clear"},
		Category:    CatInfo,
	},
	// --- LEADERBOARD COMMANDS ---
	"rankings": {
		Command:     "rankings",
		Aliases:     []string{"leaderboard", "top"},
		Description: "View server leaderboards.",
		Usage:       "rankings [category]",
		Examples:    []string{"rankings", "top kills", "leaderboard pvp"},
		Category:    CatInfo,
	},
	"stats": {
		Command:     "stats",
		Description: "View your personal statistics.",
		Usage:       "stats",
		Examples:    []string{"stats"},
		Category:    CatInfo,
	},
	// --- TRAINING COMMANDS ---
	"train": {
		Command:     "train",
		Aliases:     []string{"training"},
		Description: "Enter training programs to practice combat.",
		Usage:       "train [start|join|leave|complete] [program]",
		Examples:    []string{"train", "train start combat_basic", "train complete"},
		Category:    CatCombat,
	},
	"programs": {
		Command:     "programs",
		Description: "List available training programs.",
		Usage:       "programs",
		Examples:    []string{"programs"},
		Category:    CatCombat,
	},
	"challenges": {
		Command:     "challenges",
		Description: "View combat challenges and records.",
		Usage:       "challenges",
		Examples:    []string{"challenges"},
		Category:    CatTraining,
	},
}

// GetHelp returns the help entry for a command, checking aliases
func GetHelp(cmd string) *Entry {
	// Direct lookup
	if entry, ok := Entries[cmd]; ok {
		return entry
	}

	// Check aliases
	for _, entry := range Entries {
		for _, alias := range entry.Aliases {
			if alias == cmd {
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
		CatSystem,
	}
}
