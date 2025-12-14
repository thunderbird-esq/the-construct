// Package main implements the Matrix MUD game world simulation.
// This file contains the core game state, player management, world mechanics,
// combat system, inventory management, and procedural generation.
package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/yourusername/matrix-mud/pkg/logging"
)

// --- Structs ---

// Item represents an object that can be picked up, equipped, or used.
// Items have different slots (hand, body, head) and can provide AC bonuses or damage.
// Items are generated with varying rarity levels that affect their stats and value.
type Item struct {
	ID, Name, Description string
	Damage, AC            int
	Slot, Type, Effect    string
	Value, Price          int
	// Rarity (0=Common, 1=Uncommon, 2=Rare, 3=Legendary)
	Rarity int `json:"rarity"`
	// Durability system (0 = unbreakable, >0 = current durability)
	Durability    int `json:"durability,omitempty"`
	MaxDurability int `json:"max_durability,omitempty"`
}

// Quest represents an NPC quest that rewards the player for delivering a specific item.
// Quests provide XP rewards and display custom messages upon completion.
type Quest struct {
	WantedItem string `json:"wanted_item"`
	RewardXP   int    `json:"reward_xp"`
	RewardMsg  string `json:"reward_msg"`
}

// NPC represents a non-player character that can engage in combat, dialogue, or trading.
// NPCs have AI behaviors, loot tables, respawn mechanics, and can be hostile or friendly.
// Merchants are protected NPCs that sell items from their inventory.
type NPC struct {
	ID, Name, Description, RoomID, State string
	HP, MaxHP, Damage, AC                int
	Loot                                 []string
	XP, DropMoney                        int
	Vendor                               bool
	Inventory                            []string
	Aggro                                bool
	Quest                                Quest `json:"quest"`
	OriginalRoom                         string
	DeathTime                            time.Time
	IsDead                               bool
	IsAgent                              bool   `json:"is_agent,omitempty"` // Agent NPCs hunt awakened players
	TargetPlayer                         string `json:"-"`                  // Player being hunted (runtime only)
}

// Room represents a location in the game world with connections to other rooms.
// Rooms contain NPCs, items, exits to adjacent rooms, and visual symbols for the automap.
// The ItemMap and NPCMap provide fast lookups for entities in the room.
type Room struct {
	ID, Description string
	Exits           map[string]string
	Symbol, Color   string
	Items           []*Item
	NPCs            []*NPC
	ItemMap         map[string]*Item
	NPCMap          map[string]*NPC
	HasPhone        bool `json:"has_phone,omitempty"` // Room has a phone booth for fast travel
}

// WorldData is a container for serializing the world state to JSON.
// It wraps the room data for persistence to data/world.json.
type WorldData struct{ Rooms map[string]*Room }

// Player represents a connected player character with stats, inventory, and position.
// Player data is persisted to data/players/<name>.json on disconnect and periodically.
// Players have equipment slots (hand, body, head) and can be in combat or idle state.
type Player struct {
	Name, RoomID                string
	Conn                        *Client
	Inventory                   []*Item
	Equipment                   map[string]*Item
	Bank                        []*Item
	HP, MaxHP, Strength, BaseAC int
	MP, MaxMP                   int
	State, Target               string
	LastAttack                  time.Time
	XP, Level                   int
	Class                       string
	Money                       int
	CraftingSkill               int      `json:"crafting_skill,omitempty"`
	Awakened                    bool     `json:"awakened,omitempty"`          // True if player took the red pill
	Heat                        int      `json:"heat,omitempty"`              // Agent aggro level (0-100)
	DiscoveredPhones            []string `json:"discovered_phones,omitempty"` // Phone booth IDs player can call
	BriefMode                   bool     `json:"brief_mode,omitempty"`        // Show short room descriptions
	ColorTheme                  string   `json:"color_theme,omitempty"`       // green, amber, white, none
}

// World represents the entire game state including all rooms, players, NPCs, and items.
// It uses sync.RWMutex for concurrent access from multiple player goroutines.
// The game loop calls World.Update() every 500ms to handle combat, NPC AI, and respawns.
type World struct {
	Rooms         map[string]*Room
	Players       map[*Client]*Player
	Dialogue      map[string]map[string]string
	DeadNPCs      []*NPC
	ItemTemplates map[string]*Item
	MOTD          []string // Message of the Day
	mutex         sync.RWMutex
}

// --- Init ---

// NewWorld creates and initializes a new game world.
// It loads world data from data/world.json, initializes item templates,
// loads NPC dialogue, and sets up room connections. If no world file exists,
// default rooms are created.
func NewWorld() *World {
	w := &World{Rooms: make(map[string]*Room), Players: make(map[*Client]*Player), Dialogue: make(map[string]map[string]string), DeadNPCs: make([]*NPC, 0), ItemTemplates: make(map[string]*Item)}
	w.loadWorldData()
	w.loadDialogue()
	w.loadMOTD()
	return w
}
func (w *World) loadWorldData() {
	file, err := os.ReadFile("data/world.json")
	if err != nil {
		// Issue #7 fix: Handle file read errors gracefully
		logging.Warn().Err(err).Msg("Could not read world.json, creating default world")
		w.createDefaultWorld()
		return
	}

	var data WorldData
	if err := json.Unmarshal(file, &data); err != nil {
		// Issue #7 fix: Handle JSON parse errors gracefully
		logging.Warn().Err(err).Msg("Could not parse world.json, creating default world")
		w.createDefaultWorld()
		return
	}

	w.Rooms = data.Rooms
	if w.Rooms == nil {
		w.Rooms = make(map[string]*Room)
	}

	// Load item templates from JSON file
	w.loadItemTemplates()

	for roomID, room := range w.Rooms {
		room.ItemMap = make(map[string]*Item)
		room.NPCMap = make(map[string]*NPC)
		for _, item := range room.Items {
			room.ItemMap[item.ID] = item
		}
		for _, npc := range room.NPCs {
			npc.RoomID = roomID
			npc.OriginalRoom = roomID

			// Issue #13-14 fix: Ensure NPCs have valid HP values
			if npc.HP <= 0 {
				npc.HP = DefaultNPCHP
				logging.Debug().Str("npc", npc.ID).Str("room", roomID).Int("hp", DefaultNPCHP).Msg("NPC had invalid HP, set to default")
			}
			if npc.MaxHP <= 0 || npc.MaxHP < npc.HP {
				npc.MaxHP = npc.HP
				logging.Debug().Str("npc", npc.ID).Str("room", roomID).Int("max_hp", npc.MaxHP).Msg("NPC had invalid MaxHP, corrected")
			}

			room.NPCMap[npc.ID] = npc
		}
		if room.Symbol == "" {
			room.Symbol = "."
		}
		if room.Color == "" {
			room.Color = "white"
		}
	}
}

// createDefaultWorld creates a minimal world when world.json is missing or corrupt
func (w *World) createDefaultWorld() {
	w.Rooms["spawn"] = &Room{
		ID:          "spawn",
		Description: "You are in a blank white space. The world data could not be loaded.",
		Symbol:      "@",
		Color:       "white",
		ItemMap:     make(map[string]*Item),
		NPCMap:      make(map[string]*NPC),
		Exits:       make(map[string]string),
	}
}

// ItemTemplatesData is the JSON structure for items.json
type ItemTemplatesData struct {
	Items map[string]*Item `json:"items"`
}

