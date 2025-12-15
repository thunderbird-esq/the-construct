package main

import (
	"strings"
	"testing"

	"github.com/yourusername/matrix-mud/pkg/dialogue"
	"github.com/yourusername/matrix-mud/pkg/instance"
)

func TestHandleTalkCommand(t *testing.T) {
	world := NewWorld()
	player := &Player{
		Name:   "TestPlayer",
		RoomID: "dojo", // Morpheus is here
	}

	// Empty argument
	result := HandleTalkCommand(world, player, "")
	if !strings.Contains(result, "whom") {
		t.Errorf("Should ask who to talk to: %s", result)
	}

	// Unknown NPC
	result = HandleTalkCommand(world, player, "nonexistent")
	if !strings.Contains(result, "no") {
		t.Errorf("Should indicate NPC not found: %s", result)
	}

	// Valid NPC (Morpheus in dojo)
	result = HandleTalkCommand(world, player, "morpheus")
	if result == "" {
		t.Error("Should return dialogue")
	}

	// Clean up dialogue session
	dialogue.GlobalDialogue.EndDialogue(player.Name)
}

func TestHandleDialogueChoice(t *testing.T) {
	world := NewWorld()
	player := &Player{
		Name:   "TestPlayer2",
		RoomID: "dojo",
	}

	// Not in dialogue
	result := HandleDialogueChoice(world, player, "1")
	if result != "" {
		t.Error("Should return empty when not in dialogue")
	}

	// Start dialogue first
	dialogue.GlobalDialogue.StartDialogue(player.Name, "morpheus")

	// Valid choice
	result = HandleDialogueChoice(world, player, "1")
	if result == "" {
		t.Error("Should return next dialogue node")
	}

	// Clean up
	dialogue.GlobalDialogue.EndDialogue(player.Name)
}

func TestHandleByeCommand(t *testing.T) {
	player := &Player{Name: "TestPlayer3"}

	// Not in dialogue
	result := HandleByeCommand(player)
	if result != "" {
		t.Error("Should return empty when not in dialogue")
	}

	// Start dialogue
	dialogue.GlobalDialogue.StartDialogue(player.Name, "morpheus")

	// Now say bye
	result = HandleByeCommand(player)
	if !strings.Contains(result, "end") {
		t.Errorf("Should indicate conversation ended: %s", result)
	}

	if dialogue.GlobalDialogue.IsInDialogue(player.Name) {
		t.Error("Should no longer be in dialogue")
	}
}

func TestIsInDialogue(t *testing.T) {
	playerName := "TestPlayer4"

	if IsInDialogue(playerName) {
		t.Error("Should not be in dialogue initially")
	}

	dialogue.GlobalDialogue.StartDialogue(playerName, "morpheus")

	if !IsInDialogue(playerName) {
		t.Error("Should be in dialogue after starting")
	}

	dialogue.GlobalDialogue.EndDialogue(playerName)
}

func TestIsInInstance(t *testing.T) {
	playerName := "TestPlayer5"

	if IsInInstance(playerName) {
		t.Error("Should not be in instance initially")
	}
}

func TestHandleInstanceCommand(t *testing.T) {
	world := NewWorld()
	player := &Player{
		Name:  "TestPlayer6",
		Level: 5,
	}

	// List command
	result := HandleInstanceCommand(world, player, "list")
	if !strings.Contains(result, "INSTANCE") || !strings.Contains(result, "training") {
		t.Errorf("Should list available instances: %s", result)
	}

	// Help (empty command)
	result = HandleInstanceCommand(world, player, "")
	if !strings.Contains(result, "INSTANCE COMMANDS") {
		t.Errorf("Should show help: %s", result)
	}

	// Create instance
	result = HandleInstanceCommand(world, player, "create training_gauntlet")
	if !strings.Contains(result, "ENTERING") {
		t.Errorf("Should enter instance: %s", result)
	}

	// Look in instance
	result = HandleInstanceCommand(world, player, "look")
	if !strings.Contains(result, "Trial") {
		t.Errorf("Should show instance room: %s", result)
	}

	// Leave instance
	result = HandleInstanceCommand(world, player, "leave")
	if !strings.Contains(result, "leave") {
		t.Errorf("Should confirm leaving: %s", result)
	}
}

func TestHandleInstanceMove(t *testing.T) {
	player := &Player{Name: "TestPlayer7", Level: 5}

	// Not in instance
	result, handled := HandleInstanceMove(player, "north")
	if handled {
		t.Error("Should not handle when not in instance")
	}
	if result != "" {
		t.Error("Should return empty when not in instance")
	}
}

func TestHandleInstanceAttack(t *testing.T) {
	world := NewWorld()
	player := &Player{
		Name:     "TestPlayer8",
		Level:    5,
		Strength: 10,
	}

	// Not in instance
	result, handled := HandleInstanceAttack(world, player, "target")
	if handled {
		t.Error("Should not handle when not in instance")
	}
	if result != "" {
		t.Error("Should return empty when not in instance")
	}
}

func TestHandleInstanceLook(t *testing.T) {
	player := &Player{Name: "TestPlayer9"}

	// Not in instance
	result, handled := HandleInstanceLook(player)
	if handled {
		t.Error("Should not handle when not in instance")
	}
	if result != "" {
		t.Error("Should return empty when not in instance")
	}

	// Create instance and look
	instance.GlobalInstance.CreateInstance("training_gauntlet", player.Name, 5)
	result, handled = HandleInstanceLook(player)
	if !handled {
		t.Error("Should handle when in instance")
	}
	if !strings.Contains(result, "Trial") {
		t.Errorf("Should show room info: %s", result)
	}

	// Clean up
	instance.GlobalInstance.LeaveInstance(player.Name)
}

func TestCheckLevelUp(t *testing.T) {
	player := &Player{
		Level: 1,
		XP:    0,
		HP:    100,
		MaxHP: 100,
	}

	// No level up
	checkLevelUp(player)
	if player.Level != 1 {
		t.Error("Should not level up with 0 XP")
	}

	// Set XP for level up
	player.XP = 100
	checkLevelUp(player)
	if player.Level != 2 {
		t.Errorf("Should level up to 2, got %d", player.Level)
	}
	if player.MaxHP != 110 {
		t.Errorf("MaxHP should increase to 110, got %d", player.MaxHP)
	}

	// Multiple level ups
	player.XP = 1000
	checkLevelUp(player)
	if player.Level <= 2 {
		t.Error("Should have leveled up multiple times")
	}
}

func TestInstanceHelp(t *testing.T) {
	help := instanceHelp()
	if !strings.Contains(help, "INSTANCE COMMANDS") {
		t.Error("Should contain header")
	}
	if !strings.Contains(help, "list") {
		t.Error("Should mention list command")
	}
	if !strings.Contains(help, "create") {
		t.Error("Should mention create command")
	}
}
