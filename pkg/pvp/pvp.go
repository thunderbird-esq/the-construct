// Package pvp implements PvP arena and ranked combat for Matrix MUD.
// Supports duels, team battles, free-for-all, and tournaments.
package pvp

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

// ArenaType defines the type of arena match
type ArenaType string

const (
	ArenaDuel ArenaType = "duel" // 1v1
	ArenaTeam ArenaType = "team" // 2v2 or 3v3
	ArenaFFA  ArenaType = "ffa"  // Free-for-all
	ArenaKOTH ArenaType = "koth" // King of the Hill
)

// ArenaState tracks match state
type ArenaState string

const (
	StateWaiting  ArenaState = "waiting"
	StateStarting ArenaState = "starting"
	StateActive   ArenaState = "active"
	StateEnded    ArenaState = "ended"
)

// RankTier defines PvP ranking tiers
type RankTier string

const (
	TierBronze   RankTier = "Bronze"
	TierSilver   RankTier = "Silver"
	TierGold     RankTier = "Gold"
	TierPlatinum RankTier = "Platinum"
	TierDiamond  RankTier = "Diamond"
	TierTheOne   RankTier = "The One"
)

// ArenaPlayer represents a player in an arena match
type ArenaPlayer struct {
	Name      string
	Team      int
	HP        int
	MaxHP     int
	Damage    int
	Kills     int
	Deaths    int
	Assists   int
	Score     int
	IsAlive   bool
	LastDamageBy string
	JoinedAt  time.Time
}

// Arena represents an active arena match
type Arena struct {
	ID          string
	Type        ArenaType
	State       ArenaState
	Players     map[string]*ArenaPlayer
	Teams       map[int][]string // team number -> player names
	MaxPlayers  int
	MinPlayers  int
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	Winner      string // player name or "Team X"
	WinnerTeam  int
	Ranked      bool
	mu          sync.RWMutex
}

// PlayerStats tracks a player's PvP statistics
type PlayerStats struct {
	Name           string
	Rating         int
	Tier           RankTier
	Wins           int
	Losses         int
	Kills          int
	Deaths         int
	Assists        int
	WinStreak      int
	BestWinStreak  int
	TotalMatches   int
	LastMatchTime  time.Time
	SeasonRating   int
	SeasonWins     int
	SeasonLosses   int
}

// QueueEntry represents a player in the matchmaking queue
type QueueEntry struct {
	PlayerName string
	Rating     int
	ArenaType  ArenaType
	TeamSize   int
	QueuedAt   time.Time
}

// Tournament represents a PvP tournament
type Tournament struct {
	ID           string
	Name         string
	Type         ArenaType
	State        string // "registration", "active", "completed"
	MaxPlayers   int
	Participants []string
	Bracket      [][]*TournamentMatch
	CurrentRound int
	Winner       string
	Rewards      TournamentRewards
	StartTime    time.Time
	CreatedAt    time.Time
	mu           sync.RWMutex
}

// TournamentMatch represents a match in a tournament bracket
type TournamentMatch struct {
	ID       string
	Player1  string
	Player2  string
	Winner   string
	Score1   int
	Score2   int
	ArenaID  string
	Played   bool
}

// TournamentRewards defines tournament prizes
type TournamentRewards struct {
	FirstXP     int
	FirstMoney  int
	FirstTitle  string
	SecondXP    int
	SecondMoney int
	ThirdXP     int
	ThirdMoney  int
}

// Manager handles all PvP operations
type Manager struct {
	mu          sync.RWMutex
	Arenas      map[string]*Arena
	PlayerStats map[string]*PlayerStats
	Queue       map[ArenaType][]*QueueEntry
	Tournaments map[string]*Tournament
	nextArenaID int
	nextTournID int
}

// NewManager creates a new PvP manager
func NewManager() *Manager {
	return &Manager{
		Arenas:      make(map[string]*Arena),
		PlayerStats: make(map[string]*PlayerStats),
		Queue: map[ArenaType][]*QueueEntry{
			ArenaDuel: {},
			ArenaTeam: {},
			ArenaFFA:  {},
		},
		Tournaments: make(map[string]*Tournament),
		nextArenaID: 1,
		nextTournID: 1,
	}
}