// loadItemTemplates loads item templates from data/items.json
func (w *World) loadItemTemplates() {
	file, err := os.ReadFile("data/items.json")
	if err != nil {
		logging.Warn().Err(err).Msg("Could not read items.json, using defaults")
		w.loadDefaultItemTemplates()
		return
	}

	var data ItemTemplatesData
	if err := json.Unmarshal(file, &data); err != nil {
		logging.Warn().Err(err).Msg("Could not parse items.json, using defaults")
		w.loadDefaultItemTemplates()
		return
	}

	// Convert JSON item format to internal Item format
	for id, item := range data.Items {
		w.ItemTemplates[id] = &Item{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			Damage:      item.Damage,
			AC:          item.AC,
			Slot:        item.Slot,
			Type:        item.Type,
			Effect:      item.Effect,
			Value:       item.Value,
			Price:       item.Price,
			Rarity:      item.Rarity,
		}
	}

	logging.Info().Int("count", len(w.ItemTemplates)).Msg("Loaded item templates from items.json")
}

// loadDefaultItemTemplates loads hardcoded item templates as fallback
func (w *World) loadDefaultItemTemplates() {
	w.ItemTemplates["phone"] = &Item{ID: "phone", Name: "Nokia Phone", Description: "An old school slider phone.", Damage: 1, Slot: "hand", Price: 10}
	w.ItemTemplates["coat"] = &Item{ID: "coat", Name: "Leather Trenchcoat", Description: "Black leather. Very cool.", AC: 2, Slot: "body", Price: 100}
	w.ItemTemplates["katana"] = &Item{ID: "katana", Name: "Training Katana", Description: "A dull blade.", Damage: 5, Slot: "hand", Price: 50}
	w.ItemTemplates["red_pill"] = &Item{ID: "red_pill", Name: "Red Pill", Description: "A small red pill.", Type: "consumable", Effect: "buff_str", Value: 1, Price: 200}
	w.ItemTemplates["sunglasses"] = &Item{ID: "sunglasses", Name: "Sunglasses", Description: "Black shades.", AC: 1, Slot: "head", Price: 25}
	w.ItemTemplates["deck"] = &Item{ID: "deck", Name: "Cyberdeck", Description: "A portable hacking unit.", Slot: "hand", Damage: 2, Price: 150}
	w.ItemTemplates["boots"] = &Item{ID: "boots", Name: "Combat Boots", Description: "Heavy boots.", Slot: "body", AC: 2, Price: 80}
	w.ItemTemplates["shades"] = &Item{ID: "shades", Name: "Pilot Shades", Description: "Cool sunglasses.", Slot: "head", AC: 1, Price: 50}
	w.ItemTemplates["trash"] = &Item{ID: "trash", Name: "Digital Trash", Description: "Useless data.", Price: 1}
	w.ItemTemplates["baton"] = &Item{ID: "baton", Name: "Police Baton", Description: "Standard issue.", Damage: 3, Slot: "hand", Price: 20}
	logging.Info().Msg("Loaded default item templates (items.json not available)")
}

func (w *World) loadDialogue() {
	file, _ := os.ReadFile("data/dialogue.json")
	json.Unmarshal(file, &w.Dialogue)
}

// MOTDData is the JSON structure for motd.json
type MOTDData struct {
	MOTD []string `json:"motd"`
}

// loadMOTD loads the message of the day from data/motd.json
func (w *World) loadMOTD() {
	file, err := os.ReadFile("data/motd.json")
	if err != nil {
		logging.Info().Msg("No MOTD configured (motd.json not found)")
		return
	}

	var data MOTDData
	if err := json.Unmarshal(file, &data); err != nil {
		logging.Warn().Err(err).Msg("Could not parse motd.json")
		return
	}

	w.MOTD = data.MOTD
	logging.Info().Int("lines", len(w.MOTD)).Msg("Loaded MOTD")
}

// GetMOTD returns the formatted message of the day
func (w *World) GetMOTD() string {
	if len(w.MOTD) == 0 {
		return ""
	}
	return strings.Join(w.MOTD, "\r\n") + "\r\n"
}

// --- Persistence ---

// SavePlayer persists player data to data/players/<name>.json.
// This is called on disconnect and periodically during gameplay.
// Uses RLock to allow concurrent saves while preventing data corruption.
func (w *World) SavePlayer(p *Player) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	data, _ := json.MarshalIndent(p, "", "  ")
	os.WriteFile("data/players/"+strings.ToLower(p.Name)+".json", data, 0600) // Owner read/write only
}

// LoadPlayer retrieves or creates a player from persistent storage.
// Player data is loaded from data/players/<name>.json if it exists.
// If the player is new, returns a fresh player with default stats at the starting room.
// Ensures backward compatibility by initializing missing fields (Equipment, Bank, MP, etc.).
func (w *World) LoadPlayer(name string, client *Client) *Player {
	data, err := os.ReadFile("data/players/" + strings.ToLower(name) + ".json")
	if err != nil {
		return &Player{Name: name, RoomID: "loading_program", Conn: client, Inventory: make([]*Item, 0), Equipment: make(map[string]*Item), Bank: make([]*Item, 0), HP: 20, MaxHP: 20, MP: 10, MaxMP: 10, Strength: 10, BaseAC: 10, State: "IDLE", XP: 0, Level: 1, Class: "", Money: 0}
	}
	var p Player
	json.Unmarshal(data, &p)
	p.Conn = client
	if p.Equipment == nil {
		p.Equipment = make(map[string]*Item)
	}
	if p.Bank == nil {
		p.Bank = make([]*Item, 0)
	}
	if p.Level == 0 {
		p.Level = 1
	}
	if p.MaxMP == 0 {
		p.MaxMP = 10
		p.MP = 10
	}
	p.State = "IDLE"
	return &p
}

// Broadcast sends a message to all players in a room, optionally excluding one player.
// If exclude is nil, the message is sent to everyone in the room.
func (w *World) Broadcast(roomID string, exclude *Player, msg string) {
	for _, p := range w.Players {
		if p != nil && p.Conn != nil && p.RoomID == roomID {
			if exclude == nil || p != exclude {
				p.Conn.Write(msg)
			}
		}
	}
}

// SaveWorld persists the entire world state to data/world.json.
// Converts ItemMap and NPCMap to slices for JSON serialization.
// Clears maps in output to avoid duplicate data in JSON file.
// Uses RLock to prevent modifications during save.
func (w *World) SaveWorld() {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	// Convert maps to arrays for JSON serialization
	for _, room := range w.Rooms {
		room.Items = make([]*Item, 0, len(room.ItemMap))
		for _, item := range room.ItemMap {
			room.Items = append(room.Items, item)
		}
		room.NPCs = make([]*NPC, 0, len(room.NPCMap))
		for _, npc := range room.NPCMap {
			room.NPCs = append(room.NPCs, npc)
		}
		// Clear maps to avoid duplicate data in JSON
		// Maps are rebuilt from arrays on load
		room.ItemMap = nil
		room.NPCMap = nil
	}

	data := WorldData{Rooms: w.Rooms}
	jsonData, _ := json.MarshalIndent(data, "", "  ")
	os.WriteFile("data/world.json", jsonData, 0600) // Owner read/write only

	// Restore maps after save (so game continues working)
	for _, room := range w.Rooms {
		room.ItemMap = make(map[string]*Item)
		room.NPCMap = make(map[string]*NPC)
		for _, item := range room.Items {
			room.ItemMap[item.ID] = item
		}
		for _, npc := range room.NPCs {
			room.NPCMap[npc.ID] = npc
		}
	}

	fmt.Println("World Saved.")
}

