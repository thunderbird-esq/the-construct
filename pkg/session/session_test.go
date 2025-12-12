package session

import (
	"strings"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.sessions == nil {
		t.Error("sessions map not initialized")
	}
	if m.byPlayer == nil {
		t.Error("byPlayer map not initialized")
	}
}

func TestCreateSession(t *testing.T) {
	m := NewManager()

	session := m.CreateSession("testplayer", "loading_program", 100, 50)

	if session == nil {
		t.Fatal("CreateSession returned nil")
	}
	if session.PlayerName != "testplayer" {
		t.Errorf("PlayerName = %q, want testplayer", session.PlayerName)
	}
	if session.RoomID != "loading_program" {
		t.Errorf("RoomID = %q, want loading_program", session.RoomID)
	}
	if session.HP != 100 {
		t.Errorf("HP = %d, want 100", session.HP)
	}
	if session.MP != 50 {
		t.Errorf("MP = %d, want 50", session.MP)
	}
	if !session.IsConnected {
		t.Error("New session should be connected")
	}
	if len(session.Token) != 64 {
		t.Errorf("Token length = %d, want 64", len(session.Token))
	}
}

func TestCreateSessionExisting(t *testing.T) {
	m := NewManager()

	s1 := m.CreateSession("testplayer", "room1", 100, 50)
	token1 := s1.Token

	s2 := m.CreateSession("testplayer", "room2", 80, 40)

	if s2.Token != token1 {
		t.Error("Existing session should keep same token")
	}
	if s2.RoomID != "room2" {
		t.Errorf("RoomID should be updated to room2, got %s", s2.RoomID)
	}
	if s2.HP != 80 {
		t.Errorf("HP should be updated to 80, got %d", s2.HP)
	}
}

func TestDisconnect(t *testing.T) {
	m := NewManager()

	m.CreateSession("testplayer", "room", 100, 50)
	m.Disconnect("testplayer")

	session := m.GetSession("testplayer")
	if session.IsConnected {
		t.Error("Session should be disconnected")
	}
	if session.Disconnected.IsZero() {
		t.Error("Disconnected time should be set")
	}
}

func TestDisconnectNonexistent(t *testing.T) {
	m := NewManager()
	// Should not panic
	m.Disconnect("nonexistent")
}

func TestReconnect(t *testing.T) {
	m := NewManager()

	original := m.CreateSession("testplayer", "dojo", 80, 40)
	originalToken := original.Token
	m.Disconnect("testplayer")

	session := m.Reconnect("testplayer")

	if session == nil {
		t.Fatal("Reconnect should return session")
	}
	if session.Token != originalToken {
		t.Error("Should return same session")
	}
	if !session.IsConnected {
		t.Error("Session should be reconnected")
	}
	if session.RoomID != "dojo" {
		t.Errorf("RoomID = %q, want dojo", session.RoomID)
	}
}

func TestReconnectAlreadyConnected(t *testing.T) {
	m := NewManager()

	m.CreateSession("testplayer", "room", 100, 50)

	session := m.Reconnect("testplayer")
	if session != nil {
		t.Error("Should not allow reconnect when already connected")
	}
}

func TestReconnectExpired(t *testing.T) {
	m := NewManager()

	m.CreateSession("testplayer", "room", 100, 50)

	session := m.GetSession("testplayer")
	session.IsConnected = false
	session.Disconnected = time.Now().Add(-SessionDuration - time.Minute)

	result := m.Reconnect("testplayer")
	if result != nil {
		t.Error("Should not allow reconnect for expired session")
	}
}

func TestReconnectNonexistent(t *testing.T) {
	m := NewManager()
	result := m.Reconnect("nonexistent")
	if result != nil {
		t.Error("Should return nil for nonexistent session")
	}
}

func TestIsReconnectable(t *testing.T) {
	m := NewManager()

	if m.IsReconnectable("nonexistent") {
		t.Error("Nonexistent player should not be reconnectable")
	}

	m.CreateSession("connected", "room", 100, 50)
	if m.IsReconnectable("connected") {
		t.Error("Connected player should not be reconnectable")
	}

	m.CreateSession("disconnected", "room", 100, 50)
	m.Disconnect("disconnected")
	if !m.IsReconnectable("disconnected") {
		t.Error("Disconnected player should be reconnectable")
	}
}

func TestTimeUntilExpiry(t *testing.T) {
	m := NewManager()

	m.CreateSession("testplayer", "room", 100, 50)
	m.Disconnect("testplayer")

	remaining := m.TimeUntilExpiry("testplayer")
	if remaining <= 0 || remaining > SessionDuration {
		t.Errorf("TimeUntilExpiry = %v, should be between 0 and %v", remaining, SessionDuration)
	}
}

func TestTimeUntilExpiryConnected(t *testing.T) {
	m := NewManager()

	m.CreateSession("testplayer", "room", 100, 50)
	// Don't disconnect

	remaining := m.TimeUntilExpiry("testplayer")
	if remaining != 0 {
		t.Errorf("TimeUntilExpiry for connected = %v, want 0", remaining)
	}
}

func TestTimeUntilExpiryNonexistent(t *testing.T) {
	m := NewManager()

	remaining := m.TimeUntilExpiry("nonexistent")
	if remaining != 0 {
		t.Errorf("TimeUntilExpiry for nonexistent = %v, want 0", remaining)
	}
}

func TestActiveCount(t *testing.T) {
	m := NewManager()

	if m.ActiveCount() != 0 {
		t.Error("Should start with 0 active sessions")
	}

	m.CreateSession("player1", "room", 100, 50)
	m.CreateSession("player2", "room", 100, 50)

	if m.ActiveCount() != 2 {
		t.Errorf("ActiveCount = %d, want 2", m.ActiveCount())
	}

	m.Disconnect("player1")

	if m.ActiveCount() != 1 {
		t.Errorf("ActiveCount = %d, want 1", m.ActiveCount())
	}
}

func TestGenerateToken(t *testing.T) {
	tokens := make(map[string]bool)

	for i := 0; i < 100; i++ {
		token := generateToken()
		if len(token) != 64 {
			t.Errorf("Token length = %d, want 64", len(token))
		}
		if tokens[token] {
			t.Error("Duplicate token generated")
		}
		tokens[token] = true

		for _, c := range strings.ToLower(token) {
			if !strings.ContainsRune("0123456789abcdef", c) {
				t.Errorf("Invalid character in token: %c", c)
			}
		}
	}
}

func TestSessionFields(t *testing.T) {
	m := NewManager()

	s := m.CreateSession("test", "room", 100, 50)

	if s.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
	if s.LastActive.IsZero() {
		t.Error("LastActive should be set")
	}
	if !s.Disconnected.IsZero() {
		t.Error("Disconnected should be zero for connected session")
	}
}

func TestCleanupRemovesExpired(t *testing.T) {
	m := NewManager()

	m.CreateSession("expired", "room", 100, 50)
	session := m.GetSession("expired")
	session.IsConnected = false
	session.Disconnected = time.Now().Add(-SessionDuration - time.Hour)

	m.CreateSession("active", "room", 100, 50)

	m.cleanup()

	if m.GetSession("expired") != nil {
		t.Error("Expired session should be cleaned up")
	}
	if m.GetSession("active") == nil {
		t.Error("Active session should remain")
	}
}

func TestCleanupKeepsRecent(t *testing.T) {
	m := NewManager()

	m.CreateSession("recent", "room", 100, 50)
	m.Disconnect("recent")

	m.cleanup()

	if m.GetSession("recent") == nil {
		t.Error("Recently disconnected session should remain")
	}
}