// GetOrCreateStats gets or creates player stats
func (m *Manager) GetOrCreateStats(playerName string) *PlayerStats {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.getOrCreateStatsLocked(playerName)
}

// getOrCreateStatsLocked is internal - caller must hold lock
func (m *Manager) getOrCreateStatsLocked(playerName string) *PlayerStats {
	name := strings.ToLower(playerName)
	if stats, ok := m.PlayerStats[name]; ok {
		return stats
	}

	stats := &PlayerStats{
		Name:   playerName,
		Rating: 1000, // Starting ELO
		Tier:   TierBronze,
	}
	m.PlayerStats[name] = stats
	return stats
}

// QueueForArena adds a player to the matchmaking queue
func (m *Manager) QueueForArena(playerName string, arenaType ArenaType, teamSize int) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := strings.ToLower(playerName)

	// Check if already queued
	for _, queue := range m.Queue {
		for _, entry := range queue {
			if entry.PlayerName == name {
				return "", fmt.Errorf("you are already in queue")
			}
		}
	}

	// Check if already in arena
	for _, arena := range m.Arenas {
		if _, ok := arena.Players[name]; ok {
			return "", fmt.Errorf("you are already in an arena")
		}
	}

	// Get player rating
	stats := m.PlayerStats[name]
	rating := 1000
	if stats != nil {
		rating = stats.Rating
	}

	entry := &QueueEntry{
		PlayerName: name,
		Rating:     rating,
		ArenaType:  arenaType,
		TeamSize:   teamSize,
		QueuedAt:   time.Now(),
	}

	m.Queue[arenaType] = append(m.Queue[arenaType], entry)

	// Try to create a match
	arenaID := m.tryCreateMatch(arenaType)
	if arenaID != "" {
		return arenaID, nil
	}

	return "", nil // Still queued
}

// LeaveQueue removes a player from the queue
func (m *Manager) LeaveQueue(playerName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := strings.ToLower(playerName)
	for arenaType, queue := range m.Queue {
		for i, entry := range queue {
			if entry.PlayerName == name {
				m.Queue[arenaType] = append(queue[:i], queue[i+1:]...)
				return nil
			}
		}
	}
	return fmt.Errorf("you are not in queue")
}

// tryCreateMatch attempts to create a match from queued players
func (m *Manager) tryCreateMatch(arenaType ArenaType) string {
	queue := m.Queue[arenaType]

	var requiredPlayers int
	switch arenaType {
	case ArenaDuel:
		requiredPlayers = 2
	case ArenaTeam:
		requiredPlayers = 4 // 2v2
	case ArenaFFA:
		requiredPlayers = 4
	default:
		return ""
	}

	if len(queue) < requiredPlayers {
		return ""
	}

	// Sort by rating for better matchmaking
	sort.Slice(queue, func(i, j int) bool {
		return queue[i].Rating < queue[j].Rating
	})

	// Take first requiredPlayers players
	players := queue[:requiredPlayers]
	m.Queue[arenaType] = queue[requiredPlayers:]

	// Create arena
	arena := m.createArena(arenaType, players)
	return arena.ID
}

// createArena creates a new arena with the given players
func (m *Manager) createArena(arenaType ArenaType, entries []*QueueEntry) *Arena {
	arenaID := fmt.Sprintf("arena_%d", m.nextArenaID)
	m.nextArenaID++

	arena := &Arena{
		ID:         arenaID,
		Type:       arenaType,
		State:      StateStarting,
		Players:    make(map[string]*ArenaPlayer),
		Teams:      make(map[int][]string),
		MaxPlayers: len(entries),
		MinPlayers: len(entries),
		StartTime:  time.Now().Add(10 * time.Second), // 10 second countdown
		Duration:   5 * time.Minute,
		Ranked:     true,
	}

	// Assign players to teams
	for i, entry := range entries {
		team := 0
		if arenaType == ArenaTeam {
			team = i%2 + 1
		} else if arenaType == ArenaDuel {
			team = i + 1
		}

		player := &ArenaPlayer{
			Name:     entry.PlayerName,
			Team:     team,
			HP:       100,
			MaxHP:    100,
			Damage:   10,
			IsAlive:  true,
			JoinedAt: time.Now(),
		}
		arena.Players[entry.PlayerName] = player

		if team > 0 {
			arena.Teams[team] = append(arena.Teams[team], entry.PlayerName)
		}
	}

	m.Arenas[arenaID] = arena
	return arena
}

