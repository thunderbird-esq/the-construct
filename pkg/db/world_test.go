package db

import (
	"database/sql"
	"testing"
	"time"
)

func setupWorldTestDB(t *testing.T) (*DB, *WorldRepository) {
	db, err := NewMemory()
	if err != nil {
		t.Fatalf("NewMemory failed: %v", err)
	}
	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}
	return db, NewWorldRepository(db)
}

func TestWorldStateSetGet(t *testing.T) {
	db, repo := setupWorldTestDB(t)
	defer db.Close()

	err := repo.SetState("test_key", "test_value")
	if err != nil {
		t.Fatalf("SetState failed: %v", err)
	}

	value, err := repo.GetState("test_key")
	if err != nil {
		t.Fatalf("GetState failed: %v", err)
	}
	if value != "test_value" {
		t.Errorf("Value = %s, want test_value", value)
	}
}

func TestWorldStateUpdate(t *testing.T) {
	db, repo := setupWorldTestDB(t)
	defer db.Close()

	repo.SetState("update_key", "initial")
	repo.SetState("update_key", "updated")

	value, _ := repo.GetState("update_key")
	if value != "updated" {
		t.Errorf("Value = %s, want updated", value)
	}
}

func TestWorldStateDelete(t *testing.T) {
	db, repo := setupWorldTestDB(t)
	defer db.Close()

	repo.SetState("delete_key", "value")
	repo.DeleteState("delete_key")

	value, _ := repo.GetState("delete_key")
	if value != "" {
		t.Error("Value should be empty after delete")
	}
}

func TestWorldStateGetAll(t *testing.T) {
	db, repo := setupWorldTestDB(t)
	defer db.Close()

	repo.SetState("key1", "value1")
	repo.SetState("key2", "value2")
	repo.SetState("key3", "value3")

	state, err := repo.GetAllState()
	if err != nil {
		t.Fatalf("GetAllState failed: %v", err)
	}
	if len(state) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(state))
	}
	if state["key1"] != "value1" {
		t.Error("key1 value incorrect")
	}
}

func TestWorldStateGetNonExistent(t *testing.T) {
	db, repo := setupWorldTestDB(t)
	defer db.Close()

	value, err := repo.GetState("nonexistent")
	if err != nil {
		t.Errorf("Should not error on non-existent key: %v", err)
	}
	if value != "" {
		t.Errorf("Value should be empty for non-existent key, got %s", value)
	}
}

func TestNPCStateSave(t *testing.T) {
	db, repo := setupWorldTestDB(t)
	defer db.Close()

	state := &NPCState{
		NPCID:      "npc_morpheus",
		RoomID:     "dojo",
		HP:         100,
		IsDead:     false,
		CustomData: map[string]interface{}{"dialogue_state": "intro"},
	}

	err := repo.SaveNPCState(state)
	if err != nil {
		t.Fatalf("SaveNPCState failed: %v", err)
	}
}

func TestNPCStateGet(t *testing.T) {
	db, repo := setupWorldTestDB(t)
	defer db.Close()

	original := &NPCState{
		NPCID:      "npc_oracle",
		RoomID:     "oracle_apartment",
		HP:         50,
		IsDead:     false,
		CustomData: map[string]interface{}{"cookies_given": true},
	}
	repo.SaveNPCState(original)

	state, err := repo.GetNPCState("npc_oracle")
	if err != nil {
		t.Fatalf("GetNPCState failed: %v", err)
	}
	if state == nil {
		t.Fatal("NPC state not found")
	}
	if state.NPCID != "npc_oracle" {
		t.Errorf("NPCID = %s, want npc_oracle", state.NPCID)
	}
	if state.HP != 50 {
		t.Errorf("HP = %d, want 50", state.HP)
	}
}

func TestNPCStateGetNonExistent(t *testing.T) {
	db, repo := setupWorldTestDB(t)
	defer db.Close()

	state, err := repo.GetNPCState("nonexistent")
	if err != nil {
		t.Errorf("Should not error: %v", err)
	}
	if state != nil {
		t.Error("Should return nil for non-existent NPC")
	}
}

func TestNPCStateUpdate(t *testing.T) {
	db, repo := setupWorldTestDB(t)
	defer db.Close()

	state := &NPCState{
		NPCID:  "npc_update",
		RoomID: "room1",
		HP:     100,
		IsDead: false,
	}
	repo.SaveNPCState(state)

	// Update
	state.HP = 0
	state.IsDead = true
	state.RespawnAt = sql.NullTime{Time: time.Now().Add(time.Minute), Valid: true}
	repo.SaveNPCState(state)

	updated, _ := repo.GetNPCState("npc_update")
	if updated.HP != 0 {
		t.Errorf("HP = %d, want 0", updated.HP)
	}
	if !updated.IsDead {
		t.Error("IsDead should be true")
	}
}

func TestNPCStateDelete(t *testing.T) {
	db, repo := setupWorldTestDB(t)
	defer db.Close()

	repo.SaveNPCState(&NPCState{NPCID: "delete_npc", RoomID: "room"})
	repo.DeleteNPCState("delete_npc")

	state, _ := repo.GetNPCState("delete_npc")
	if state != nil {
		t.Error("NPC state should be deleted")
	}
}

