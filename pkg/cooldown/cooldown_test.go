package cooldown

import (
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.cooldowns == nil {
		t.Error("cooldowns map not initialized")
	}
}

func TestUseAndIsReady(t *testing.T) {
	m := NewManager()
	
	// Should be ready before use
	if !m.IsReady("player1", "glitch") {
		t.Error("Ability should be ready before first use")
	}
	
	// Use the ability
	m.Use("player1", "glitch")
	
	// Should not be ready immediately after
	if m.IsReady("player1", "glitch") {
		t.Error("Ability should not be ready immediately after use")
	}
}

func TestTimeRemaining(t *testing.T) {
	m := NewManager()
	
	// No cooldown if not used
	if m.TimeRemaining("player1", "glitch") != 0 {
		t.Error("TimeRemaining should be 0 for unused ability")
	}
	
	m.Use("player1", "glitch")
	
	remaining := m.TimeRemaining("player1", "glitch")
	if remaining <= 0 {
		t.Error("TimeRemaining should be positive after use")
	}
	
	cooldown := AbilityCooldowns["glitch"]
	if remaining > cooldown {
		t.Errorf("TimeRemaining = %v, should not exceed cooldown %v", remaining, cooldown)
	}
}

func TestCooldownExpires(t *testing.T) {
	m := NewManager()
	
	// Use an ability with short cooldown
	m.Use("player1", "use") // 2 second cooldown
	
	if m.IsReady("player1", "use") {
		t.Error("Should not be ready immediately")
	}
	
	// Wait for cooldown
	time.Sleep(2100 * time.Millisecond)
	
	if !m.IsReady("player1", "use") {
		t.Error("Should be ready after cooldown expires")
	}
}

func TestGetAllCooldowns(t *testing.T) {
	m := NewManager()
	
	// Empty at start
	cds := m.GetAllCooldowns("player1")
	if len(cds) != 0 {
		t.Errorf("Should have 0 cooldowns, got %d", len(cds))
	}
	
	// Use some abilities
	m.Use("player1", "glitch")
	m.Use("player1", "patch")
	
	cds = m.GetAllCooldowns("player1")
	if len(cds) != 2 {
		t.Errorf("Should have 2 cooldowns, got %d", len(cds))
	}
	
	if _, ok := cds["glitch"]; !ok {
		t.Error("glitch should be on cooldown")
	}
	if _, ok := cds["patch"]; !ok {
		t.Error("patch should be on cooldown")
	}
}

func TestReset(t *testing.T) {
	m := NewManager()
	
	m.Use("player1", "glitch")
	m.Use("player1", "patch")
	
	m.Reset("player1")
	
	if !m.IsReady("player1", "glitch") {
		t.Error("glitch should be ready after reset")
	}
	if !m.IsReady("player1", "patch") {
		t.Error("patch should be ready after reset")
	}
}

func TestResetAbility(t *testing.T) {
	m := NewManager()
	
	m.Use("player1", "glitch")
	m.Use("player1", "patch")
	
	m.ResetAbility("player1", "glitch")
	
	if !m.IsReady("player1", "glitch") {
		t.Error("glitch should be ready after reset")
	}
	if m.IsReady("player1", "patch") {
		t.Error("patch should still be on cooldown")
	}
}

func TestMultiplePlayers(t *testing.T) {
	m := NewManager()
	
	m.Use("player1", "glitch")
	m.Use("player2", "smash")
	
	// Player 1's glitch should be on cooldown
	if m.IsReady("player1", "glitch") {
		t.Error("player1 glitch should be on cooldown")
	}
	
	// Player 2's glitch should be ready (they didn't use it)
	if !m.IsReady("player2", "glitch") {
		t.Error("player2 glitch should be ready")
	}
	
	// Player 1's smash should be ready (they didn't use it)
	if !m.IsReady("player1", "smash") {
		t.Error("player1 smash should be ready")
	}
}

func TestAbilityCooldownsExist(t *testing.T) {
	expectedAbilities := []string{
		"glitch", "patch", "overflow", "backdoor",
		"smash", "fortify", "rampage", "ironwall",
		"strike", "vanish", "assassinate", "shadowstep",
		"flee", "use",
	}
	
	for _, ability := range expectedAbilities {
		if _, ok := AbilityCooldowns[ability]; !ok {
			t.Errorf("Missing cooldown for %s", ability)
		}
	}
}

func TestGlobalCD(t *testing.T) {
	// GlobalCD should be initialized
	if GlobalCD == nil {
		t.Error("GlobalCD should be initialized")
	}
	
	// Should work
	GlobalCD.Use("globaltest", "glitch")
	if GlobalCD.IsReady("globaltest", "glitch") {
		t.Error("Global cooldown should work")
	}
}

func TestCleanup(t *testing.T) {
	m := NewManager()
	
	// Add some entries
	m.Use("player1", "glitch")
	
	// Cleanup shouldn't crash
	m.Cleanup()
	
	// Recent cooldowns should still exist
	if m.IsReady("player1", "glitch") {
		t.Error("Recent cooldown should survive cleanup")
	}
}