// StartArena transitions an arena to active state
func (m *Manager) StartArena(arenaID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	arena, ok := m.Arenas[arenaID]
	if !ok {
		return fmt.Errorf("arena not found")
	}

	arena.mu.Lock()
	defer arena.mu.Unlock()

	if arena.State != StateStarting {
		return fmt.Errorf("arena is not in starting state")
	}

	arena.State = StateActive
	arena.StartTime = time.Now()
	arena.EndTime = time.Now().Add(arena.Duration)
	return nil
}

// AttackPlayer handles combat in an arena
func (m *Manager) AttackPlayer(arenaID, attackerName, targetName string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	arena, ok := m.Arenas[arenaID]
	if !ok {
		return "", fmt.Errorf("arena not found")
	}

	arena.mu.Lock()
	defer arena.mu.Unlock()

	if arena.State != StateActive {
		return "", fmt.Errorf("match is not active")
	}

	attacker, ok := arena.Players[strings.ToLower(attackerName)]
	if !ok || !attacker.IsAlive {
		return "", fmt.Errorf("you cannot attack")
	}

	target, ok := arena.Players[strings.ToLower(targetName)]
	if !ok {
		return "", fmt.Errorf("target not found")
	}

	if !target.IsAlive {
		return "", fmt.Errorf("target is already dead")
	}

	// Can't attack teammates
	if arena.Type == ArenaTeam && attacker.Team == target.Team {
		return "", fmt.Errorf("you cannot attack your teammate")
	}

	// Deal damage
	damage := attacker.Damage
	target.HP -= damage
	target.LastDamageBy = attacker.Name

	var msg string
	if target.HP <= 0 {
		target.HP = 0
		target.IsAlive = false
		target.Deaths++
		attacker.Kills++
		attacker.Score += 100

		msg = fmt.Sprintf("You killed %s! (+100 points)", target.Name)

		// Check for match end
		if m.checkMatchEnd(arena) {
			msg += "\r\nMATCH OVER!"
		}
	} else {
		msg = fmt.Sprintf("You hit %s for %d damage. (%d/%d HP)", target.Name, damage, target.HP, target.MaxHP)
	}

	return msg, nil
}

// checkMatchEnd checks if the match should end
func (m *Manager) checkMatchEnd(arena *Arena) bool {
	switch arena.Type {
	case ArenaDuel:
		// One player left
		alive := 0
		var winner string
		for _, p := range arena.Players {
			if p.IsAlive {
				alive++
				winner = p.Name
			}
		}
		if alive <= 1 {
			arena.State = StateEnded
			arena.EndTime = time.Now()
			arena.Winner = winner
			m.processMatchEnd(arena)
			return true
		}

	case ArenaTeam:
		// One team eliminated
		for teamNum, members := range arena.Teams {
			teamAlive := false
			for _, name := range members {
				if p := arena.Players[name]; p != nil && p.IsAlive {
					teamAlive = true
					break
				}
			}
			if !teamAlive {
				// Other team wins
				arena.State = StateEnded
				arena.EndTime = time.Now()
				if teamNum == 1 {
					arena.WinnerTeam = 2
					arena.Winner = "Team 2"
				} else {
					arena.WinnerTeam = 1
					arena.Winner = "Team 1"
				}
				m.processMatchEnd(arena)
				return true
			}
		}

	case ArenaFFA:
		// One player left
		alive := 0
		var winner string
		for _, p := range arena.Players {
			if p.IsAlive {
				alive++
				winner = p.Name
			}
		}
		if alive <= 1 {
			arena.State = StateEnded
			arena.EndTime = time.Now()
			arena.Winner = winner
			m.processMatchEnd(arena)
			return true
		}
	}

	return false
}

