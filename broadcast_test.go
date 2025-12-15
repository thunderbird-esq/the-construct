package main

import (
	"testing"
)

// TestBroadcastNilSafety verifies broadcast handles nil values safely
func TestBroadcastNilSafety(t *testing.T) {
	world := NewWorld()

	// Test with nil sender - should return early
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("broadcast panicked with nil sender: %v", r)
		}
	}()

	broadcast(world, nil, "test message")
}

// TestBroadcastEmptyWorld verifies broadcast with no players
func TestBroadcastEmptyWorld(t *testing.T) {
	world := NewWorld()
	world.Players = make(map[*Client]*Player)

	sender := &Player{
		Name:   "Sender",
		RoomID: "dojo",
	}

	// Should not panic with empty world
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("broadcast panicked with empty world: %v", r)
		}
	}()

	broadcast(world, sender, "hello everyone")
}
