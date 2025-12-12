// Package game implements core game mechanics for Matrix MUD.
// This package contains the combat system, inventory management,
// NPC behaviors, room interactions, and player actions.
package game

import (
	"sync"
	"time"
)

// Item represents an object that can be picked up, equipped, or used.
// Items have different slots (hand, body, head) and can provide AC bonuses or damage.
// Items are generated with varying rarity levels that affect their stats and value.
type Item struct {
	ID, Name, Description string
	Damage, AC            int
	Slot, Type, Effect    string
	Value, Price          int
	Rarity                int `json:"rarity"` // 0=Common, 1=Uncommon, 2=Rare, 3=Legendary
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
}

// Player represents a connected player character with stats, inventory, and position.
// Player data is persisted to data/players/<n>.json on disconnect and periodically.
// Players have equipment slots (hand, body, head) and can be in combat or idle state.
type Player struct {
	Name, RoomID                string
	Conn                        interface{} // *Client from main package
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
}

// WorldData is a container for serializing the world state to JSON.
// It wraps the room data for persistence to data/world.json.
type WorldData struct {
	Rooms map[string]*Room
}

// World represents the entire game state including all rooms, players, NPCs, and items.
// It uses sync.RWMutex for concurrent access from multiple player goroutines.
type World struct {
	Rooms         map[string]*Room
	Players       map[interface{}]*Player // map[*Client]*Player
	Dialogue      map[string]map[string]string
	DeadNPCs      []*NPC
	ItemTemplates map[string]*Item
	Mutex         sync.RWMutex
}
