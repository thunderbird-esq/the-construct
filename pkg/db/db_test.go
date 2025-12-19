package db

import (
	"database/sql"
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	db, err := NewMemory()
	if err != nil {
		t.Fatalf("NewMemory failed: %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Fatal("Database is nil")
	}

	if err := db.Health(); err != nil {
		t.Errorf("Health check failed: %v", err)
	}
}

func TestRunMigrations(t *testing.T) {
	db, err := NewMemory()
	if err != nil {
		t.Fatalf("NewMemory failed: %v", err)
	}
	defer db.Close()

	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	// Running again should be idempotent
	if err := db.RunMigrations(); err != nil {
		t.Fatalf("Second RunMigrations failed: %v", err)
	}

	// Check tables exist
	tables := []string{"players", "player_inventory", "player_quests", 
		"player_achievements", "world_state", "audit_log", "sessions"}
	
	for _, table := range tables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("Table %s not found: %v", table, err)
		}
	}
}

func TestTransaction(t *testing.T) {
	db, err := NewMemory()
	if err != nil {
		t.Fatalf("NewMemory failed: %v", err)
	}
	defer db.Close()
	db.RunMigrations()

	// Successful transaction
	err = db.Transaction(func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO world_state (key, value) VALUES ('test', 'value')")
		return err
	})
	if err != nil {
		t.Errorf("Transaction failed: %v", err)
	}

	// Verify data was inserted
	var value string
	db.QueryRow("SELECT value FROM world_state WHERE key='test'").Scan(&value)
	if value != "value" {
		t.Errorf("Transaction did not commit")
	}

	// Failed transaction should rollback
	err = db.Transaction(func(tx *sql.Tx) error {
		tx.Exec("INSERT INTO world_state (key, value) VALUES ('rollback', 'test')")
		return fmt.Errorf("intentional error")
	})
	if err == nil {
		t.Error("Transaction should have failed")
	}

	// Verify rollback
	var count int
	db.QueryRow("SELECT COUNT(*) FROM world_state WHERE key='rollback'").Scan(&count)
	if count != 0 {
		t.Error("Transaction did not rollback")
	}
}

func TestStats(t *testing.T) {
	db, err := NewMemory()
	if err != nil {
		t.Fatalf("NewMemory failed: %v", err)
	}
	defer db.Close()

	stats := db.Stats()
	if stats.MaxOpenConnections == 0 {
		t.Error("MaxOpenConnections should be set")
	}
}

func TestEncodeDecodeJSON(t *testing.T) {
	input := map[string]int{"a": 1, "b": 2}
	encoded := encodeJSON(input)
	
	var decoded map[string]int
	err := decodeJSON(encoded, &decoded)
	if err != nil {
		t.Errorf("decodeJSON failed: %v", err)
	}
	
	if decoded["a"] != 1 || decoded["b"] != 2 {
		t.Error("JSON round-trip failed")
	}
}
