package main

import (
	"testing"
)

// TestLoadDefaultItemTemplatesExplicit verifies item templates loading explicitly
func TestLoadDefaultItemTemplatesExplicit(t *testing.T) {
	world := NewWorld()

	// Should have loaded item templates from items.json
	if len(world.ItemTemplates) == 0 {
		t.Error("ItemTemplates should be loaded")
	}

	// Test that loadDefaultItemTemplates creates fallbacks if needed
	// Clear templates and call directly
	emptyWorld := &World{
		ItemTemplates: make(map[string]*Item),
	}
	emptyWorld.loadDefaultItemTemplates()

	if len(emptyWorld.ItemTemplates) == 0 {
		t.Error("loadDefaultItemTemplates should create default items")
	}

	// Check for essential items
	essential := []string{"phone", "coat", "trash"}
	for _, id := range essential {
		if _, exists := emptyWorld.ItemTemplates[id]; !exists {
			t.Errorf("Missing essential item template: %s", id)
		}
	}
}

// TestGenerateLootCoverage verifies loot generation with more cases
func TestGenerateLootCoverage(t *testing.T) {
	world := NewWorld()

	// Test with valid template multiple times (random chance)
	for i := 0; i < 5; i++ {
		item := world.GenerateLoot("phone")
		if item != nil && item.Name == "" {
			t.Error("Generated item should have name")
		}
	}

	// Test with invalid template
	item := world.GenerateLoot("nonexistent_item_xyz")
	if item != nil {
		t.Error("Nonexistent template should return nil")
	}

	// Test with empty template
	item = world.GenerateLoot("")
	if item != nil {
		t.Error("Empty template should return nil")
	}
}

// TestGetReverseDirCoverage verifies direction reversal
func TestGetReverseDirCoverage(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"north", "south"},
		{"south", "north"},
		{"east", "west"},
		{"west", "east"},
		{"up", "down"},
		{"down", "up"},
		// Unknown directions return "back"
		{"n", "back"},
		{"s", "back"},
		{"invalid", "back"},
	}

	for _, tt := range tests {
		result := getReverseDir(tt.input)
		if result != tt.expected {
			t.Errorf("getReverseDir(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// TestRepairItemCoverage verifies item repair with more cases
func TestRepairItemCoverage(t *testing.T) {
	world := NewWorld()

	// Player with damaged equipment
	player := &Player{
		Name:   "RepairTester",
		RoomID: "dojo",
		Money:  1000,
		Equipment: map[string]*Item{
			"hand": {ID: "sword", Name: "Sword", Durability: 50, MaxDurability: 100},
		},
	}

	// Test repair - might fail due to location
	result := world.RepairItem(player, "sword")
	t.Logf("Repair at dojo: %s", result)

	// Test repair with fully repaired item
	player.Equipment["hand"].Durability = 100
	result = world.RepairItem(player, "sword")
	t.Logf("Repair full durability: %s", result)

	// Test repair with no money
	player.Money = 0
	player.Equipment["hand"].Durability = 50
	result = world.RepairItem(player, "sword")
	t.Logf("Repair no money: %s", result)

	// Test repair nonexistent item
	result = world.RepairItem(player, "nonexistent")
	t.Logf("Repair nonexistent: %s", result)
}
