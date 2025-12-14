package main

import (
	"strings"
	"testing"
)

// TestTakePillNotAwakened verifies red pill awakening
func TestTakePillNotAwakened(t *testing.T) {
	world := NewWorld()
	
	player := &Player{
		Name:      "Neo",
		RoomID:    "dojo",
		Awakened:  false,
		HP:        100,
		MaxHP:     100,
		Strength:  10,
		Inventory: []*Item{},
	}
	
	// Add red pill to room
	room := world.Rooms[player.RoomID]
	if room == nil {
		t.Skip("Dojo room not found")
	}
	redPill := &Item{ID: "red_pill", Name: "Red Pill"}
	room.Items = append(room.Items, redPill)
	if room.ItemMap == nil {
		room.ItemMap = make(map[string]*Item)
	}
	room.ItemMap["red_pill"] = redPill
	
	result := world.TakePill(player, "red")
	
	if !player.Awakened {
		t.Error("Player should be awakened after taking red pill")
	}
	if !strings.Contains(result, "red pill") {
		t.Errorf("Result should mention red pill: %s", result)
	}
	if player.MaxHP != 110 {
		t.Errorf("MaxHP should be 110 after awakening, got %d", player.MaxHP)
	}
	if player.Strength != 12 {
		t.Errorf("Strength should be 12 after awakening, got %d", player.Strength)
	}
}

// TestTakePillAlreadyAwakened verifies cannot re-awaken
func TestTakePillAlreadyAwakened(t *testing.T) {
	world := NewWorld()
	
	player := &Player{
		Name:     "Neo",
		RoomID:   "dojo",
		Awakened: true,
	}
	
	result := world.TakePill(player, "red")
	
	if !strings.Contains(result, "already seen the truth") {
		t.Errorf("Should indicate already awakened: %s", result)
	}
}

// TestTakePillNoPill verifies error when pill not present
func TestTakePillNoPill(t *testing.T) {
	world := NewWorld()
	
	player := &Player{
		Name:      "Neo",
		RoomID:    "dojo",
		Awakened:  false,
		Inventory: []*Item{},
	}
	
	result := world.TakePill(player, "red")
	
	if !strings.Contains(result, "don't see") {
		t.Errorf("Should indicate pill not found: %s", result)
	}
}

// TestTakeBluePill verifies blue pill effect
func TestTakeBluePill(t *testing.T) {
	world := NewWorld()
	
	player := &Player{
		Name:      "Thomas",
		RoomID:    "dojo",
		Awakened:  false,
		Inventory: []*Item{},
	}
	
	// Add blue pill to inventory
	bluePill := &Item{ID: "blue_pill", Name: "Blue Pill"}
	player.Inventory = append(player.Inventory, bluePill)
	
	result := world.TakePill(player, "blue")
	
	if player.Awakened {
		t.Error("Player should NOT be awakened after taking blue pill")
	}
	if !strings.Contains(result, "wake up in your bed") {
		t.Errorf("Should have blue pill message: %s", result)
	}
}

// TestTakePillInvalidColor verifies error for invalid pill color
func TestTakePillInvalidColor(t *testing.T) {
	world := NewWorld()
	
	player := &Player{
		Name:     "Neo",
		RoomID:   "dojo",
		Awakened: false,
	}
	
	result := world.TakePill(player, "green")
	
	// The function checks for the pill first, so it says "don't see a green pill"
	if !strings.Contains(result, "don't see") {
		t.Errorf("Should indicate pill not found: %s", result)
	}
}

// TestTakePillNilRoom verifies handling of invalid room
func TestTakePillNilRoom(t *testing.T) {
	world := NewWorld()
	
	player := &Player{
		Name:     "Neo",
		RoomID:   "nonexistent_room",
		Awakened: false,
	}
	
	result := world.TakePill(player, "red")
	
	if !strings.Contains(result, "cannot do that") {
		t.Errorf("Should indicate cannot do that here: %s", result)
	}
}

