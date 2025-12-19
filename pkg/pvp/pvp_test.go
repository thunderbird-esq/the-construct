package pvp

import (
	"strings"
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.Arenas == nil {
		t.Error("Arenas map is nil")
	}
	if m.PlayerStats == nil {
		t.Error("PlayerStats map is nil")
	}
	if m.Queue == nil {
		t.Error("Queue map is nil")
	}
}

func TestGetOrCreateStats(t *testing.T) {
	m := NewManager()

	stats := m.GetOrCreateStats("TestPlayer")
	if stats == nil {
		t.Fatal("GetOrCreateStats returned nil")
	}
	if stats.Rating != 1000 {
		t.Errorf("Initial rating should be 1000, got %d", stats.Rating)
	}
	if stats.Tier != TierBronze {
		t.Errorf("Initial tier should be Bronze, got %s", stats.Tier)
	}

	// Get same player again
	stats2 := m.GetOrCreateStats("TestPlayer")
	if stats != stats2 {
		t.Error("Should return same stats object")
	}
}

func TestQueueForArena(t *testing.T) {
	m := NewManager()

	arenaID, err := m.QueueForArena("Player1", ArenaDuel, 1)
	if err != nil {
		t.Fatalf("QueueForArena failed: %v", err)
	}

	// Should be empty string (no match yet)
	if arenaID != "" {
		t.Error("Should not create arena with only 1 player")
	}

	// Queue should have 1 player
	if len(m.Queue[ArenaDuel]) != 1 {
		t.Errorf("Queue should have 1 player, got %d", len(m.Queue[ArenaDuel]))
	}
}

func TestQueueForArenaTwice(t *testing.T) {
	m := NewManager()

	m.QueueForArena("Player1", ArenaDuel, 1)
	_, err := m.QueueForArena("Player1", ArenaDuel, 1)
	if err == nil {
		t.Error("Should not allow queuing twice")
	}
}

func TestMatchCreation(t *testing.T) {
	m := NewManager()

	// Queue two players for duel
	m.QueueForArena("Player1", ArenaDuel, 1)
	arenaID, err := m.QueueForArena("Player2", ArenaDuel, 1)

	if err != nil {
		t.Fatalf("Second queue failed: %v", err)
	}

	if arenaID == "" {
		t.Error("Should create arena with 2 players")
	}

	// Queue should be empty
	if len(m.Queue[ArenaDuel]) != 0 {
		t.Error("Queue should be empty after match creation")
	}

	// Arena should exist
	arena := m.Arenas[arenaID]
	if arena == nil {
		t.Fatal("Arena should exist")
	}

	if len(arena.Players) != 2 {
		t.Errorf("Arena should have 2 players, got %d", len(arena.Players))
	}
}

func TestLeaveQueue(t *testing.T) {
	m := NewManager()

	m.QueueForArena("Player1", ArenaDuel, 1)
	err := m.LeaveQueue("Player1")
	if err != nil {
		t.Fatalf("LeaveQueue failed: %v", err)
	}

	if len(m.Queue[ArenaDuel]) != 0 {
		t.Error("Queue should be empty")
	}
}

func TestLeaveQueueNotIn(t *testing.T) {
	m := NewManager()

	err := m.LeaveQueue("NotQueued")
	if err == nil {
		t.Error("Should fail when not in queue")
	}
}

func TestStartArena(t *testing.T) {
	m := NewManager()

	// Create arena with 2 players
	m.QueueForArena("Player1", ArenaDuel, 1)
	arenaID, _ := m.QueueForArena("Player2", ArenaDuel, 1)

	err := m.StartArena(arenaID)
	if err != nil {
		t.Fatalf("StartArena failed: %v", err)
	}

	arena := m.Arenas[arenaID]
	if arena.State != StateActive {
		t.Errorf("Arena state should be active, got %s", arena.State)
	}
}

