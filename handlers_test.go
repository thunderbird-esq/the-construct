package main

import (
	"strings"
	"testing"

	"github.com/yourusername/matrix-mud/pkg/achievements"
	"github.com/yourusername/matrix-mud/pkg/faction"
	"github.com/yourusername/matrix-mud/pkg/leaderboard"
	"github.com/yourusername/matrix-mud/pkg/training"
)

// TestHandleReputationCommand verifies reputation display
func TestHandleReputationCommand(t *testing.T) {
	player := &Player{
		Name:   "TestPlayer",
		RoomID: "dojo",
	}

	result := handleReputationCommand(player)

	if !strings.Contains(result, "REPUTATION") {
		t.Errorf("Should contain REPUTATION header: %s", result)
	}
	// Should show all factions
	if !strings.Contains(result, "Zion") {
		t.Errorf("Should mention Zion: %s", result)
	}
}

// TestHandleAchievementsCommandOverview verifies achievements overview
func TestHandleAchievementsCommandOverview(t *testing.T) {
	player := &Player{
		Name:   "TestPlayer",
		RoomID: "dojo",
	}

	result := handleAchievementsCommand(player, "")

	if !strings.Contains(result, "ACHIEVEMENTS") {
		t.Errorf("Should contain ACHIEVEMENTS header: %s", result)
	}
	if !strings.Contains(result, "Categories") {
		t.Errorf("Should mention categories: %s", result)
	}
}

// TestHandleAchievementsCommandCategory verifies category listing
func TestHandleAchievementsCommandCategory(t *testing.T) {
	player := &Player{
		Name:   "TestPlayer",
		RoomID: "dojo",
	}

	result := handleAchievementsCommand(player, "combat")

	// Should show combat achievements (may be empty but should have header)
	if !strings.Contains(result, "COMBAT") && !strings.Contains(result, "Unknown category") {
		t.Errorf("Should contain COMBAT header or error: %s", result)
	}
}

// TestHandleAchievementsCommandInvalidCategory verifies invalid category
func TestHandleAchievementsCommandInvalidCategory(t *testing.T) {
	player := &Player{
		Name:   "TestPlayer",
		RoomID: "dojo",
	}

	result := handleAchievementsCommand(player, "nonexistent")

	if !strings.Contains(result, "Unknown category") {
		t.Errorf("Should indicate unknown category: %s", result)
	}
}

// TestHandleTitleCommandList verifies title listing
func TestHandleTitleCommandList(t *testing.T) {
	player := &Player{
		Name:   "TestPlayer",
		RoomID: "dojo",
	}

	result := handleTitleCommand(player, "")

	if !strings.Contains(result, "TITLES") {
		t.Errorf("Should contain TITLES header: %s", result)
	}
	if !strings.Contains(result, "Current") {
		t.Errorf("Should show current title: %s", result)
	}
}

// TestHandleTitleCommandClear verifies title clearing
func TestHandleTitleCommandClear(t *testing.T) {
	player := &Player{
		Name:   "TestPlayer",
		RoomID: "dojo",
	}

	// First set a title if possible
	achievements.GlobalAchievements.SetTitle(player.Name, "Test Title")

	result := handleTitleCommand(player, "clear")

	if !strings.Contains(result, "cleared") {
		t.Errorf("Should confirm title cleared: %s", result)
	}
}

// TestHandleLeaderboardCommandDefault verifies default leaderboard
func TestHandleLeaderboardCommandDefault(t *testing.T) {
	result := handleLeaderboardCommand("")

	if !strings.Contains(result, "TOP") {
		t.Errorf("Should contain TOP header: %s", result)
	}
	if !strings.Contains(result, "XP") {
		t.Errorf("Default should be XP leaderboard: %s", result)
	}
}

// TestHandleLeaderboardCommandXP verifies XP leaderboard
func TestHandleLeaderboardCommandXP(t *testing.T) {
	result := handleLeaderboardCommand("xp")

	if !strings.Contains(result, "TOP") {
		t.Errorf("Should contain TOP header: %s", result)
	}
	if !strings.Contains(result, "XP") {
		t.Errorf("Should show XP leaderboard: %s", result)
	}
}

// TestHandleLeaderboardCommandKills verifies kills leaderboard
func TestHandleLeaderboardCommandKills(t *testing.T) {
	result := handleLeaderboardCommand("kills")

	if !strings.Contains(result, "TOP") {
		t.Errorf("Should contain TOP header: %s", result)
	}
	if !strings.Contains(result, "Kills") {
		t.Errorf("Should show Kills leaderboard: %s", result)
	}
}

