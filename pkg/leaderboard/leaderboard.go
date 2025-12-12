// Package leaderboard implements player rankings for Matrix MUD.
// Tracks various statistics and provides sorted leaderboards.
package leaderboard

import (
	"encoding/json"
	"os"
	"sort"
	"sync"
)

// StatType represents a tracked statistic
type StatType string

const (
	StatXP          StatType = "xp"
	StatLevel       StatType = "level"
	StatKills       StatType = "kills"
	StatDeaths      StatType = "deaths"
	StatQuestsCompleted StatType = "quests"
	StatMoney       StatType = "money"
	StatPvPWins     StatType = "pvp_wins"
	StatPvPLosses   StatType = "pvp_losses"
	StatPlayTime    StatType = "play_time"
	StatAchievements StatType = "achievements"
)

// PlayerStats holds all tracked statistics for a player
type PlayerStats struct {
	Name            string `json:"name"`
	XP              int    `json:"xp"`
	Level           int    `json:"level"`
	Kills           int    `json:"kills"`
	Deaths          int    `json:"deaths"`
	QuestsCompleted int    `json:"quests_completed"`
	Money           int    `json:"money"`
	PvPWins         int    `json:"pvp_wins"`
	PvPLosses       int    `json:"pvp_losses"`
	PlayTimeMinutes int    `json:"play_time_minutes"`
	Achievements    int    `json:"achievements"`
}

// Entry represents a leaderboard entry
type Entry struct {
	Rank  int    `json:"rank"`
	Name  string `json:"name"`
	Value int    `json:"value"`
}

// Manager handles leaderboard operations
type Manager struct {
	mu      sync.RWMutex
	Players map[string]*PlayerStats
}

// NewManager creates a new leaderboard manager
func NewManager() *Manager {
	return &Manager{
		Players: make(map[string]*PlayerStats),
	}
}

// GetStats returns a player's stats, creating if needed
func (m *Manager) GetStats(playerName string) *PlayerStats {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ps, ok := m.Players[playerName]; ok {
		return ps
	}

	ps := &PlayerStats{Name: playerName}
	m.Players[playerName] = ps
	return ps
}

// UpdateStat updates a specific stat for a player
func (m *Manager) UpdateStat(playerName string, stat StatType, value int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ps := m.Players[playerName]
	if ps == nil {
		ps = &PlayerStats{Name: playerName}
		m.Players[playerName] = ps
	}

	switch stat {
	case StatXP:
		ps.XP = value
	case StatLevel:
		ps.Level = value
	case StatKills:
		ps.Kills = value
	case StatDeaths:
		ps.Deaths = value
	case StatQuestsCompleted:
		ps.QuestsCompleted = value
	case StatMoney:
		ps.Money = value
	case StatPvPWins:
		ps.PvPWins = value
	case StatPvPLosses:
		ps.PvPLosses = value
	case StatPlayTime:
		ps.PlayTimeMinutes = value
	case StatAchievements:
		ps.Achievements = value
	}
}

// IncrementStat increments a stat by delta
func (m *Manager) IncrementStat(playerName string, stat StatType, delta int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ps := m.Players[playerName]
	if ps == nil {
		ps = &PlayerStats{Name: playerName}
		m.Players[playerName] = ps
	}

	switch stat {
	case StatXP:
		ps.XP += delta
	case StatLevel:
		ps.Level += delta
	case StatKills:
		ps.Kills += delta
	case StatDeaths:
		ps.Deaths += delta
	case StatQuestsCompleted:
		ps.QuestsCompleted += delta
	case StatMoney:
		ps.Money += delta
	case StatPvPWins:
		ps.PvPWins += delta
	case StatPvPLosses:
		ps.PvPLosses += delta
	case StatPlayTime:
		ps.PlayTimeMinutes += delta
	case StatAchievements:
		ps.Achievements += delta
	}
}

// GetLeaderboard returns the top N players for a stat
func (m *Manager) GetLeaderboard(stat StatType, limit int) []Entry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Collect all entries
	entries := make([]Entry, 0, len(m.Players))
	for _, ps := range m.Players {
		var value int
		switch stat {
		case StatXP:
			value = ps.XP
		case StatLevel:
			value = ps.Level
		case StatKills:
			value = ps.Kills
		case StatDeaths:
			value = ps.Deaths
		case StatQuestsCompleted:
			value = ps.QuestsCompleted
		case StatMoney:
			value = ps.Money
		case StatPvPWins:
			value = ps.PvPWins
		case StatPvPLosses:
			value = ps.PvPLosses
		case StatPlayTime:
			value = ps.PlayTimeMinutes
		case StatAchievements:
			value = ps.Achievements
		}

		entries = append(entries, Entry{Name: ps.Name, Value: value})
	}

	// Sort descending
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Value > entries[j].Value
	})

	// Limit and add ranks
	if limit > len(entries) {
		limit = len(entries)
	}
	result := entries[:limit]
	for i := range result {
		result[i].Rank = i + 1
	}

	return result
}

// GetRank returns a player's rank for a specific stat
func (m *Manager) GetRank(playerName string, stat StatType) int {
	board := m.GetLeaderboard(stat, len(m.Players))
	for _, entry := range board {
		if entry.Name == playerName {
			return entry.Rank
		}
	}
	return 0
}

// GetAllStats returns all stats for a player
func (m *Manager) GetAllStats(playerName string) *PlayerStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.Players[playerName]
}

// Save persists leaderboard data to disk
func (m *Manager) Save() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := json.MarshalIndent(m.Players, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("data/leaderboard.json", data, 0644)
}

// Load reads leaderboard data from disk
func (m *Manager) Load() error {
	data, err := os.ReadFile("data/leaderboard.json")
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	return json.Unmarshal(data, &m.Players)
}

// GlobalLeaderboard is a global leaderboard manager instance
var GlobalLeaderboard = NewManager()
