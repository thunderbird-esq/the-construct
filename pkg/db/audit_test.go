package db

import (
	"testing"
	"time"
)

func setupAuditTestDB(t *testing.T) (*DB, *AuditRepository) {
	db, err := NewMemory()
	if err != nil {
		t.Fatalf("NewMemory failed: %v", err)
	}
	if err := db.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}
	return db, NewAuditRepository(db)
}

func TestAuditLog(t *testing.T) {
	db, repo := setupAuditTestDB(t)
	defer db.Close()

	entry := &AuditEntry{
		PlayerID:   1,
		PlayerName: "TestPlayer",
		Action:     AuditLogin,
		Details:    "User logged in",
		IPAddress:  "127.0.0.1",
	}

	err := repo.Log(entry)
	if err != nil {
		t.Fatalf("Log failed: %v", err)
	}
}

func TestAuditLogAction(t *testing.T) {
	db, repo := setupAuditTestDB(t)
	defer db.Close()

	err := repo.LogAction(1, "TestPlayer", AuditKill, "Killed Agent Smith")
	if err != nil {
		t.Fatalf("LogAction failed: %v", err)
	}
}

func TestAuditGetRecent(t *testing.T) {
	db, repo := setupAuditTestDB(t)
	defer db.Close()

	repo.LogAction(1, "Player1", AuditLogin, "Login")
	repo.LogAction(2, "Player2", AuditLevelUp, "Level up to 5")
	repo.LogAction(3, "Player3", AuditDeath, "Died")

	logs, err := repo.GetRecent(2)
	if err != nil {
		t.Fatalf("GetRecent failed: %v", err)
	}
	if len(logs) != 2 {
		t.Errorf("Expected 2 logs, got %d", len(logs))
	}
}

func TestAuditGetByPlayer(t *testing.T) {
	db, repo := setupAuditTestDB(t)
	defer db.Close()

	repo.LogAction(1, "Player1", AuditLogin, "Login 1")
	repo.LogAction(1, "Player1", AuditLogout, "Logout 1")
	repo.LogAction(2, "Player2", AuditLogin, "Login 2")

	logs, err := repo.GetByPlayer(1, 10)
	if err != nil {
		t.Fatalf("GetByPlayer failed: %v", err)
	}
	if len(logs) != 2 {
		t.Errorf("Expected 2 logs for player 1, got %d", len(logs))
	}
}

func TestAuditGetByPlayerName(t *testing.T) {
	db, repo := setupAuditTestDB(t)
	defer db.Close()

	repo.LogAction(1, "NameTest", AuditLogin, "Login")
	repo.LogAction(1, "NameTest", AuditKill, "Kill")

	logs, err := repo.GetByPlayerName("NameTest", 10)
	if err != nil {
		t.Fatalf("GetByPlayerName failed: %v", err)
	}
	if len(logs) != 2 {
		t.Errorf("Expected 2 logs, got %d", len(logs))
	}
}

func TestAuditGetByAction(t *testing.T) {
	db, repo := setupAuditTestDB(t)
	defer db.Close()

	repo.LogAction(1, "Player1", AuditLogin, "Login 1")
	repo.LogAction(2, "Player2", AuditLogin, "Login 2")
	repo.LogAction(3, "Player3", AuditLogout, "Logout")

	logs, err := repo.GetByAction(AuditLogin, 10)
	if err != nil {
		t.Fatalf("GetByAction failed: %v", err)
	}
	if len(logs) != 2 {
		t.Errorf("Expected 2 login logs, got %d", len(logs))
	}
}

func TestAuditGetInTimeRange(t *testing.T) {
	db, repo := setupAuditTestDB(t)
	defer db.Close()

	start := time.Now().Add(-time.Hour)
	repo.LogAction(1, "Player1", AuditLogin, "In range")
	end := time.Now().Add(time.Hour)

	logs, err := repo.GetInTimeRange(start, end, 10)
	if err != nil {
		t.Fatalf("GetInTimeRange failed: %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("Expected 1 log in range, got %d", len(logs))
	}
}

func TestAuditSearch(t *testing.T) {
	db, repo := setupAuditTestDB(t)
	defer db.Close()

	repo.LogAction(1, "Player1", AuditKill, "Killed Agent Smith")
	repo.LogAction(2, "Player2", AuditKill, "Killed Security Guard")
	repo.LogAction(3, "Player3", AuditLevelUp, "Level up")

	logs, err := repo.Search("Agent", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("Expected 1 log matching 'Agent', got %d", len(logs))
	}
}

func TestAuditGetCount(t *testing.T) {
	db, repo := setupAuditTestDB(t)
	defer db.Close()

	count, _ := repo.GetCount()
	if count != 0 {
		t.Errorf("Initial count should be 0, got %d", count)
	}

	repo.LogAction(1, "Player", AuditLogin, "Log 1")
	repo.LogAction(1, "Player", AuditLogout, "Log 2")

	count, _ = repo.GetCount()
	if count != 2 {
		t.Errorf("Count should be 2, got %d", count)
	}
}

func TestAuditGetCountByAction(t *testing.T) {
	db, repo := setupAuditTestDB(t)
	defer db.Close()

	repo.LogAction(1, "P1", AuditLogin, "L1")
	repo.LogAction(2, "P2", AuditLogin, "L2")
	repo.LogAction(3, "P3", AuditLogout, "L3")

	count, err := repo.GetCountByAction(AuditLogin)
	if err != nil {
		t.Fatalf("GetCountByAction failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 logins, got %d", count)
	}
}

func TestAuditCleanup(t *testing.T) {
	db, repo := setupAuditTestDB(t)
	defer db.Close()

	repo.LogAction(1, "Player", AuditLogin, "Old log")
	
	// Cleanup logs older than 1 hour (our logs are < 1 second old)
	deleted, err := repo.Cleanup(time.Hour)
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}
	if deleted != 0 {
		t.Errorf("Should not delete recent logs, deleted %d", deleted)
	}

	// All logs still exist
	count, _ := repo.GetCount()
	if count != 1 {
		t.Errorf("Log should still exist, count = %d", count)
	}
}

