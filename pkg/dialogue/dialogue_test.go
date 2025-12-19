package dialogue

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.Trees == nil {
		t.Error("Trees map is nil")
	}
	if m.Sessions == nil {
		t.Error("Sessions map is nil")
	}
}

func TestDefaultTrees(t *testing.T) {
	m := NewManager()

	// Check that major NPCs have dialogue trees
	npcs := []string{"morpheus", "oracle", "architect", "merovingian"}
	for _, npc := range npcs {
		if _, ok := m.Trees[npc]; !ok {
			t.Errorf("Missing dialogue tree for %s", npc)
		}
	}
}

func TestStartDialogue(t *testing.T) {
	m := NewManager()

	// Start dialogue with Morpheus
	node, err := m.StartDialogue("TestPlayer", "morpheus")
	if err != nil {
		t.Fatalf("StartDialogue failed: %v", err)
	}
	if node == nil {
		t.Fatal("StartDialogue returned nil node")
	}
	if node.Speaker != "Morpheus" {
		t.Errorf("Expected speaker Morpheus, got %s", node.Speaker)
	}

	// Check session was created
	if !m.IsInDialogue("TestPlayer") {
		t.Error("Player should be in dialogue")
	}
}

func TestSelectChoice(t *testing.T) {
	m := NewManager()

	// Start dialogue
	m.StartDialogue("TestPlayer", "morpheus")

	// Select first choice
	node, action, err := m.SelectChoice("TestPlayer", 0)
	if err != nil {
		t.Fatalf("SelectChoice failed: %v", err)
	}
	if node == nil {
		t.Fatal("SelectChoice returned nil node")
	}

	// Verify we moved to a new node
	if node.ID == "greeting" {
		t.Error("Should have moved to new node")
	}

	_ = action // Action may be nil for some choices
}

func TestInvalidChoice(t *testing.T) {
	m := NewManager()

	// Start dialogue
	m.StartDialogue("TestPlayer", "morpheus")

	// Select invalid choice (too high)
	node, _, _ := m.SelectChoice("TestPlayer", 999)
	if node == nil {
		t.Fatal("Invalid choice should return current node")
	}
	if node.ID != "greeting" {
		t.Error("Should stay on same node for invalid choice")
	}
}

func TestEndDialogue(t *testing.T) {
	m := NewManager()

	m.StartDialogue("TestPlayer", "morpheus")
	if !m.IsInDialogue("TestPlayer") {
		t.Error("Player should be in dialogue")
	}

	m.EndDialogue("TestPlayer")
	if m.IsInDialogue("TestPlayer") {
		t.Error("Player should not be in dialogue after EndDialogue")
	}
}

func TestDialogueWithUnknownNPC(t *testing.T) {
	m := NewManager()

	node, err := m.StartDialogue("TestPlayer", "unknown_npc")
	if err != nil {
		t.Errorf("Should not error for unknown NPC: %v", err)
	}
	if node != nil {
		t.Error("Should return nil for unknown NPC")
	}
}

func TestGetCurrentNode(t *testing.T) {
	m := NewManager()

	// No dialogue started
	node := m.GetCurrentNode("TestPlayer")
	if node != nil {
		t.Error("Should return nil when no dialogue active")
	}

	// Start dialogue
	m.StartDialogue("TestPlayer", "morpheus")
	node = m.GetCurrentNode("TestPlayer")
	if node == nil {
		t.Error("Should return current node")
	}
}

func TestFormatNode(t *testing.T) {
	node := &Node{
		ID:      "test",
		Type:    NodeChoice,
		Speaker: "Test NPC",
		Text:    "Hello there!",
		Choices: []Choice{
			{Text: "Choice 1", NextNodeID: "next1"},
			{Text: "Choice 2", NextNodeID: "next2"},
		},
	}

	output := FormatNode(node)
	if output == "" {
		t.Error("FormatNode returned empty string")
	}
	if len(output) < 20 {
		t.Error("FormatNode output too short")
	}
}

func TestFormatNodeNil(t *testing.T) {
	output := FormatNode(nil)
	if output != "" {
		t.Error("FormatNode should return empty for nil")
	}
}

func TestDialogueToEndNode(t *testing.T) {
	m := NewManager()

	// Start dialogue with Morpheus and navigate to farewell
	m.StartDialogue("TestPlayer", "morpheus")

	// Select "I should go" (choice index 3)
	node, _, _ := m.SelectChoice("TestPlayer", 3)

	// Should be at farewell node (end type)
	if node == nil {
		// Dialogue ended, session should be cleared
		if m.IsInDialogue("TestPlayer") {
			t.Error("Session should be cleared after end node")
		}
	}
}

func TestMultiplePlayers(t *testing.T) {
	m := NewManager()

	// Two players start dialogues
	m.StartDialogue("Player1", "morpheus")
	m.StartDialogue("Player2", "oracle")

	// Both should be in dialogue
	if !m.IsInDialogue("Player1") {
		t.Error("Player1 should be in dialogue")
	}
	if !m.IsInDialogue("Player2") {
		t.Error("Player2 should be in dialogue")
	}

	// End Player1's dialogue
	m.EndDialogue("Player1")

	// Only Player2 should still be in dialogue
	if m.IsInDialogue("Player1") {
		t.Error("Player1 should not be in dialogue")
	}
	if !m.IsInDialogue("Player2") {
		t.Error("Player2 should still be in dialogue")
	}
}

func TestCaseInsensitivePlayerName(t *testing.T) {
	m := NewManager()

	m.StartDialogue("TestPlayer", "morpheus")

	// Check with different cases
	if !m.IsInDialogue("testplayer") {
		t.Error("Should be case insensitive")
	}
	if !m.IsInDialogue("TESTPLAYER") {
		t.Error("Should be case insensitive")
	}
}

func TestChoiceAction(t *testing.T) {
	m := NewManager()

	// Oracle's farewell has an action to complete objective
	m.StartDialogue("TestPlayer", "oracle")

	// Navigate to farewell: greeting -> no_cookie -> after_cookie -> farewell
	m.SelectChoice("TestPlayer", 1)                    // No, thank you -> no_cookie
	m.SelectChoice("TestPlayer", 0)                    // Continue (text node auto-advance)
	node, action, _ := m.SelectChoice("TestPlayer", 2) // I need to go -> farewell

	_ = node
	if action != nil {
		if action.Type != "complete_objective" {
			t.Logf("Action type: %s", action.Type)
		}
	}
}

func TestNodeTypes(t *testing.T) {
	// Test node type constants
	if NodeText != "text" {
		t.Error("NodeText should be 'text'")
	}
	if NodeChoice != "choice" {
		t.Error("NodeChoice should be 'choice'")
	}
	if NodeAction != "action" {
		t.Error("NodeAction should be 'action'")
	}
	if NodeEnd != "end" {
		t.Error("NodeEnd should be 'end'")
	}
}
