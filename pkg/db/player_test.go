package db

import (
	"testing"
	"time"
)

func setupTestDB(t *testing.T) (*DB, *PlayerRepository) {
	db, err := NewMemory()
	if err != nil {
		t.Fatalf("NewMemory failed: %v", err)
	}
	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}
	return db, NewPlayerRepository(db)
}

func TestPlayerCreate(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	player := &Player{
		Name:         "TestPlayer",
		PasswordHash: "hash123",
		Class:        "Hacker",
		Level:        1,
		XP:           0,
		HP:           100,
		MaxHP:        100,
		MP:           50,
		MaxMP:        50,
		Strength:     10,
		AC:           0,
		Money:        100,
		RoomID:       "construct_entrance",
		State:        "IDLE",
		Inventory:    []string{"item1", "item2"},
		Equipment:    map[string]string{"weapon": "sword"},
		Skills:       map[string]int{"hacking": 1},
		Titles:       []string{"Newbie"},
		CurrentTitle: "Newbie",
	}

	err := repo.Create(player)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if player.ID == 0 {
		t.Error("Player ID should be set after create")
	}
}

func TestPlayerCreateDuplicate(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	player := &Player{
		Name:         "DuplicateTest",
		PasswordHash: "hash",
		Class:        "Hacker",
	}

	repo.Create(player)
	err := repo.Create(player)
	if err == nil {
		t.Error("Should fail on duplicate name")
	}
}

func TestPlayerGetByName(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	// Create player
	original := &Player{
		Name:         "GetByNameTest",
		PasswordHash: "hash",
		Class:        "Runner",
		Level:        5,
		Money:        500,
		Inventory:    []string{"item1"},
	}
	repo.Create(original)

	// Retrieve by name
	player, err := repo.GetByName("GetByNameTest")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if player == nil {
		t.Fatal("Player not found")
	}

	if player.Name != "GetByNameTest" {
		t.Errorf("Name = %s, want GetByNameTest", player.Name)
	}
	if player.Class != "Runner" {
		t.Errorf("Class = %s, want Runner", player.Class)
	}
	if player.Level != 5 {
		t.Errorf("Level = %d, want 5", player.Level)
	}
}

func TestPlayerGetByNameCaseInsensitive(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	repo.Create(&Player{Name: "CaseTest", PasswordHash: "hash", Class: "Hacker"})

	// Should find with different cases
	player, _ := repo.GetByName("casetest")
	if player == nil {
		t.Error("Should find player with lowercase")
	}

	player, _ = repo.GetByName("CASETEST")
	if player == nil {
		t.Error("Should find player with uppercase")
	}
}

func TestPlayerGetByNameNotFound(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	player, err := repo.GetByName("NonExistent")
	if err != nil {
		t.Errorf("Should not error on not found: %v", err)
	}
	if player != nil {
		t.Error("Should return nil for non-existent player")
	}
}