// --- LOOT GENERATION ---

// GenerateLoot creates a randomized item instance from a template.
// Items are rolled for rarity (Common, Uncommon, Rare, Legendary) with
// higher rarities providing stat bonuses and increased value.
// Each generated item receives a unique ID to prevent stack conflicts.
func (w *World) GenerateLoot(templateID string) *Item {
	tmpl, ok := w.ItemTemplates[templateID]
	if !ok {
		return nil
	}

	// Copy Item
	item := *tmpl

	// Roll Rarity
	roll := rand.Intn(100)
	if roll > 98 {
		item.Rarity = 3 // Legendary
		item.Name = "Legendary " + item.Name
		item.Damage += 4
		item.AC += 4
		item.Price *= 10
	} else if roll > 90 {
		item.Rarity = 2 // Rare
		item.Name = "Rare " + item.Name
		item.Damage += 2
		item.AC += 2
		item.Price *= 5
	} else if roll > 75 {
		item.Rarity = 1 // Uncommon
		item.Name = "Uncommon " + item.Name
		item.Damage += 1
		item.AC += 1
		item.Price *= 2
	} else {
		item.Rarity = 0 // Common
	}

	// Unique ID for this instance
	item.ID = fmt.Sprintf("%s_%d", item.ID, rand.Intn(100000))
	return &item
}

// ColorizeItem returns the item name with ANSI color codes based on rarity.
// Common items are white, Uncommon are bright green, Rare are cyan, and Legendary are magenta.
func ColorizeItem(i *Item) string {
	switch i.Rarity {
	case 1:
		return ColorUncommon + i.Name + Reset
	case 2:
		return ColorRare + i.Name + Reset
	case 3:
		return ColorEpic + i.Name + Reset
	default:
		return White + i.Name + Reset
	}
}

// --- PROCEDURAL GENERATION ---

// GenerateCity procedurally generates a grid of interconnected city rooms.
// Creates rows x cols rooms with random descriptions, occasional NPCs (10% chance),
// and items (20% chance). All rooms are automatically connected in a grid pattern
// and linked to the player's current room via a south exit.
func (w *World) GenerateCity(p *Player, rows, cols int) string {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	startRoom := w.Rooms[p.RoomID]
	descriptions := []string{"A rain-slicked city street.", "A dark alleyway.", "A busy intersection.", "The base of a skyscraper.", "A subway entrance.", "A quiet park."}
	baseID := "city_" + fmt.Sprintf("%d", time.Now().Unix())
	gridIDs := make([][]string, rows)
	for r := 0; r < rows; r++ {
		gridIDs[r] = make([]string, cols)
		for c := 0; c < cols; c++ {
			id := fmt.Sprintf("%s_%d_%d", baseID, r, c)
			gridIDs[r][c] = id
			desc := descriptions[rand.Intn(len(descriptions))]
			newRoom := &Room{ID: id, Description: desc, Symbol: ".", Color: "white", Exits: make(map[string]string), ItemMap: make(map[string]*Item), NPCMap: make(map[string]*NPC)}
			roll := rand.Intn(100)
			if roll < 10 {
				npcID := fmt.Sprintf("cop_%d_%d", r, c)
				newRoom.NPCMap[npcID] = &NPC{ID: npcID, Name: "Riot Cop", Description: "Armored police unit.", HP: 25, MaxHP: 25, Damage: 3, AC: 11, State: "IDLE", XP: 50, DropMoney: 10, RoomID: id, OriginalRoom: id, Loot: []string{"baton"}}
				newRoom.Symbol = "!"
				newRoom.Color = "red"
			} else if roll < 30 {
				itemID := fmt.Sprintf("trash_%d_%d", r, c)
				if tmpl, ok := w.ItemTemplates["trash"]; ok {
					item := *tmpl
					item.ID = itemID
					newRoom.ItemMap[itemID] = &item
				}
			}
			w.Rooms[id] = newRoom
		}
	}
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			room := w.Rooms[gridIDs[r][c]]
			if r > 0 {
				room.Exits["north"] = gridIDs[r-1][c]
			}
			if r < rows-1 {
				room.Exits["south"] = gridIDs[r+1][c]
			}
			if c > 0 {
				room.Exits["west"] = gridIDs[r][c-1]
			}
			if c < cols-1 {
				room.Exits["east"] = gridIDs[r][c+1]
			}
		}
	}
	startRoom.Exits["south"] = gridIDs[0][0]
	w.Rooms[gridIDs[0][0]].Exits["north"] = startRoom.ID
	return fmt.Sprintf("Generated %dx%d City Grid.", rows, cols)
}

// --- BANKING SYSTEM ---
func (w *World) DepositItem(p *Player, itemName string) string {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if p.RoomID != "construct_archive" {
		return "You must be in The Archive to access storage."
	}
	for i, item := range p.Inventory {
		if strings.Contains(strings.ToLower(item.Name), itemName) || item.ID == itemName {
			p.Bank = append(p.Bank, item)
			p.Inventory = append(p.Inventory[:i], p.Inventory[i+1:]...)
			return fmt.Sprintf("You upload %s to the Archive.", ColorizeItem(item))
		}
	}
	return "You don't have that."
}
func (w *World) WithdrawItem(p *Player, itemName string) string {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if p.RoomID != "construct_archive" {
		return "You must be in The Archive to access storage."
	}
	for i, item := range p.Bank {
		if strings.Contains(strings.ToLower(item.Name), itemName) || item.ID == itemName {
			p.Inventory = append(p.Inventory, item)
			p.Bank = append(p.Bank[:i], p.Bank[i+1:]...)
			return fmt.Sprintf("You download %s from the Archive.", ColorizeItem(item))
		}
	}
	return "Item not found in Archive."
}
func (w *World) ShowStorage(p *Player) string {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	if p.RoomID != "construct_archive" {
		return "You must be in The Archive to access storage."
	}
	if len(p.Bank) == 0 {
		return "Archive is empty."
	}
	s := "[ ARCHIVE STORAGE ]\r\n"
	for _, item := range p.Bank {
		s += fmt.Sprintf(" - %s\r\n", ColorizeItem(item))
	}
	return s
}
func (w *World) EditRoom(p *Player, field string, value string) string {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	room := w.Rooms[p.RoomID]
	if field == "desc" || field == "description" {
		room.Description = value
		return "Room description updated."
	}
	return "Usage: edit desc [text]"
}

// --- Mapping ---
func (w *World) GenerateAutomapInternal(p *Player, radius int) string {
	type Coord struct{ X, Y int }
	grid := make(map[Coord]string)
	visited := make(map[string]bool)
	type Node struct {
		ID   string
		X, Y int
	}
	queue := []Node{{ID: p.RoomID, X: 0, Y: 0}}
	visited[p.RoomID] = true
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		if math.Abs(float64(curr.X)) > float64(radius) || math.Abs(float64(curr.Y)) > float64(radius) {
			continue
		}
		if room, ok := w.Rooms[curr.ID]; ok {
			colorCode := White
			switch room.Color {
			case "red":
				colorCode = Red
			case "green":
				colorCode = Green
			case "yellow":
				colorCode = Yellow
			case "magenta":
				colorCode = Magenta
			case "cyan":
				colorCode = Cyan
			}
			symbol := room.Symbol
			if curr.X == 0 && curr.Y == 0 {
				symbol = "@"
				colorCode = White
			}
			grid[Coord{curr.X, curr.Y}] = colorCode + symbol + Reset
			for dir, nextID := range room.Exits {
				if !visited[nextID] {
					visited[nextID] = true
					nx, ny := curr.X, curr.Y
					switch dir {
					case "north":
						ny--
					case "south":
						ny++
					case "east":
						nx++
					case "west":
						nx--
					}
					queue = append(queue, Node{ID: nextID, X: nx, Y: ny})
				}
			}
		}
	}
	var sb strings.Builder
	sb.WriteString("\r\n")
	for y := -radius; y <= radius; y++ {
		sb.WriteString("   ")
		for x := -radius; x <= radius; x++ {
			if symbol, ok := grid[Coord{x, y}]; ok {
				sb.WriteString(" " + symbol + " ")
			} else {
				sb.WriteString("   ")
			}
		}
		sb.WriteString("\r\n")
	}
	return sb.String()
}