func TestGetDeadNPCs(t *testing.T) {
	db, repo := setupWorldTestDB(t)
	defer db.Close()

	repo.SaveNPCState(&NPCState{NPCID: "alive", RoomID: "room", HP: 100, IsDead: false})
	repo.SaveNPCState(&NPCState{NPCID: "dead1", RoomID: "room", HP: 0, IsDead: true})
	repo.SaveNPCState(&NPCState{NPCID: "dead2", RoomID: "room", HP: 0, IsDead: true})

	dead, err := repo.GetDeadNPCs()
	if err != nil {
		t.Fatalf("GetDeadNPCs failed: %v", err)
	}
	if len(dead) != 2 {
		t.Errorf("Expected 2 dead NPCs, got %d", len(dead))
	}
}

func TestGetNPCsToRespawn(t *testing.T) {
	db, repo := setupWorldTestDB(t)
	defer db.Close()

	// Dead NPC ready to respawn
	repo.SaveNPCState(&NPCState{
		NPCID:     "respawn_ready",
		RoomID:    "room",
		HP:        0,
		IsDead:    true,
		RespawnAt: sql.NullTime{Time: time.Now().Add(-time.Second), Valid: true},
	})

	// Dead NPC not ready yet
	repo.SaveNPCState(&NPCState{
		NPCID:     "respawn_later",
		RoomID:    "room",
		HP:        0,
		IsDead:    true,
		RespawnAt: sql.NullTime{Time: time.Now().Add(time.Hour), Valid: true},
	})

	ready, err := repo.GetNPCsToRespawn()
	if err != nil {
		t.Fatalf("GetNPCsToRespawn failed: %v", err)
	}
	if len(ready) != 1 {
		t.Errorf("Expected 1 NPC ready to respawn, got %d", len(ready))
	}
	if ready[0].NPCID != "respawn_ready" {
		t.Error("Wrong NPC returned")
	}
}

func TestRoomStateSave(t *testing.T) {
	db, repo := setupWorldTestDB(t)
	defer db.Close()

	state := &RoomState{
		RoomID:     "test_room",
		Items:      []string{"item1", "item2"},
		CustomData: map[string]interface{}{"locked": false},
	}

	err := repo.SaveRoomState(state)
	if err != nil {
		t.Fatalf("SaveRoomState failed: %v", err)
	}
}

func TestRoomStateGet(t *testing.T) {
	db, repo := setupWorldTestDB(t)
	defer db.Close()

	original := &RoomState{
		RoomID: "get_room",
		Items:  []string{"sword", "shield"},
	}
	repo.SaveRoomState(original)

	state, err := repo.GetRoomState("get_room")
	if err != nil {
		t.Fatalf("GetRoomState failed: %v", err)
	}
	if state == nil {
		t.Fatal("Room state not found")
	}
	if len(state.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(state.Items))
	}
}

func TestRoomStateGetNonExistent(t *testing.T) {
	db, repo := setupWorldTestDB(t)
	defer db.Close()

	state, err := repo.GetRoomState("nonexistent")
	if err != nil {
		t.Errorf("Should not error: %v", err)
	}
	if state != nil {
		t.Error("Should return nil for non-existent room")
	}
}

func TestRoomStateUpdate(t *testing.T) {
	db, repo := setupWorldTestDB(t)
	defer db.Close()

	state := &RoomState{
		RoomID: "update_room",
		Items:  []string{"item1"},
	}
	repo.SaveRoomState(state)

	// Update
	state.Items = []string{"item1", "item2", "item3"}
	repo.SaveRoomState(state)

	updated, _ := repo.GetRoomState("update_room")
	if len(updated.Items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(updated.Items))
	}
}

func TestRoomStateDelete(t *testing.T) {
	db, repo := setupWorldTestDB(t)
	defer db.Close()

	repo.SaveRoomState(&RoomState{RoomID: "delete_room"})
	repo.DeleteRoomState("delete_room")

	state, _ := repo.GetRoomState("delete_room")
	if state != nil {
		t.Error("Room state should be deleted")
	}
}

func TestGetAllRoomStates(t *testing.T) {
	db, repo := setupWorldTestDB(t)
	defer db.Close()

	repo.SaveRoomState(&RoomState{RoomID: "room1", Items: []string{"a"}})
	repo.SaveRoomState(&RoomState{RoomID: "room2", Items: []string{"b"}})
	repo.SaveRoomState(&RoomState{RoomID: "room3", Items: []string{"c"}})

	states, err := repo.GetAllRoomStates()
	if err != nil {
		t.Fatalf("GetAllRoomStates failed: %v", err)
	}
	if len(states) != 3 {
		t.Errorf("Expected 3 room states, got %d", len(states))
	}
}

func TestClearAllState(t *testing.T) {
	db, repo := setupWorldTestDB(t)
	defer db.Close()

	repo.SetState("key", "value")
	repo.SaveNPCState(&NPCState{NPCID: "npc", RoomID: "room"})
	repo.SaveRoomState(&RoomState{RoomID: "room"})

	err := repo.ClearAllState()
	if err != nil {
		t.Fatalf("ClearAllState failed: %v", err)
	}

	state, _ := repo.GetAllState()
	if len(state) != 0 {
		t.Error("World state should be cleared")
	}

	npc, _ := repo.GetNPCState("npc")
	if npc != nil {
		t.Error("NPC state should be cleared")
	}

	room, _ := repo.GetRoomState("room")
	if room != nil {
		t.Error("Room state should be cleared")
	}
}
