package main

import (
	"testing"
)

// TestPhase1 verifies items can be picked up and dropped
func TestPhase1_Inventory(t *testing.T) {
	w := NewWorld()
	p := &Player{
		Name:      "Tester",
		RoomID:    "loading_program",
		Inventory: make([]*Item, 0),
	}

	// 1. Verify 'phone' is in loading_program
	room := w.Rooms["loading_program"]
	if _, ok := room.ItemMap["phone"]; !ok {
		t.Fatal("Expected phone in loading_program, found nothing")
	}

	// 2. Pick up phone
	res := w.GetItem(p, "phone")
	if len(p.Inventory) != 1 {
		t.Errorf("Inventory count expected 1, got %d", len(p.Inventory))
	}
	if _, ok := room.ItemMap["phone"]; ok {
		t.Error("Phone should be gone from room, but it's still there")
	}
	t.Logf("Get Result: %s", res)

	// 3. Drop phone
	res = w.DropItem(p, "phone")
	if len(p.Inventory) != 0 {
		t.Errorf("Inventory count expected 0, got %d", len(p.Inventory))
	}
	if _, ok := room.ItemMap["phone"]; !ok {
		t.Error("Phone should be back in room, but it's not")
	}
	t.Logf("Drop Result: %s", res)
}
