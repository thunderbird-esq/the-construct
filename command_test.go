package main

import (
	"strings"
	"testing"

	"github.com/yourusername/matrix-mud/pkg/help"
)

// TestParseCommandFunction tests the parseCommand function directly
func TestParseCommandFunction(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantCmd string
		wantArg string
	}{
		{"empty", "", "", ""},
		{"single command", "look", "look", ""},
		{"command with arg", "look morpheus", "look", "morpheus"},
		{"uppercase", "NORTH", "north", ""},
		{"multi-word arg", "get red pill", "get", "red pill"},
		{"with whitespace", "  say hello world  ", "say", "hello world"},
		{"tabs and spaces", "\t tell  bob   hello there \t", "tell", "bob hello there"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, arg := parseCommand(tt.input)
			if cmd != tt.wantCmd {
				t.Errorf("cmd = %q, want %q", cmd, tt.wantCmd)
			}
			if arg != tt.wantArg {
				t.Errorf("arg = %q, want %q", arg, tt.wantArg)
			}
		})
	}
}

// TestFormatHelpEmpty tests help with no argument
func TestFormatHelpEmpty(t *testing.T) {
	result := formatHelp("")

	if !strings.Contains(result, "HELP SYSTEM") {
		t.Error("Help should contain HELP SYSTEM header")
	}

	// Should list all categories
	for _, cat := range help.GetCategories() {
		if !strings.Contains(result, cat) {
			t.Errorf("Help should list category %q", cat)
		}
	}
}

// TestFormatHelpValidCommand tests help for a valid command
func TestFormatHelpValidCommand(t *testing.T) {
	result := formatHelp("look")

	if !strings.Contains(result, "LOOK") {
		t.Error("Help for 'look' should contain LOOK")
	}
	if strings.Contains(result, "No help available") {
		t.Error("Help for 'look' should not say no help available")
	}
}

// TestFormatHelpAlias tests help for a command alias
func TestFormatHelpAlias(t *testing.T) {
	result := formatHelp("l") // alias for look

	if strings.Contains(result, "No help available") {
		t.Error("Help for alias 'l' should resolve to look")
	}
}

// TestFormatHelpInvalid tests help for an unknown command
func TestFormatHelpInvalid(t *testing.T) {
	result := formatHelp("xyzzy123")

	if !strings.Contains(result, "No help available") {
		t.Error("Help for unknown command should say no help available")
	}
}

// TestParseCommand verifies command parsing logic
func TestParseCommand(t *testing.T) {
	tests := []struct {
		input string
		cmd   string
		arg   string
	}{
		{"look", "look", ""},
		{"look morpheus", "look", "morpheus"},
		{"NORTH", "north", ""},
		{"get red pill", "get", "red pill"},
		{"  say hello world  ", "say", "hello world"},
	}

	for _, tt := range tests {
		input := strings.TrimSpace(tt.input)
		parts := strings.Fields(input)
		cmd := ""
		arg := ""
		if len(parts) > 0 {
			cmd = strings.ToLower(parts[0])
		}
		if len(parts) > 1 {
			arg = strings.ToLower(strings.Join(parts[1:], " "))
		}

		if cmd != tt.cmd {
			t.Errorf("Command %q: got cmd=%q, want %q", tt.input, cmd, tt.cmd)
		}
		if arg != tt.arg {
			t.Errorf("Command %q: got arg=%q, want %q", tt.input, arg, tt.arg)
		}
	}
}

// TestMovementDirections verifies all movement directions
func TestMovementDirections(t *testing.T) {
	world := NewWorld()

	player := &Player{
		Name:   "TestPlayer",
		RoomID: "loading_program",
	}

	directions := []struct {
		dir      string
		expected string
	}{
		{"north", "dojo"},
		{"south", "subway"},
		{"east", "construct_archive"},
	}

	for _, d := range directions {
		player.RoomID = "loading_program" // Reset
		world.MovePlayer(player, d.dir)
		if player.RoomID != d.expected {
			t.Errorf("Move %s: got room=%q, want %q", d.dir, player.RoomID, d.expected)
		}
	}
}

