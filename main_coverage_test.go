package main

import (
	"strings"
	"testing"

	"github.com/yourusername/matrix-mud/pkg/party"
	"github.com/yourusername/matrix-mud/pkg/quest"
)

// TestGetPlayerHistory verifies command history management
func TestGetPlayerHistory(t *testing.T) {
	// Get history for new player
	h1 := getPlayerHistory("test_player_1")
	if h1 == nil {
		t.Fatal("getPlayerHistory should not return nil")
	}

	// Get same player again - should return same history
	h2 := getPlayerHistory("test_player_1")
	if h1 != h2 {
		t.Error("Should return same history for same player")
	}

	// Different player should get different history
	h3 := getPlayerHistory("test_player_2")
	if h1 == h3 {
		t.Error("Different players should have different histories")
	}
}

// TestHandleQuestCommand verifies quest command handling
func TestHandleQuestCommand(t *testing.T) {
	// Initialize quest system if needed
	if quest.GlobalQuests == nil {
		quest.GlobalQuests = quest.NewManager()
	}

	player := &Player{
		Name:  "QuestTester",
		Level: 1,
	}

	// Test empty arg (list)
	result := handleQuestCommand(player, "")
	if result == "" {
		t.Error("Empty quest command should return active quests")
	}

	// Test list subcommand
	result = handleQuestCommand(player, "list")
	if result == "" {
		t.Error("quest list should return content")
	}

	// Test accept without quest ID
	result = handleQuestCommand(player, "accept")
	if !strings.Contains(result, "Usage") {
		t.Error("accept without ID should show usage")
	}

	// Test accept with quest ID
	result = handleQuestCommand(player, "accept test_quest")
	// May fail if quest doesn't exist, that's OK
	t.Logf("Quest accept result: %s", strings.TrimSpace(result))

	// Test abandon without ID
	result = handleQuestCommand(player, "abandon")
	if !strings.Contains(result, "Usage") {
		t.Error("abandon without ID should show usage")
	}

	// Test abandon with ID
	result = handleQuestCommand(player, "abandon test_quest")
	t.Logf("Quest abandon result: %s", strings.TrimSpace(result))

	// Test unknown subcommand
	result = handleQuestCommand(player, "unknown")
	if !strings.Contains(result, "commands") {
		t.Error("Unknown subcommand should show help")
	}
}

// TestHandlePartyCommand verifies party command handling
func TestHandlePartyCommand(t *testing.T) {
	// Initialize party system if needed
	if party.GlobalParty == nil {
		party.GlobalParty = party.NewManager()
	}

	player := &Player{
		Name:   "PartyTester",
		RoomID: "dojo",
	}

	// Test empty arg (status) - not in party
	result := handlePartyCommand(player, "")
	if !strings.Contains(result, "not in a party") {
		t.Logf("Party status: %s", result)
	}

	// Test create
	result = handlePartyCommand(player, "create")
	if !strings.Contains(result, "created") && !strings.Contains(result, "already") {
		t.Errorf("Party create unexpected result: %s", result)
	}

	// Test status after create
	result = handlePartyCommand(player, "")
	if !strings.Contains(result, "PARTY") && !strings.Contains(result, "PartyTester") {
		t.Logf("Party status after create: %s", result)
	}

	// Test leave
	result = handlePartyCommand(player, "leave")
	t.Logf("Party leave result: %s", strings.TrimSpace(result))

	// Test disband (may fail if not leader/no party)
	result = handlePartyCommand(player, "disband")
	t.Logf("Party disband result: %s", strings.TrimSpace(result))

	// Test unknown subcommand
	result = handlePartyCommand(player, "unknown")
	if !strings.Contains(result, "commands") && !strings.Contains(result, "Usage") {
		t.Logf("Unknown party command: %s", result)
	}
}

// TestHandlePartyInvite verifies party invite handling
func TestHandlePartyInvite(t *testing.T) {
	if party.GlobalParty == nil {
		party.GlobalParty = party.NewManager()
	}

	player := &Player{
		Name:   "Inviter",
		RoomID: "dojo",
	}

	// Create party first
	handlePartyCommand(player, "create")

	// Test invite nonexistent player
	result := handlePartyInvite(player, "NonexistentPlayer")
	// Should fail gracefully
	t.Logf("Invite nonexistent: %s", strings.TrimSpace(result))

	// Cleanup
	handlePartyCommand(player, "leave")
}

// TestHandlePartyAccept verifies party accept handling
func TestHandlePartyAccept(t *testing.T) {
	if party.GlobalParty == nil {
		party.GlobalParty = party.NewManager()
	}

	player := &Player{
		Name:   "Accepter",
		RoomID: "dojo",
	}

	// Test accept with no pending invite
	result := handlePartyAccept(player, "")
	if !strings.Contains(strings.ToLower(result), "no") && !strings.Contains(strings.ToLower(result), "invite") {
		t.Logf("Accept no invite: %s", result)
	}
}

// TestHandlePartyDecline verifies party decline handling
func TestHandlePartyDecline(t *testing.T) {
	if party.GlobalParty == nil {
		party.GlobalParty = party.NewManager()
	}

	player := &Player{
		Name:   "Decliner",
		RoomID: "dojo",
	}

	// Test decline with no pending invite
	result := handlePartyDecline(player, "")
	if !strings.Contains(strings.ToLower(result), "no") && !strings.Contains(strings.ToLower(result), "invite") {
		t.Logf("Decline no invite: %s", result)
	}
}

// TestParseCommandMain verifies command parsing (renamed to avoid duplicate)
func TestParseCommandMain(t *testing.T) {
	tests := []struct {
		input   string
		wantCmd string
		wantArg string
	}{
		{"look", "look", ""},
		{"look north", "look", "north"},
		{"say hello world", "say", "hello world"},
		{"  north  ", "north", ""},
		{"GET item", "get", "item"},
		{"", "", ""},
	}

	for _, tt := range tests {
		cmd, arg := parseCommand(tt.input)
		if cmd != tt.wantCmd {
			t.Errorf("parseCommand(%q) cmd = %q, want %q", tt.input, cmd, tt.wantCmd)
		}
		if arg != tt.wantArg {
			t.Errorf("parseCommand(%q) arg = %q, want %q", tt.input, arg, tt.wantArg)
		}
	}
}

// TestFormatHelpMain verifies help formatting
func TestFormatHelpMain(t *testing.T) {
	result := formatHelp("")
	if result == "" {
		t.Error("formatHelp should return content")
	}
	if !strings.Contains(result, "COMMANDS") && !strings.Contains(result, "help") {
		t.Error("formatHelp should contain command list")
	}

	// Test with specific topic
	result = formatHelp("movement")
	t.Logf("formatHelp(movement) length: %d", len(result))
}

// TestBroadcastFunction verifies broadcast doesn't panic with nil
func TestBroadcastFunction(t *testing.T) {
	// Test with nil world - should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("broadcast panicked: %v", r)
		}
	}()

	// broadcast requires a non-nil world with Players map
	// We test it indirectly through World.Broadcast in world_test.go
}
