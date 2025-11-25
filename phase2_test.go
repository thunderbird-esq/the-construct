package main

import (
	"testing"
)

func TestPhase2_NPCs(t *testing.T) {
	w := NewWorld()

	// 1. Check for Morpheus in the Dojo
	dojo := w.Rooms["dojo"]
	if _, ok := dojo.NPCMap["morpheus"]; !ok {
		t.Fatal("Morpheus is missing from the Dojo!")
	}

	// 2. Test Ticker (Simulated)
	w.Update()

	// 3. Test Look NPC
	p := &Player{Name: "Neo", RoomID: "dojo"}
	desc := w.Look(p, "morpheus")
	if len(desc) < 10 {
		t.Error("Looking at Morpheus returned too little text")
	}
	t.Logf("Look Morpheus Result: %s", desc)
}
