package instance

import (
	"strings"
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.Templates == nil {
		t.Error("Templates map is nil")
	}
	if m.Instances == nil {
		t.Error("Instances map is nil")
	}
}

func TestDefaultTemplates(t *testing.T) {
	m := NewManager()

	// Check default templates exist
	templates := []string{"gov_raid", "club_depths", "training_gauntlet"}
	for _, id := range templates {
		if _, ok := m.Templates[id]; !ok {
			t.Errorf("Missing template: %s", id)
		}
	}
}

func TestCreateInstance(t *testing.T) {
	m := NewManager()

	// Create training gauntlet (level 1 required)
	instance, err := m.CreateInstance("training_gauntlet", "TestPlayer", 5)
	if err != "" {
		t.Fatalf("CreateInstance failed: %s", err)
	}
	if instance == nil {
		t.Fatal("Instance is nil")
	}
	if instance.Owner != "TestPlayer" {
		t.Errorf("Wrong owner: %s", instance.Owner)
	}
	if instance.State != "active" {
		t.Errorf("Wrong state: %s", instance.State)
	}
}

func TestCreateInstanceLevelRequirement(t *testing.T) {
	m := NewManager()

	// Try to create gov_raid (level 3 required) at level 1
	_, err := m.CreateInstance("gov_raid", "LowLevel", 1)
	if err == "" {
		t.Error("Should fail level requirement")
	}
	if !strings.Contains(err, "level") {
		t.Errorf("Error should mention level: %s", err)
	}
}

func TestCreateInstanceInvalidTemplate(t *testing.T) {
	m := NewManager()

	_, err := m.CreateInstance("invalid_template", "TestPlayer", 10)
	if err == "" {
		t.Error("Should fail for invalid template")
	}
}

func TestGetPlayerInstance(t *testing.T) {
	m := NewManager()

	// Not in instance
	inst := m.GetPlayerInstance("NoInstance")
	if inst != nil {
		t.Error("Should return nil when not in instance")
	}

	// Create instance
	m.CreateInstance("training_gauntlet", "TestPlayer", 5)
	inst = m.GetPlayerInstance("TestPlayer")
	if inst == nil {
		t.Error("Should return instance")
	}
}

func TestMoveInInstance(t *testing.T) {
	m := NewManager()

	// Create instance
	m.CreateInstance("training_gauntlet", "TestPlayer", 5)

	// Try to move before clearing
	msg, ok := m.MoveInInstance("TestPlayer", "north")
	if ok {
		t.Error("Should not be able to move before clearing enemies")
	}
	if !strings.Contains(msg, "defeat") {
		t.Errorf("Message should mention defeating enemies: %s", msg)
	}
}

func TestAttackInInstance(t *testing.T) {
	m := NewManager()

	// Create instance
	m.CreateInstance("training_gauntlet", "TestPlayer", 5)

	// Attack training bot
	msg, ok := m.AttackInInstance("TestPlayer", "training", 100)
	if !ok {
		t.Errorf("Attack failed: %s", msg)
	}
	if !strings.Contains(msg, "defeat") && !strings.Contains(msg, "hit") {
		t.Errorf("Unexpected message: %s", msg)
	}
}

func TestAttackInvalidTarget(t *testing.T) {
	m := NewManager()

	m.CreateInstance("training_gauntlet", "TestPlayer", 5)

	msg, ok := m.AttackInInstance("TestPlayer", "nonexistent", 10)
	if ok {
		t.Error("Should fail for invalid target")
	}
	if !strings.Contains(msg, "not found") {
		t.Errorf("Message should say not found: %s", msg)
	}
}

func TestLeaveInstance(t *testing.T) {
	m := NewManager()

	m.CreateInstance("training_gauntlet", "TestPlayer", 5)
	if !m.IsInInstance("TestPlayer") {
		t.Error("Should be in instance")
	}

	m.LeaveInstance("TestPlayer")
	if m.IsInInstance("TestPlayer") {
		t.Error("Should not be in instance after leaving")
	}
}

func TestLeaveInstanceNotIn(t *testing.T) {
	m := NewManager()

	msg := m.LeaveInstance("NotInInstance")
	if !strings.Contains(msg, "not in") {
		t.Errorf("Message should say not in instance: %s", msg)
	}
}

func TestListTemplates(t *testing.T) {
	m := NewManager()

	output := m.ListTemplates(10)
	if !strings.Contains(output, "Training Gauntlet") {
		t.Error("Should list training gauntlet")
	}
	if !strings.Contains(output, "Government Building") {
		t.Error("Should list gov raid")
	}
}