func TestAttackPlayer(t *testing.T) {
	m := NewManager()

	m.QueueForArena("Player1", ArenaDuel, 1)
	arenaID, _ := m.QueueForArena("Player2", ArenaDuel, 1)
	m.StartArena(arenaID)

	msg, err := m.AttackPlayer(arenaID, "Player1", "Player2")
	if err != nil {
		t.Fatalf("AttackPlayer failed: %v", err)
	}

	if !strings.Contains(msg, "hit") && !strings.Contains(msg, "killed") {
		t.Errorf("Unexpected attack message: %s", msg)
	}
}

func TestAttackPlayerNotInArena(t *testing.T) {
	m := NewManager()

	_, err := m.AttackPlayer("fake_arena", "Player1", "Player2")
	if err == nil {
		t.Error("Should fail for invalid arena")
	}
}

func TestAttackTeammate(t *testing.T) {
	m := NewManager()

	// Create team arena
	m.QueueForArena("Player1", ArenaTeam, 2)
	m.QueueForArena("Player2", ArenaTeam, 2)
	m.QueueForArena("Player3", ArenaTeam, 2)
	arenaID, _ := m.QueueForArena("Player4", ArenaTeam, 2)
	m.StartArena(arenaID)

	arena := m.Arenas[arenaID]

	// Find teammates
	var teammate1, teammate2 string
	for name, p := range arena.Players {
		if p.Team == 1 {
			if teammate1 == "" {
				teammate1 = name
			} else {
				teammate2 = name
			}
		}
	}

	_, err := m.AttackPlayer(arenaID, teammate1, teammate2)
	if err == nil {
		t.Error("Should not allow attacking teammates")
	}
}

func TestMatchEnd(t *testing.T) {
	m := NewManager()

	m.QueueForArena("Player1", ArenaDuel, 1)
	arenaID, _ := m.QueueForArena("Player2", ArenaDuel, 1)
	m.StartArena(arenaID)

	// Kill player2 (attack until dead)
	for i := 0; i < 20; i++ {
		msg, _ := m.AttackPlayer(arenaID, "Player1", "Player2")
		if strings.Contains(msg, "MATCH OVER") {
			break
		}
	}

	arena := m.Arenas[arenaID]
	if arena.State != StateEnded {
		t.Error("Match should have ended")
	}
	if arena.Winner != "player1" {
		t.Errorf("Winner should be player1, got %s", arena.Winner)
	}
}

func TestRatingChange(t *testing.T) {
	m := NewManager()

	// Create and complete a match
	m.QueueForArena("Winner", ArenaDuel, 1)
	arenaID, _ := m.QueueForArena("Loser", ArenaDuel, 1)
	m.StartArena(arenaID)

	// Kill loser
	for i := 0; i < 20; i++ {
		m.AttackPlayer(arenaID, "Winner", "Loser")
	}

	winnerStats := m.GetOrCreateStats("Winner")
	loserStats := m.GetOrCreateStats("Loser")

	if winnerStats.Rating <= 1000 {
		t.Error("Winner rating should increase")
	}
	if loserStats.Rating >= 1000 {
		t.Error("Loser rating should decrease")
	}
}

func TestCalculateTier(t *testing.T) {
	tests := []struct {
		rating int
		tier   RankTier
	}{
		{500, TierBronze},
		{1000, TierSilver},
		{1300, TierGold},
		{1600, TierPlatinum},
		{2000, TierDiamond},
		{2500, TierTheOne},
		{3000, TierTheOne},
	}

	for _, tt := range tests {
		got := calculateTier(tt.rating)
		if got != tt.tier {
			t.Errorf("calculateTier(%d) = %s, want %s", tt.rating, got, tt.tier)
		}
	}
}

func TestGetStats(t *testing.T) {
	m := NewManager()

	output := m.GetStats("TestPlayer")
	if !strings.Contains(output, "Rating") {
		t.Error("Should show rating")
	}
	if !strings.Contains(output, "Bronze") {
		t.Error("Should show tier")
	}
}

