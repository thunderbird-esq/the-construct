package tutorial

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if len(m.tutorials) == 0 {
		t.Error("Default tutorials should be registered")
	}
}

func TestRegisterTutorial(t *testing.T) {
	m := NewManager()
	
	tutorial := &Tutorial{
		ID:   "test_tutorial",
		Name: "Test Tutorial",
		Steps: []TutorialStep{
			{ID: "step1", Title: "Step 1", Trigger: TriggerCommand},
		},
	}
	
	m.RegisterTutorial(tutorial)
	
	got := m.GetTutorial("test_tutorial")
	if got == nil {
		t.Fatal("Tutorial not registered")
	}
	if got.Name != "Test Tutorial" {
		t.Errorf("Name = %s, want Test Tutorial", got.Name)
	}
}

func TestListTutorials(t *testing.T) {
	m := NewManager()
	
	tutorials := m.ListTutorials()
	if len(tutorials) == 0 {
		t.Error("Should have default tutorials")
	}
}

func TestStartTutorial(t *testing.T) {
	m := NewManager()
	
	tutorial, err := m.StartTutorial("player1", "new_player")
	if err != nil {
		t.Fatalf("StartTutorial failed: %v", err)
	}
	if tutorial == nil {
		t.Fatal("Tutorial should not be nil")
	}
	if tutorial.ID != "new_player" {
		t.Errorf("Tutorial ID = %s, want new_player", tutorial.ID)
	}
	
	// Check progress was created
	progress := m.GetProgress("player1")
	if progress == nil {
		t.Fatal("Progress should be created")
	}
	if _, ok := progress["new_player"]; !ok {
		t.Error("Progress for new_player should exist")
	}
}

func TestStartTutorialNotFound(t *testing.T) {
	m := NewManager()
	
	_, err := m.StartTutorial("player1", "nonexistent")
	if err == nil {
		t.Error("Should error for nonexistent tutorial")
	}
}

func TestStartTutorialPrerequisites(t *testing.T) {
	m := NewManager()
	
	// Try to start combat_mastery without completing new_player
	_, err := m.StartTutorial("player1", "combat_mastery")
	if err == nil {
		t.Error("Should error when prerequisites not met")
	}
}

func TestGetCurrentStep(t *testing.T) {
	m := NewManager()
	m.StartTutorial("player1", "new_player")
	
	tutorial, step, stepNum := m.GetCurrentStep("player1")
	if tutorial == nil {
		t.Fatal("Tutorial should not be nil")
	}
	if step == nil {
		t.Fatal("Step should not be nil")
	}
	if stepNum != 0 {
		t.Errorf("Step number = %d, want 0", stepNum)
	}
	if step.ID != "welcome" {
		t.Errorf("Step ID = %s, want welcome", step.ID)
	}
}

func TestGetCurrentStepNoTutorial(t *testing.T) {
	m := NewManager()
	
	tutorial, step, stepNum := m.GetCurrentStep("nonexistent")
	if tutorial != nil || step != nil || stepNum != -1 {
		t.Error("Should return nil for player with no tutorial")
	}
}

func TestCheckProgress(t *testing.T) {
	m := NewManager()
	m.StartTutorial("player1", "new_player")
	
	// Complete first step (look command)
	stepComplete, tutorialComplete, msg := m.CheckProgress("player1", TriggerCommand, map[string]interface{}{
		"command": "look",
	})
	
	if !stepComplete {
		t.Error("Step should be complete")
	}
	if tutorialComplete {
		t.Error("Tutorial should not be complete yet")
	}
	if msg == "" {
		t.Error("Should have a message")
	}
	
	// Verify progress
	_, step, stepNum := m.GetCurrentStep("player1")
	if stepNum != 1 {
		t.Errorf("Step number = %d, want 1", stepNum)
	}
	if step.ID != "movement" {
		t.Errorf("Step ID = %s, want movement", step.ID)
	}
}

func TestCheckProgressWrongTrigger(t *testing.T) {
	m := NewManager()
	m.StartTutorial("player1", "new_player")
	
	// Try wrong trigger
	stepComplete, _, _ := m.CheckProgress("player1", TriggerMove, nil)
	if stepComplete {
		t.Error("Wrong trigger should not complete step")
	}
}

func TestCheckProgressWrongData(t *testing.T) {
	m := NewManager()
	m.StartTutorial("player1", "new_player")
	
	// Try right trigger, wrong command
	stepComplete, _, _ := m.CheckProgress("player1", TriggerCommand, map[string]interface{}{
		"command": "help",
	})
	if stepComplete {
		t.Error("Wrong trigger data should not complete step")
	}
}

