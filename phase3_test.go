package main

import (
	"strings"
	"testing"

	"github.com/yourusername/matrix-mud/pkg/achievements"
	"github.com/yourusername/matrix-mud/pkg/faction"
	"github.com/yourusername/matrix-mud/pkg/leaderboard"
	"github.com/yourusername/matrix-mud/pkg/training"
)

// TestPhase3FactionSystem tests the faction system integration
func TestPhase3FactionSystem(t *testing.T) {
	// Test faction manager
	fm := faction.NewManager()
	
	// Join faction
	msg, ok := fm.Join("testplayer", faction.FactionZion)
	if !ok {
		t.Errorf("Join failed: %s", msg)
	}
	
	// Verify faction
	pf := fm.GetPlayerFaction("testplayer")
	if pf.Faction != faction.FactionZion {
		t.Error("Player should be in Zion")
	}
	
	// Test reputation
	fm.AdjustReputation("testplayer", faction.FactionZion, 100)
	rep := fm.GetReputation("testplayer", faction.FactionZion)
	if rep < 100 {
		t.Errorf("Reputation should be at least 100, got %d", rep)
	}
	
	// Opposing faction effect
	machineRep := fm.GetReputation("testplayer", faction.FactionMachines)
	if machineRep >= 0 {
		t.Error("Machine rep should decrease when Zion increases")
	}
	
	// Standing name
	standing := faction.GetStandingName(rep)
	if standing == "" {
		t.Error("Standing name should not be empty")
	}
}

// TestPhase3AchievementSystem tests the achievement system integration
func TestPhase3AchievementSystem(t *testing.T) {
	am := achievements.NewManager()
	
	// Award achievement
	ach := am.Award("testplayer", achievements.AchFirstBlood)
	if ach == nil {
		t.Fatal("Award should return achievement")
	}
	if ach.Name != "First Blood" {
		t.Errorf("Name = %s, want First Blood", ach.Name)
	}
	
	// Check points
	points := am.GetTotalPoints("testplayer")
	if points != ach.Points {
		t.Errorf("Points = %d, want %d", points, ach.Points)
	}
	
	// Award achievement with title
	am.Award("testplayer", achievements.AchAgentSlayer)
	
	// Check available titles
	titles := am.GetAvailableTitles("testplayer")
	found := false
	for _, title := range titles {
		if title == "Agent Slayer" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should have Agent Slayer title")
	}
	
	// Set title
	ok := am.SetTitle("testplayer", "Agent Slayer")
	if !ok {
		t.Error("SetTitle should succeed")
	}
	
	title := am.GetTitle("testplayer")
	if title != "Agent Slayer" {
		t.Errorf("Title = %s, want Agent Slayer", title)
	}
}

// TestPhase3LeaderboardSystem tests the leaderboard system integration
func TestPhase3LeaderboardSystem(t *testing.T) {
	lm := leaderboard.NewManager()
	
	// Update stats for multiple players
	lm.UpdateStat("player1", leaderboard.StatXP, 1000)
	lm.UpdateStat("player2", leaderboard.StatXP, 2000)
	lm.UpdateStat("player3", leaderboard.StatXP, 500)
	
	// Get leaderboard
	board := lm.GetLeaderboard(leaderboard.StatXP, 10)
	if len(board) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(board))
	}
	
	// Verify order
	if board[0].Name != "player2" {
		t.Error("First place should be player2")
	}
	if board[0].Rank != 1 {
		t.Errorf("First place rank should be 1, got %d", board[0].Rank)
	}
	
	// Get rank
	rank := lm.GetRank("player1", leaderboard.StatXP)
	if rank != 2 {
		t.Errorf("player1 rank = %d, want 2", rank)
	}
}

// TestPhase3TrainingSystem tests the training system integration
func TestPhase3TrainingSystem(t *testing.T) {
	tm := training.NewManager()
	
	// List programs
	programs := tm.ListPrograms()
	if len(programs) == 0 {
		t.Error("Should have training programs")
	}
	
	// Start program
	instance, err := tm.StartProgram("testplayer", "combat_basic")
	if err != nil {
		t.Fatalf("StartProgram failed: %v", err)
	}
	if instance == nil {
		t.Fatal("Instance should not be nil")
	}
	
	// Verify player is in program
	if !tm.IsInProgram("testplayer") {
		t.Error("Player should be in program")
	}
	
	// Record score
	tm.RecordScore("testplayer", 100)
	
	// Complete program
	rewards, score, err := tm.CompleteProgram("testplayer")
	if err != nil {
		t.Fatalf("CompleteProgram failed: %v", err)
	}
	if rewards == nil {
		t.Error("Rewards should not be nil")
	}
	if score != 100 {
		t.Errorf("Score = %d, want 100", score)
	}
	
	// Player should be removed
	if tm.IsInProgram("testplayer") {
		t.Error("Player should not be in program after completion")
	}
}

// TestPhase3PvPArena tests the PvP arena functionality
func TestPhase3PvPArena(t *testing.T) {
	tm := training.NewManager()
	
	// Start PvP program
	instance, err := tm.StartProgram("player1", "pvp_arena")
	if err != nil {
		t.Fatalf("StartProgram failed: %v", err)
	}
	
	// Second player joins
	err = tm.JoinProgram("player2", instance.ID)
	if err != nil {
		t.Fatalf("JoinProgram failed: %v", err)
	}
	
	// Verify both in arena
	if len(instance.Players) != 2 {
		t.Errorf("Expected 2 players, got %d", len(instance.Players))
	}
	
	// Third player cannot join (full)
	err = tm.JoinProgram("player3", instance.ID)
	if err == nil {
		t.Error("Third player should not be able to join")
	}
	
	// Clean up
	tm.LeaveProgram("player1")
	tm.LeaveProgram("player2")
}