func TestGetRankings(t *testing.T) {
	m := NewManager()

	// Create some players with different ratings
	stats1 := m.GetOrCreateStats("HighRated")
	stats1.Rating = 2000
	stats1.Wins = 100
	stats1.Losses = 20

	stats2 := m.GetOrCreateStats("LowRated")
	stats2.Rating = 800
	stats2.Wins = 10
	stats2.Losses = 50

	output := m.GetRankings(10)
	if !strings.Contains(output, "HighRated") {
		t.Error("Should show HighRated player")
	}
	if !strings.Contains(output, "LowRated") {
		t.Error("Should show LowRated player")
	}
}

func TestCreateTournament(t *testing.T) {
	m := NewManager()

	rewards := TournamentRewards{
		FirstXP:    1000,
		FirstMoney: 500,
		FirstTitle: "Champion",
	}

	tournament, err := m.CreateTournament("Test Tournament", ArenaDuel, 8, rewards)
	if err != nil {
		t.Fatalf("CreateTournament failed: %v", err)
	}

	if tournament == nil {
		t.Fatal("Tournament is nil")
	}
	if tournament.State != "registration" {
		t.Error("Initial state should be registration")
	}
}

func TestCreateTournamentInvalidSize(t *testing.T) {
	m := NewManager()

	_, err := m.CreateTournament("Test", ArenaDuel, 5, TournamentRewards{})
	if err == nil {
		t.Error("Should fail for non-power-of-2 size")
	}
}

func TestJoinTournament(t *testing.T) {
	m := NewManager()

	tournament, _ := m.CreateTournament("Test", ArenaDuel, 4, TournamentRewards{})

	err := m.JoinTournament(tournament.ID, "Player1")
	if err != nil {
		t.Fatalf("JoinTournament failed: %v", err)
	}

	if len(tournament.Participants) != 1 {
		t.Error("Should have 1 participant")
	}
}

func TestJoinTournamentTwice(t *testing.T) {
	m := NewManager()

	tournament, _ := m.CreateTournament("Test", ArenaDuel, 4, TournamentRewards{})

	m.JoinTournament(tournament.ID, "Player1")
	err := m.JoinTournament(tournament.ID, "Player1")
	if err == nil {
		t.Error("Should not allow joining twice")
	}
}

func TestJoinTournamentFull(t *testing.T) {
	m := NewManager()

	tournament, _ := m.CreateTournament("Test", ArenaDuel, 2, TournamentRewards{})

	m.JoinTournament(tournament.ID, "Player1")
	m.JoinTournament(tournament.ID, "Player2")
	err := m.JoinTournament(tournament.ID, "Player3")
	if err == nil {
		t.Error("Should not allow joining full tournament")
	}
}

func TestStartTournament(t *testing.T) {
	m := NewManager()

	tournament, _ := m.CreateTournament("Test", ArenaDuel, 4, TournamentRewards{})

	m.JoinTournament(tournament.ID, "Player1")
	m.JoinTournament(tournament.ID, "Player2")
	m.JoinTournament(tournament.ID, "Player3")
	m.JoinTournament(tournament.ID, "Player4")

	err := m.StartTournament(tournament.ID)
	if err != nil {
		t.Fatalf("StartTournament failed: %v", err)
	}

	if tournament.State != "active" {
		t.Error("State should be active")
	}
	if len(tournament.Bracket) == 0 {
		t.Error("Bracket should be generated")
	}
}

func TestStartTournamentNotEnoughPlayers(t *testing.T) {
	m := NewManager()

	tournament, _ := m.CreateTournament("Test", ArenaDuel, 4, TournamentRewards{})
	m.JoinTournament(tournament.ID, "Player1")

	err := m.StartTournament(tournament.ID)
	if err == nil {
		t.Error("Should not start with < 2 players")
	}
}

