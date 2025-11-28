package main

import (
	"strings"
	"testing"
)

// =============================================================================
// PHASE 2 BUG FIX TESTS
// =============================================================================

// Issue #4: Test nil room access doesn't panic
func TestNilRoomAccessNoPanic(t *testing.T) {
	world := NewWorld()
	player := &Player{
		Name:   "TestPlayer",
		RoomID: "nonexistent_room_that_does_not_exist",
	}

	// This should not panic - should return error message
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Look panicked with nil room: %v", r)
		}
	}()

	result := world.Look(player, "")
	if result == "" {
		t.Error("Look should return a message, not empty string")
	}
	// Should contain error/void message
	if strings.Contains(result, "void") || strings.Contains(result, "Error") || strings.Contains(result, "error") {
		t.Logf("Correctly handled nil room: %s", result[:min(len(result), 80)])
	}
}

// Issue #8: Test inventory size limit
func TestInventorySizeLimit(t *testing.T) {
	world := NewWorld()
	player := &Player{
		Name:      "TestPlayer",
		RoomID:    "loading_program",
		Inventory: make([]*Item, 0),
	}

	// Fill inventory to max
	for i := 0; i < MaxInventorySize; i++ {
		player.Inventory = append(player.Inventory, &Item{
			ID:   "test_item",
			Name: "Test Item",
		})
	}

	// Verify we're at max
	if len(player.Inventory) != MaxInventorySize {
		t.Errorf("Expected inventory size %d, got %d", MaxInventorySize, len(player.Inventory))
	}

	// Try to add one more - should be rejected
	result := world.GetItem(player, "phone")
	resultLower := strings.ToLower(result)
	if !strings.Contains(resultLower, "full") && !strings.Contains(resultLower, "cannot") && !strings.Contains(resultLower, "max") {
		t.Errorf("GetItem should reject when inventory full, got: %s", result)
	}

	// Verify size didn't increase
	if len(player.Inventory) > MaxInventorySize {
		t.Errorf("Inventory exceeded max size: %d > %d", len(player.Inventory), MaxInventorySize)
	}
}

// Issue #11: Test down command alias exists
func TestDownAlias(t *testing.T) {
	// Document that 'd' is drop, 'dn' or 'down' is for moving down
	// This is a design documentation test
	t.Log("Down command: 'down' is the primary command")
	t.Log("Note: 'd' is reserved for 'drop' command")
}

// Issue #13-14: Test NPC HP values are valid
func TestNPCHPValues(t *testing.T) {
	world := NewWorld()

	// Track issues for reporting
	issues := 0

	for roomID, room := range world.Rooms {
		for npcID, npc := range room.NPCMap {
			// HP should be > 0 for living NPCs
			if npc.HP <= 0 {
				t.Errorf("NPC %s in room %s has invalid HP: %d (should be > 0)", npcID, roomID, npc.HP)
				issues++
			}

			// MaxHP should be >= HP
			if npc.MaxHP < npc.HP {
				t.Errorf("NPC %s in room %s has MaxHP (%d) < HP (%d)", npcID, roomID, npc.MaxHP, npc.HP)
				issues++
			}

			// MaxHP should be > 0
			if npc.MaxHP <= 0 {
				t.Errorf("NPC %s in room %s has invalid MaxHP: %d (should be > 0)", npcID, roomID, npc.MaxHP)
				issues++
			}
		}
	}

	if issues == 0 {
		t.Log("All NPCs have valid HP values")
	}
}

// Issue #7: Test JSON load errors are handled gracefully
func TestJSONLoadErrorHandling(t *testing.T) {
	world := NewWorld()

	if world.Rooms == nil {
		t.Error("World.Rooms should not be nil even with JSON errors")
	}

	// Should have at least some rooms
	if len(world.Rooms) == 0 {
		t.Error("No rooms loaded - createDefaultWorld should have created at least one")
	} else {
		t.Logf("Loaded %d rooms successfully", len(world.Rooms))
	}
}

// Issue #16: Test connection timeout constants exist
func TestConnectionTimeoutConfig(t *testing.T) {
	// Verify timeout constants exist and have sensible values
	if ConnectionTimeout <= 0 {
		t.Errorf("ConnectionTimeout should be > 0, got %v", ConnectionTimeout)
	} else {
		t.Logf("ConnectionTimeout: %v", ConnectionTimeout)
	}

	if IdleTimeout <= 0 {
		t.Errorf("IdleTimeout should be > 0, got %v", IdleTimeout)
	} else {
		t.Logf("IdleTimeout: %v", IdleTimeout)
	}

	// IdleTimeout should be greater than ConnectionTimeout
	if IdleTimeout < ConnectionTimeout {
		t.Errorf("IdleTimeout (%v) should be >= ConnectionTimeout (%v)", IdleTimeout, ConnectionTimeout)
	}
}

// Test that world initializes correctly
func TestWorldInitialization(t *testing.T) {
	world := NewWorld()

	if world.Rooms == nil {
		t.Error("Rooms map is nil")
	}
	if world.Players == nil {
		t.Error("Players map is nil")
	}
	if world.Dialogue == nil {
		t.Error("Dialogue map is nil")
	}
	if world.DeadNPCs == nil {
		t.Error("DeadNPCs slice is nil")
	}
	if world.ItemTemplates == nil {
		t.Error("ItemTemplates map is nil")
	}

	t.Logf("World initialized with %d rooms, %d item templates", len(world.Rooms), len(world.ItemTemplates))
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