func TestLookInInstance(t *testing.T) {
	m := NewManager()

	// Not in instance
	output := m.LookInInstance("NoInstance")
	if output != "" {
		t.Error("Should return empty when not in instance")
	}

	// In instance
	m.CreateInstance("training_gauntlet", "TestPlayer", 5)
	output = m.LookInInstance("TestPlayer")
	if !strings.Contains(output, "Trial") {
		t.Error("Should show room name")
	}
	if !strings.Contains(output, "ENEMY") {
		t.Error("Should show enemies")
	}
}

func TestIsInInstance(t *testing.T) {
	m := NewManager()

	if m.IsInInstance("Nobody") {
		t.Error("Should not be in instance")
	}

	m.CreateInstance("training_gauntlet", "TestPlayer", 5)
	if !m.IsInInstance("TestPlayer") {
		t.Error("Should be in instance")
	}
	if !m.IsInInstance("testplayer") { // case insensitive
		t.Error("Should be case insensitive")
	}
}

func TestDuplicateInstance(t *testing.T) {
	m := NewManager()

	// Create first instance
	m.CreateInstance("training_gauntlet", "TestPlayer", 5)

	// Try to create second
	_, err := m.CreateInstance("gov_raid", "TestPlayer", 5)
	if err == "" {
		t.Error("Should not allow second instance")
	}
	if !strings.Contains(err, "already") {
		t.Errorf("Error should mention already in instance: %s", err)
	}
}

func TestGetRewards(t *testing.T) {
	m := NewManager()

	// Not in instance
	_, err := m.GetRewards("Nobody")
	if err == "" {
		t.Error("Should fail when not in instance")
	}

	// In instance but not completed
	m.CreateInstance("training_gauntlet", "TestPlayer", 5)
	_, err = m.GetRewards("TestPlayer")
	if err == "" || !strings.Contains(err, "not completed") {
		t.Error("Should fail when instance not completed")
	}
}

func TestDifficultyName(t *testing.T) {
	tests := []struct {
		diff Difficulty
		want string
	}{
		{DiffEasy, "Easy"},
		{DiffNormal, "Normal"},
		{DiffHard, "Hard"},
		{DiffBoss, "Boss"},
		{Difficulty(99), "Unknown"},
	}

	for _, tt := range tests {
		got := difficultyName(tt.diff)
		if got != tt.want {
			t.Errorf("difficultyName(%d) = %s, want %s", tt.diff, got, tt.want)
		}
	}
}

func TestFormatNPCName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"security_guard", "Security Guard"},
		{"riot_cop", "Riot Cop"},
		{"agent", "Agent"},
	}

	for _, tt := range tests {
		got := formatNPCName(tt.input)
		if got != tt.want {
			t.Errorf("formatNPCName(%s) = %s, want %s", tt.input, got, tt.want)
		}
	}
}

func TestGetNPCHP(t *testing.T) {
	// Agent should have higher HP
	agentHP := getNPCHP("agent", DiffNormal)
	guardHP := getNPCHP("guard", DiffNormal)

	if agentHP <= guardHP {
		t.Error("Agent should have more HP than guard")
	}

	// Difficulty should scale HP
	easyHP := getNPCHP("guard", DiffEasy)
	hardHP := getNPCHP("guard", DiffHard)
	if hardHP <= easyHP {
		t.Error("Hard difficulty should have more HP")
	}
}

func TestGetNPCDamage(t *testing.T) {
	// Agent should have higher damage
	agentDmg := getNPCDamage("agent", DiffNormal)
	botDmg := getNPCDamage("bot", DiffNormal)

	if agentDmg <= botDmg {
		t.Error("Agent should have more damage than bot")
	}
}

func TestInstanceRoomClearing(t *testing.T) {
	m := NewManager()
	m.CreateInstance("training_gauntlet", "TestPlayer", 5)

	// Kill all enemies in first room (one training bot)
	for i := 0; i < 10; i++ {
		msg, _ := m.AttackInInstance("TestPlayer", "training", 100)
		if strings.Contains(msg, "cleared") {
			break
		}
	}

	// Should now be able to move
	msg, ok := m.MoveInInstance("TestPlayer", "north")
	if !ok {
		t.Errorf("Should be able to move after clearing: %s", msg)
	}
}

func TestGetCurrentRoom(t *testing.T) {
	m := NewManager()

	// Not in instance
	room := m.GetCurrentRoom("Nobody")
	if room != nil {
		t.Error("Should return nil when not in instance")
	}

	// In instance
	m.CreateInstance("training_gauntlet", "TestPlayer", 5)
	room = m.GetCurrentRoom("TestPlayer")
	if room == nil {
		t.Error("Should return current room")
	}
	if room.Name != "First Trial" {
		t.Errorf("Wrong room name: %s", room.Name)
	}
}
