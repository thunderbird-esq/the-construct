// Package db provides database abstraction for Matrix MUD.
// This file contains world state database operations.
package db

import (
	"database/sql"
	"encoding/json"
	"time"
)

// WorldRepository handles world state database operations
type WorldRepository struct {
	db *DB
}

// NewWorldRepository creates a new world repository
func NewWorldRepository(db *DB) *WorldRepository {
	return &WorldRepository{db: db}
}

// NPCState represents persistent NPC state
type NPCState struct {
	ID         int64
	NPCID      string
	RoomID     string
	HP         int
	IsDead     bool
	RespawnAt  sql.NullTime
	CustomData map[string]interface{}
	UpdatedAt  time.Time
}

// RoomState represents persistent room modifications
type RoomState struct {
	ID         int64
	RoomID     string
	Items      []string
	CustomData map[string]interface{}
	UpdatedAt  time.Time
}

// SetState sets a world state key-value pair
func (r *WorldRepository) SetState(key, value string) error {
	_, err := r.db.Exec(`
		INSERT INTO world_state (key, value, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at
	`, key, value, time.Now())
	return err
}

// GetState gets a world state value by key
func (r *WorldRepository) GetState(key string) (string, error) {
	var value string
	err := r.db.QueryRow("SELECT value FROM world_state WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// DeleteState deletes a world state key
func (r *WorldRepository) DeleteState(key string) error {
	_, err := r.db.Exec("DELETE FROM world_state WHERE key = ?", key)
	return err
}

// GetAllState returns all world state as a map
func (r *WorldRepository) GetAllState() (map[string]string, error) {
	rows, err := r.db.Query("SELECT key, value FROM world_state")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	state := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		state[key] = value
	}
	return state, nil
}

// SaveNPCState saves or updates NPC state
func (r *WorldRepository) SaveNPCState(state *NPCState) error {
	customData, _ := json.Marshal(state.CustomData)
	
	_, err := r.db.Exec(`
		INSERT INTO npc_state (npc_id, room_id, hp, is_dead, respawn_at, custom_data, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(npc_id) DO UPDATE SET
			room_id = excluded.room_id,
			hp = excluded.hp,
			is_dead = excluded.is_dead,
			respawn_at = excluded.respawn_at,
			custom_data = excluded.custom_data,
			updated_at = excluded.updated_at
	`, state.NPCID, state.RoomID, state.HP, state.IsDead, state.RespawnAt, string(customData), time.Now())
	return err
}

// GetNPCState retrieves NPC state by ID
func (r *WorldRepository) GetNPCState(npcID string) (*NPCState, error) {
	state := &NPCState{}
	var customData string
	var isDead int

	err := r.db.QueryRow(`
		SELECT id, npc_id, room_id, hp, is_dead, respawn_at, custom_data, updated_at
		FROM npc_state WHERE npc_id = ?
	`, npcID).Scan(
		&state.ID, &state.NPCID, &state.RoomID, &state.HP,
		&isDead, &state.RespawnAt, &customData, &state.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	state.IsDead = isDead != 0
	json.Unmarshal([]byte(customData), &state.CustomData)

	return state, nil
}

// GetDeadNPCs returns all dead NPCs
func (r *WorldRepository) GetDeadNPCs() ([]*NPCState, error) {
	rows, err := r.db.Query(`
		SELECT id, npc_id, room_id, hp, is_dead, respawn_at, custom_data, updated_at
		FROM npc_state WHERE is_dead = 1
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var npcs []*NPCState
	for rows.Next() {
		state := &NPCState{}
		var customData string
		var isDead int
		err := rows.Scan(
			&state.ID, &state.NPCID, &state.RoomID, &state.HP,
			&isDead, &state.RespawnAt, &customData, &state.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		state.IsDead = isDead != 0
		json.Unmarshal([]byte(customData), &state.CustomData)
		npcs = append(npcs, state)
	}
	return npcs, nil
}

// GetNPCsToRespawn returns NPCs ready to respawn
func (r *WorldRepository) GetNPCsToRespawn() ([]*NPCState, error) {
	rows, err := r.db.Query(`
		SELECT id, npc_id, room_id, hp, is_dead, respawn_at, custom_data, updated_at
		FROM npc_state WHERE is_dead = 1 AND respawn_at <= ?
	`, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var npcs []*NPCState
	for rows.Next() {
		state := &NPCState{}
		var customData string
		var isDead int
		err := rows.Scan(
			&state.ID, &state.NPCID, &state.RoomID, &state.HP,
			&isDead, &state.RespawnAt, &customData, &state.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		state.IsDead = isDead != 0
		json.Unmarshal([]byte(customData), &state.CustomData)
		npcs = append(npcs, state)
	}
	return npcs, nil
}

// DeleteNPCState deletes NPC state
func (r *WorldRepository) DeleteNPCState(npcID string) error {
	_, err := r.db.Exec("DELETE FROM npc_state WHERE npc_id = ?", npcID)
	return err
}

// SaveRoomState saves or updates room state
func (r *WorldRepository) SaveRoomState(state *RoomState) error {
	items, _ := json.Marshal(state.Items)
	customData, _ := json.Marshal(state.CustomData)
	
	_, err := r.db.Exec(`
		INSERT INTO room_state (room_id, items, custom_data, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(room_id) DO UPDATE SET
			items = excluded.items,
			custom_data = excluded.custom_data,
			updated_at = excluded.updated_at
	`, state.RoomID, string(items), string(customData), time.Now())
	return err
}

// GetRoomState retrieves room state by ID
func (r *WorldRepository) GetRoomState(roomID string) (*RoomState, error) {
	state := &RoomState{}
	var items, customData string

	err := r.db.QueryRow(`
		SELECT id, room_id, items, custom_data, updated_at
		FROM room_state WHERE room_id = ?
	`, roomID).Scan(&state.ID, &state.RoomID, &items, &customData, &state.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(items), &state.Items)
	json.Unmarshal([]byte(customData), &state.CustomData)

	return state, nil
}

// GetAllRoomStates returns all room states
func (r *WorldRepository) GetAllRoomStates() ([]*RoomState, error) {
	rows, err := r.db.Query("SELECT id, room_id, items, custom_data, updated_at FROM room_state")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var states []*RoomState
	for rows.Next() {
		state := &RoomState{}
		var items, customData string
		err := rows.Scan(&state.ID, &state.RoomID, &items, &customData, &state.UpdatedAt)
		if err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(items), &state.Items)
		json.Unmarshal([]byte(customData), &state.CustomData)
		states = append(states, state)
	}
	return states, nil
}

// DeleteRoomState deletes room state
func (r *WorldRepository) DeleteRoomState(roomID string) error {
	_, err := r.db.Exec("DELETE FROM room_state WHERE room_id = ?", roomID)
	return err
}

// ClearAllState clears all world state (for testing/reset)
func (r *WorldRepository) ClearAllState() error {
	_, err := r.db.Exec("DELETE FROM world_state")
	if err != nil {
		return err
	}
	_, err = r.db.Exec("DELETE FROM npc_state")
	if err != nil {
		return err
	}
	_, err = r.db.Exec("DELETE FROM room_state")
	return err
}