func TestPlayerGetByID(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	original := &Player{Name: "IDTest", PasswordHash: "hash", Class: "Operator"}
	repo.Create(original)

	player, err := repo.GetByID(original.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if player == nil {
		t.Fatal("Player not found by ID")
	}
	if player.Name != "IDTest" {
		t.Errorf("Name = %s, want IDTest", player.Name)
	}
}

func TestPlayerUpdate(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	player := &Player{
		Name:         "UpdateTest",
		PasswordHash: "hash",
		Class:        "Hacker",
		Level:        1,
		Money:        100,
	}
	repo.Create(player)

	// Update player
	player.Level = 10
	player.Money = 1000
	player.Class = "Runner"

	err := repo.Update(player)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	updated, _ := repo.GetByID(player.ID)
	if updated.Level != 10 {
		t.Errorf("Level = %d, want 10", updated.Level)
	}
	if updated.Money != 1000 {
		t.Errorf("Money = %d, want 1000", updated.Money)
	}
	if updated.Class != "Runner" {
		t.Errorf("Class = %s, want Runner", updated.Class)
	}
}

func TestPlayerDelete(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	player := &Player{Name: "DeleteTest", PasswordHash: "hash", Class: "Hacker"}
	repo.Create(player)

	err := repo.Delete(player.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	deleted, _ := repo.GetByID(player.ID)
	if deleted != nil {
		t.Error("Player should be deleted")
	}
}

func TestPlayerExists(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	exists, _ := repo.Exists("ExistsTest")
	if exists {
		t.Error("Should not exist initially")
	}

	repo.Create(&Player{Name: "ExistsTest", PasswordHash: "hash", Class: "Hacker"})

	exists, _ = repo.Exists("ExistsTest")
	if !exists {
		t.Error("Should exist after create")
	}
}

func TestPlayerList(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	// Create multiple players
	for i := 0; i < 5; i++ {
		repo.Create(&Player{
			Name:         "ListTest" + string(rune('A'+i)),
			PasswordHash: "hash",
			Class:        "Hacker",
			Level:        i + 1,
		})
	}

	players, err := repo.List(3, 0)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(players) != 3 {
		t.Errorf("Expected 3 players, got %d", len(players))
	}

	// Test offset
	players, _ = repo.List(10, 2)
	if len(players) != 3 {
		t.Errorf("Expected 3 players with offset, got %d", len(players))
	}
}

func TestPlayerCount(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	count, _ := repo.Count()
	if count != 0 {
		t.Errorf("Initial count should be 0, got %d", count)
	}

	repo.Create(&Player{Name: "CountTest1", PasswordHash: "hash", Class: "Hacker"})
	repo.Create(&Player{Name: "CountTest2", PasswordHash: "hash", Class: "Hacker"})

	count, _ = repo.Count()
	if count != 2 {
		t.Errorf("Count should be 2, got %d", count)
	}
}

func TestPlayerQuests(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	player := &Player{Name: "QuestTest", PasswordHash: "hash", Class: "Hacker"}
	repo.Create(player)

	// Save quest
	err := repo.SaveQuest(player.ID, "white_rabbit", 2)
	if err != nil {
		t.Fatalf("SaveQuest failed: %v", err)
	}

	// Get quests
	quests, err := repo.GetQuests(player.ID)
	if err != nil {
		t.Fatalf("GetQuests failed: %v", err)
	}
	if len(quests) != 1 {
		t.Errorf("Expected 1 quest, got %d", len(quests))
	}
	if quests[0].QuestID != "white_rabbit" {
		t.Errorf("QuestID = %s, want white_rabbit", quests[0].QuestID)
	}
	if quests[0].Stage != 2 {
		t.Errorf("Stage = %d, want 2", quests[0].Stage)
	}

	// Update quest stage
	repo.SaveQuest(player.ID, "white_rabbit", 3)
	quests, _ = repo.GetQuests(player.ID)
	if quests[0].Stage != 3 {
		t.Errorf("Updated stage = %d, want 3", quests[0].Stage)
	}

	// Complete quest
	repo.CompleteQuest(player.ID, "white_rabbit")
	quests, _ = repo.GetQuests(player.ID)
	if !quests[0].CompletedAt.Valid {
		t.Error("Quest should be marked completed")
	}
}

func TestPlayerAchievements(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	player := &Player{Name: "AchievementTest", PasswordHash: "hash", Class: "Hacker"}
	repo.Create(player)

	// Add achievements
	repo.AddAchievement(player.ID, "first_steps")
	repo.AddAchievement(player.ID, "level_5")
	repo.AddAchievement(player.ID, "first_steps") // Duplicate should be ignored

	achievements, err := repo.GetAchievements(player.ID)
	if err != nil {
		t.Fatalf("GetAchievements failed: %v", err)
	}
	if len(achievements) != 2 {
		t.Errorf("Expected 2 achievements, got %d", len(achievements))
	}
}

func TestPlayerFactions(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	player := &Player{Name: "FactionTest", PasswordHash: "hash", Class: "Hacker"}
	repo.Create(player)

	// Set faction reputation
	repo.SetFactionReputation(player.ID, "zion", 100)
	repo.SetFactionReputation(player.ID, "exile", -50)

	// Get single faction
	rep, err := repo.GetFactionReputation(player.ID, "zion")
	if err != nil {
		t.Fatalf("GetFactionReputation failed: %v", err)
	}
	if rep != 100 {
		t.Errorf("Zion rep = %d, want 100", rep)
	}

	// Get non-existent faction
	rep, _ = repo.GetFactionReputation(player.ID, "nonexistent")
	if rep != 0 {
		t.Errorf("Non-existent faction rep = %d, want 0", rep)
	}

	// Get all factions
	factions, err := repo.GetAllFactions(player.ID)
	if err != nil {
		t.Fatalf("GetAllFactions failed: %v", err)
	}
	if len(factions) != 2 {
		t.Errorf("Expected 2 factions, got %d", len(factions))
	}
	if factions["zion"] != 100 {
		t.Errorf("Zion rep = %d, want 100", factions["zion"])
	}
	if factions["exile"] != -50 {
		t.Errorf("Exile rep = %d, want -50", factions["exile"])
	}
}

func TestPlayerUpdateLastLogin(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	player := &Player{Name: "LoginTest", PasswordHash: "hash", Class: "Hacker"}
	repo.Create(player)

	time.Sleep(10 * time.Millisecond)
	repo.UpdateLastLogin(player.ID)

	updated, _ := repo.GetByID(player.ID)
	if updated.LastLogin.Before(player.CreatedAt) {
		t.Error("LastLogin should be updated")
	}
}

func TestPlayerGetTopByMoney(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	repo.Create(&Player{Name: "Rich", PasswordHash: "hash", Class: "Hacker", Money: 10000})
	repo.Create(&Player{Name: "Poor", PasswordHash: "hash", Class: "Hacker", Money: 10})
	repo.Create(&Player{Name: "Medium", PasswordHash: "hash", Class: "Hacker", Money: 1000})

	players, err := repo.GetTopByMoney(2)
	if err != nil {
		t.Fatalf("GetTopByMoney failed: %v", err)
	}
	if len(players) != 2 {
		t.Errorf("Expected 2 players, got %d", len(players))
	}
	if players[0].Name != "Rich" {
		t.Error("First player should be Rich")
	}
}
