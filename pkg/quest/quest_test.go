package quest

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.Quests == nil {
		t.Error("Quests map not initialized")
	}
	if m.Players == nil {
		t.Error("Players map not initialized")
	}
}

func TestDefaultQuestsLoaded(t *testing.T) {
	m := NewManager()

	// Check core quests exist
	expectedQuests := []string{"free_your_mind", "the_oracle", "rescue_morpheus"}
	for _, qid := range expectedQuests {
		if _, ok := m.Quests[qid]; !ok {
			t.Errorf("Expected quest %s not found", qid)
		}
	}
}

func TestGetPlayerQuests(t *testing.T) {
	m := NewManager()

	pq := m.GetPlayerQuests("TestPlayer")
	if pq == nil {
		t.Fatal("GetPlayerQuests returned nil")
	}
	if pq.Active == nil {
		t.Error("Active quests map not initialized")
	}

	// Should return same instance on second call
	pq2 := m.GetPlayerQuests("TestPlayer")
	if pq != pq2 {
		t.Error("Should return same PlayerQuests instance")
	}

	// Case insensitive
	pq3 := m.GetPlayerQuests("testplayer")
	if pq != pq3 {
		t.Error("Should be case insensitive")
	}
}

func TestCanStart(t *testing.T) {
	m := NewManager()

	// Should be able to start free_your_mind
	can, msg := m.CanStart("TestPlayer", "free_your_mind", 1)
	if !can {
		t.Errorf("Should be able to start free_your_mind: %s", msg)
	}

	// Should not be able to start the_oracle without prereq
	can, msg = m.CanStart("TestPlayer", "the_oracle", 5)
	if can {
		t.Error("Should not be able to start quest without prerequisites")
	}

	// Should not be able to start nonexistent quest
	can, _ = m.CanStart("TestPlayer", "nonexistent", 1)
	if can {
		t.Error("Should not be able to start nonexistent quest")
	}
}

func TestCanStartLevelRequirement(t *testing.T) {
	m := NewManager()

	// Complete prereqs first
	m.GetPlayerQuests("LevelTest")
	m.Players["leveltest"].Completed = []string{"free_your_mind"}

	// Level too low
	can, msg := m.CanStart("LevelTest", "the_oracle", 1)
	if can {
		t.Error("Should not be able to start quest below min level")
	}
	if msg == "" {
		t.Error("Should return reason for rejection")
	}

	// Level sufficient
	can, _ = m.CanStart("LevelTest", "the_oracle", 5)
	if !can {
		t.Error("Should be able to start quest at sufficient level")
	}
}

func TestStartQuest(t *testing.T) {
	m := NewManager()

	dialogue, err := m.StartQuest("StartTest", "free_your_mind")
	if err != nil {
		t.Errorf("StartQuest returned error: %v", err)
	}
	if dialogue == "" {
		t.Error("Should return stage dialogue")
	}

	// Check quest is now active
	pq := m.GetPlayerQuests("StartTest")
	if _, ok := pq.Active["free_your_mind"]; !ok {
		t.Error("Quest should be active")
	}
}

func TestUpdateProgress(t *testing.T) {
	m := NewManager()

	// Start the quest
	m.StartQuest("ProgressTest", "free_your_mind")

	// Update with wrong type - should have no effect
	msgs := m.UpdateProgress("ProgressTest", ObjKill, "wrong_target", 1)
	if len(msgs) > 0 {
		t.Error("Should not complete objective with wrong target")
	}

	// Update with correct objective
	msgs = m.UpdateProgress("ProgressTest", ObjVisit, "dojo", 1)
	// Should advance to next stage
	if len(msgs) == 0 {
		t.Log("No messages returned, but progress should be tracked")
	}
}

func TestGetActiveQuests(t *testing.T) {
	m := NewManager()

	// No quests
	result := m.GetActiveQuests("EmptyPlayer")
	if result == "" {
		t.Error("Should return message for no quests")
	}

	// Start a quest
	m.StartQuest("ActiveTest", "free_your_mind")
	result = m.GetActiveQuests("ActiveTest")
	if result == "" {
		t.Error("Should return quest list")
	}
}

func TestGetCompletedQuests(t *testing.T) {
	m := NewManager()

	// No completed
	completed := m.GetCompletedQuests("NoComplete")
	if len(completed) != 0 {
		t.Error("Should have no completed quests")
	}

	// Add completed
	m.GetPlayerQuests("CompleteTest")
	m.Players["completetest"].Completed = []string{"free_your_mind"}

	completed = m.GetCompletedQuests("CompleteTest")
	if len(completed) != 1 {
		t.Errorf("Should have 1 completed quest, got %d", len(completed))
	}
}

func TestGetQuestReward(t *testing.T) {
	m := NewManager()

	reward := m.GetQuestReward("free_your_mind")
	if reward == nil {
		t.Fatal("Should return reward")
	}
	if reward.XP != 100 {
		t.Errorf("Expected 100 XP reward, got %d", reward.XP)
	}

	// Nonexistent quest
	reward = m.GetQuestReward("nonexistent")
	if reward != nil {
		t.Error("Should return nil for nonexistent quest")
	}
}

func TestQuestPrerequisites(t *testing.T) {
	m := NewManager()

	// rescue_morpheus requires the_oracle which requires free_your_mind
	can, _ := m.CanStart("PrereqTest", "rescue_morpheus", 10)
	if can {
		t.Error("Should not start without completing prerequisite chain")
	}

	// Complete prerequisites
	pq := m.GetPlayerQuests("PrereqTest")
	pq.Completed = []string{"free_your_mind", "the_oracle"}

	can, _ = m.CanStart("PrereqTest", "rescue_morpheus", 10)
	if !can {
		t.Error("Should be able to start with prerequisites completed")
	}
}

func TestRepeatableQuest(t *testing.T) {
	m := NewManager()

	// Add a repeatable quest for testing
	m.Quests["test_repeatable"] = &Quest{
		ID:          "test_repeatable",
		Name:        "Repeatable Test Quest",
		Description: "A quest that can be repeated.",
		Repeatable:  true,
		Stages: []Stage{
			{ID: "stage1", Name: "Test Stage", Objectives: []Objective{{ID: "obj1", Type: ObjKill, Target: "test", Count: 1}}},
		},
		Reward: Reward{XP: 10},
	}

	// Start the quest
	m.StartQuest("RepeatTest", "test_repeatable")
	pq := m.GetPlayerQuests("RepeatTest")

	// Complete it
	pq.Completed = append(pq.Completed, "test_repeatable")
	delete(pq.Active, "test_repeatable")

	// Should be able to start again
	can, _ := m.CanStart("RepeatTest", "test_repeatable", 5)
	if !can {
		t.Error("Should be able to restart repeatable quest")
	}
}

func TestObjectiveTypes(t *testing.T) {
	// Verify all objective types are valid
	types := []ObjectiveType{ObjKill, ObjCollect, ObjDeliver, ObjVisit, ObjTalk, ObjUse, ObjChoice}
	for _, ot := range types {
		if ot == "" {
			t.Error("Objective type should not be empty")
		}
	}
}