func TestGetTournamentBracket(t *testing.T) {
	m := NewManager()

	tournament, _ := m.CreateTournament("Test Tournament", ArenaDuel, 4, TournamentRewards{})
	m.JoinTournament(tournament.ID, "Player1")
	m.JoinTournament(tournament.ID, "Player2")
	m.JoinTournament(tournament.ID, "Player3")
	m.JoinTournament(tournament.ID, "Player4")
	m.StartTournament(tournament.ID)

	output := m.GetTournamentBracket(tournament.ID)
	if !strings.Contains(output, "Round") {
		t.Error("Should show rounds")
	}
	if !strings.Contains(output, "vs") {
		t.Error("Should show matchups")
	}
}

func TestListTournaments(t *testing.T) {
	m := NewManager()

	m.CreateTournament("Tournament 1", ArenaDuel, 4, TournamentRewards{})
	m.CreateTournament("Tournament 2", ArenaTeam, 8, TournamentRewards{})

	output := m.ListTournaments()
	if !strings.Contains(output, "Tournament 1") {
		t.Error("Should list Tournament 1")
	}
	if !strings.Contains(output, "Tournament 2") {
		t.Error("Should list Tournament 2")
	}
}

func TestIsQueued(t *testing.T) {
	m := NewManager()

	if m.IsQueued("Player1") {
		t.Error("Should not be queued initially")
	}

	m.QueueForArena("Player1", ArenaDuel, 1)

	if !m.IsQueued("Player1") {
		t.Error("Should be queued after QueueForArena")
	}
}

func TestIsInArena(t *testing.T) {
	m := NewManager()

	if m.IsInArena("Player1") {
		t.Error("Should not be in arena initially")
	}

	m.QueueForArena("Player1", ArenaDuel, 1)
	m.QueueForArena("Player2", ArenaDuel, 1)

	if !m.IsInArena("Player1") {
		t.Error("Should be in arena after match created")
	}
}

func TestLeaveArena(t *testing.T) {
	m := NewManager()

	m.QueueForArena("Player1", ArenaDuel, 1)
	m.QueueForArena("Player2", ArenaDuel, 1)

	err := m.LeaveArena("Player1")
	if err != nil {
		t.Fatalf("LeaveArena failed: %v", err)
	}
}

func TestLeaveArenaNotIn(t *testing.T) {
	m := NewManager()

	err := m.LeaveArena("NotInArena")
	if err == nil {
		t.Error("Should fail when not in arena")
	}
}

func TestGetPlayerArena(t *testing.T) {
	m := NewManager()

	if m.GetPlayerArena("Player1") != nil {
		t.Error("Should return nil when not in arena")
	}

	m.QueueForArena("Player1", ArenaDuel, 1)
	m.QueueForArena("Player2", ArenaDuel, 1)

	arena := m.GetPlayerArena("Player1")
	if arena == nil {
		t.Error("Should return arena when in one")
	}
}

func TestGenerateBracket(t *testing.T) {
	participants := []string{"p1", "p2", "p3", "p4"}
	bracket := generateBracket(participants)

	if len(bracket) == 0 {
		t.Error("Bracket should not be empty")
	}

	// First round should have 2 matches
	if len(bracket[0]) != 2 {
		t.Errorf("First round should have 2 matches, got %d", len(bracket[0]))
	}
}

func TestFFAMatch(t *testing.T) {
	m := NewManager()

	// Queue 4 players for FFA
	m.QueueForArena("Player1", ArenaFFA, 1)
	m.QueueForArena("Player2", ArenaFFA, 1)
	m.QueueForArena("Player3", ArenaFFA, 1)
	arenaID, _ := m.QueueForArena("Player4", ArenaFFA, 1)

	if arenaID == "" {
		t.Fatal("Should create FFA arena")
	}

	arena := m.Arenas[arenaID]
	if arena.Type != ArenaFFA {
		t.Errorf("Arena type should be FFA, got %s", arena.Type)
	}
	if len(arena.Players) != 4 {
		t.Errorf("FFA arena should have 4 players, got %d", len(arena.Players))
	}
}