func TestCheckProgressCaseInsensitive(t *testing.T) {
	m := NewManager()
	m.StartTutorial("player1", "new_player")
	
	// Try uppercase
	stepComplete, _, _ := m.CheckProgress("player1", TriggerCommand, map[string]interface{}{
		"command": "LOOK",
	})
	if !stepComplete {
		t.Error("Command matching should be case-insensitive")
	}
}

func TestTutorialCompletion(t *testing.T) {
	m := NewManager()
	
	// Create a simple 2-step tutorial
	m.RegisterTutorial(&Tutorial{
		ID:   "simple",
		Name: "Simple Tutorial",
		Steps: []TutorialStep{
			{ID: "step1", Trigger: TriggerCommand, TriggerData: map[string]interface{}{"command": "a"}},
			{ID: "step2", Trigger: TriggerCommand, TriggerData: map[string]interface{}{"command": "b"}},
		},
		Rewards: TutorialRewards{XP: 100},
	})
	
	m.StartTutorial("player1", "simple")
	
	// Complete step 1
	m.CheckProgress("player1", TriggerCommand, map[string]interface{}{"command": "a"})
	
	// Complete step 2
	_, tutorialComplete, _ := m.CheckProgress("player1", TriggerCommand, map[string]interface{}{"command": "b"})
	
	if !tutorialComplete {
		t.Error("Tutorial should be complete")
	}
	
	// Verify completion
	if !m.HasCompletedTutorial("player1", "simple") {
		t.Error("Tutorial should be marked as completed")
	}
	
	completed := m.GetCompletedTutorials("player1")
	if len(completed) != 1 || completed[0] != "simple" {
		t.Error("Completed tutorials list incorrect")
	}
}

func TestSkipStep(t *testing.T) {
	m := NewManager()
	m.StartTutorial("player1", "new_player")
	
	success, msg := m.SkipStep("player1")
	if !success {
		t.Error("Skip should succeed")
	}
	if msg == "" {
		t.Error("Should have message")
	}
	
	_, step, stepNum := m.GetCurrentStep("player1")
	if stepNum != 1 {
		t.Errorf("Step number = %d, want 1", stepNum)
	}
	if step.ID != "movement" {
		t.Errorf("Step ID = %s, want movement", step.ID)
	}
}

func TestSkipStepNoTutorial(t *testing.T) {
	m := NewManager()
	
	success, _ := m.SkipStep("nonexistent")
	if success {
		t.Error("Skip should fail for player with no tutorial")
	}
}

func TestSkipTutorial(t *testing.T) {
	m := NewManager()
	m.StartTutorial("player1", "new_player")
	
	success, msg := m.SkipTutorial("player1", "new_player")
	if !success {
		t.Error("Skip should succeed")
	}
	if msg == "" {
		t.Error("Should have message")
	}
	
	if !m.HasCompletedTutorial("player1", "new_player") {
		t.Error("Skipped tutorial should be marked complete")
	}
}

func TestSkipTutorialNotSkippable(t *testing.T) {
	m := NewManager()
	
	m.RegisterTutorial(&Tutorial{
		ID:        "unskippable",
		Name:      "Cannot Skip",
		Skippable: false,
		Steps:     []TutorialStep{{ID: "step1", Trigger: TriggerAny}},
	})
	
	m.StartTutorial("player1", "unskippable")
	
	success, _ := m.SkipTutorial("player1", "unskippable")
	if success {
		t.Error("Should not be able to skip unskippable tutorial")
	}
}

func TestGetAvailableTutorials(t *testing.T) {
	m := NewManager()
	
	available := m.GetAvailableTutorials("newplayer")
	
	// Should include new_player (no prereqs)
	found := false
	for _, tut := range available {
		if tut.ID == "new_player" {
			found = true
			break
		}
	}
	if !found {
		t.Error("new_player should be available")
	}
	
	// Should NOT include combat_mastery (requires new_player)
	for _, tut := range available {
		if tut.ID == "combat_mastery" {
			t.Error("combat_mastery should not be available without new_player complete")
		}
	}
}

func TestGetAvailableTutorialsAfterCompletion(t *testing.T) {
	m := NewManager()
	
	// Mark new_player as completed
	m.completedTutorials["player1"] = []string{"new_player"}
	
	available := m.GetAvailableTutorials("player1")
	
	// Should NOT include new_player (already completed)
	for _, tut := range available {
		if tut.ID == "new_player" {
			t.Error("new_player should not be available (already completed)")
		}
	}
	
	// Should include combat_mastery (prereq now met)
	found := false
	for _, tut := range available {
		if tut.ID == "combat_mastery" {
			found = true
			break
		}
	}
	if !found {
		t.Error("combat_mastery should be available after new_player completion")
	}
}