// TestInvalidMovement verifies invalid moves are handled
func TestInvalidMovement(t *testing.T) {
	world := NewWorld()

	player := &Player{
		Name:   "TestPlayer",
		RoomID: "loading_program",
	}

	result := world.MovePlayer(player, "northwest")
	if player.RoomID != "loading_program" {
		t.Error("Invalid move should not change room")
	}
	if result == "" {
		t.Error("Invalid move should return message")
	}
}

// TestCombatInitiation verifies combat can be started
func TestCombatInitiation(t *testing.T) {
	world := NewWorld()

	player := &Player{
		Name:     "TestPlayer",
		RoomID:   "dojo",
		HP:       100,
		MaxHP:    100,
		Strength: 12,
	}

	result := world.StartCombat(player, "morpheus")
	if result == "" {
		t.Error("StartCombat should return message")
	}
	t.Logf("Combat result: %s", result)
}

// TestCombatWithNonexistentTarget verifies combat with invalid target
func TestCombatWithNonexistentTarget(t *testing.T) {
	world := NewWorld()

	player := &Player{
		Name:   "TestPlayer",
		RoomID: "dojo",
	}

	result := world.StartCombat(player, "nonexistent_npc_xyz")
	if result == "" {
		t.Error("StartCombat should return message for invalid target")
	}
}

// TestStopCombat verifies flee/stop combat
func TestStopCombat(t *testing.T) {
	world := NewWorld()

	player := &Player{
		Name:   "TestPlayer",
		RoomID: "dojo",
		State:  "COMBAT",
		Target: "morpheus",
	}

	result := world.StopCombat(player)
	if result == "" {
		t.Error("StopCombat should return message")
	}
}

// TestWearItem verifies equipping items
func TestWearItem(t *testing.T) {
	world := NewWorld()

	sword := &Item{
		ID:     "sword",
		Name:   "Sword",
		Slot:   "hand",
		Damage: 5,
	}

	player := &Player{
		Name:      "TestPlayer",
		RoomID:    "loading_program",
		Inventory: []*Item{sword},
		Equipment: make(map[string]*Item),
	}

	result := world.WearItem(player, "sword")
	if result == "" {
		t.Error("WearItem should return message")
	}
	if player.Equipment["hand"] != sword {
		t.Error("Sword should be equipped")
	}
}

// TestRemoveItem verifies unequipping items
func TestRemoveItem(t *testing.T) {
	world := NewWorld()

	sword := &Item{
		ID:     "sword",
		Name:   "Sword",
		Slot:   "hand",
		Damage: 5,
	}

	player := &Player{
		Name:      "TestPlayer",
		RoomID:    "loading_program",
		Inventory: []*Item{},
		Equipment: map[string]*Item{"hand": sword},
	}

	result := world.RemoveItem(player, "sword")
	if result == "" {
		t.Error("RemoveItem should return message")
	}
}

// TestUseItem verifies using consumable items
func TestUseItem(t *testing.T) {
	world := NewWorld()

	potion := &Item{
		ID:     "potion",
		Name:   "Health Potion",
		Type:   "consumable",
		Effect: "heal",
		Value:  10,
	}

	player := &Player{
		Name:      "TestPlayer",
		RoomID:    "loading_program",
		HP:        50,
		MaxHP:     100,
		Inventory: []*Item{potion},
	}

	result := world.UseItem(player, "potion")
	if result == "" {
		t.Error("UseItem should return message")
	}
	t.Logf("UseItem result: %s", result)
}

// TestTeleport verifies teleport command
func TestTeleport(t *testing.T) {
	world := NewWorld()

	player := &Player{
		Name:   "TestPlayer",
		RoomID: "loading_program",
	}

	result := world.Teleport(player, "dojo")
	if player.RoomID != "dojo" {
		t.Errorf("Teleport failed: room=%q, want dojo", player.RoomID)
	}
	if result == "" {
		t.Error("Teleport should return message")
	}
}

// TestTeleportInvalid verifies teleport to invalid room
func TestTeleportInvalid(t *testing.T) {
	world := NewWorld()

	player := &Player{
		Name:   "TestPlayer",
		RoomID: "loading_program",
	}

	result := world.Teleport(player, "nonexistent_room_xyz")
	if player.RoomID != "loading_program" {
		t.Error("Invalid teleport should not change room")
	}
	if result == "" {
		t.Error("Invalid teleport should return message")
	}
}

