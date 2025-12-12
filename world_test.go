package main

import (
	"testing"
)

// TestNewWorld verifies world creation
func TestNewWorld(t *testing.T) {
	world := NewWorld()

	if world == nil {
		t.Fatal("NewWorld should not return nil")
	}
	if world.Rooms == nil {
		t.Error("Rooms map should be initialized")
	}
	if world.Players == nil {
		t.Error("Players map should be initialized")
	}
	if world.Dialogue == nil {
		t.Error("Dialogue map should be initialized")
	}
	if world.DeadNPCs == nil {
		t.Error("DeadNPCs slice should be initialized")
	}
	if world.ItemTemplates == nil {
		t.Error("ItemTemplates map should be initialized")
	}

	t.Logf("World created with %d rooms", len(world.Rooms))
}

// TestWorldRoomsHaveExits verifies all rooms have valid exit references
func TestWorldRoomsHaveExits(t *testing.T) {
	world := NewWorld()

	for roomID, room := range world.Rooms {
		if room.Exits == nil {
			t.Errorf("Room %s has nil Exits map", roomID)
			continue
		}

		for dir, targetID := range room.Exits {
			if targetID == "" {
				t.Errorf("Room %s has empty exit target for %s", roomID, dir)
			}
			if _, exists := world.Rooms[targetID]; !exists {
				t.Errorf("Room %s exit %s points to nonexistent room %s", roomID, dir, targetID)
			}
		}
	}
}

// TestWorldItemTemplates verifies item templates are loaded
func TestWorldItemTemplates(t *testing.T) {
	world := NewWorld()

	expectedItems := []string{"phone", "coat", "katana", "red_pill", "sunglasses", "deck", "boots", "shades", "trash", "baton"}

	for _, itemID := range expectedItems {
		if _, exists := world.ItemTemplates[itemID]; !exists {
			t.Errorf("Missing item template: %s", itemID)
		}
	}

	t.Logf("World has %d item templates", len(world.ItemTemplates))
}

// TestLookAtRoom verifies look command
func TestLookAtRoom(t *testing.T) {
	world := NewWorld()

	player := &Player{
		Name:   "TestPlayer",
		RoomID: "loading_program",
	}

	result := world.Look(player, "")

	if result == "" {
		t.Error("Look should return room description")
	}

	// Should contain room description elements
	if len(result) < 20 {
		t.Error("Look result seems too short")
	}

	t.Logf("Look result length: %d chars", len(result))
}

// TestLookAtNonexistentRoom verifies look handles invalid rooms
func TestLookAtNonexistentRoom(t *testing.T) {
	world := NewWorld()

	player := &Player{
		Name:   "TestPlayer",
		RoomID: "nonexistent_room_xyz",
	}

	result := world.Look(player, "")

	// Should return error message, not panic
	if result == "" {
		t.Error("Look should return error for nonexistent room")
	}

	// Should mention error or void
	if len(result) < 10 {
		t.Error("Error message seems too short")
	}
}

// TestMovePlayer verifies movement
func TestMovePlayer(t *testing.T) {
	world := NewWorld()

	player := &Player{
		Name:   "TestPlayer",
		RoomID: "loading_program",
	}

	// Move south (should work - leads to subway)
	result := world.MovePlayer(player, "south")

	if player.RoomID != "subway" {
		t.Errorf("Player should be in subway, got %s", player.RoomID)
	}

	// Move invalid direction
	player.RoomID = "loading_program"
	result = world.MovePlayer(player, "invalid")

	if player.RoomID != "loading_program" {
		t.Error("Invalid direction should not move player")
	}
	if result == "" {
		t.Error("Invalid move should return message")
	}
}

// TestGetItem verifies item pickup
func TestGetItemWorld(t *testing.T) {
	world := NewWorld()

	player := &Player{
		Name:      "TestPlayer",
		RoomID:    "loading_program",
		Inventory: make([]*Item, 0),
	}

	// Pick up phone
	result := world.GetItem(player, "phone")

	if len(player.Inventory) == 0 {
		t.Error("Player should have item in inventory")
	}

	t.Logf("GetItem result: %s", result)
}