// TestPhase3FactionChat tests faction chat filtering
func TestPhase3FactionChat(t *testing.T) {
	fm := faction.NewManager()
	
	fm.Join("player1", faction.FactionZion)
	fm.Join("player2", faction.FactionZion)
	fm.Join("player3", faction.FactionMachines)
	
	// Same faction check
	if !fm.IsSameFaction("player1", "player2") {
		t.Error("player1 and player2 should be same faction")
	}
	
	// Different faction check
	if fm.IsSameFaction("player1", "player3") {
		t.Error("player1 and player3 should be different factions")
	}
}

// TestPhase3AllFactions tests all faction definitions
func TestPhase3AllFactions(t *testing.T) {
	fm := faction.NewManager()
	factions := fm.GetAllFactions()
	
	if len(factions) != 3 {
		t.Errorf("Expected 3 factions, got %d", len(factions))
	}
	
	// Verify faction data
	zion := fm.GetFaction(faction.FactionZion)
	if zion == nil {
		t.Fatal("Zion faction not found")
	}
	if zion.Leader != "Morpheus" {
		t.Errorf("Zion leader = %s, want Morpheus", zion.Leader)
	}
	
	machines := fm.GetFaction(faction.FactionMachines)
	if machines == nil {
		t.Fatal("Machines faction not found")
	}
	if machines.Leader != "The Architect" {
		t.Errorf("Machines leader = %s, want The Architect", machines.Leader)
	}
	
	exiles := fm.GetFaction(faction.FactionExiles)
	if exiles == nil {
		t.Fatal("Exiles faction not found")
	}
	if exiles.Leader != "The Merovingian" {
		t.Errorf("Exiles leader = %s, want The Merovingian", exiles.Leader)
	}
}

// TestPhase3AllAchievements tests all achievement definitions
func TestPhase3AllAchievements(t *testing.T) {
	am := achievements.NewManager()
	
	// Verify all achievement constants exist
	ids := []achievements.AchievementID{
		achievements.AchFirstBlood,
		achievements.AchAgentSlayer,
		achievements.AchAwakened,
		achievements.AchReachZion,
		achievements.AchMeetOracle,
		achievements.AchPhoneMaster,
		achievements.AchPartyAnimal,
		achievements.AchCrafter,
		achievements.AchMillionaire,
		achievements.AchLevel10,
		achievements.AchLevel25,
		achievements.AchQuestComplete5,
		achievements.AchQuestComplete10,
		achievements.AchSurvivor,
		achievements.AchTheOne,
	}
	
	for _, id := range ids {
		if am.Achievements[id] == nil {
			t.Errorf("Achievement %s not found", id)
		}
	}
}

// TestPhase3AllStatTypes tests all leaderboard stat types
func TestPhase3AllStatTypes(t *testing.T) {
	lm := leaderboard.NewManager()
	
	stats := []leaderboard.StatType{
		leaderboard.StatXP,
		leaderboard.StatLevel,
		leaderboard.StatKills,
		leaderboard.StatDeaths,
		leaderboard.StatQuestsCompleted,
		leaderboard.StatMoney,
		leaderboard.StatPvPWins,
		leaderboard.StatPvPLosses,
		leaderboard.StatPlayTime,
		leaderboard.StatAchievements,
	}
	
	for _, stat := range stats {
		lm.UpdateStat("testplayer", stat, 100)
	}
	
	ps := lm.GetStats("testplayer")
	if ps.XP != 100 || ps.Level != 100 || ps.Kills != 100 {
		t.Error("Stats not updated correctly")
	}
}

// TestPhase3TrainingPrograms tests all training program definitions
func TestPhase3TrainingPrograms(t *testing.T) {
	tm := training.NewManager()
	
	programs := []string{
		"combat_basic",
		"combat_advanced",
		"survival_wave",
		"pvp_arena",
		"trial_speed",
		"kung_fu",
	}
	
	for _, id := range programs {
		p := tm.GetProgram(id)
		if p == nil {
			t.Errorf("Program %s not found", id)
		}
	}
}

// TestPhase3CommandHandlers tests that Phase 3 command handlers produce output
func TestPhase3CommandHandlers(t *testing.T) {
	// Create a mock player for testing
	player := &Player{Name: "testcmd"}
	
	// Test handleFactionCommand with "list"
	output := handleFactionCommand(player, "list")
	if !strings.Contains(output, "Zion") {
		t.Error("faction list should mention Zion")
	}
	
	// Test handleProgramsCommand
	output = handleProgramsCommand()
	if !strings.Contains(output, "Training") && !strings.Contains(output, "Program") {
		t.Error("handleProgramsCommand should mention Training or Program")
	}
	
	// Test handleChallengesCommand
	output = handleChallengesCommand()
	if !strings.Contains(output, "Challenge") && !strings.Contains(output, "CHALLENGES") {
		t.Error("handleChallengesCommand should mention Challenges")
	}
}