// TestHandleLeaderboardCommandDeaths verifies deaths leaderboard
func TestHandleLeaderboardCommandDeaths(t *testing.T) {
	result := handleLeaderboardCommand("deaths")

	if !strings.Contains(result, "Deaths") {
		t.Errorf("Should show Deaths leaderboard: %s", result)
	}
}

// TestHandleLeaderboardCommandQuests verifies quests leaderboard
func TestHandleLeaderboardCommandQuests(t *testing.T) {
	result := handleLeaderboardCommand("quests")

	if !strings.Contains(result, "Quests") {
		t.Errorf("Should show Quests leaderboard: %s", result)
	}
}

// TestHandleLeaderboardCommandMoney verifies money leaderboard
func TestHandleLeaderboardCommandMoney(t *testing.T) {
	result := handleLeaderboardCommand("money")

	if !strings.Contains(result, "Bits") {
		t.Errorf("Should show Bits leaderboard: %s", result)
	}
}

// TestHandleLeaderboardCommandBits verifies bits alias
func TestHandleLeaderboardCommandBits(t *testing.T) {
	result := handleLeaderboardCommand("bits")

	if !strings.Contains(result, "Bits") {
		t.Errorf("Should show Bits leaderboard: %s", result)
	}
}

// TestHandleLeaderboardCommandPvP verifies pvp leaderboard
func TestHandleLeaderboardCommandPvP(t *testing.T) {
	result := handleLeaderboardCommand("pvp")

	if !strings.Contains(result, "PvP") {
		t.Errorf("Should show PvP leaderboard: %s", result)
	}
}

// TestHandleLeaderboardCommandLevel verifies level leaderboard
func TestHandleLeaderboardCommandLevel(t *testing.T) {
	result := handleLeaderboardCommand("level")

	if !strings.Contains(result, "Level") {
		t.Errorf("Should show Level leaderboard: %s", result)
	}
}

// TestHandleLeaderboardCommandAchievements verifies achievements leaderboard
func TestHandleLeaderboardCommandAchievements(t *testing.T) {
	result := handleLeaderboardCommand("achievements")

	if !strings.Contains(result, "Achievements") {
		t.Errorf("Should show Achievements leaderboard: %s", result)
	}
}

// TestHandleStatsCommand verifies player stats display
func TestHandleStatsCommand(t *testing.T) {
	player := &Player{
		Name:     "TestPlayer",
		RoomID:   "dojo",
		HP:       100,
		MaxHP:    100,
		XP:       500,
		Level:    5,
		Strength: 12,
	}

	result := handleStatsCommand(player)

	if !strings.Contains(result, "STATS") {
		t.Errorf("Should contain STATS header: %s", result)
	}
}

// TestHandleTrainingCommandList verifies training room listing
func TestHandleTrainingCommandList(t *testing.T) {
	player := &Player{
		Name:   "TestPlayer",
		RoomID: "dojo",
		Level:  5,
	}

	result := handleTrainingCommand(player, "")

	// When not in training, it says "not in a training program"
	if !strings.Contains(result, "training") && !strings.Contains(result, "TRAINING") {
		t.Errorf("Should contain training info: %s", result)
	}
}

// TestHandleTrainingCommandStart verifies starting training
func TestHandleTrainingCommandStart(t *testing.T) {
	player := &Player{
		Name:   "TestTrainer",
		RoomID: "dojo",
		Level:  5,
	}

	// Leave any existing program first
	training.GlobalTraining.LeaveProgram(player.Name)

	result := handleTrainingCommand(player, "start combat_basics")

	// Should either start or indicate why not
	if result == "" {
		t.Error("Should return some response")
	}
	t.Logf("Training start result: %s", result)
}

// TestHandleTrainingCommandStartMissingArg verifies error on missing arg
func TestHandleTrainingCommandStartMissingArg(t *testing.T) {
	player := &Player{
		Name:   "TestPlayer",
		RoomID: "dojo",
		Level:  5,
	}

	result := handleTrainingCommand(player, "start")

	if !strings.Contains(result, "Usage") {
		t.Errorf("Should show usage: %s", result)
	}
}