// TestGetNonexistentItem verifies get handles missing items
func TestGetNonexistentItem(t *testing.T) {
	world := NewWorld()

	player := &Player{
		Name:      "TestPlayer",
		RoomID:    "loading_program",
		Inventory: make([]*Item, 0),
	}

	result := world.GetItem(player, "nonexistent_item_xyz")

	if len(player.Inventory) != 0 {
		t.Error("Inventory should be empty")
	}
	if result == "" {
		t.Error("Should return message for nonexistent item")
	}
}

// TestDropItem verifies item dropping
func TestDropItemWorld(t *testing.T) {
	world := NewWorld()

	item := &Item{
		ID:   "test_drop",
		Name: "Test Item",
	}

	player := &Player{
		Name:      "TestPlayer",
		RoomID:    "loading_program",
		Inventory: []*Item{item},
	}

	result := world.DropItem(player, "test_drop")

	if len(player.Inventory) != 0 {
		t.Error("Inventory should be empty after drop")
	}

	t.Logf("DropItem result: %s", result)
}

// TestShowInventory verifies inventory display
func TestShowInventoryWorld(t *testing.T) {
	world := NewWorld()

	player := &Player{
		Name: "TestPlayer",
		Inventory: []*Item{
			{ID: "item1", Name: "Item One"},
			{ID: "item2", Name: "Item Two"},
		},
	}

	result := world.ShowInventory(player)

	if result == "" {
		t.Error("ShowInventory should return content")
	}

	t.Logf("Inventory display: %s", result)
}

// TestShowScore verifies score display
func TestShowScoreWorld(t *testing.T) {
	world := NewWorld()

	player := &Player{
		Name:     "TestPlayer",
		HP:       50,
		MaxHP:    100,
		MP:       25,
		MaxMP:    50,
		XP:       1000,
		Level:    5,
		Class:    "Hacker",
		Money:    500,
		Strength: 12,
		BaseAC:   10,
	}

	result := world.ShowScore(player)

	if result == "" {
		t.Error("ShowScore should return content")
	}

	t.Logf("Score display: %s", result)
}

// TestCreateDefaultWorld verifies fallback world creation
func TestCreateDefaultWorld(t *testing.T) {
	world := &World{
		Rooms: make(map[string]*Room),
	}

	world.createDefaultWorld()

	if _, exists := world.Rooms["spawn"]; !exists {
		t.Error("Default world should have spawn room")
	}
}

// TestRoomMaps verifies ItemMap and NPCMap are populated
func TestRoomMaps(t *testing.T) {
	world := NewWorld()

	for roomID, room := range world.Rooms {
		if room.ItemMap == nil {
			t.Errorf("Room %s has nil ItemMap", roomID)
		}
		if room.NPCMap == nil {
			t.Errorf("Room %s has nil NPCMap", roomID)
		}

		// Verify Items are in ItemMap
		for _, item := range room.Items {
			if _, exists := room.ItemMap[item.ID]; !exists {
				t.Errorf("Room %s: item %s not in ItemMap", roomID, item.ID)
			}
		}

		// Verify NPCs are in NPCMap
		for _, npc := range room.NPCs {
			if _, exists := room.NPCMap[npc.ID]; !exists {
				t.Errorf("Room %s: NPC %s not in NPCMap", roomID, npc.ID)
			}
		}
	}
}

// TestNPCsHaveValidHP verifies all NPCs have proper HP values
func TestNPCsHaveValidHP(t *testing.T) {
	world := NewWorld()

	for roomID, room := range world.Rooms {
		for _, npc := range room.NPCs {
			if npc.HP <= 0 {
				t.Errorf("NPC %s in %s has invalid HP: %d", npc.ID, roomID, npc.HP)
			}
			if npc.MaxHP <= 0 {
				t.Errorf("NPC %s in %s has invalid MaxHP: %d", npc.ID, roomID, npc.MaxHP)
			}
			if npc.HP > npc.MaxHP {
				t.Errorf("NPC %s in %s has HP > MaxHP: %d > %d", npc.ID, roomID, npc.HP, npc.MaxHP)
			}
		}
	}
}