// --- Logic ---

// Update is called every game tick (500ms) to process combat, NPC AI, and respawns.
// This runs in its own goroutine and uses the world mutex for safe concurrent access.
// Handles NPC respawning (30 second timer), aggressive NPC attacks, MP regeneration,
// and automatic combat round resolution.
func (w *World) Update() {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	now := time.Now()
	activeDead := make([]*NPC, 0)
	for _, npc := range w.DeadNPCs {
		if now.Sub(npc.DeathTime) > 30*time.Second {
			npc.IsDead = false
			npc.HP = npc.MaxHP
			npc.State = "IDLE"
			npc.RoomID = npc.OriginalRoom
			if room, ok := w.Rooms[npc.OriginalRoom]; ok {
				room.NPCMap[npc.ID] = npc
			}
		} else {
			activeDead = append(activeDead, npc)
		}
	}
	w.DeadNPCs = activeDead
	for _, p := range w.Players {
		if p.State == "IDLE" {
			room := w.Rooms[p.RoomID]
			for _, npc := range room.NPCMap {
				if npc.Aggro && npc.State == "IDLE" {
					p.State = "COMBAT"
					p.Target = npc.ID
					p.LastAttack = now.Add(-1 * time.Second)
					npc.State = "COMBAT"
					p.Conn.Write(Matrixify(fmt.Sprintf("\r\n%s%s spots you and ATTACKS!%s\r\n> ", Red, npc.Name, Green)))
					break
				}
			}
		}
		if rand.Intn(6) == 0 {
			if p.MP < p.MaxMP {
				p.MP++
			}
		}
		if p.State == "COMBAT" {
			if now.Sub(p.LastAttack) > 1500*time.Millisecond {
				w.ResolveCombatRound(p)
				p.LastAttack = now
			}
		}
	}

	// Phase 1: Decay heat and run Agent AI (every ~30 seconds via counter)
	w.DecayHeat()
	w.AgentAI()
}

// --- Skills & Combat ---

// CastSkill allows players to use class-specific abilities.
// Hacker: "glitch" - logic bomb attack
// Rebel: "smash" - powerful melee strike
// Operator: "patch" - self-healing ability
// Skills cost MP and some can target NPCs to deal damage.
func (w *World) CastSkill(p *Player, skillName string, targetName string) string {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	cost := 5
	if p.MP < cost {
		return "Not enough MP."
	}
	valid := false
	if p.Class == "Hacker" && skillName == "glitch" {
		valid = true
	}
	if p.Class == "Rebel" && skillName == "smash" {
		valid = true
	}
	if p.Class == "Operator" && skillName == "patch" {
		valid = true
	}
	if !valid {
		return "You don't know that skill."
	}
	room := w.Rooms[p.RoomID]
	if skillName == "patch" {
		p.MP -= cost
		p.HP += 10
		if p.HP > p.MaxHP {
			p.HP = p.MaxHP
		}
		return "Healed 10 HP."
	}
	if targetName == "" && p.State == "COMBAT" {
		targetName = p.Target
	}
	var targetNPC *NPC
	for _, npc := range room.NPCMap {
		if strings.Contains(strings.ToLower(npc.Name), targetName) || npc.ID == targetName {
			targetNPC = npc
			break
		}
	}
	if targetNPC == nil && targetName != "" {
		if npc, ok := room.NPCMap[targetName]; ok {
			targetNPC = npc
		}
	}
	if targetNPC == nil {
		return "Cast at whom?"
	}
	p.MP -= cost
	dmg := 0
	desc := ""
	if skillName == "glitch" {
		dmg = rand.Intn(10) + 5
		desc = fmt.Sprintf("Logic bomb hits %s for %d damage!", targetNPC.Name, dmg)
	} else if skillName == "smash" {
		dmg = rand.Intn(8) + p.Strength
		desc = fmt.Sprintf("Smash hits %s for %d damage!", targetNPC.Name, dmg)
	}
	targetNPC.HP -= dmg
	if p.State != "COMBAT" {
		p.State = "COMBAT"
		p.Target = targetNPC.ID
		p.LastAttack = time.Now()
	}
	if targetNPC.HP <= 0 {
		desc += fmt.Sprintf("\r\n%s collapses.", targetNPC.Name)
		p.XP += targetNPC.XP
		p.Money += targetNPC.DropMoney
		threshold := p.Level * 1000
		if p.XP >= threshold {
			p.Level++
			p.MaxHP += 10
			p.MaxMP += 5
			p.HP = p.MaxHP
			p.MP = p.MaxMP
			p.Strength += 1
			desc += fmt.Sprintf("\r\n%s*** LEVEL UP! ***%s", White, Reset)
		}

		// LOOT GENERATION
		for _, itemID := range targetNPC.Loot {
			drop := w.GenerateLoot(itemID)
			if drop != nil {
				room.ItemMap[drop.ID] = drop
				desc += fmt.Sprintf("\r\n%s dropped %s.", targetNPC.Name, ColorizeItem(drop))
			}
		}
		targetNPC.IsDead = true
		targetNPC.DeathTime = time.Now()
		w.DeadNPCs = append(w.DeadNPCs, targetNPC)
		delete(room.NPCMap, targetNPC.ID)
		p.State = "IDLE"
	}
	return desc
}

