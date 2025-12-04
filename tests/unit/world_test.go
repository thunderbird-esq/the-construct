package unit

import (
	"encoding/json"
	"os"
	"testing"
)

// Room represents a room structure for testing JSON loading
type Room struct {
	ID          string            `json:"ID"`
	Description string            `json:"Description"`
	Exits       map[string]string `json:"Exits"`
	Symbol      string            `json:"Symbol"`
	Color       string            `json:"Color"`
	Items       []Item            `json:"Items"`
	NPCs        []NPC             `json:"NPCs"`
	ItemMap     map[string]*Item  `json:"ItemMap"`
	NPCMap      map[string]*NPC   `json:"NPCMap"`
}

// Item represents an item for testing
type Item struct {
	ID          string `json:"ID"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
}

// NPC represents an NPC for testing
type NPC struct {
	ID          string `json:"ID"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
	HP          int    `json:"HP"`
	MaxHP       int    `json:"MaxHP"`
}

// WorldData wraps rooms for JSON parsing
type WorldData struct {
	Rooms map[string]*Room `json:"Rooms"`
}

// TestWorldCreation tests that world.json can be loaded and parsed
func TestWorldCreation(t *testing.T) {
	// Test that world.json exists and is valid JSON
	data, err := os.ReadFile("../../data/world.json")
	if err != nil {
		t.Fatalf("Failed to read world.json: %v", err)
	}

	var world WorldData
	if err := json.Unmarshal(data, &world); err != nil {
		t.Fatalf("Failed to parse world.json: %v", err)
	}

	if len(world.Rooms) == 0 {
		t.Error("World has no rooms")
	}

	t.Logf("World loaded successfully with %d rooms", len(world.Rooms))
}

// TestPlayerActions tests player data structure
func TestPlayerActions(t *testing.T) {
	// Test that player JSON files can be loaded
	playerDir := "../../data/players"
	entries, err := os.ReadDir(playerDir)
	if err != nil {
		// Players directory might not exist yet - that's OK
		t.Log("Players directory not found - no saved players yet")
		return
	}

	playerCount := 0
	for _, entry := range entries {
		if !entry.IsDir() && len(entry.Name()) > 5 {
			playerCount++
		}
	}

	t.Logf("Found %d player save files", playerCount)
}

// TestItemManagement tests item data in world.json
func TestItemManagement(t *testing.T) {
	data, err := os.ReadFile("../../data/world.json")
	if err != nil {
		t.Fatalf("Failed to read world.json: %v", err)
	}

	var world WorldData
	if err := json.Unmarshal(data, &world); err != nil {
		t.Fatalf("Failed to parse world.json: %v", err)
	}

	itemCount := 0
	npcCount := 0

	for _, room := range world.Rooms {
		itemCount += len(room.Items)
		npcCount += len(room.NPCs)
	}

	t.Logf("World contains %d items and %d NPCs across %d rooms", itemCount, npcCount, len(world.Rooms))

	// Verify no duplicate data (ItemMap should be nil or empty after clean save)
	duplicateFound := false
	for roomID, room := range world.Rooms {
		if len(room.ItemMap) > 0 && len(room.Items) > 0 {
			duplicateFound = true
			t.Logf("Room %s has both Items array and ItemMap (duplicate data)", roomID)
		}
	}

	if duplicateFound {
		t.Log("Note: Run 'save world' in-game to clean up duplicate data")
	}
}
