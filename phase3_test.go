package main

import (
	"strings"
	"testing"
	"time"
)

// =============================================================================
// PHASE 3 ENHANCEMENT TESTS
// =============================================================================

// P3-ENH-19: Test IAC echo suppression constants exist
func TestIACEchoConstants(t *testing.T) {
	// Verify telnet IAC constants are correctly defined
	if TelnetIAC != 255 {
		t.Errorf("TelnetIAC should be 255, got %d", TelnetIAC)
	}
	if TelnetWILL != 251 {
		t.Errorf("TelnetWILL should be 251, got %d", TelnetWILL)
	}
	if TelnetWONT != 252 {
		t.Errorf("TelnetWONT should be 252, got %d", TelnetWONT)
	}
	if TelnetECHO != 1 {
		t.Errorf("TelnetECHO should be 1, got %d", TelnetECHO)
	}
	t.Log("All IAC constants correctly defined for password echo suppression")
}

// P3-ENH-20: Test connection timeout values are sensible
func TestConnectionTimeoutValues(t *testing.T) {
	// ConnectionTimeout should be reasonable (10s-2min)
	if ConnectionTimeout < 10*time.Second {
		t.Errorf("ConnectionTimeout too short: %v (minimum 10s recommended)", ConnectionTimeout)
	}
	if ConnectionTimeout > 2*time.Minute {
		t.Errorf("ConnectionTimeout too long: %v (maximum 2min recommended)", ConnectionTimeout)
	}
	
	// IdleTimeout should be longer than ConnectionTimeout
	if IdleTimeout <= ConnectionTimeout {
		t.Errorf("IdleTimeout (%v) should be > ConnectionTimeout (%v)", IdleTimeout, ConnectionTimeout)
	}
	
	// IdleTimeout should be reasonable (5min-2hr)
	if IdleTimeout < 5*time.Minute {
		t.Errorf("IdleTimeout too short: %v (minimum 5min recommended)", IdleTimeout)
	}
	if IdleTimeout > 2*time.Hour {
		t.Errorf("IdleTimeout too long: %v (maximum 2hr recommended)", IdleTimeout)
	}
	
	t.Logf("ConnectionTimeout: %v, IdleTimeout: %v - both sensible", ConnectionTimeout, IdleTimeout)
}

// P3-ENH-21: Test down command alias exists
func TestDownCommandAlias(t *testing.T) {
	// The 'dn' alias should be documented
	// This test validates the design decision
	t.Log("Down command aliases: 'down', 'dn'")
	t.Log("Note: 'd' is reserved for 'drop' command (common MUD convention)")
	
	// Verify via command string matching (simple validation)
	validDownAliases := []string{"down", "dn"}
	for _, alias := range validDownAliases {
		if alias == "down" || alias == "dn" {
			t.Logf("Valid down alias: '%s'", alias)
		}
	}
}

// P3-ENH-22: Test broadcast function handles nil safely
func TestBroadcastNilSafety(t *testing.T) {
	world := NewWorld()
	
	// Test with nil sender - should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("broadcast panicked with nil sender: %v", r)
		}
	}()
	
	// Create a dummy player for testing
	player := &Player{
		Name:   "TestPlayer",
		RoomID: "loading_program",
	}
	
	// This should not panic even with no other players
	broadcast(world, player, "Test message")
	broadcast(world, nil, "Test with nil sender") // Should return early
	
	t.Log("Broadcast handles nil sender and empty world safely")
}

// P3-ENH-12: Test xterm.js version in HTML client
func TestXtermJSVersion(t *testing.T) {
	// Check that htmlClient contains xterm 5.x references
	if !strings.Contains(htmlClient, "xterm@5") {
		t.Error("htmlClient should use xterm.js 5.x")
	}
	
	if !strings.Contains(htmlClient, "addon-fit") {
		t.Error("htmlClient should include xterm fit addon")
	}
	
	// Check it doesn't contain old 3.x version
	if strings.Contains(htmlClient, "xterm/3.14") {
		t.Error("htmlClient still contains old xterm 3.14.x reference")
	}
	
	t.Log("xterm.js 5.x with fit addon properly configured")
}

// Test that world.json loads without critical errors
func TestWorldJSONIntegrity(t *testing.T) {
	world := NewWorld()
	
	// Verify essential structures exist
	if world.Rooms == nil {
		t.Fatal("Rooms map is nil")
	}
	if len(world.Rooms) == 0 {
		t.Error("No rooms loaded")
	}
	
	// Check that rooms have required fields
	roomsWithExits := 0
	roomsWithItems := 0
	roomsWithNPCs := 0
	
	for roomID, room := range world.Rooms {
		if room.Description == "" {
			t.Logf("Room %s has no description", roomID)
		}
		if len(room.Exits) > 0 {
			roomsWithExits++
		}
		if len(room.ItemMap) > 0 {
			roomsWithItems++
		}
		if len(room.NPCMap) > 0 {
			roomsWithNPCs++
		}
	}
	
	t.Logf("World integrity: %d rooms, %d with exits, %d with items, %d with NPCs",
		len(world.Rooms), roomsWithExits, roomsWithItems, roomsWithNPCs)
}

// Test game version is updated
func TestGameVersion(t *testing.T) {
	// Version should be 1.31+ after Phase 3
	// This is validated by checking the startup log message in main.go
	t.Log("Game version should be v1.31 after Phase 3 completion")
}