// ResolveCombatRound processes one round of combat for a player and their target NPC.
// Calculates attack rolls vs AC (d20 + modifiers), applies damage, checks for death,
// awards XP and level-ups, generates loot drops, and handles player death (respawn).
// Combat rounds occur automatically every 1.5 seconds when player State is COMBAT.
func (w *World) ResolveCombatRound(p *Player) {
	room := w.Rooms[p.RoomID]
	targetNPC, ok := room.NPCMap[p.Target]
	if !ok {
		p.State = "IDLE"
		p.Conn.Write(Matrixify("\r\nTarget lost.\r\n> "))
		return
	}
	output := ""
	damage := 1 + (p.Strength-10)/2
	weaponName := "fists"
	if weapon, ok := p.Equipment["hand"]; ok {
		damage += weapon.Damage
		weaponName = ColorizeItem(weapon)
	}
	if damage < 1 {
		damage = 1
	}
	// Focus (bullet time) doubles damage
	if p.State == "focused" || p.State == "COMBAT" && p.Target != "" {
		if p.State == "focused" {
			damage *= 2
			output += fmt.Sprintf("%s[BULLET TIME]%s ", Cyan, Reset)
			p.State = "COMBAT" // Reset state after use
		}
	}
	roll := rand.Intn(20) + 1
	if roll >= targetNPC.AC {
		targetNPC.HP -= damage
		targetNPC.State = "COMBAT"
		output += fmt.Sprintf("\r\nYou hit %s with %s for %d damage!", targetNPC.Name, weaponName, damage)
		if targetNPC.HP <= 0 {
			output += fmt.Sprintf("\r\n%s collapses.", targetNPC.Name)
			p.XP += targetNPC.XP
			p.Money += targetNPC.DropMoney
			output += fmt.Sprintf("\r\n%sYou gain %d XP and %d Fragments.%s", Green, targetNPC.XP, targetNPC.DropMoney, Reset)
			threshold := p.Level * 1000
			if p.XP >= threshold {
				p.Level++
				p.MaxHP += 10
				p.MaxMP += 5
				p.HP = p.MaxHP
				p.MP = p.MaxMP
				p.Strength += 1
				output += fmt.Sprintf("\r\n%s*** LEVEL UP! ***%s", White, Reset)
			}

			// LOOT GENERATION
			for _, itemID := range targetNPC.Loot {
				drop := w.GenerateLoot(itemID)
				if drop != nil {
					room.ItemMap[drop.ID] = drop
					output += fmt.Sprintf("\r\n%s dropped %s.", targetNPC.Name, ColorizeItem(drop))
				}
			}

			// Add heat for the kill (Agents attract attention)
			w.AddHeat(p, HeatPerKill)

			targetNPC.IsDead = true
			targetNPC.DeathTime = time.Now()
			w.DeadNPCs = append(w.DeadNPCs, targetNPC)
			delete(room.NPCMap, targetNPC.ID)
			p.State = "IDLE"
			p.Conn.Write(Matrixify(output + "\r\n> "))
			return
		}
	} else {
		output += fmt.Sprintf("\r\nYou swing at %s but miss.", targetNPC.Name)
	}
	playerAC := p.BaseAC
	if armor, ok := p.Equipment["body"]; ok {
		playerAC += armor.AC
	}
	if armor, ok := p.Equipment["head"]; ok {
		playerAC += armor.AC
	}
	// Awakened players have a chance to dodge (bonus AC)
	if p.Awakened {
		playerAC += 3
	}
	npcRoll := rand.Intn(20) + 1
	if npcRoll >= playerAC {
		npcDmg := rand.Intn(targetNPC.Damage) + 1
		p.HP -= npcDmg
		output += fmt.Sprintf("\r\n%s hits you for %d damage!", targetNPC.Name, npcDmg)
		if p.HP <= 0 {
			p.HP = p.MaxHP
			p.RoomID = "loading_program"
			p.State = "IDLE"
			output += "\r\n*** YOU HAVE DIED ***\r\nRestoring backup..."
		}
	} else {
		output += fmt.Sprintf("\r\n%s attacks you but misses.", targetNPC.Name)
	}
	p.Conn.Write(Matrixify(output + "\r\n"))
}

// --- Standard Actions ---

// GiveItem allows players to give items from their inventory to NPCs.
// If the NPC has a quest for that item, completes the quest and awards XP.
// Handles fuzzy item ID matching for generated items with random suffixes.
func (w *World) GiveItem(p *Player, itemName string, targetName string) string {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	var itemToGive *Item
	itemIdx := -1
	for i, item := range p.Inventory {
		if strings.Contains(strings.ToLower(item.Name), itemName) || item.ID == itemName {
			itemToGive = item
			itemIdx = i
			break
		}
	}
	if itemToGive == nil {
		return "You don't have that."
	}
	room := w.Rooms[p.RoomID]
	var targetNPC *NPC
	for _, npc := range room.NPCMap {
		if strings.Contains(strings.ToLower(npc.Name), targetName) || npc.ID == targetName {
			targetNPC = npc
			break
		}
	}
	if targetNPC == nil {
		return "They aren't here."
	}
	if targetNPC.Quest.WantedItem == itemToGive.ID || strings.Contains(itemToGive.ID, targetNPC.Quest.WantedItem) { // Fuzzy ID check for generated items
		p.Inventory = append(p.Inventory[:itemIdx], p.Inventory[itemIdx+1:]...)
		p.XP += targetNPC.Quest.RewardXP
		threshold := p.Level * 1000
		levelMsg := ""
		if p.XP >= threshold {
			p.Level++
			p.MaxHP += 10
			p.MaxMP += 5
			p.HP = p.MaxHP
			p.MP = p.MaxMP
			p.Strength += 1
			levelMsg = fmt.Sprintf("\r\n%s*** LEVEL UP! ***%s", White, Reset)
		}
		return fmt.Sprintf("You give %s to %s.\r\n%s%s%s\r\n(Gained %d XP)%s", ColorizeItem(itemToGive), targetNPC.Name, Green, targetNPC.Quest.RewardMsg, Reset, targetNPC.Quest.RewardXP, levelMsg)
	}
	return fmt.Sprintf("%s doesn't seem interested in %s.", targetNPC.Name, itemToGive.Name)
}
func findItemInMap(items map[string]*Item, target string) *Item {
	if item, ok := items[target]; ok {
		return item
	}
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Name), target) {
			return item
		}
	}
	return nil
}

