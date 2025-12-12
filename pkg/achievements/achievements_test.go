package achievements

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if len(m.Achievements) == 0 {
		t.Error("Achievements should be loaded")
	}
}

func TestGetPlayerAchievements(t *testing.T) {
	m := NewManager()
	pa := m.GetPlayerAchievements("testplayer")

	if pa == nil {
		t.Fatal("GetPlayerAchievements returned nil")
	}
	if pa.Earned == nil {
		t.Error("Earned map should be initialized")
	}
}

func TestAward(t *testing.T) {
	m := NewManager()

	ach := m.Award("player1", AchFirstBlood)
	if ach == nil {
		t.Fatal("Award should return achievement")
	}
	if ach.Name != "First Blood" {
		t.Errorf("Name = %s, want First Blood", ach.Name)
	}

	// Second award should return nil
	ach2 := m.Award("player1", AchFirstBlood)
	if ach2 != nil {
		t.Error("Second award should return nil")
	}
}

func TestAwardInvalid(t *testing.T) {
	m := NewManager()

	ach := m.Award("player1", AchievementID("nonexistent"))
	if ach != nil {
		t.Error("Invalid achievement should return nil")
	}
}

func TestHasAchievement(t *testing.T) {
	m := NewManager()

	if m.HasAchievement("player1", AchFirstBlood) {
		t.Error("Should not have achievement before earning")
	}

	m.Award("player1", AchFirstBlood)

	if !m.HasAchievement("player1", AchFirstBlood) {
		t.Error("Should have achievement after earning")
	}
}

func TestGetEarnedAchievements(t *testing.T) {
	m := NewManager()

	m.Award("player1", AchFirstBlood)
	m.Award("player1", AchAwakened)

	earned := m.GetEarnedAchievements("player1")
	if len(earned) != 2 {
		t.Errorf("Expected 2 achievements, got %d", len(earned))
	}
}

func TestTotalPoints(t *testing.T) {
	m := NewManager()

	m.Award("player1", AchFirstBlood)   // 10 points
	m.Award("player1", AchAwakened)     // 25 points

	points := m.GetTotalPoints("player1")
	if points != 35 {
		t.Errorf("Total points = %d, want 35", points)
	}
}

func TestTitles(t *testing.T) {
	m := NewManager()

	// Award achievement with title
	m.Award("player1", AchAgentSlayer) // Has title "Agent Slayer"

	titles := m.GetAvailableTitles("player1")
	if len(titles) != 1 {
		t.Errorf("Expected 1 title, got %d", len(titles))
	}

	// Set title
	ok := m.SetTitle("player1", "Agent Slayer")
	if !ok {
		t.Error("SetTitle should succeed")
	}

	title := m.GetTitle("player1")
	if title != "Agent Slayer" {
		t.Errorf("Title = %s, want Agent Slayer", title)
	}
}

func TestSetTitleUnearned(t *testing.T) {
	m := NewManager()
	m.GetPlayerAchievements("player1") // Initialize

	ok := m.SetTitle("player1", "Agent Slayer")
	if ok {
		t.Error("Should not set unearned title")
	}
}

func TestSetTitleEmpty(t *testing.T) {
	m := NewManager()
	m.Award("player1", AchAgentSlayer)
	m.SetTitle("player1", "Agent Slayer")

	// Clear title
	ok := m.SetTitle("player1", "")
	if !ok {
		t.Error("Should be able to clear title")
	}

	title := m.GetTitle("player1")
	if title != "" {
		t.Error("Title should be cleared")
	}
}

func TestGetByCategory(t *testing.T) {
	m := NewManager()

	combat := m.GetByCategory(CatCombat)
	if len(combat) == 0 {
		t.Error("Should have combat achievements")
	}

	for _, ach := range combat {
		if ach.Category != CatCombat {
			t.Errorf("Achievement %s has wrong category", ach.ID)
		}
	}
}

func TestHiddenAchievements(t *testing.T) {
	m := NewManager()

	// Hidden achievements should not appear in category list
	secrets := m.GetByCategory(CatSecret)
	for _, ach := range secrets {
		if ach.Hidden {
			t.Errorf("Hidden achievement %s should not be listed", ach.ID)
		}
	}
}

func TestAchievementConstants(t *testing.T) {
	m := NewManager()

	// Verify all constants have corresponding achievements
	ids := []AchievementID{
		AchFirstBlood, AchAgentSlayer, AchAwakened, AchReachZion,
		AchMeetOracle, AchPhoneMaster, AchPartyAnimal, AchCrafter,
		AchMillionaire, AchLevel10, AchLevel25, AchQuestComplete5,
		AchQuestComplete10, AchSurvivor, AchTheOne,
	}

	for _, id := range ids {
		if m.Achievements[id] == nil {
			t.Errorf("Achievement %s not found", id)
		}
	}
}

func TestAchievementPoints(t *testing.T) {
	m := NewManager()

	for id, ach := range m.Achievements {
		if ach.Points <= 0 {
			t.Errorf("Achievement %s has no points", id)
		}
	}
}
