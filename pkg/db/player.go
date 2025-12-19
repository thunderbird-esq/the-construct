// Package db provides database abstraction for Matrix MUD.
// This file contains player-related database operations.
package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// PlayerRepository handles player database operations
type PlayerRepository struct {
	db *DB
}

// NewPlayerRepository creates a new player repository
func NewPlayerRepository(db *DB) *PlayerRepository {
	return &PlayerRepository{db: db}
}

// Create creates a new player
func (r *PlayerRepository) Create(p *Player) error {
	inventory, _ := json.Marshal(p.Inventory)
	equipment, _ := json.Marshal(p.Equipment)
	skills, _ := json.Marshal(p.Skills)
	titles, _ := json.Marshal(p.Titles)

	result, err := r.db.Exec(`
		INSERT INTO players (
			name, password_hash, class, level, xp, hp, max_hp, mp, max_mp,
			strength, ac, money, room_id, state, inventory, equipment, 
			skills, titles, current_title, created_at, updated_at, last_login
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		p.Name, p.PasswordHash, p.Class, p.Level, p.XP, p.HP, p.MaxHP,
		p.MP, p.MaxMP, p.Strength, p.AC, p.Money, p.RoomID, p.State,
		string(inventory), string(equipment), string(skills), string(titles),
		p.CurrentTitle, time.Now(), time.Now(), time.Now(),
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			return fmt.Errorf("player '%s' already exists", p.Name)
		}
		return fmt.Errorf("failed to create player: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get player ID: %w", err)
	}
	p.ID = id
	return nil
}

// GetByName retrieves a player by name
func (r *PlayerRepository) GetByName(name string) (*Player, error) {
	p := &Player{}
	var inventory, equipment, skills, titles string
	var completedAt sql.NullTime

	err := r.db.QueryRow(`
		SELECT id, name, password_hash, class, level, xp, hp, max_hp, mp, max_mp,
			   strength, ac, money, room_id, state, inventory, equipment, 
			   skills, titles, current_title, created_at, updated_at, last_login
		FROM players WHERE name = ? COLLATE NOCASE
	`, name).Scan(
		&p.ID, &p.Name, &p.PasswordHash, &p.Class, &p.Level, &p.XP,
		&p.HP, &p.MaxHP, &p.MP, &p.MaxMP, &p.Strength, &p.AC, &p.Money,
		&p.RoomID, &p.State, &inventory, &equipment, &skills, &titles,
		&p.CurrentTitle, &p.CreatedAt, &p.UpdatedAt, &p.LastLogin,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	_ = completedAt // Silence unused warning
	json.Unmarshal([]byte(inventory), &p.Inventory)
	json.Unmarshal([]byte(equipment), &p.Equipment)
	json.Unmarshal([]byte(skills), &p.Skills)
	json.Unmarshal([]byte(titles), &p.Titles)

	return p, nil
}

// GetByID retrieves a player by ID
func (r *PlayerRepository) GetByID(id int64) (*Player, error) {
	p := &Player{}
	var inventory, equipment, skills, titles string

	err := r.db.QueryRow(`
		SELECT id, name, password_hash, class, level, xp, hp, max_hp, mp, max_mp,
			   strength, ac, money, room_id, state, inventory, equipment, 
			   skills, titles, current_title, created_at, updated_at, last_login
		FROM players WHERE id = ?
	`, id).Scan(
		&p.ID, &p.Name, &p.PasswordHash, &p.Class, &p.Level, &p.XP,
		&p.HP, &p.MaxHP, &p.MP, &p.MaxMP, &p.Strength, &p.AC, &p.Money,
		&p.RoomID, &p.State, &inventory, &equipment, &skills, &titles,
		&p.CurrentTitle, &p.CreatedAt, &p.UpdatedAt, &p.LastLogin,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	json.Unmarshal([]byte(inventory), &p.Inventory)
	json.Unmarshal([]byte(equipment), &p.Equipment)
	json.Unmarshal([]byte(skills), &p.Skills)
	json.Unmarshal([]byte(titles), &p.Titles)

	return p, nil
}

// Update updates a player
func (r *PlayerRepository) Update(p *Player) error {
	inventory, _ := json.Marshal(p.Inventory)
	equipment, _ := json.Marshal(p.Equipment)
	skills, _ := json.Marshal(p.Skills)
	titles, _ := json.Marshal(p.Titles)

	_, err := r.db.Exec(`
		UPDATE players SET
			password_hash = ?, class = ?, level = ?, xp = ?, hp = ?, max_hp = ?,
			mp = ?, max_mp = ?, strength = ?, ac = ?, money = ?, room_id = ?,
			state = ?, inventory = ?, equipment = ?, skills = ?, titles = ?,
			current_title = ?, updated_at = ?
		WHERE id = ?
	`,
		p.PasswordHash, p.Class, p.Level, p.XP, p.HP, p.MaxHP,
		p.MP, p.MaxMP, p.Strength, p.AC, p.Money, p.RoomID,
		p.State, string(inventory), string(equipment), string(skills),
		string(titles), p.CurrentTitle, time.Now(), p.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update player: %w", err)
	}
	return nil
}

// Delete deletes a player
func (r *PlayerRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM players WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete player: %w", err)
	}
	return nil
}

// UpdateLastLogin updates the last login timestamp
func (r *PlayerRepository) UpdateLastLogin(id int64) error {
	_, err := r.db.Exec("UPDATE players SET last_login = ? WHERE id = ?", time.Now(), id)
	return err
}

// Exists checks if a player exists by name
func (r *PlayerRepository) Exists(name string) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM players WHERE name = ? COLLATE NOCASE", name).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// List returns all players with optional filters
func (r *PlayerRepository) List(limit, offset int) ([]*Player, error) {
	rows, err := r.db.Query(`
		SELECT id, name, class, level, xp, hp, max_hp, money, room_id, 
			   current_title, last_login
		FROM players
		ORDER BY level DESC, xp DESC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list players: %w", err)
	}
	defer rows.Close()

	var players []*Player
	for rows.Next() {
		p := &Player{}
		err := rows.Scan(
			&p.ID, &p.Name, &p.Class, &p.Level, &p.XP, &p.HP, &p.MaxHP,
			&p.Money, &p.RoomID, &p.CurrentTitle, &p.LastLogin,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan player: %w", err)
		}
		players = append(players, p)
	}
	return players, nil
}

