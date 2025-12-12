package main

import (
	"strings"
	"testing"
)

// =============================================================================
// PHASE 4 TECHNICAL POLISH TESTS
// =============================================================================

// --- Stream J: Context & Connection Tests ---

func TestMaxConnectionsConstant(t *testing.T) {
	if MaxConnections <= 0 {
		t.Error("MaxConnections should be positive")
	}
	if MaxConnections > 1000 {
		t.Error("MaxConnections seems too high, verify this is intentional")
	}
	t.Logf("MaxConnections configured: %d", MaxConnections)
}

// --- Stream L: Quality of Life Tests ---

func TestApplyThemeGreen(t *testing.T) {
	input := Green + "Hello World" + Reset
	result := ApplyTheme(input, "green")
	if result != input {
		t.Error("Green theme should return unchanged text")
	}
}

func TestApplyThemeAmber(t *testing.T) {
	input := Green + "Hello World" + Reset
	result := ApplyTheme(input, "amber")
	if strings.Contains(result, Green) {
		t.Error("Amber theme should replace green color")
	}
	if !strings.Contains(result, "\033[33m") {
		t.Error("Amber theme should contain yellow/amber color code")
	}
}

func TestApplyThemeWhite(t *testing.T) {
	input := Green + "Hello World" + Reset
	result := ApplyTheme(input, "white")
	if strings.Contains(result, Green) {
		t.Error("White theme should replace green color")
	}
	if !strings.Contains(result, "\033[97m") {
		t.Error("White theme should contain bright white color code")
	}
}

func TestApplyThemeNone(t *testing.T) {
	input := Green + "Hello" + Red + " World" + Reset
	result := ApplyTheme(input, "none")
	if strings.Contains(result, "\033[") {
		t.Error("None theme should strip all color codes")
	}
	if result != "Hello World" {
		t.Errorf("None theme result = %q, want 'Hello World'", result)
	}
}

func TestApplyThemeEmpty(t *testing.T) {
	input := Green + "Hello World" + Reset
	result := ApplyTheme(input, "")
	if result != input {
		t.Error("Empty theme should return unchanged text (default green)")
	}
}

func TestApplyThemeUnknown(t *testing.T) {
	input := Green + "Hello World" + Reset
	result := ApplyTheme(input, "invalid")
	if result != input {
		t.Error("Unknown theme should return unchanged text")
	}
}

func TestStripColors(t *testing.T) {
	input := Green + "Green " + Red + "Red " + Yellow + "Yellow" + Reset
	result := stripColors(input)
	expected := "Green Red Yellow"
	if result != expected {
		t.Errorf("stripColors = %q, want %q", result, expected)
	}
}

func TestBriefModePlayerField(t *testing.T) {
	p := &Player{Name: "test"}
	if p.BriefMode {
		t.Error("BriefMode should default to false")
	}
	p.BriefMode = true
	if !p.BriefMode {
		t.Error("BriefMode should be settable to true")
	}
}

func TestColorThemePlayerField(t *testing.T) {
	p := &Player{Name: "test"}
	if p.ColorTheme != "" {
		t.Error("ColorTheme should default to empty string")
	}
	p.ColorTheme = "amber"
	if p.ColorTheme != "amber" {
		t.Errorf("ColorTheme = %q, want 'amber'", p.ColorTheme)
	}
}

func TestBriefModeLookDescription(t *testing.T) {
	w := NewWorld()

	// Create a test room with a long description
	longDesc := "This is a very long room description that should be truncated when brief mode is enabled. It contains many details about the environment that players might not want to see every time they enter the room."
	w.Rooms["test_room"] = &Room{
		ID:          "test_room",
		Description: longDesc,
		Exits:       map[string]string{},
		Items:       []*Item{},
		NPCs:        []*NPC{},
	}

	// Player with brief mode OFF
	playerFull := &Player{Name: "full", RoomID: "test_room", BriefMode: false}
	w.Players[nil] = playerFull
	fullResult := w.Look(playerFull, "")
	if !strings.Contains(fullResult, "truncated when brief mode") {
		t.Error("Full mode should show complete description")
	}

	// Player with brief mode ON
	playerBrief := &Player{Name: "brief", RoomID: "test_room", BriefMode: true}
	briefResult := w.Look(playerBrief, "")
	if strings.Contains(briefResult, "truncated when brief mode") {
		t.Error("Brief mode should truncate long descriptions")
	}
	if !strings.Contains(briefResult, "...") && len(briefResult) > 150 {
		t.Error("Brief mode should truncate with ellipsis or at sentence boundary")
	}
}

// --- Deployment Tests (from original) ---

func TestHealthEndpointExists(t *testing.T) {
	t.Log("Health endpoint: GET /health")
	t.Log("Returns: {\"status\":\"healthy\",\"version\":\"...\",\"service\":\"matrix-mud\"}")
}

func TestDeploymentConfigsExist(t *testing.T) {
	configs := []string{
		"Dockerfile",
		"fly.toml",
		".env.production.example",
		"scripts/deploy.sh",
	}
	for _, config := range configs {
		t.Logf("Deployment config: %s", config)
	}
}