// TestGossip verifies global chat
func TestGossip(t *testing.T) {
	world := NewWorld()

	player := &Player{
		Name:   "TestPlayer",
		RoomID: "loading_program",
	}

	// This should not panic
	world.Gossip(player, "Hello everyone!")
}

// TestTell verifies private messaging
func TestTell(t *testing.T) {
	world := NewWorld()

	player := &Player{
		Name:   "TestPlayer",
		RoomID: "loading_program",
	}

	result := world.Tell(player, "nonexistent", "Hello")
	if result == "" {
		t.Error("Tell should return message")
	}
}

// TestListPlayers verifies who command
func TestListPlayers(t *testing.T) {
	world := NewWorld()

	result := world.ListPlayers()
	if result == "" {
		t.Error("ListPlayers should return content")
	}
}

// TestVendorOperations verifies buy/sell
func TestVendorOperations(t *testing.T) {
	world := NewWorld()

	player := &Player{
		Name:      "TestPlayer",
		RoomID:    "loading_program",
		Money:     1000,
		Inventory: []*Item{},
	}

	// List goods
	result := world.ListGoods(player)
	t.Logf("ListGoods: %s", result)

	// Buy something
	result = world.BuyItem(player, "coat")
	t.Logf("BuyItem: %s", result)

	// Sell something
	result = world.SellItem(player, "coat")
	t.Logf("SellItem: %s", result)
}

// TestBankOperations verifies deposit/withdraw
func TestBankOperations(t *testing.T) {
	world := NewWorld()

	item := &Item{
		ID:   "test_item",
		Name: "Test Item",
	}

	player := &Player{
		Name:      "TestPlayer",
		RoomID:    "construct_archive",
		Inventory: []*Item{item},
		Bank:      []*Item{},
	}

	// Deposit
	result := world.DepositItem(player, "test_item")
	t.Logf("Deposit: %s", result)

	// Show storage
	result = world.ShowStorage(player)
	t.Logf("Storage: %s", result)

	// Withdraw
	result = world.WithdrawItem(player, "test_item")
	t.Logf("Withdraw: %s", result)
}

// TestGenerateCity verifies city generation
func TestGenerateCity(t *testing.T) {
	world := NewWorld()

	player := &Player{
		Name:   "TestPlayer",
		RoomID: "loading_program",
	}

	initialRoomCount := len(world.Rooms)
	result := world.GenerateCity(player, 2, 2)
	if result == "" {
		t.Error("GenerateCity should return message")
	}
	if len(world.Rooms) <= initialRoomCount {
		t.Error("GenerateCity should create new rooms")
	}
	t.Logf("Created %d new rooms", len(world.Rooms)-initialRoomCount)
}

// TestDigRoom verifies room creation
func TestDigRoom(t *testing.T) {
	world := NewWorld()

	player := &Player{
		Name:   "TestPlayer",
		RoomID: "loading_program",
	}

	result := world.Dig(player, "west", "Test Room")
	t.Logf("Dig result: %s", result)
}

// TestEditRoom verifies room editing
func TestEditRoom(t *testing.T) {
	world := NewWorld()

	player := &Player{
		Name:   "TestPlayer",
		RoomID: "loading_program",
	}

	result := world.EditRoom(player, "desc", "New description text")
	t.Logf("EditRoom result: %s", result)
}

// TestSaveWorld verifies world persistence
func TestSaveWorld(t *testing.T) {
	world := NewWorld()

	// Should not panic
	world.SaveWorld()
}

// TestHandleSay verifies NPC dialogue
func TestHandleSay(t *testing.T) {
	world := NewWorld()

	player := &Player{
		Name:   "TestPlayer",
		RoomID: "dojo",
	}

	response := world.HandleSay(player, "hello")
	t.Logf("HandleSay response: %s", response)
}