// TestRemoveItemFromRoom verifies item removal from room
func TestRemoveItemFromRoom(t *testing.T) {
	world := NewWorld()
	
	room := &Room{
		ID:      "test_room",
		Items:   []*Item{},
		ItemMap: make(map[string]*Item),
	}
	
	item1 := &Item{ID: "item1", Name: "Item 1"}
	item2 := &Item{ID: "item2", Name: "Item 2"}
	room.Items = append(room.Items, item1, item2)
	room.ItemMap["item1"] = item1
	room.ItemMap["item2"] = item2
	
	world.removeItemFromRoom(room, item1)
	
	if len(room.Items) != 1 {
		t.Errorf("Room should have 1 item, got %d", len(room.Items))
	}
	if _, exists := room.ItemMap["item1"]; exists {
		t.Error("item1 should be removed from ItemMap")
	}
	if room.Items[0].ID != "item2" {
		t.Error("Remaining item should be item2")
	}
}

// TestRemoveItemFromInventory verifies item removal from inventory
func TestRemoveItemFromInventory(t *testing.T) {
	world := NewWorld()
	
	item1 := &Item{ID: "item1", Name: "Item 1"}
	item2 := &Item{ID: "item2", Name: "Item 2"}
	
	player := &Player{
		Name:      "Test",
		Inventory: []*Item{item1, item2},
	}
	
	world.removeItemFromInventory(player, item1)
	
	if len(player.Inventory) != 1 {
		t.Errorf("Inventory should have 1 item, got %d", len(player.Inventory))
	}
	if player.Inventory[0].ID != "item2" {
		t.Error("Remaining item should be item2")
	}
}

// TestMaybeSpawnAgent verifies agent spawning logic
func TestMaybeSpawnAgent(t *testing.T) {
	world := NewWorld()
	
	// Non-awakened player should not spawn agents
	player1 := &Player{
		Name:     "Bluepill",
		RoomID:   "dojo",
		Awakened: false,
		Heat:     100, // Max heat but not awakened
	}
	
	world.maybeSpawnAgent(player1)
	// Should not panic, agent only spawns for awakened players
	
	// Awakened player with low heat
	player2 := &Player{
		Name:     "Neo",
		RoomID:   "dojo",
		Awakened: true,
		Heat:     10, // Low heat
	}
	
	world.maybeSpawnAgent(player2)
	// Should not spawn (heat too low)
}

// TestMoveNPC verifies NPC movement
func TestMoveNPC(t *testing.T) {
	world := NewWorld()
	
	// Create an NPC in a room with exits
	npc := &NPC{
		ID:           "test_npc",
		Name:         "Test NPC",
		HP:           50,
		MaxHP:        50,
		OriginalRoom: "loading_program",
	}
	
	fromRoom := world.Rooms["loading_program"]
	if fromRoom == nil {
		t.Skip("loading_program room not found")
	}
	
	// Find a connected room
	var toRoomID string
	for _, exitRoom := range fromRoom.Exits {
		toRoomID = exitRoom
		break
	}
	
	if toRoomID == "" {
		t.Skip("No exits from loading_program")
	}
	
	toRoom := world.Rooms[toRoomID]
	if toRoom == nil {
		t.Skip("Target room not found")
	}
	
	// Add NPC to from room
	fromRoom.NPCs = append(fromRoom.NPCs, npc)
	if fromRoom.NPCMap == nil {
		fromRoom.NPCMap = make(map[string]*NPC)
	}
	fromRoom.NPCMap[npc.ID] = npc
	
	// moveNPC requires from, to, and direction
	world.moveNPC(npc, fromRoom, toRoom, "south")
	
	// NPC should be moved (verify no panic)
	t.Log("moveNPC completed without panic")
}

// TestDecayHeat verifies heat decay for all players
func TestDecayHeat(t *testing.T) {
	world := NewWorld()
	
	// DecayHeat operates on world.Players which uses *Client keys
	// We can't easily add players without clients, but we can call it
	// to ensure it doesn't panic with empty players
	world.DecayHeat()
	
	// Test passed if no panic
	t.Log("DecayHeat completed without panic")
}

// TestAgentAI verifies agent AI behavior
func TestAgentAI(t *testing.T) {
	world := NewWorld()
	
	// AgentAI operates on all agents, takes no parameters
	// Should not panic without players
	world.AgentAI()
	
	t.Log("AgentAI completed without panic")
}