func TestSessionCreate(t *testing.T) {
	db, repo := setupAuditTestDB(t)
	defer db.Close()

	session := &Session{
		PlayerID:     1,
		SessionToken: "token123",
		IPAddress:    "127.0.0.1",
		UserAgent:    "TestAgent",
		ExpiresAt:    time.Now().Add(time.Hour),
	}

	err := repo.CreateSession(session)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	if session.ID == 0 {
		t.Error("Session ID should be set")
	}
}

func TestSessionGet(t *testing.T) {
	db, repo := setupAuditTestDB(t)
	defer db.Close()

	original := &Session{
		PlayerID:     1,
		SessionToken: "gettoken",
		IPAddress:    "192.168.1.1",
		UserAgent:    "Browser",
		ExpiresAt:    time.Now().Add(time.Hour),
	}
	repo.CreateSession(original)

	session, err := repo.GetSession("gettoken")
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if session == nil {
		t.Fatal("Session not found")
	}
	if session.PlayerID != 1 {
		t.Errorf("PlayerID = %d, want 1", session.PlayerID)
	}
	if session.IPAddress != "192.168.1.1" {
		t.Errorf("IPAddress = %s, want 192.168.1.1", session.IPAddress)
	}
}

func TestSessionGetNotFound(t *testing.T) {
	db, repo := setupAuditTestDB(t)
	defer db.Close()

	session, err := repo.GetSession("nonexistent")
	if err != nil {
		t.Errorf("Should not error: %v", err)
	}
	if session != nil {
		t.Error("Should return nil for non-existent session")
	}
}

func TestSessionDelete(t *testing.T) {
	db, repo := setupAuditTestDB(t)
	defer db.Close()

	repo.CreateSession(&Session{
		PlayerID:     1,
		SessionToken: "deletetoken",
		ExpiresAt:    time.Now().Add(time.Hour),
	})

	err := repo.DeleteSession("deletetoken")
	if err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}

	session, _ := repo.GetSession("deletetoken")
	if session != nil {
		t.Error("Session should be deleted")
	}
}

func TestSessionDeleteExpired(t *testing.T) {
	db, repo := setupAuditTestDB(t)
	defer db.Close()

	// Create expired session
	repo.CreateSession(&Session{
		PlayerID:     1,
		SessionToken: "expired",
		ExpiresAt:    time.Now().Add(-time.Hour), // Already expired
	})

	// Create valid session
	repo.CreateSession(&Session{
		PlayerID:     2,
		SessionToken: "valid",
		ExpiresAt:    time.Now().Add(time.Hour),
	})

	deleted, err := repo.DeleteExpiredSessions()
	if err != nil {
		t.Fatalf("DeleteExpiredSessions failed: %v", err)
	}
	if deleted != 1 {
		t.Errorf("Expected 1 deleted, got %d", deleted)
	}

	// Verify
	expired, _ := repo.GetSession("expired")
	if expired != nil {
		t.Error("Expired session should be deleted")
	}

	valid, _ := repo.GetSession("valid")
	if valid == nil {
		t.Error("Valid session should still exist")
	}
}

func TestGetPlayerSessions(t *testing.T) {
	db, repo := setupAuditTestDB(t)
	defer db.Close()

	repo.CreateSession(&Session{PlayerID: 1, SessionToken: "p1s1", ExpiresAt: time.Now().Add(time.Hour)})
	repo.CreateSession(&Session{PlayerID: 1, SessionToken: "p1s2", ExpiresAt: time.Now().Add(time.Hour)})
	repo.CreateSession(&Session{PlayerID: 2, SessionToken: "p2s1", ExpiresAt: time.Now().Add(time.Hour)})

	sessions, err := repo.GetPlayerSessions(1)
	if err != nil {
		t.Fatalf("GetPlayerSessions failed: %v", err)
	}
	if len(sessions) != 2 {
		t.Errorf("Expected 2 sessions for player 1, got %d", len(sessions))
	}
}

func TestAuditActions(t *testing.T) {
	// Verify all action constants are defined
	actions := []AuditAction{
		AuditLogin, AuditLogout, AuditCreate, AuditDelete,
		AuditLevelUp, AuditDeath, AuditKill, AuditTrade,
		AuditPurchase, AuditSale, AuditQuestStart, AuditQuestEnd,
		AuditAchievement, AuditAdminAction, AuditChat, AuditPvP,
		AuditBan, AuditUnban, AuditWarning,
	}

	for _, action := range actions {
		if action == "" {
			t.Error("Action should not be empty")
		}
	}
}