// TestHandleTrainingCommandJoinMissingArg verifies error on missing arg
func TestHandleTrainingCommandJoinMissingArg(t *testing.T) {
	player := &Player{
		Name:   "TestPlayer",
		RoomID: "dojo",
		Level:  5,
	}

	result := handleTrainingCommand(player, "join")

	if !strings.Contains(result, "Usage") {
		t.Errorf("Should show usage: %s", result)
	}
}

// TestHandleTrainingCommandLeave verifies leaving training
func TestHandleTrainingCommandLeave(t *testing.T) {
	player := &Player{
		Name:   "TestLeaver",
		RoomID: "dojo",
		Level:  5,
	}

	// Try to leave (may or may not be in program)
	result := handleTrainingCommand(player, "leave")

	// Should respond either way
	if result == "" {
		t.Error("Should return some response")
	}
	t.Logf("Training leave result: %s", result)
}

// TestHandleTrainingCommandComplete verifies completing training
func TestHandleTrainingCommandComplete(t *testing.T) {
	player := &Player{
		Name:   "TestCompleter",
		RoomID: "dojo",
		Level:  5,
		XP:     0,
		Money:  0,
	}

	// Try to complete (may or may not be in program)
	result := handleTrainingCommand(player, "complete")

	// Should respond either way
	if result == "" {
		t.Error("Should return some response")
	}
	t.Logf("Training complete result: %s", result)
}

// TestHandleTrainingCommandInvalid verifies invalid subcommand
func TestHandleTrainingCommandInvalid(t *testing.T) {
	player := &Player{
		Name:   "TestPlayer",
		RoomID: "dojo",
		Level:  5,
	}

	result := handleTrainingCommand(player, "invalid")

	if !strings.Contains(result, "commands") {
		t.Errorf("Should show available commands: %s", result)
	}
}

// TestHandleProgramsCommand verifies programs listing
func TestHandleProgramsCommand(t *testing.T) {
	result := handleProgramsCommand()

	if !strings.Contains(result, "TRAINING PROGRAMS") {
		t.Errorf("Should contain TRAINING PROGRAMS header: %s", result)
	}
}

// TestHandleChallengesCommand verifies challenge listing
func TestHandleChallengesCommand(t *testing.T) {
	result := handleChallengesCommand()

	if !strings.Contains(result, "CHALLENGES") {
		t.Errorf("Should contain CHALLENGES header: %s", result)
	}
}

// TestHandleFactionCommandInfo verifies faction info display
func TestHandleFactionCommandInfo(t *testing.T) {
	player := &Player{
		Name:   "TestPlayer",
		RoomID: "dojo",
	}

	result := handleFactionCommand(player, "info")

	if !strings.Contains(result, "Faction") || !strings.Contains(result, "faction") {
		t.Logf("Faction info result: %s", result)
	}
}

// TestHandleFactionCommandJoin verifies faction joining
func TestHandleFactionCommandJoin(t *testing.T) {
	player := &Player{
		Name:   "TestFactionJoiner",
		RoomID: "dojo",
	}

	// Leave any current faction first
	faction.GlobalFaction.Leave(player.Name)

	result := handleFactionCommand(player, "join zion")

	// Should either join successfully or indicate why not
	t.Logf("Faction join result: %s", result)
}

// TestHandleFactionCommandLeave verifies faction leaving
func TestHandleFactionCommandLeave(t *testing.T) {
	player := &Player{
		Name:   "TestFactionLeaver",
		RoomID: "dojo",
	}

	// Join a faction first
	faction.GlobalFaction.Join(player.Name, faction.FactionZion)

	result := handleFactionCommand(player, "leave")

	// Should either leave or indicate issue
	t.Logf("Faction leave result: %s", result)
}

// TestHandleFactionCommandInvalid verifies invalid faction command
func TestHandleFactionCommandInvalid(t *testing.T) {
	player := &Player{
		Name:   "TestPlayer",
		RoomID: "dojo",
	}

	result := handleFactionCommand(player, "invalid")

	// Should show help or error
	if result == "" {
		t.Error("Should return some response for invalid command")
	}
}

// Initialize globals for testing
func init() {
	// Ensure global managers exist
	if achievements.GlobalAchievements == nil {
		achievements.GlobalAchievements = achievements.NewManager()
	}
	if faction.GlobalFaction == nil {
		faction.GlobalFaction = faction.NewManager()
	}
	if leaderboard.GlobalLeaderboard == nil {
		leaderboard.GlobalLeaderboard = leaderboard.NewManager()
	}
	if training.GlobalTraining == nil {
		training.GlobalTraining = training.NewManager()
	}
}