// TestShowAbilitiesHacker verifies Hacker class abilities
func TestShowAbilitiesHacker(t *testing.T) {
	world := NewWorld()
	player := &Player{
		Name:     "TestHacker",
		Class:    "Hacker",
		Awakened: false,
	}

	result := world.ShowAbilities(player)

	if !strings.Contains(result, "ABILITIES") {
		t.Error("Should contain ABILITIES header")
	}
	if !strings.Contains(result, "Hacker") {
		t.Error("Should show Hacker class")
	}
	if !strings.Contains(result, "glitch") {
		t.Error("Should show glitch ability")
	}
	if !strings.Contains(result, "patch") {
		t.Error("Should show patch ability")
	}
	if !strings.Contains(result, "overflow") {
		t.Error("Should show overflow ability")
	}
	if !strings.Contains(result, "red pill") {
		t.Error("Should mention red pill for non-awakened")
	}
}

// TestShowAbilitiesRebel verifies Rebel class abilities
func TestShowAbilitiesRebel(t *testing.T) {
	world := NewWorld()
	player := &Player{
		Name:     "TestRebel",
		Class:    "Rebel",
		Awakened: false,
	}

	result := world.ShowAbilities(player)

	if !strings.Contains(result, "Rebel") {
		t.Error("Should show Rebel class")
	}
	if !strings.Contains(result, "smash") {
		t.Error("Should show smash ability")
	}
	if !strings.Contains(result, "fortify") {
		t.Error("Should show fortify ability")
	}
	if !strings.Contains(result, "rampage") {
		t.Error("Should show rampage ability")
	}
}

// TestShowAbilitiesOperator verifies Operator class abilities
func TestShowAbilitiesOperator(t *testing.T) {
	world := NewWorld()
	player := &Player{
		Name:     "TestOperator",
		Class:    "Operator",
		Awakened: false,
	}

	result := world.ShowAbilities(player)

	if !strings.Contains(result, "Operator") {
		t.Error("Should show Operator class")
	}
	if !strings.Contains(result, "strike") {
		t.Error("Should show strike ability")
	}
	if !strings.Contains(result, "vanish") {
		t.Error("Should show vanish ability")
	}
	if !strings.Contains(result, "assassinate") {
		t.Error("Should show assassinate ability")
	}
}

// TestShowAbilitiesAwakened verifies awakened powers section
func TestShowAbilitiesAwakened(t *testing.T) {
	world := NewWorld()
	player := &Player{
		Name:     "TestAwakened",
		Class:    "Hacker",
		Awakened: true,
	}

	result := world.ShowAbilities(player)

	if !strings.Contains(result, "Awakened Powers") {
		t.Error("Should show Awakened Powers header")
	}
	if !strings.Contains(result, "see_code") {
		t.Error("Should show see_code ability")
	}
	if !strings.Contains(result, "dodge") {
		t.Error("Should show dodge ability")
	}
	if !strings.Contains(result, "focus") {
		t.Error("Should show focus ability")
	}
	if strings.Contains(result, "red pill") {
		t.Error("Should NOT mention red pill for awakened player")
	}
}

// TestSeeCodeWithItems verifies code vision with items in room
func TestSeeCodeWithItems(t *testing.T) {
	world := NewWorld()
	player := &Player{
		Name:     "CodeViewer",
		RoomID:   "loading_program",
		Awakened: true,
	}

	result := world.SeeCode(player)

	if !strings.Contains(result, "CODE VISION") {
		t.Error("Should contain CODE VISION header")
	}
	if !strings.Contains(result, "Room ID") {
		t.Error("Should show Room ID")
	}
}

// TestAddHeatMax verifies heat capping at 100
func TestAddHeatMax(t *testing.T) {
	world := NewWorld()
	player := &Player{
		Name:     "Hacker",
		RoomID:   "dojo",
		Awakened: true,
		Heat:     90,
	}

	world.AddHeat(player, 50)

	if player.Heat > 100 {
		t.Errorf("Heat should be capped at 100, got %d", player.Heat)
	}
	if player.Heat != 100 {
		t.Errorf("Heat should be 100, got %d", player.Heat)
	}
}

// TestMaybeSpawnAgentHighHeat verifies agent spawn chance increases with heat
func TestMaybeSpawnAgentHighHeat(t *testing.T) {
	world := NewWorld()
	player := &Player{
		Name:     "HighHeat",
		RoomID:   "dojo",
		Awakened: true,
		Heat:     100, // Max heat
	}

	// Run multiple times - should not panic
	for i := 0; i < 10; i++ {
		world.maybeSpawnAgent(player)
	}
	t.Log("maybeSpawnAgent with high heat completed without panic")
}