// processMatchEnd updates stats and ratings after a match
func (m *Manager) processMatchEnd(arena *Arena) {
	if !arena.Ranked {
		return
	}

	// Determine winners and losers
	var winners, losers []*ArenaPlayer
	for _, p := range arena.Players {
		if arena.Type == ArenaTeam {
			if p.Team == arena.WinnerTeam {
				winners = append(winners, p)
			} else {
				losers = append(losers, p)
			}
		} else {
			if p.Name == arena.Winner {
				winners = append(winners, p)
			} else {
				losers = append(losers, p)
			}
		}
	}

	// Calculate average ratings
	winnerRating := 0
	loserRating := 0
	for _, p := range winners {
		stats := m.getOrCreateStatsLocked(p.Name)
		winnerRating += stats.Rating
	}
	for _, p := range losers {
		stats := m.getOrCreateStatsLocked(p.Name)
		loserRating += stats.Rating
	}
	if len(winners) > 0 {
		winnerRating /= len(winners)
	}
	if len(losers) > 0 {
		loserRating /= len(losers)
	}

	// ELO calculation
	K := 32.0
	expectedWinner := 1.0 / (1.0 + math.Pow(10, float64(loserRating-winnerRating)/400.0))
	ratingChange := int(K * (1.0 - expectedWinner))

	// Update winner stats
	for _, p := range winners {
		stats := m.getOrCreateStatsLocked(p.Name)
		stats.Rating += ratingChange
		stats.Wins++
		stats.SeasonWins++
		stats.Kills += p.Kills
		stats.Deaths += p.Deaths
		stats.Assists += p.Assists
		stats.TotalMatches++
		stats.WinStreak++
		if stats.WinStreak > stats.BestWinStreak {
			stats.BestWinStreak = stats.WinStreak
		}
		stats.LastMatchTime = time.Now()
		stats.Tier = calculateTier(stats.Rating)
	}

	// Update loser stats
	for _, p := range losers {
		stats := m.getOrCreateStatsLocked(p.Name)
		stats.Rating -= ratingChange
		if stats.Rating < 0 {
			stats.Rating = 0
		}
		stats.Losses++
		stats.SeasonLosses++
		stats.Kills += p.Kills
		stats.Deaths += p.Deaths
		stats.Assists += p.Assists
		stats.TotalMatches++
		stats.WinStreak = 0
		stats.LastMatchTime = time.Now()
		stats.Tier = calculateTier(stats.Rating)
	}
}

// calculateTier determines rank tier from rating
func calculateTier(rating int) RankTier {
	switch {
	case rating >= 2500:
		return TierTheOne
	case rating >= 2000:
		return TierDiamond
	case rating >= 1600:
		return TierPlatinum
	case rating >= 1300:
		return TierGold
	case rating >= 1000:
		return TierSilver
	default:
		return TierBronze
	}
}

// GetPlayerArena returns the arena a player is in
func (m *Manager) GetPlayerArena(playerName string) *Arena {
	m.mu.RLock()
	defer m.mu.RUnlock()

	name := strings.ToLower(playerName)
	for _, arena := range m.Arenas {
		if _, ok := arena.Players[name]; ok {
			return arena
		}
	}
	return nil
}

// LeaveArena removes a player from their arena
func (m *Manager) LeaveArena(playerName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := strings.ToLower(playerName)
	for arenaID, arena := range m.Arenas {
		if player, ok := arena.Players[name]; ok {
			arena.mu.Lock()
			
			// Mark as dead/disconnected
			player.IsAlive = false
			player.Deaths++
			
			// Check if match should end
			m.checkMatchEnd(arena)
			
			// Remove from arena if ended
			if arena.State == StateEnded {
				delete(m.Arenas, arenaID)
			}
			
			arena.mu.Unlock()
			return nil
		}
	}
	return fmt.Errorf("you are not in an arena")
}

