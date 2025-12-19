// Package db provides database abstraction for Matrix MUD.
// Supports SQLite for persistent storage with migration support.
package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Config holds database configuration
type Config struct {
	Driver   string // "sqlite3" or "memory"
	DSN      string // Data source name (file path for sqlite)
	MaxConns int    // Maximum connections
}

// DB wraps the database connection with helper methods
type DB struct {
	*sql.DB
	config Config
	mu     sync.RWMutex
}

// Player represents a player record in the database
type Player struct {
	ID           int64
	Name         string
	PasswordHash string
	Class        string
	Level        int
	XP           int
	HP           int
	MaxHP        int
	MP           int
	MaxMP        int
	Strength     int
	AC           int
	Money        int
	RoomID       string
	State        string
	Inventory    []string // JSON encoded
	Equipment    map[string]string
	Skills       map[string]int
	Titles       []string
	CurrentTitle string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLogin    time.Time
}

// PlayerQuest represents quest progress
type PlayerQuest struct {
	ID          int64
	PlayerID    int64
	QuestID     string
	Stage       int
	StartedAt   time.Time
	CompletedAt sql.NullTime
}

// PlayerAchievement represents earned achievements
type PlayerAchievement struct {
	ID            int64
	PlayerID      int64
	AchievementID string
	EarnedAt      time.Time
}

// PlayerFaction represents faction standing
type PlayerFaction struct {
	ID         int64
	PlayerID   int64
	FactionID  string
	Reputation int
	JoinedAt   time.Time
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID        int64
	PlayerID  int64
	Action    string
	Details   string
	IPAddress string
	Timestamp time.Time
}

// WorldState represents persistent world state
type WorldState struct {
	ID        int64
	Key       string
	Value     string
	UpdatedAt time.Time
}

// New creates a new database connection
func New(config Config) (*DB, error) {
	if config.Driver == "" {
		config.Driver = "sqlite3"
	}
	if config.DSN == "" {
		config.DSN = ":memory:"
	}
	if config.MaxConns == 0 {
		config.MaxConns = 10
	}

	db, err := sql.Open(config.Driver, config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(config.MaxConns)
	db.SetMaxIdleConns(config.MaxConns / 2)
	db.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{DB: db, config: config}, nil
}

// NewMemory creates an in-memory database (for testing)
func NewMemory() (*DB, error) {
	return New(Config{
		Driver: "sqlite3",
		DSN:    ":memory:",
	})
}

// RunMigrations runs all pending migrations
func (db *DB) RunMigrations() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Create migrations table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Run migrations in order
	migrations := []struct {
		name string
		sql  string
	}{
		{"001_initial_schema", migration001},
		{"002_world_state", migration002},
		{"003_audit_log", migration003},
	}

	for _, m := range migrations {
		// Check if already applied
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM migrations WHERE name = ?", m.name).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration %s: %w", m.name, err)
		}
		if count > 0 {
			continue
		}

		// Apply migration
		_, err = db.Exec(m.sql)
		if err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", m.name, err)
		}

		// Record migration
		_, err = db.Exec("INSERT INTO migrations (name) VALUES (?)", m.name)
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", m.name, err)
		}
	}

	return nil
}

