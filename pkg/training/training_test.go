package training

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if len(m.Programs) == 0 {
		t.Error("Programs should be loaded")
	}
}

func TestListPrograms(t *testing.T) {
	m := NewManager()
	programs := m.ListPrograms()
	if len(programs) == 0 {
		t.Error("Should have programs")
	}
}

func TestGetProgram(t *testing.T) {
	m := NewManager()

	p := m.GetProgram("combat_basic")
	if p == nil {
		t.Fatal("GetProgram returned nil")
	}
	if p.Name != "Basic Combat Training" {
		t.Errorf("Name = %s, want Basic Combat Training", p.Name)
	}
}

func TestStartProgram(t *testing.T) {
	m := NewManager()

	instance, err := m.StartProgram("player1", "combat_basic")
	if err != nil {
		t.Fatalf("StartProgram failed: %v", err)
	}
	if instance == nil {
		t.Fatal("Instance is nil")
	}
	if !instance.IsActive {
		t.Error("Instance should be active")
	}
	if _, ok := instance.Players["player1"]; !ok {
		t.Error("Player should be in instance")
	}
}

func TestStartProgramAlreadyIn(t *testing.T) {
	m := NewManager()

	m.StartProgram("player1", "combat_basic")
	_, err := m.StartProgram("player1", "combat_advanced")
	if err == nil {
		t.Error("Should error when already in program")
	}
}

func TestStartProgramInvalid(t *testing.T) {
	m := NewManager()

	_, err := m.StartProgram("player1", "nonexistent")
	if err == nil {
		t.Error("Should error for invalid program")
	}
}

func TestJoinProgram(t *testing.T) {
	m := NewManager()

	// Start PvP program
	instance, _ := m.StartProgram("player1", "pvp_arena")

	err := m.JoinProgram("player2", instance.ID)
	if err != nil {
		t.Fatalf("JoinProgram failed: %v", err)
	}

	if len(instance.Players) != 2 {
		t.Errorf("Expected 2 players, got %d", len(instance.Players))
	}
}

func TestJoinProgramNonPvP(t *testing.T) {
	m := NewManager()

	instance, _ := m.StartProgram("player1", "combat_basic")

	err := m.JoinProgram("player2", instance.ID)
	if err == nil {
		t.Error("Should error when joining non-PvP program")
	}
}

func TestLeaveProgram(t *testing.T) {
	m := NewManager()

	m.StartProgram("player1", "combat_basic")

	err := m.LeaveProgram("player1")
	if err != nil {
		t.Fatalf("LeaveProgram failed: %v", err)
	}

	if m.IsInProgram("player1") {
		t.Error("Player should not be in program")
	}
}

func TestLeaveProgramNotIn(t *testing.T) {
	m := NewManager()

	err := m.LeaveProgram("player1")
	if err == nil {
		t.Error("Should error when not in program")
	}
}

func TestIsInProgram(t *testing.T) {
	m := NewManager()

	if m.IsInProgram("player1") {
		t.Error("Should not be in program initially")
	}

	m.StartProgram("player1", "combat_basic")

	if !m.IsInProgram("player1") {
		t.Error("Should be in program after starting")
	}
}

func TestGetPlayerInstance(t *testing.T) {
	m := NewManager()

	m.StartProgram("player1", "combat_basic")

	instance := m.GetPlayerInstance("player1")
	if instance == nil {
		t.Fatal("GetPlayerInstance returned nil")
	}

	instance2 := m.GetPlayerInstance("player2")
	if instance2 != nil {
		t.Error("Should return nil for player not in program")
	}
}

func TestRecordScore(t *testing.T) {
	m := NewManager()

	m.StartProgram("player1", "combat_basic")
	m.RecordScore("player1", 100)
	m.RecordScore("player1", 50)

	instance := m.GetPlayerInstance("player1")
	if instance.Scores["player1"] != 150 {
		t.Errorf("Score = %d, want 150", instance.Scores["player1"])
	}
}

func TestCompleteProgram(t *testing.T) {
	m := NewManager()

	m.StartProgram("player1", "combat_basic")
	m.RecordScore("player1", 100)

	rewards, score, err := m.CompleteProgram("player1")
	if err != nil {
		t.Fatalf("CompleteProgram failed: %v", err)
	}
	if rewards == nil {
		t.Fatal("Rewards is nil")
	}
	if score != 100 {
		t.Errorf("Score = %d, want 100", score)
	}
	if m.IsInProgram("player1") {
		t.Error("Player should be removed from program")
	}
}

func TestCompleteProgramNotIn(t *testing.T) {
	m := NewManager()

	_, _, err := m.CompleteProgram("player1")
	if err == nil {
		t.Error("Should error when not in program")
	}
}

func TestListChallenges(t *testing.T) {
	m := NewManager()
	challenges := m.ListChallenges()
	if len(challenges) == 0 {
		t.Error("Should have challenges")
	}
}

func TestGetChallenge(t *testing.T) {
	m := NewManager()

	c := m.GetChallenge("speed_trial")
	if c == nil {
		t.Fatal("GetChallenge returned nil")
	}
	if c.Name != "Speed Trial" {
		t.Errorf("Name = %s, want Speed Trial", c.Name)
	}
}

func TestPvPArenaFull(t *testing.T) {
	m := NewManager()

	instance, _ := m.StartProgram("player1", "pvp_arena")
	m.JoinProgram("player2", instance.ID)

	err := m.JoinProgram("player3", instance.ID)
	if err == nil {
		t.Error("Should error when arena is full")
	}
}

func TestProgramTypes(t *testing.T) {
	if ProgramCombat != "combat" {
		t.Error("ProgramCombat should be 'combat'")
	}
	if ProgramSurvival != "survival" {
		t.Error("ProgramSurvival should be 'survival'")
	}
	if ProgramPvP != "pvp" {
		t.Error("ProgramPvP should be 'pvp'")
	}
	if ProgramTrial != "trial" {
		t.Error("ProgramTrial should be 'trial'")
	}
}