func TestGetHint(t *testing.T) {
	m := NewManager()
	m.StartTutorial("player1", "new_player")
	
	hint := m.GetHint("player1")
	if hint == "" {
		t.Error("Should have a hint")
	}
}

func TestGetHintNoTutorial(t *testing.T) {
	m := NewManager()
	
	hint := m.GetHint("nonexistent")
	if hint == "" {
		t.Error("Should have a default message")
	}
}

func TestFormatStepDisplay(t *testing.T) {
	m := NewManager()
	m.StartTutorial("player1", "new_player")
	
	display := m.FormatStepDisplay("player1")
	if display == "" {
		t.Error("Should have display text")
	}
	if len(display) < 50 {
		t.Error("Display seems too short")
	}
}

func TestFormatStepDisplayNoTutorial(t *testing.T) {
	m := NewManager()
	
	display := m.FormatStepDisplay("nonexistent")
	if display != "" {
		t.Error("Should be empty for player with no tutorial")
	}
}

func TestTriggerAny(t *testing.T) {
	m := NewManager()
	
	m.RegisterTutorial(&Tutorial{
		ID:   "any_trigger",
		Name: "Any Trigger Test",
		Steps: []TutorialStep{
			{ID: "step1", Trigger: TriggerAny},
		},
	})
	
	m.StartTutorial("player1", "any_trigger")
	
	// Any trigger should work
	stepComplete, _, _ := m.CheckProgress("player1", TriggerMove, nil)
	if !stepComplete {
		t.Error("TriggerAny should complete on any action")
	}
}

func TestOnCompleteCallback(t *testing.T) {
	m := NewManager()
	
	var called bool
	m.RegisterTutorial(&Tutorial{
		ID:   "callback_test",
		Name: "Callback Test",
		Steps: []TutorialStep{
			{
				ID:         "step1",
				Trigger:    TriggerAny,
				OnComplete: func(data interface{}) { called = true },
			},
		},
	})
	
	m.StartTutorial("player1", "callback_test")
	m.CheckProgress("player1", TriggerAny, nil)
	
	if !called {
		t.Error("OnComplete callback should have been called")
	}
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		text     string
		width    int
		expected int // number of lines
	}{
		{"short", 20, 1},
		{"this is a longer text that should wrap", 10, 5},
		{"", 10, 1},
		{"one two three four five", 15, 2},
	}
	
	for _, tt := range tests {
		lines := wrapText(tt.text, tt.width)
		if len(lines) != tt.expected {
			t.Errorf("wrapText(%q, %d) = %d lines, want %d", tt.text, tt.width, len(lines), tt.expected)
		}
	}
}

func TestDefaultTutorials(t *testing.T) {
	m := NewManager()
	
	expectedTutorials := []string{
		"new_player",
		"combat_mastery",
		"crafting_101",
		"faction_guide",
		"party_play",
	}
	
	for _, id := range expectedTutorials {
		if m.GetTutorial(id) == nil {
			t.Errorf("Default tutorial %s should exist", id)
		}
	}
}

func TestGlobalManager(t *testing.T) {
	if GlobalManager == nil {
		t.Error("GlobalManager should be initialized")
	}
}

func TestMatchesTriggerDataNil(t *testing.T) {
	m := NewManager()
	
	// nil required should match anything
	if !m.matchesTriggerData(nil, map[string]interface{}{"key": "value"}) {
		t.Error("nil required should match")
	}
	
	// empty required should match anything
	if !m.matchesTriggerData(map[string]interface{}{}, map[string]interface{}{"key": "value"}) {
		t.Error("empty required should match")
	}
}

func TestMatchesTriggerDataMissing(t *testing.T) {
	m := NewManager()
	
	required := map[string]interface{}{"key": "value"}
	actual := map[string]interface{}{"other": "value"}
	
	if m.matchesTriggerData(required, actual) {
		t.Error("Missing key should not match")
	}
}

func TestHasCompletedTutorial(t *testing.T) {
	m := NewManager()
	
	if m.HasCompletedTutorial("player1", "new_player") {
		t.Error("Should not be completed initially")
	}
	
	m.completedTutorials["player1"] = []string{"new_player"}
	
	if !m.HasCompletedTutorial("player1", "new_player") {
		t.Error("Should be completed after adding to list")
	}
}