// Migration 001: Initial schema
const migration001 = `
-- Players table
CREATE TABLE IF NOT EXISTS players (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT UNIQUE NOT NULL COLLATE NOCASE,
	password_hash TEXT NOT NULL,
	class TEXT NOT NULL DEFAULT 'Hacker',
	level INTEGER NOT NULL DEFAULT 1,
	xp INTEGER NOT NULL DEFAULT 0,
	hp INTEGER NOT NULL DEFAULT 100,
	max_hp INTEGER NOT NULL DEFAULT 100,
	mp INTEGER NOT NULL DEFAULT 50,
	max_mp INTEGER NOT NULL DEFAULT 50,
	strength INTEGER NOT NULL DEFAULT 10,
	ac INTEGER NOT NULL DEFAULT 0,
	money INTEGER NOT NULL DEFAULT 100,
	room_id TEXT NOT NULL DEFAULT 'construct_entrance',
	state TEXT NOT NULL DEFAULT 'IDLE',
	inventory TEXT DEFAULT '[]',
	equipment TEXT DEFAULT '{}',
	skills TEXT DEFAULT '{}',
	titles TEXT DEFAULT '[]',
	current_title TEXT DEFAULT '',
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	last_login TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Player inventory (normalized)
CREATE TABLE IF NOT EXISTS player_inventory (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	player_id INTEGER NOT NULL,
	item_id TEXT NOT NULL,
	slot INTEGER DEFAULT 0,
	quantity INTEGER DEFAULT 1,
	FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE
);

-- Player equipment
CREATE TABLE IF NOT EXISTS player_equipment (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	player_id INTEGER NOT NULL,
	slot TEXT NOT NULL,
	item_id TEXT NOT NULL,
	UNIQUE(player_id, slot),
	FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE
);

-- Player bank/storage
CREATE TABLE IF NOT EXISTS player_bank (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	player_id INTEGER NOT NULL,
	item_id TEXT NOT NULL,
	quantity INTEGER DEFAULT 1,
	FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE
);

-- Player quests
CREATE TABLE IF NOT EXISTS player_quests (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	player_id INTEGER NOT NULL,
	quest_id TEXT NOT NULL,
	stage INTEGER DEFAULT 0,
	started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	completed_at TIMESTAMP,
	UNIQUE(player_id, quest_id),
	FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE
);

-- Player achievements
CREATE TABLE IF NOT EXISTS player_achievements (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	player_id INTEGER NOT NULL,
	achievement_id TEXT NOT NULL,
	earned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	UNIQUE(player_id, achievement_id),
	FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE
);

-- Player factions
CREATE TABLE IF NOT EXISTS player_factions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	player_id INTEGER NOT NULL,
	faction_id TEXT NOT NULL,
	reputation INTEGER DEFAULT 0,
	joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	UNIQUE(player_id, faction_id),
	FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_players_name ON players(name);
CREATE INDEX IF NOT EXISTS idx_player_inventory_player ON player_inventory(player_id);
CREATE INDEX IF NOT EXISTS idx_player_quests_player ON player_quests(player_id);
CREATE INDEX IF NOT EXISTS idx_player_achievements_player ON player_achievements(player_id);
`

// Migration 002: World state
const migration002 = `
-- World state key-value store
CREATE TABLE IF NOT EXISTS world_state (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	key TEXT UNIQUE NOT NULL,
	value TEXT NOT NULL,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- NPC state (for persistent NPC modifications)
CREATE TABLE IF NOT EXISTS npc_state (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	npc_id TEXT UNIQUE NOT NULL,
	room_id TEXT NOT NULL,
	hp INTEGER NOT NULL,
	is_dead INTEGER DEFAULT 0,
	respawn_at TIMESTAMP,
	custom_data TEXT DEFAULT '{}',
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Room state (for persistent room modifications)  
CREATE TABLE IF NOT EXISTS room_state (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	room_id TEXT UNIQUE NOT NULL,
	items TEXT DEFAULT '[]',
	custom_data TEXT DEFAULT '{}',
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_world_state_key ON world_state(key);
CREATE INDEX IF NOT EXISTS idx_npc_state_room ON npc_state(room_id);
`

// Migration 003: Audit log
const migration003 = `
-- Audit log for tracking player actions
CREATE TABLE IF NOT EXISTS audit_log (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	player_id INTEGER,
	player_name TEXT,
	action TEXT NOT NULL,
	details TEXT,
	ip_address TEXT,
	timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE SET NULL
);

-- Session tracking
CREATE TABLE IF NOT EXISTS sessions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	player_id INTEGER NOT NULL,
	session_token TEXT UNIQUE NOT NULL,
	ip_address TEXT,
	user_agent TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	expires_at TIMESTAMP NOT NULL,
	FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_audit_log_player ON audit_log(player_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_action ON audit_log(action);
CREATE INDEX IF NOT EXISTS idx_audit_log_timestamp ON audit_log(timestamp);
CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_sessions_player ON sessions(player_id);
`

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// Health checks database health
func (db *DB) Health() error {
	return db.Ping()
}

// Stats returns database statistics
func (db *DB) Stats() sql.DBStats {
	return db.DB.Stats()
}

// Transaction executes a function within a transaction
func (db *DB) Transaction(fn func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// Helper to encode JSON
func encodeJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

// Helper to decode JSON
func decodeJSON(data string, v interface{}) error {
	return json.Unmarshal([]byte(data), v)
}
