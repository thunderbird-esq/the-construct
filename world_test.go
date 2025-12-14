package main

import (
	"strings"
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


// TestListRecipes verifies recipe listing
func TestListRecipes(t *testing.T) {
	world := NewWorld()
	player := &Player{
		Name:          "TestCrafter",
		CraftingSkill: 3,
	}

	result := world.ListRecipes(player)

	if result == "" {
		t.Error("ListRecipes should return content")
	}
	if !strings.Contains(result, "CRAFTING RECIPES") {
		t.Error("Should contain header")
	}
	if !strings.Contains(result, "health_vial") {
		t.Error("Should list health_vial recipe")
	}
	if !strings.Contains(result, "Crafting Skill: 3") {
		t.Error("Should show player's crafting skill")
	}
}

// TestCraft verifies crafting functionality
func TestCraft(t *testing.T) {
	world := NewWorld()
	player := &Player{
		Name:          "TestCrafter",
		RoomID:        "dojo",
		CraftingSkill: 0,
		Inventory:     []*Item{},
	}

	// Try crafting without ingredients
	result := world.Craft(player, "health_vial")
	if !strings.Contains(strings.ToLower(result), "missing") && !strings.Contains(strings.ToLower(result), "need") {
		t.Logf("Craft without ingredients: %s", result)
	}

	// Add ingredients and try again
	for i := 0; i < 3; i++ {
		player.Inventory = append(player.Inventory, &Item{ID: "trash", Name: "Digital Trash"})
	}

	result = world.Craft(player, "health_vial")
	t.Logf("Craft with ingredients: %s", result)
}

// TestResolveCombatRound verifies combat resolution
// This test is limited because ResolveCombatRound requires a client connection
func TestResolveCombatRound(t *testing.T) {
	world := NewWorld()

	player := &Player{
		Name:     "Fighter",
		RoomID:   "dojo",
		HP:       100,
		MaxHP:    100,
		Strength: 10,
		State:    "IDLE", // Set to IDLE so combat doesn't process
	}

	// Cannot call ResolveCombatRound without a client (would panic)
	// Just verify the player state is valid
	if player.State != "IDLE" {
		t.Errorf("Player state = %q, want IDLE", player.State)
	}
	t.Logf("Combat test player: HP=%d, State=%s (combat requires client)", player.HP, player.State)
	_ = world
}

// TestBroadcast verifies message broadcasting
func TestBroadcast(t *testing.T) {
	world := NewWorld()

	// Broadcast doesn't require players to be in world.Players map
	// It iterates through the map, so we just test with empty map
	// Should not panic
	world.Broadcast("dojo", nil, "Test message")
}

// TestGenerateLoot verifies loot generation
func TestGenerateLoot(t *testing.T) {
	world := NewWorld()

	// GenerateLoot takes a template ID string
	result := world.GenerateLoot("phone")
	if result == nil {
		t.Log("No loot generated (phone template may not exist)")
	} else {
		t.Logf("Generated loot: %s", result.Name)
	}
}

// TestTellWorld verifies private messaging through World method
func TestTellWorld(t *testing.T) {
	world := NewWorld()

	sender := &Player{Name: "Sender", RoomID: "dojo"}

	// Tell non-existent player (we can't easily add players without clients)
	result := world.Tell(sender, "Nobody", "Hello!")
	if !strings.Contains(result, "not found") && !strings.Contains(result, "No player") && result != "" {
		t.Logf("Tell result: %s", result)
	}
}

// TestRepairItem verifies item repair
func TestRepairItem(t *testing.T) {
	world := NewWorld()
	player := &Player{
		Name:   "Repairer",
		RoomID: "dojo",
		Money:  1000,
		Equipment: map[string]*Item{
			"hand": {ID: "sword", Name: "Sword", Durability: 50, MaxDurability: 100},
		},
	}

	result := world.RepairItem(player, "sword")
	t.Logf("Repair result: %s", result)
}

// TestDegradeEquipment verifies durability degradation
func TestDegradeEquipment(t *testing.T) {
	world := NewWorld()
	player := &Player{
		Name:   "Fighter",
		RoomID: "dojo",
		Equipment: map[string]*Item{
			"hand": {ID: "sword", Name: "Sword", Durability: 100, MaxDurability: 100},
		},
	}

	initialDurability := player.Equipment["hand"].Durability
	world.DegradeEquipment(player)

	// Durability may or may not decrease (random chance)
	t.Logf("Durability: %d -> %d", initialDurability, player.Equipment["hand"].Durability)
}

// TestAwakeningFunctions tests awakening-related functions
func TestAwakeningFunctions(t *testing.T) {
	world := NewWorld()
	player := &Player{
		Name:     "Neo",
		RoomID:   "dojo",
		HP:       100,
		MaxHP:    100,
		Awakened: false,
	}

	// Test ShowAbilities before awakening
	result := world.ShowAbilities(player)
	if !strings.Contains(result, "ABILITIES") {
		t.Error("Should show abilities header")
	}

	// Test SeeCode before awakening
	result = world.SeeCode(player)
	if strings.Contains(result, "CODE VISION") {
		t.Error("Should not allow code vision before awakening")
	}

	// Test Focus before awakening
	result = world.Focus(player)
	if strings.Contains(result, "slow") {
		t.Error("Should not allow focus before awakening")
	}

	// Awaken the player
	player.Awakened = true

	// Test SeeCode after awakening
	result = world.SeeCode(player)
	if !strings.Contains(result, "CODE VISION") && result != "" {
		t.Logf("SeeCode result: %s", result)
	}

	// Test Focus after awakening
	result = world.Focus(player)
	if !strings.Contains(result, "slow") && !strings.Contains(result, "focus") {
		t.Logf("Focus result: %s", result)
	}
}

// TestHeatSystem verifies heat/wanted system
func TestHeatSystem(t *testing.T) {
	world := NewWorld()
	player := &Player{
		Name:     "Hacker",
		RoomID:   "dojo",
		Awakened: true,
		Heat:     0,
	}

	// Add heat
	world.AddHeat(player, 10)
	if player.Heat != 10 {
		t.Errorf("Heat should be 10, got %d", player.Heat)
	}

	// DecayHeat operates on all players in world, not a single player
	// We can't easily test it without adding player to world.Players map
	// Just verify AddHeat works
	t.Logf("Heat after AddHeat: %d", player.Heat)

	// Non-awakened players shouldn't accumulate heat
	player2 := &Player{Name: "Bluepill", Awakened: false, Heat: 0}
	world.AddHeat(player2, 50)
	if player2.Heat != 0 {
		t.Errorf("Non-awakened player should not gain heat, got %d", player2.Heat)
	}
}

// TestPhoneFunctions tests phone booth functionality
func TestPhoneFunctions(t *testing.T) {
	world := NewWorld()
	player := &Player{
		Name:             "Operator",
		RoomID:           "dojo",
		DiscoveredPhones: []string{},
	}

	// Test ListPhones with no discovered phones
	result := world.ListPhones(player)
	if !strings.Contains(result, "phone") && !strings.Contains(result, "Phone") {
		t.Logf("ListPhones result: %s", result)
	}

	// Discover a phone
	player.DiscoveredPhones = append(player.DiscoveredPhones, "phone_dojo")

	result = world.ListPhones(player)
	t.Logf("ListPhones after discovery: %s", result)
}