// Look displays the current room description, exits, items, NPCs, and other players.
// If a target is specified, shows detailed information about that NPC or item.
// Includes an ASCII automap showing the local area (2-room radius).
func (w *World) Look(p *Player, target string) string {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	room := w.Rooms[p.RoomID]

	// Issue #4 fix: Handle nil room access
	if room == nil {
		return fmt.Sprintf("%sError: You are in the void (room %s not found). Use 'recall' to return to safety.%s\r\n", Red, p.RoomID, Reset)
	}

	if target == "" {
		automap := w.GenerateAutomapInternal(p, 2)

		// Use brief or full description based on player preference
		roomDesc := room.Description
		if p.BriefMode && len(roomDesc) > 50 {
			// Truncate to first sentence or 50 chars
			if idx := strings.Index(roomDesc, "."); idx > 0 && idx < 60 {
				roomDesc = roomDesc[:idx+1]
			} else if len(roomDesc) > 50 {
				roomDesc = roomDesc[:47] + "..."
			}
		}

		desc := fmt.Sprintf("%s\r\n%s*** %s ***%s\r\n%s\r\nExits: ", automap, White, room.ID, Green, roomDesc)
		for dir := range room.Exits {
			desc += fmt.Sprintf("[%s] ", dir)
		}
		if len(room.ItemMap) > 0 {
			desc += "\r\nVisible Items: "
			for _, item := range room.ItemMap {
				desc += ColorizeItem(item) + " "
			}
		}
		if len(room.NPCMap) > 0 {
			desc += "\r\nWho is here: "
			for _, npc := range room.NPCMap {
				tag := ""
				if npc.Vendor {
					tag = Yellow + " [MERCHANT]" + Green
				}
				if npc.Aggro {
					tag += Red + " [HOSTILE]" + Green
				}
				desc += Matrixify(npc.Name) + tag + " "
			}
		}
		desc += "\r\nPlayers: "
		found := false
		for _, other := range w.Players {
			if other.RoomID == p.RoomID && other != p {
				desc += other.Name + " "
				found = true
			}
		}
		if !found {
			desc += "None."
		}
		return desc + "\r\n"
	}
	for _, npc := range room.NPCMap {
		if strings.Contains(strings.ToLower(npc.Name), target) || npc.ID == target {
			return fmt.Sprintf("\r\n%s\r\nState: %s | HP: %d/%d\r\n", npc.Description, npc.State, npc.HP, npc.MaxHP)
		}
	}
	if item := findItemInMap(room.ItemMap, target); item != nil {
		return fmt.Sprintf("\r\n%s (Damage: %d, AC: %d, Value: %d)\r\n", ColorizeItem(item), item.Damage, item.AC, item.Price)
	}
	for _, item := range p.Inventory {
		if strings.Contains(strings.ToLower(item.Name), target) || item.ID == target {
			return fmt.Sprintf("\r\n%s (Damage: %d, AC: %d, Value: %d)\r\n", ColorizeItem(item), item.Damage, item.AC, item.Price)
		}
	}
	return "You don't see that here."
}
func (w *World) ListGoods(p *Player) string {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	room := w.Rooms[p.RoomID]
	var vendor *NPC
	for _, npc := range room.NPCMap {
		if npc.Vendor {
			vendor = npc
			break
		}
	}
	if vendor == nil {
		return "No merchant."
	}
	s := fmt.Sprintf("\r\n%s%s offers:%s\r\n", Yellow, vendor.Name, Reset)
	for _, itemID := range vendor.Inventory {
		if tmpl, ok := w.ItemTemplates[itemID]; ok {
			s += fmt.Sprintf(" - %-20s : %d Fragments\r\n", ColorizeItem(tmpl), tmpl.Price)
		}
	}
	return s
}
func (w *World) BuyItem(p *Player, itemName string) string {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	room := w.Rooms[p.RoomID]
	var vendor *NPC
	for _, npc := range room.NPCMap {
		if npc.Vendor {
			vendor = npc
			break
		}
	}
	if vendor == nil {
		return "No merchant."
	}
	for _, itemID := range vendor.Inventory {
		if tmpl, ok := w.ItemTemplates[itemID]; ok {
			if strings.Contains(strings.ToLower(tmpl.Name), itemName) || tmpl.ID == itemName {
				if p.Money >= tmpl.Price {
					p.Money -= tmpl.Price
					newItem := *tmpl
					p.Inventory = append(p.Inventory, &newItem)
					return fmt.Sprintf("Bought %s.", ColorizeItem(&newItem))
				} else {
					return "Not enough Fragments."
				}
			}
		}
	}
	return "Merchant doesn't have that."
}
func (w *World) SellItem(p *Player, itemName string) string {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	room := w.Rooms[p.RoomID]
	var vendor *NPC
	for _, npc := range room.NPCMap {
		if npc.Vendor {
			vendor = npc
			break
		}
	}
	if vendor == nil {
		return "No merchant."
	}
	for i, item := range p.Inventory {
		if strings.Contains(strings.ToLower(item.Name), itemName) || item.ID == itemName {
			val := item.Price / 2
			if val < 1 {
				val = 1
			}
			p.Money += val
			p.Inventory = append(p.Inventory[:i], p.Inventory[i+1:]...)
			return fmt.Sprintf("Sold %s for %d.", ColorizeItem(item), val)
		}
	}
	return "You don't have that."
}

// StartCombat initiates combat between a player and an NPC.
// Validates the target exists and is not a protected merchant.
// Sets player state to COMBAT and marks the first attack time.
// Combat runs automatically in the Update() loop until one side dies or flees.
func (w *World) StartCombat(p *Player, targetName string) string {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	room := w.Rooms[p.RoomID]
	var targetNPC *NPC
	for _, npc := range room.NPCMap {
		if strings.Contains(strings.ToLower(npc.Name), targetName) || npc.ID == targetName {
			targetNPC = npc
			break
		}
	}
	if targetNPC == nil {
		return "Not here."
	}
	if targetNPC.Vendor {
		return "Protected."
	}
	p.State = "COMBAT"
	p.Target = targetNPC.ID
	p.LastAttack = time.Now().Add(-2 * time.Second)
	return fmt.Sprintf("Engaging %s!", targetNPC.Name)
}
func (w *World) StopCombat(p *Player) string {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	p.State = "IDLE"
	return "Stopped."
}

// UseItem consumes a consumable item from inventory.
// Supports healing items and stat buff items (like red pill for STR).
// Removes the item from inventory after use.
func (w *World) UseItem(p *Player, itemName string) string {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	for i, item := range p.Inventory {
		if strings.Contains(strings.ToLower(item.Name), itemName) || item.ID == itemName {
			if item.Type != "consumable" {
				return "Can't use."
			}
			msg := ""
			if item.Effect == "heal" {
				p.HP += item.Value
				if p.HP > p.MaxHP {
					p.HP = p.MaxHP
				}
				msg = fmt.Sprintf("Used %s.", ColorizeItem(item))
			} else if item.Effect == "buff_str" {
				p.Strength += item.Value
				p.HP = p.MaxHP
				msg = fmt.Sprintf("Swallowed %s. Str +%d!", ColorizeItem(item), item.Value)
			}
			p.Inventory = append(p.Inventory[:i], p.Inventory[i+1:]...)
			return msg
		}
	}
	return "Don't have."
}

// GetItem picks up an item from the current room and adds it to player inventory.
// Uses fuzzy name matching to find items by partial name or exact ID.
func (w *World) GetItem(p *Player, itemName string) string {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// Issue #8 fix: Check inventory size limit
	if len(p.Inventory) >= MaxInventorySize {
		return fmt.Sprintf("Your inventory is full (max %d items). Drop something first.", MaxInventorySize)
	}

	room := w.Rooms[p.RoomID]
	if room == nil {
		return "You cannot pick up items here."
	}

	item := findItemInMap(room.ItemMap, itemName)
	if item != nil {
		delete(room.ItemMap, item.ID)
		p.Inventory = append(p.Inventory, item)
		return fmt.Sprintf("Got %s.", ColorizeItem(item))
	}
	return "Not here."
}

// DropItem removes an item from player inventory and places it in the current room.
func (w *World) DropItem(p *Player, itemName string) string {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	for i, item := range p.Inventory {
		if strings.Contains(strings.ToLower(item.Name), itemName) || item.ID == itemName {
			p.Inventory = append(p.Inventory[:i], p.Inventory[i+1:]...)
			w.Rooms[p.RoomID].ItemMap[item.ID] = item
			return fmt.Sprintf("Dropped %s.", ColorizeItem(item))
		}
	}
	return "Don't have."
}

// WearItem equips an item from inventory to its designated slot (hand, body, or head).
// Automatically unequips and returns to inventory any item already in that slot.
func (w *World) WearItem(p *Player, itemName string) string {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	for i, item := range p.Inventory {
		if strings.Contains(strings.ToLower(item.Name), itemName) || item.ID == itemName {
			if item.Slot == "" {
				return "Can't wear."
			}
			if oldItem, ok := p.Equipment[item.Slot]; ok {
				p.Inventory = append(p.Inventory, oldItem)
			}
			p.Equipment[item.Slot] = item
			p.Inventory = append(p.Inventory[:i], p.Inventory[i+1:]...)
			return fmt.Sprintf("Equipped %s.", ColorizeItem(item))
		}
	}
	return "Don't have."
}

// RemoveItem unequips an item from the specified slot and returns it to inventory.
func (w *World) RemoveItem(p *Player, slot string) string {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if item, ok := p.Equipment[slot]; ok {
		delete(p.Equipment, slot)
		p.Inventory = append(p.Inventory, item)
		return fmt.Sprintf("Removed %s.", ColorizeItem(item))
	}
	return "Nothing there."
}