// GetStats returns a player's PvP stats formatted for display
func (m *Manager) GetStats(playerName string) string {
	stats := m.GetOrCreateStats(playerName)

	var sb strings.Builder
	sb.WriteString("=== PVP STATISTICS ===\r\n\r\n")
	sb.WriteString(fmt.Sprintf("Rating: %d (%s)\r\n", stats.Rating, stats.Tier))
	sb.WriteString(fmt.Sprintf("Record: %d-%d (%.1f%% win rate)\r\n", 
		stats.Wins, stats.Losses, 
		float64(stats.Wins)/float64(max(1, stats.Wins+stats.Losses))*100))
	sb.WriteString(fmt.Sprintf("K/D/A: %d/%d/%d\r\n", stats.Kills, stats.Deaths, stats.Assists))
	sb.WriteString(fmt.Sprintf("Best Win Streak: %d\r\n", stats.BestWinStreak))
	sb.WriteString(fmt.Sprintf("Total Matches: %d\r\n", stats.TotalMatches))

	return sb.String()
}

// GetRankings returns the PvP leaderboard
func (m *Manager) GetRankings(limit int) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Collect and sort stats
	var allStats []*PlayerStats
	for _, stats := range m.PlayerStats {
		allStats = append(allStats, stats)
	}
	sort.Slice(allStats, func(i, j int) bool {
		return allStats[i].Rating > allStats[j].Rating
	})

	if limit > len(allStats) {
		limit = len(allStats)
	}

	var sb strings.Builder
	sb.WriteString("=== PVP LEADERBOARD ===\r\n\r\n")
	sb.WriteString("Rank  Player          Rating    W-L      Tier\r\n")
	sb.WriteString("----  --------------  ------    -----    --------\r\n")

	for i := 0; i < limit && i < len(allStats); i++ {
		s := allStats[i]
		sb.WriteString(fmt.Sprintf("%-4d  %-14s  %-6d    %d-%-4d   %s\r\n",
			i+1, s.Name, s.Rating, s.Wins, s.Losses, s.Tier))
	}

	return sb.String()
}

// CreateTournament creates a new tournament
func (m *Manager) CreateTournament(name string, arenaType ArenaType, maxPlayers int, rewards TournamentRewards) (*Tournament, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate max players (must be power of 2)
	valid := false
	for p := 2; p <= 64; p *= 2 {
		if maxPlayers == p {
			valid = true
			break
		}
	}
	if !valid {
		return nil, fmt.Errorf("max players must be a power of 2 (2, 4, 8, 16, 32, 64)")
	}

	tournID := fmt.Sprintf("tourn_%d", m.nextTournID)
	m.nextTournID++

	tournament := &Tournament{
		ID:           tournID,
		Name:         name,
		Type:         arenaType,
		State:        "registration",
		MaxPlayers:   maxPlayers,
		Participants: make([]string, 0),
		Bracket:      make([][]*TournamentMatch, 0),
		Rewards:      rewards,
		CreatedAt:    time.Now(),
	}

	m.Tournaments[tournID] = tournament
	return tournament, nil
}

// JoinTournament adds a player to a tournament
func (m *Manager) JoinTournament(tournamentID, playerName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := strings.ToLower(playerName)
	tournament, ok := m.Tournaments[tournamentID]
	if !ok {
		return fmt.Errorf("tournament not found")
	}

	tournament.mu.Lock()
	defer tournament.mu.Unlock()

	if tournament.State != "registration" {
		return fmt.Errorf("tournament registration is closed")
	}

	// Check if already registered
	for _, p := range tournament.Participants {
		if p == name {
			return fmt.Errorf("you are already registered")
		}
	}

	if len(tournament.Participants) >= tournament.MaxPlayers {
		return fmt.Errorf("tournament is full")
	}

	tournament.Participants = append(tournament.Participants, name)
	return nil
}

// StartTournament begins a tournament and generates the bracket
func (m *Manager) StartTournament(tournamentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tournament, ok := m.Tournaments[tournamentID]
	if !ok {
		return fmt.Errorf("tournament not found")
	}

	tournament.mu.Lock()
	defer tournament.mu.Unlock()

	if tournament.State != "registration" {
		return fmt.Errorf("tournament is not in registration")
	}

	if len(tournament.Participants) < 2 {
		return fmt.Errorf("need at least 2 participants")
	}

	// Generate bracket
	tournament.Bracket = generateBracket(tournament.Participants)
	tournament.State = "active"
	tournament.CurrentRound = 0
	tournament.StartTime = time.Now()

	return nil
}

