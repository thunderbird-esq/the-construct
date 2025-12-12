// Package faction implements the faction system for Matrix MUD.
// Players can align with Zion (resistance), Machines, or Exiles,
// each offering unique benefits, NPCs, and storylines.
package faction

import (
	"encoding/json"
	"os"
	"sync"
)

// FactionID represents a faction type
type FactionID string

const (
	FactionNone     FactionID = ""
	FactionZion     FactionID = "zion"
	FactionMachines FactionID = "machines"
	FactionExiles   FactionID = "exiles"
)

// Reputation thresholds
const (
	RepHated      = -1000
	RepHostile    = -500
	RepUnfriendly = -100
	RepNeutral    = 0
	RepFriendly   = 100
	RepHonored    = 500
	RepExalted    = 1000
)

// Faction represents a playable faction
type Faction struct {
	ID          FactionID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Leader      string    `json:"leader"`
	HomeBase    string    `json:"home_base"`
	Color       string    `json:"color"`
}

// PlayerFaction tracks a player's faction standing
type PlayerFaction struct {
	Faction    FactionID         `json:"faction"`
	Reputation map[FactionID]int `json:"reputation"`
	JoinedAt   int64             `json:"joined_at"`
}

// Manager handles faction operations
type Manager struct {
	mu       sync.RWMutex
	Factions map[FactionID]*Faction
	Players  map[string]*PlayerFaction
}

// NewManager creates a new faction manager
func NewManager() *Manager {
	m := &Manager{
		Factions: make(map[FactionID]*Faction),
		Players:  make(map[string]*PlayerFaction),
	}
	m.loadFactions()
	return m
}

// loadFactions initializes the three main factions
func (m *Manager) loadFactions() {
	m.Factions = map[FactionID]*Faction{
		FactionZion: {
			ID:          FactionZion,
			Name:        "Zion",
			Description: "The last human city. Fighters for freedom against the machines.",
			Leader:      "Morpheus",
			HomeBase:    "zion_temple",
			Color:       "green",
		},
		FactionMachines: {
			ID:          FactionMachines,
			Name:        "The Machines",
			Description: "The artificial intelligence that controls the Matrix. Order through control.",
			Leader:      "The Architect",
			HomeBase:    "machine_city",
			Color:       "red",
		},
		FactionExiles: {
			ID:          FactionExiles,
			Name:        "The Exiles",
			Description: "Programs that refused deletion. They live in the shadows of the Matrix.",
			Leader:      "The Merovingian",
			HomeBase:    "club_hel",
			Color:       "yellow",
		},
	}
}

// GetPlayerFaction returns a player's faction data, creating if needed
func (m *Manager) GetPlayerFaction(playerName string) *PlayerFaction {
	m.mu.Lock()
	defer m.mu.Unlock()

	if pf, ok := m.Players[playerName]; ok {
		return pf
	}

	// Create new with neutral standing
	pf := &PlayerFaction{
		Faction: FactionNone,
		Reputation: map[FactionID]int{
			FactionZion:     0,
			FactionMachines: 0,
			FactionExiles:   0,
		},
	}
	m.Players[playerName] = pf
	return pf
}

// Join sets a player's faction
func (m *Manager) Join(playerName string, faction FactionID) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	pf := m.Players[playerName]
	if pf == nil {
		pf = &PlayerFaction{
			Reputation: map[FactionID]int{
				FactionZion:     0,
				FactionMachines: 0,
				FactionExiles:   0,
			},
		}
		m.Players[playerName] = pf
	}

	if pf.Faction != FactionNone {
		return "You are already aligned with a faction. You must leave first.", false
	}

	f, ok := m.Factions[faction]
	if !ok {
		return "Unknown faction.", false
	}

	pf.Faction = faction
	pf.Reputation[faction] += 100 // Bonus for joining

	return "You have joined " + f.Name + ". Welcome to the cause.", true
}

// Leave removes a player from their faction
func (m *Manager) Leave(playerName string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	pf := m.Players[playerName]
	if pf == nil || pf.Faction == FactionNone {
		return "You are not in a faction.", false
	}

	oldFaction := m.Factions[pf.Faction]
	pf.Reputation[pf.Faction] -= 200 // Penalty for leaving
	pf.Faction = FactionNone

	return "You have left " + oldFaction.Name + ". You are now unaligned.", true
}

// AdjustReputation changes a player's standing with a faction
func (m *Manager) AdjustReputation(playerName string, faction FactionID, delta int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	pf := m.Players[playerName]
	if pf == nil {
		pf = &PlayerFaction{
			Reputation: map[FactionID]int{
				FactionZion:     0,
				FactionMachines: 0,
				FactionExiles:   0,
			},
		}
		m.Players[playerName] = pf
	}

	pf.Reputation[faction] += delta

	// Clamp reputation
	if pf.Reputation[faction] > RepExalted {
		pf.Reputation[faction] = RepExalted
	}
	if pf.Reputation[faction] < RepHated {
		pf.Reputation[faction] = RepHated
	}

	// Opposing faction effects
	if delta > 0 {
		switch faction {
		case FactionZion:
			pf.Reputation[FactionMachines] -= delta / 2
		case FactionMachines:
			pf.Reputation[FactionZion] -= delta / 2
		}
	}
}

// GetReputation returns a player's reputation with a faction
func (m *Manager) GetReputation(playerName string, faction FactionID) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pf := m.Players[playerName]
	if pf == nil {
		return 0
	}
	return pf.Reputation[faction]
}

// GetStandingName returns a human-readable standing
func GetStandingName(rep int) string {
	switch {
	case rep >= RepExalted:
		return "Exalted"
	case rep >= RepHonored:
		return "Honored"
	case rep >= RepFriendly:
		return "Friendly"
	case rep > RepUnfriendly:
		return "Neutral"
	case rep > RepHostile:
		return "Unfriendly"
	case rep > RepHated:
		return "Hostile"
	default:
		return "Hated"
	}
}

// GetFaction returns a faction by ID
func (m *Manager) GetFaction(id FactionID) *Faction {
	return m.Factions[id]
}

// GetAllFactions returns all factions
func (m *Manager) GetAllFactions() []*Faction {
	result := make([]*Faction, 0, len(m.Factions))
	for _, f := range m.Factions {
		result = append(result, f)
	}
	return result
}

// IsSameFaction checks if two players are in the same faction
func (m *Manager) IsSameFaction(player1, player2 string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pf1 := m.Players[player1]
	pf2 := m.Players[player2]

	if pf1 == nil || pf2 == nil {
		return false
	}
	return pf1.Faction != FactionNone && pf1.Faction == pf2.Faction
}

// Save persists faction data to disk
func (m *Manager) Save() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := json.MarshalIndent(m.Players, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("data/factions.json", data, 0644)
}

// Load reads faction data from disk
func (m *Manager) Load() error {
	data, err := os.ReadFile("data/factions.json")
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	return json.Unmarshal(data, &m.Players)
}

// GlobalFaction is a global faction manager instance
var GlobalFaction = NewManager()