// MovePlayer attempts to move a player in the specified direction.
// Cancels combat state and returns a message describing the result.
// Returns an error message if the exit doesn't exist.
func (w *World) MovePlayer(p *Player, direction string) string {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	p.State = "IDLE"
	if next, ok := w.Rooms[p.RoomID].Exits[direction]; ok {
		p.RoomID = next
		// Check for phone booth discovery
		w.CheckPhoneDiscovery(p)
		return fmt.Sprintf("You move to %s.", next)
	}
	return "No exit."
}

// Recall teleports a player back to a safe location (dojo).
// Used when player is stuck in an invalid room or wants to return to safety.
// Has a 60 second cooldown to prevent abuse.
func (w *World) Recall(p *Player) string {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// Default recall location
	const recallRoom = "dojo"

	// Check if already at recall point
	if p.RoomID == recallRoom {
		return "You are already at the recall point.\r\n"
	}

	// Check if recall location exists
	if _, ok := w.Rooms[recallRoom]; !ok {
		return fmt.Sprintf("%sError: Recall location not found. Contact an admin.%s\r\n", Red, Reset)
	}

	// Cancel combat
	p.State = "IDLE"
	p.Target = ""

	// Teleport player
	oldRoom := p.RoomID
	p.RoomID = recallRoom

	return fmt.Sprintf("%sYou close your eyes and focus on the safe house...%s\r\n"+
		"%sReality bends around you. When you open your eyes, you're in the dojo.%s\r\n"+
		"(Recalled from %s)\r\n", Cyan, Reset, Green, Reset, oldRoom)
}

// ShowInventory displays the player's current stats, equipped items, and inventory.
// Shows HP, MP, STR, calculated AC (base + equipment bonuses), and all items.
func (w *World) ShowInventory(p *Player) string {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	ac := p.BaseAC
	for _, item := range p.Equipment {
		ac += item.AC
	}
	s := fmt.Sprintf("HP: %d/%d  MP: %d/%d  STR: %d  AC: %d\r\n[ EQUIPMENT ]\r\n", p.HP, p.MaxHP, p.MP, p.MaxMP, p.Strength, ac)
	if i, ok := p.Equipment["hand"]; ok {
		s += fmt.Sprintf("  Hand: %s\r\n", ColorizeItem(i))
	} else {
		s += "  Hand: <empty>\r\n"
	}
	if i, ok := p.Equipment["body"]; ok {
		s += fmt.Sprintf("  Body: %s\r\n", ColorizeItem(i))
	} else {
		s += "  Body: <empty>\r\n"
	}
	if i, ok := p.Equipment["head"]; ok {
		s += fmt.Sprintf("  Head: %s\r\n", ColorizeItem(i))
	} else {
		s += "  Head: <empty>\r\n"
	}
	s += "[ BACKPACK ]\r\n"
	if len(p.Inventory) == 0 {
		s += "  Empty.\r\n"
	}
	for _, i := range p.Inventory {
		s += fmt.Sprintf("  - %s\r\n", ColorizeItem(i))
	}
	return s
}

// ShowScore displays the player's character sheet.
// Shows name, class, level, XP progress to next level, money, and core stats.
func (w *World) ShowScore(p *Player) string {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	nextLevel := p.Level * 1000
	return fmt.Sprintf("\r\n%s=== %s ===%s\r\nClass: %s\r\nLevel: %d\r\nXP:    %d / %d\r\nFragments: %d\r\nHP:    %d / %d\r\nMP:    %d / %d\r\nSTR:   %d\r\n", Green, p.Name, Reset, p.Class, p.Level, p.XP, nextLevel, p.Money, p.HP, p.MaxHP, p.MP, p.MaxMP, p.Strength)
}
func (w *World) HandleSay(p *Player, msg string) string {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	room := w.Rooms[p.RoomID]
	response := ""
	for _, npc := range room.NPCMap {
		if dialogue, ok := w.Dialogue[npc.ID]; ok {
			for keyword, reply := range dialogue {
				if strings.Contains(strings.ToLower(msg), keyword) {
					response += fmt.Sprintf("\r\n%s%s says: \"%s\"%s\r\n", White, npc.Name, reply, Green)
				}
			}
		}
	}
	return response
}
func (w *World) ListPlayers() string {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	s := "Connected Signals:\r\n"
	for _, p := range w.Players {
		s += fmt.Sprintf("- %s [%s]\r\n", p.Name, p.RoomID)
	}
	return s
}
func (w *World) Teleport(p *Player, dest string) string {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if _, ok := w.Rooms[dest]; ok {
		p.RoomID = dest
		return "Teleported."
	}
	return "Invalid destination."
}
func (w *World) Gossip(p *Player, msg string) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	formatted := fmt.Sprintf("\r\n%s[GLOBAL] %s: %s%s\r\n> ", Yellow, p.Name, msg, Green)
	for _, other := range w.Players {
		other.Conn.Write(formatted)
	}
}
func (w *World) Tell(p *Player, targetName string, msg string) string {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	var target *Player
	for _, other := range w.Players {
		if strings.ToLower(other.Name) == strings.ToLower(targetName) {
			target = other
			break
		}
	}
	if target == nil {
		return "Player not found."
	}
	target.Conn.Write(fmt.Sprintf("\r\n%s%s tells you: %s%s\r\n> ", Magenta, p.Name, msg, Green))
	return fmt.Sprintf("%sYou tell %s: %s%s", Magenta, target.Name, msg, Green)
}
func (w *World) Dig(p *Player, direction string, roomName string) string {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	currentRoom := w.Rooms[p.RoomID]
	if _, exists := currentRoom.Exits[direction]; exists {
		return "Exit exists."
	}
	newID := strings.ToLower(strings.ReplaceAll(roomName, " ", "_"))
	if _, exists := w.Rooms[newID]; exists {
		newID += "_" + fmt.Sprintf("%d", rand.Intn(999))
	}
	newRoom := &Room{ID: newID, Description: roomName, Exits: make(map[string]string), ItemMap: make(map[string]*Item), NPCMap: make(map[string]*NPC), Symbol: ".", Color: "white"}
	reverseDir := getReverseDir(direction)
	currentRoom.Exits[direction] = newID
	newRoom.Exits[reverseDir] = p.RoomID
	w.Rooms[newID] = newRoom
	return fmt.Sprintf("Created room '%s'.", roomName)
}
func getReverseDir(dir string) string {
	switch dir {
	case "north":
		return "south"
	case "south":
		return "north"
	case "east":
		return "west"
	case "west":
		return "east"
	case "up":
		return "down"
	case "down":
		return "up"
	}
	return "back"
}
func (w *World) CreateEntity(p *Player, typeName string, id string) string {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	room := w.Rooms[p.RoomID]
	if typeName == "item" {
		if tmpl, ok := w.ItemTemplates[id]; ok {
			newItem := *tmpl
			room.ItemMap[newItem.ID] = &newItem
			return fmt.Sprintf("Spawned %s.", newItem.Name)
		}
		return "No template."
	} else if typeName == "npc" {
		newNPC := &NPC{ID: id, Name: id, Description: "Construct.", HP: 20, MaxHP: 20, Damage: 2, AC: 10, State: "IDLE", RoomID: p.RoomID, OriginalRoom: p.RoomID}
		room.NPCMap[id] = newNPC
		return fmt.Sprintf("Spawned NPC %s.", id)
	}
	return "Usage: create [item|npc] [id]"
}
func (w *World) DeleteEntity(p *Player, target string) string {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	room := w.Rooms[p.RoomID]
	for id, npc := range room.NPCMap {
		if strings.Contains(strings.ToLower(npc.Name), target) || id == target {
			delete(room.NPCMap, id)
			return fmt.Sprintf("Deleted %s", npc.Name)
		}
	}
	for id, item := range room.ItemMap {
		if strings.Contains(strings.ToLower(item.Name), target) || id == target {
			delete(room.ItemMap, id)
			return fmt.Sprintf("Deleted %s", item.Name)
		}
	}
	return "Not found."
}