// generateBracket creates a single-elimination bracket
func generateBracket(participants []string) [][]*TournamentMatch {
	// Shuffle participants
	shuffled := make([]string, len(participants))
	copy(shuffled, participants)
	// Simple shuffle (not cryptographically secure, but fine for games)
	for i := len(shuffled) - 1; i > 0; i-- {
		j := int(time.Now().UnixNano()) % (i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	// Calculate number of rounds
	numRounds := int(math.Ceil(math.Log2(float64(len(shuffled)))))
	bracket := make([][]*TournamentMatch, numRounds)

	// First round
	firstRound := make([]*TournamentMatch, 0)
	for i := 0; i < len(shuffled); i += 2 {
		match := &TournamentMatch{
			ID:      fmt.Sprintf("match_%d", i/2+1),
			Player1: shuffled[i],
		}
		if i+1 < len(shuffled) {
			match.Player2 = shuffled[i+1]
		} else {
			// Bye
			match.Winner = shuffled[i]
			match.Played = true
		}
		firstRound = append(firstRound, match)
	}
	bracket[0] = firstRound

	// Subsequent rounds (empty placeholders)
	for r := 1; r < numRounds; r++ {
		roundMatches := make([]*TournamentMatch, len(bracket[r-1])/2)
		for i := range roundMatches {
			roundMatches[i] = &TournamentMatch{
				ID: fmt.Sprintf("r%d_match_%d", r+1, i+1),
			}
		}
		bracket[r] = roundMatches
	}

	return bracket
}

// GetTournamentBracket returns the bracket display
func (m *Manager) GetTournamentBracket(tournamentID string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tournament, ok := m.Tournaments[tournamentID]
	if !ok {
		return "Tournament not found."
	}

	tournament.mu.RLock()
	defer tournament.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== %s ===\r\n\r\n", tournament.Name))
	sb.WriteString(fmt.Sprintf("Status: %s | Participants: %d/%d\r\n\r\n",
		tournament.State, len(tournament.Participants), tournament.MaxPlayers))

	if tournament.State == "registration" {
		sb.WriteString("Registered Players:\r\n")
		for i, p := range tournament.Participants {
			sb.WriteString(fmt.Sprintf("  %d. %s\r\n", i+1, p))
		}
		return sb.String()
	}

	for r, round := range tournament.Bracket {
		sb.WriteString(fmt.Sprintf("--- Round %d ---\r\n", r+1))
		for _, match := range round {
			p1 := match.Player1
			p2 := match.Player2
			if p1 == "" {
				p1 = "TBD"
			}
			if p2 == "" {
				p2 = "BYE"
			}

			status := "Pending"
			if match.Played {
				status = fmt.Sprintf("%s wins", match.Winner)
			}

			sb.WriteString(fmt.Sprintf("  %s vs %s - %s\r\n", p1, p2, status))
		}
		sb.WriteString("\r\n")
	}

	if tournament.Winner != "" {
		sb.WriteString(fmt.Sprintf("CHAMPION: %s\r\n", tournament.Winner))
	}

	return sb.String()
}

// ListTournaments returns active tournaments
func (m *Manager) ListTournaments() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString("=== TOURNAMENTS ===\r\n\r\n")

	if len(m.Tournaments) == 0 {
		sb.WriteString("No active tournaments.\r\n")
		return sb.String()
	}

	for _, t := range m.Tournaments {
		t.mu.RLock()
		sb.WriteString(fmt.Sprintf("%s - %s (%s)\r\n", t.ID, t.Name, t.State))
		sb.WriteString(fmt.Sprintf("  Type: %s | Players: %d/%d\r\n\r\n",
			t.Type, len(t.Participants), t.MaxPlayers))
		t.mu.RUnlock()
	}

	return sb.String()
}

// IsQueued checks if a player is in queue
func (m *Manager) IsQueued(playerName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	name := strings.ToLower(playerName)
	for _, queue := range m.Queue {
		for _, entry := range queue {
			if entry.PlayerName == name {
				return true
			}
		}
	}
	return false
}

// IsInArena checks if a player is in an arena
func (m *Manager) IsInArena(playerName string) bool {
	return m.GetPlayerArena(playerName) != nil
}

// Helper function
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Global PvP manager
var GlobalPvP = NewManager()
