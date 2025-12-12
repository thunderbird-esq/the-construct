package faction

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if len(m.Factions) != 3 {
		t.Errorf("Expected 3 factions, got %d", len(m.Factions))
	}
}

func TestFactionIDs(t *testing.T) {
	if FactionZion != "zion" {
		t.Error("FactionZion should be 'zion'")
	}
	if FactionMachines != "machines" {
		t.Error("FactionMachines should be 'machines'")
	}
	if FactionExiles != "exiles" {
		t.Error("FactionExiles should be 'exiles'")
	}
}

func TestGetPlayerFaction(t *testing.T) {
	m := NewManager()
	pf := m.GetPlayerFaction("testplayer")

	if pf == nil {
		t.Fatal("GetPlayerFaction returned nil")
	}
	if pf.Faction != FactionNone {
		t.Error("New player should have no faction")
	}
	if pf.Reputation[FactionZion] != 0 {
		t.Error("New player should have 0 reputation")
	}
}

func TestJoin(t *testing.T) {
	m := NewManager()

	msg, ok := m.Join("player1", FactionZion)
	if !ok {
		t.Errorf("Join failed: %s", msg)
	}

	pf := m.GetPlayerFaction("player1")
	if pf.Faction != FactionZion {
		t.Error("Player should be in Zion")
	}
	if pf.Reputation[FactionZion] < 100 {
		t.Error("Should have reputation bonus for joining")
	}
}

func TestJoinAlreadyInFaction(t *testing.T) {
	m := NewManager()

	m.Join("player1", FactionZion)
	_, ok := m.Join("player1", FactionMachines)
	if ok {
		t.Error("Should not be able to join second faction")
	}
}

func TestJoinInvalidFaction(t *testing.T) {
	m := NewManager()

	_, ok := m.Join("player1", FactionID("invalid"))
	if ok {
		t.Error("Should not be able to join invalid faction")
	}
}

func TestLeave(t *testing.T) {
	m := NewManager()

	m.Join("player1", FactionZion)
	msg, ok := m.Leave("player1")
	if !ok {
		t.Errorf("Leave failed: %s", msg)
	}

	pf := m.GetPlayerFaction("player1")
	if pf.Faction != FactionNone {
		t.Error("Player should have no faction after leaving")
	}
}

func TestLeaveNotInFaction(t *testing.T) {
	m := NewManager()

	_, ok := m.Leave("player1")
	if ok {
		t.Error("Should not be able to leave when not in faction")
	}
}

func TestAdjustReputation(t *testing.T) {
	m := NewManager()

	m.AdjustReputation("player1", FactionZion, 50)
	rep := m.GetReputation("player1", FactionZion)
	if rep != 50 {
		t.Errorf("Reputation should be 50, got %d", rep)
	}

	// Test opposing faction effect
	machineRep := m.GetReputation("player1", FactionMachines)
	if machineRep >= 0 {
		t.Error("Machines rep should decrease when Zion increases")
	}
}

func TestReputationClamping(t *testing.T) {
	m := NewManager()

	m.AdjustReputation("player1", FactionZion, 10000)
	rep := m.GetReputation("player1", FactionZion)
	if rep > RepExalted {
		t.Error("Reputation should be clamped to Exalted")
	}

	m.AdjustReputation("player1", FactionMachines, -10000)
	rep = m.GetReputation("player1", FactionMachines)
	if rep < RepHated {
		t.Error("Reputation should be clamped to Hated")
	}
}

func TestGetStandingName(t *testing.T) {
	tests := []struct {
		rep      int
		expected string
	}{
		{RepExalted, "Exalted"},
		{RepHonored, "Honored"},
		{RepFriendly, "Friendly"},
		{0, "Neutral"},
		{RepUnfriendly, "Unfriendly"},
		{RepHostile, "Hostile"},
		{RepHated, "Hated"},
	}

	for _, tt := range tests {
		got := GetStandingName(tt.rep)
		if got != tt.expected {
			t.Errorf("GetStandingName(%d) = %s, want %s", tt.rep, got, tt.expected)
		}
	}
}

func TestIsSameFaction(t *testing.T) {
	m := NewManager()

	m.Join("player1", FactionZion)
	m.Join("player2", FactionZion)
	m.Join("player3", FactionMachines)

	if !m.IsSameFaction("player1", "player2") {
		t.Error("Players in same faction should return true")
	}
	if m.IsSameFaction("player1", "player3") {
		t.Error("Players in different factions should return false")
	}
	if m.IsSameFaction("player1", "nonexistent") {
		t.Error("Nonexistent player should return false")
	}
}

func TestGetFaction(t *testing.T) {
	m := NewManager()

	f := m.GetFaction(FactionZion)
	if f == nil {
		t.Fatal("GetFaction returned nil")
	}
	if f.Name != "Zion" {
		t.Errorf("Name = %s, want Zion", f.Name)
	}
	if f.Leader != "Morpheus" {
		t.Errorf("Leader = %s, want Morpheus", f.Leader)
	}
}

func TestGetAllFactions(t *testing.T) {
	m := NewManager()

	factions := m.GetAllFactions()
	if len(factions) != 3 {
		t.Errorf("Expected 3 factions, got %d", len(factions))
	}
}