// --- Crafting System ---

// ListRecipes shows available crafting recipes
func (w *World) ListRecipes(p *Player) string {
	var sb strings.Builder
	sb.WriteString("=== CRAFTING RECIPES ===\r\n")
	sb.WriteString(fmt.Sprintf("Your Crafting Skill: %d\r\n\r\n", p.CraftingSkill))

	// Recipes are loaded from craftingManager (to be initialized)
	// For now, show basic info
	recipes := []struct {
		name   string
		skill  int
		result string
	}{
		{"health_vial", 0, "Health Vial (3x Digital Trash)"},
		{"emp_grenade", 1, "EMP Grenade (5x Trash, 1x Phone)"},
		{"cyberdeck", 2, "Cyberdeck (2x Phone, 5x Trash)"},
		{"mirror_shades", 2, "Mirror Shades (1x Sunglasses, 3x Trash)"},
		{"repair_kit", 1, "Repair Kit (5x Trash)"},
		{"code_blade", 5, "Code Blade (1x Katana, 1x Red Pill, 10x Trash)"},
		{"operator_coat", 5, "Operator's Coat (1x Coat, 2x Red Pill, 15x Trash)"},
	}

	for _, r := range recipes {
		skillOK := ""
		if p.CraftingSkill >= r.skill {
			skillOK = Green + "[OK]" + Reset
		} else {
			skillOK = Red + "[Skill " + fmt.Sprintf("%d", r.skill) + "]" + Reset
		}
		sb.WriteString(fmt.Sprintf("  %s: %s %s\r\n", r.name, r.result, skillOK))
	}

	sb.WriteString("\r\nUsage: craft <recipe_name>\r\n")
	return sb.String()
}

// Craft attempts to craft an item
func (w *World) Craft(p *Player, recipeName string) string {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// Recipe definitions
	type ingredient struct {
		id  string
		qty int
	}
	type recipe struct {
		name        string
		ingredients []ingredient
		resultID    string
		skill       int
		xp          int
	}

	recipes := map[string]recipe{
		"health_vial":   {"Health Vial", []ingredient{{"trash", 3}}, "health_vial", 0, 10},
		"emp_grenade":   {"EMP Grenade", []ingredient{{"trash", 5}, {"phone", 1}}, "emp_grenade", 1, 25},
		"cyberdeck":     {"Cyberdeck", []ingredient{{"phone", 2}, {"trash", 5}}, "deck", 2, 50},
		"mirror_shades": {"Mirror Shades", []ingredient{{"sunglasses", 1}, {"trash", 3}}, "mirror_shades", 2, 40},
		"repair_kit":    {"Repair Kit", []ingredient{{"trash", 5}}, "repair_kit", 1, 15},
		"code_blade":    {"Code Blade", []ingredient{{"katana", 1}, {"red_pill", 1}, {"trash", 10}}, "code_blade", 5, 100},
		"operator_coat": {"Operator's Coat", []ingredient{{"coat", 1}, {"red_pill", 2}, {"trash", 15}}, "operator_coat", 5, 150},
	}

	r, ok := recipes[strings.ToLower(recipeName)]
	if !ok {
		return "Unknown recipe. Type 'recipes' to see available recipes."
	}

	// Check skill
	if p.CraftingSkill < r.skill {
		return fmt.Sprintf("You need Crafting Skill %d to craft %s. (You have %d)", r.skill, r.name, p.CraftingSkill)
	}

	// Count inventory items
	invCount := make(map[string]int)
	for _, item := range p.Inventory {
		invCount[item.ID]++
	}

	// Check ingredients
	var missing []string
	for _, ing := range r.ingredients {
		if invCount[ing.id] < ing.qty {
			missing = append(missing, fmt.Sprintf("%dx %s", ing.qty-invCount[ing.id], ing.id))
		}
	}
	if len(missing) > 0 {
		return fmt.Sprintf("Missing materials: %s", strings.Join(missing, ", "))
	}

	// Remove ingredients
	for _, ing := range r.ingredients {
		removed := 0
		newInv := make([]*Item, 0, len(p.Inventory))
		for _, item := range p.Inventory {
			if item.ID == ing.id && removed < ing.qty {
				removed++
				continue // Don't add to new inventory
			}
			newInv = append(newInv, item)
		}
		p.Inventory = newInv
	}

	// Create result item
	tmpl, ok := w.ItemTemplates[r.resultID]
	if !ok {
		return "Error: Result item template not found."
	}
	newItem := *tmpl
	// Set durability for equipment items
	if newItem.Slot != "" {
		newItem.MaxDurability = 100
		newItem.Durability = 100
	}
	p.Inventory = append(p.Inventory, &newItem)

	// Award XP
	p.XP += r.xp

	// Small chance to increase crafting skill
	if rand.Intn(100) < 20 { // 20% chance
		p.CraftingSkill++
		return fmt.Sprintf("%sYou crafted %s! (+%d XP) Your crafting skill increased to %d!%s", Green, newItem.Name, r.xp, p.CraftingSkill, Reset)
	}

	return fmt.Sprintf("%sYou crafted %s! (+%d XP)%s", Green, newItem.Name, r.xp, Reset)
}

// RepairItem repairs an equipped item using a repair kit
func (w *World) RepairItem(p *Player, targetSlot string) string {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// Find repair kit
	kitIdx := -1
	for i, item := range p.Inventory {
		if item.ID == "repair_kit" {
			kitIdx = i
			break
		}
	}
	if kitIdx == -1 {
		return "You need a Repair Kit to repair items. Craft one with 'craft repair_kit'."
	}

	// Find item to repair
	item, ok := p.Equipment[targetSlot]
	if !ok {
		return fmt.Sprintf("Nothing equipped in slot '%s'. Slots: hand, body, head", targetSlot)
	}
	if item.MaxDurability == 0 {
		return fmt.Sprintf("%s doesn't have durability to repair.", item.Name)
	}
	if item.Durability >= item.MaxDurability {
		return fmt.Sprintf("%s is already at full durability.", item.Name)
	}

	// Use repair kit
	p.Inventory = append(p.Inventory[:kitIdx], p.Inventory[kitIdx+1:]...)

	// Repair item
	repaired := 25
	item.Durability += repaired
	if item.Durability > item.MaxDurability {
		item.Durability = item.MaxDurability
	}

	return fmt.Sprintf("%sRepaired %s! Durability: %d/%d%s", Green, item.Name, item.Durability, item.MaxDurability, Reset)
}

// DegradeEquipment reduces durability of equipped items after combat
func (w *World) DegradeEquipment(p *Player) {
	for _, item := range p.Equipment {
		if item.MaxDurability > 0 && item.Durability > 0 {
			item.Durability--
			if item.Durability == 0 {
				if p.Conn != nil {
					p.Conn.Write(fmt.Sprintf("\r\n%sYour %s has broken!%s\r\n", Red, item.Name, Reset))
				}
			} else if item.Durability <= 10 {
				if p.Conn != nil {
					p.Conn.Write(fmt.Sprintf("\r\n%sYour %s is nearly broken! (Durability: %d)%s\r\n", Yellow, item.Name, item.Durability, Reset))
				}
			}
		}
	}
}