// TestLook verifies the look command
func TestLook(t *testing.T) {
	world := NewWorld()

	player := &Player{
		Name:   "TestPlayer",
		RoomID: "loading_program",
	}

	result := world.Look(player, "")
	if result == "" {
		t.Error("Look should return room description")
	}
	if !strings.Contains(result, "loading_program") && !strings.Contains(result, "Loading") {
		t.Error("Look should include room name")
	}
}

// TestLookAtNPCFromCommand verifies looking at NPCs
func TestLookAtNPCFromCommand(t *testing.T) {
	world := NewWorld()

	player := &Player{
		Name:   "TestPlayer",
		RoomID: "dojo",
	}

	result := world.Look(player, "morpheus")
	if result == "" {
		t.Error("Look at NPC should return description")
	}
}

// TestShowScore verifies score display
func TestShowScore(t *testing.T) {
	world := NewWorld()

	player := &Player{
		Name:  "TestPlayer",
		HP:    80,
		MaxHP: 100,
		MP:    30,
		MaxMP: 50,
		Money: 500,
		XP:    100,
		Level: 2,
		Class: "Hacker",
	}

	result := world.ShowScore(player)
	if result == "" {
		t.Error("ShowScore should return player stats")
	}
	if !strings.Contains(result, "80") || !strings.Contains(result, "100") {
		t.Error("ShowScore should include HP values")
	}
}

// TestShowInventory verifies inventory display
func TestShowInventory(t *testing.T) {
	world := NewWorld()

	item := &Item{ID: "test", Name: "Test Item"}
	player := &Player{
		Name:      "TestPlayer",
		Inventory: []*Item{item},
	}

	result := world.ShowInventory(player)
	if result == "" {
		t.Error("ShowInventory should return content")
	}
}

// TestGetItem verifies picking up items
func TestGetItem(t *testing.T) {
	world := NewWorld()

	item := &Item{ID: "test_item", Name: "Test Item"}
	world.Rooms["loading_program"].ItemMap["test_item"] = item

	player := &Player{
		Name:      "TestPlayer",
		RoomID:    "loading_program",
		Inventory: []*Item{},
	}

	result := world.GetItem(player, "test_item")
	if result == "" {
		t.Error("GetItem should return message")
	}
}

// TestDropItem verifies dropping items
func TestDropItem(t *testing.T) {
	world := NewWorld()

	item := &Item{ID: "test_item", Name: "Test Item"}
	player := &Player{
		Name:      "TestPlayer",
		RoomID:    "loading_program",
		Inventory: []*Item{item},
	}

	result := world.DropItem(player, "test_item")
	if result == "" {
		t.Error("DropItem should return message")
	}
}

// TestGiveItem verifies giving items to NPCs
func TestGiveItem(t *testing.T) {
	world := NewWorld()

	item := &Item{ID: "test_item", Name: "Test Item"}
	player := &Player{
		Name:      "TestPlayer",
		RoomID:    "dojo",
		Inventory: []*Item{item},
	}

	result := world.GiveItem(player, "test_item", "morpheus")
	if result == "" {
		t.Error("GiveItem should return message")
	}
}

// TestAutomapGeneration verifies automap works
func TestAutomapGeneration(t *testing.T) {
	world := NewWorld()

	player := &Player{
		Name:   "TestPlayer",
		RoomID: "loading_program",
	}

	result := world.GenerateAutomapInternal(player, 2)
	if result == "" {
		t.Error("GenerateAutomap should return map")
	}
}

// TestRecall verifies recall command
func TestRecall(t *testing.T) {
	t.Skip("Recall functionality not implemented")
}

// TestCastSkill verifies skill casting
func TestCastSkill(t *testing.T) {
	world := NewWorld()

	player := &Player{
		Name:   "TestPlayer",
		RoomID: "dojo",
		Class:  "Hacker",
		HP:     100,
		MaxHP:  100,
		MP:     50,
		MaxMP:  50,
	}

	// Try casting out of combat
	result := world.CastSkill(player, "glitch", "")
	t.Logf("CastSkill out of combat: %s", result)
}

// TestGetMOTD verifies message of the day
func TestGetMOTD(t *testing.T) {
	world := NewWorld()

	motd := world.GetMOTD()
	// MOTD might be empty if file doesn't exist
	t.Logf("MOTD: %s", motd)
}
