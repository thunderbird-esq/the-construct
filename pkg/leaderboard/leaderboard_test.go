package leaderboard

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.Players == nil {
		t.Error("Players map should be initialized")
	}
}

func TestGetStats(t *testing.T) {
	m := NewManager()
	ps := m.GetStats("player1")

	if ps == nil {
		t.Fatal("GetStats returned nil")
	}
	if ps.Name != "player1" {
		t.Errorf("Name = %s, want player1", ps.Name)
	}
}

func TestUpdateStat(t *testing.T) {
	m := NewManager()

	m.UpdateStat("player1", StatXP, 1000)
	ps := m.GetStats("player1")
	if ps.XP != 1000 {
		t.Errorf("XP = %d, want 1000", ps.XP)
	}

	m.UpdateStat("player1", StatLevel, 10)
	if ps.Level != 10 {
		t.Errorf("Level = %d, want 10", ps.Level)
	}

	m.UpdateStat("player1", StatKills, 50)
	if ps.Kills != 50 {
		t.Errorf("Kills = %d, want 50", ps.Kills)
	}
}

func TestIncrementStat(t *testing.T) {
	m := NewManager()

	m.UpdateStat("player1", StatKills, 10)
	m.IncrementStat("player1", StatKills, 5)

	ps := m.GetStats("player1")
	if ps.Kills != 15 {
		t.Errorf("Kills = %d, want 15", ps.Kills)
	}
}

func TestGetLeaderboard(t *testing.T) {
	m := NewManager()

	m.UpdateStat("player1", StatXP, 1000)
	m.UpdateStat("player2", StatXP, 2000)
	m.UpdateStat("player3", StatXP, 500)

	board := m.GetLeaderboard(StatXP, 10)
	if len(board) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(board))
	}

	// Should be sorted descending
	if board[0].Name != "player2" {
		t.Error("First place should be player2")
	}
	if board[0].Rank != 1 {
		t.Errorf("First place rank should be 1, got %d", board[0].Rank)
	}
	if board[1].Name != "player1" {
		t.Error("Second place should be player1")
	}
	if board[2].Name != "player3" {
		t.Error("Third place should be player3")
	}
}

func TestGetLeaderboardLimit(t *testing.T) {
	m := NewManager()

	for i := 0; i < 20; i++ {
		m.UpdateStat("player"+string(rune('A'+i)), StatKills, i*10)
	}

	board := m.GetLeaderboard(StatKills, 5)
	if len(board) != 5 {
		t.Errorf("Expected 5 entries, got %d", len(board))
	}
}

func TestGetRank(t *testing.T) {
	m := NewManager()

	m.UpdateStat("player1", StatXP, 1000)
	m.UpdateStat("player2", StatXP, 2000)
	m.UpdateStat("player3", StatXP, 500)

	rank := m.GetRank("player1", StatXP)
	if rank != 2 {
		t.Errorf("player1 rank = %d, want 2", rank)
	}

	rank = m.GetRank("player2", StatXP)
	if rank != 1 {
		t.Errorf("player2 rank = %d, want 1", rank)
	}

	rank = m.GetRank("nonexistent", StatXP)
	if rank != 0 {
		t.Errorf("nonexistent rank = %d, want 0", rank)
	}
}

func TestAllStatTypes(t *testing.T) {
	m := NewManager()

	// Update all stat types
	m.UpdateStat("player1", StatXP, 100)
	m.UpdateStat("player1", StatLevel, 5)
	m.UpdateStat("player1", StatKills, 10)
	m.UpdateStat("player1", StatDeaths, 2)
	m.UpdateStat("player1", StatQuestsCompleted, 3)
	m.UpdateStat("player1", StatMoney, 1000)
	m.UpdateStat("player1", StatPvPWins, 5)
	m.UpdateStat("player1", StatPvPLosses, 2)
	m.UpdateStat("player1", StatPlayTime, 60)
	m.UpdateStat("player1", StatAchievements, 3)

	ps := m.GetStats("player1")
	if ps.XP != 100 || ps.Level != 5 || ps.Kills != 10 ||
		ps.Deaths != 2 || ps.QuestsCompleted != 3 || ps.Money != 1000 ||
		ps.PvPWins != 5 || ps.PvPLosses != 2 || ps.PlayTimeMinutes != 60 ||
		ps.Achievements != 3 {
		t.Error("Not all stats were updated correctly")
	}
}

func TestGetAllStats(t *testing.T) {
	m := NewManager()

	m.UpdateStat("player1", StatXP, 100)
	m.UpdateStat("player1", StatKills, 10)

	ps := m.GetAllStats("player1")
	if ps == nil {
		t.Fatal("GetAllStats returned nil")
	}
	if ps.XP != 100 || ps.Kills != 10 {
		t.Error("Stats not returned correctly")
	}

	ps2 := m.GetAllStats("nonexistent")
	if ps2 != nil {
		t.Error("Nonexistent player should return nil")
	}
}

func TestIncrementAllStatTypes(t *testing.T) {
	m := NewManager()

	// Increment all stat types
	stats := []StatType{
		StatXP, StatLevel, StatKills, StatDeaths, StatQuestsCompleted,
		StatMoney, StatPvPWins, StatPvPLosses, StatPlayTime, StatAchievements,
	}

	for _, stat := range stats {
		m.IncrementStat("player1", stat, 5)
	}

	ps := m.GetStats("player1")
	if ps.XP != 5 || ps.Level != 5 || ps.Kills != 5 {
		t.Error("Increment didn't work for all stats")
	}
}