// Count returns the total number of players
func (r *PlayerRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM players").Scan(&count)
	return count, err
}

// GetOnlineCount returns count of players active in last N minutes
func (r *PlayerRepository) GetOnlineCount(minutes int) (int, error) {
	var count int
	cutoff := time.Now().Add(-time.Duration(minutes) * time.Minute)
	err := r.db.QueryRow("SELECT COUNT(*) FROM players WHERE last_login > ?", cutoff).Scan(&count)
	return count, err
}

// GetTopByLevel returns top players by level
func (r *PlayerRepository) GetTopByLevel(limit int) ([]*Player, error) {
	return r.List(limit, 0)
}

// GetTopByMoney returns top players by money
func (r *PlayerRepository) GetTopByMoney(limit int) ([]*Player, error) {
	rows, err := r.db.Query(`
		SELECT id, name, class, level, money, current_title
		FROM players ORDER BY money DESC LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []*Player
	for rows.Next() {
		p := &Player{}
		err := rows.Scan(&p.ID, &p.Name, &p.Class, &p.Level, &p.Money, &p.CurrentTitle)
		if err != nil {
			return nil, err
		}
		players = append(players, p)
	}
	return players, nil
}

// SaveQuest saves or updates quest progress
func (r *PlayerRepository) SaveQuest(playerID int64, questID string, stage int) error {
	_, err := r.db.Exec(`
		INSERT INTO player_quests (player_id, quest_id, stage, started_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(player_id, quest_id) DO UPDATE SET stage = excluded.stage
	`, playerID, questID, stage, time.Now())
	return err
}

// CompleteQuest marks a quest as completed
func (r *PlayerRepository) CompleteQuest(playerID int64, questID string) error {
	_, err := r.db.Exec(`
		UPDATE player_quests SET completed_at = ? 
		WHERE player_id = ? AND quest_id = ?
	`, time.Now(), playerID, questID)
	return err
}

// GetQuests returns all quests for a player
func (r *PlayerRepository) GetQuests(playerID int64) ([]*PlayerQuest, error) {
	rows, err := r.db.Query(`
		SELECT id, player_id, quest_id, stage, started_at, completed_at
		FROM player_quests WHERE player_id = ?
	`, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quests []*PlayerQuest
	for rows.Next() {
		q := &PlayerQuest{}
		err := rows.Scan(&q.ID, &q.PlayerID, &q.QuestID, &q.Stage, &q.StartedAt, &q.CompletedAt)
		if err != nil {
			return nil, err
		}
		quests = append(quests, q)
	}
	return quests, nil
}

// AddAchievement records an earned achievement
func (r *PlayerRepository) AddAchievement(playerID int64, achievementID string) error {
	_, err := r.db.Exec(`
		INSERT OR IGNORE INTO player_achievements (player_id, achievement_id, earned_at)
		VALUES (?, ?, ?)
	`, playerID, achievementID, time.Now())
	return err
}

// GetAchievements returns all achievements for a player
func (r *PlayerRepository) GetAchievements(playerID int64) ([]string, error) {
	rows, err := r.db.Query(`
		SELECT achievement_id FROM player_achievements WHERE player_id = ?
	`, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var achievements []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		achievements = append(achievements, id)
	}
	return achievements, nil
}

// SetFactionReputation sets faction reputation
func (r *PlayerRepository) SetFactionReputation(playerID int64, factionID string, rep int) error {
	_, err := r.db.Exec(`
		INSERT INTO player_factions (player_id, faction_id, reputation, joined_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(player_id, faction_id) DO UPDATE SET reputation = excluded.reputation
	`, playerID, factionID, rep, time.Now())
	return err
}

// GetFactionReputation gets faction reputation
func (r *PlayerRepository) GetFactionReputation(playerID int64, factionID string) (int, error) {
	var rep int
	err := r.db.QueryRow(`
		SELECT reputation FROM player_factions 
		WHERE player_id = ? AND faction_id = ?
	`, playerID, factionID).Scan(&rep)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return rep, err
}

// GetAllFactions returns all faction standings for a player
func (r *PlayerRepository) GetAllFactions(playerID int64) (map[string]int, error) {
	rows, err := r.db.Query(`
		SELECT faction_id, reputation FROM player_factions WHERE player_id = ?
	`, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	factions := make(map[string]int)
	for rows.Next() {
		var factionID string
		var rep int
		if err := rows.Scan(&factionID, &rep); err != nil {
			return nil, err
		}
		factions[factionID] = rep
	}
	return factions, nil
}
