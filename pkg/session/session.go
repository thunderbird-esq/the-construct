// Package session provides persistent session management for Matrix MUD.
// It allows players to reconnect to their characters within a time window
// and maintains session state across disconnections.
package session

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// SessionDuration is how long a session remains valid after disconnect
const SessionDuration = 30 * time.Minute

// Session represents an active or suspended player session
type Session struct {
	Token        string    // Unique session token
	PlayerName   string    // Player name (lowercase)
	RoomID       string    // Last known room
	HP           int       // HP at disconnect
	MP           int       // MP at disconnect
	CreatedAt    time.Time // When session was created
	LastActive   time.Time // Last activity timestamp
	Disconnected time.Time // When player disconnected (zero if connected)
	IsConnected  bool      // Whether player is currently connected
}

// Manager handles player sessions
type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*Session // token -> session
	byPlayer map[string]*Session // player name -> session
}

// NewManager creates a new session manager
func NewManager() *Manager {
	m := &Manager{
		sessions: make(map[string]*Session),
		byPlayer: make(map[string]*Session),
	}

	// Start cleanup goroutine
	go m.cleanupLoop()

	return m
}

// generateToken creates a cryptographically secure session token
func generateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// CreateSession creates a new session for a player
func (m *Manager) CreateSession(playerName, roomID string, hp, mp int) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for existing session
	if existing, ok := m.byPlayer[playerName]; ok {
		// Update existing session
		existing.RoomID = roomID
		existing.HP = hp
		existing.MP = mp
		existing.LastActive = time.Now()
		existing.IsConnected = true
		existing.Disconnected = time.Time{}
		return existing
	}

	// Create new session
	session := &Session{
		Token:       generateToken(),
		PlayerName:  playerName,
		RoomID:      roomID,
		HP:          hp,
		MP:          mp,
		CreatedAt:   time.Now(),
		LastActive:  time.Now(),
		IsConnected: true,
	}

	m.sessions[session.Token] = session
	m.byPlayer[playerName] = session

	return session
}

// Disconnect marks a session as disconnected
func (m *Manager) Disconnect(playerName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session, ok := m.byPlayer[playerName]; ok {
		session.IsConnected = false
		session.Disconnected = time.Now()
	}
}

// Reconnect attempts to reconnect a player to their session
// Returns the session if valid and reconnectable, nil otherwise
func (m *Manager) Reconnect(playerName string) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.byPlayer[playerName]
	if !ok {
		return nil
	}

	// Check if session is still valid
	if session.IsConnected {
		return nil // Already connected elsewhere
	}

	if time.Since(session.Disconnected) > SessionDuration {
		// Session expired, clean it up
		delete(m.sessions, session.Token)
		delete(m.byPlayer, playerName)
		return nil
	}

	// Reconnect
	session.IsConnected = true
	session.LastActive = time.Now()
	session.Disconnected = time.Time{}

	return session
}

// GetSession returns the session for a player if it exists
func (m *Manager) GetSession(playerName string) *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.byPlayer[playerName]
}

// IsReconnectable checks if a player can reconnect
func (m *Manager) IsReconnectable(playerName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.byPlayer[playerName]
	if !ok {
		return false
	}

	if session.IsConnected {
		return false
	}

	return time.Since(session.Disconnected) <= SessionDuration
}

// TimeUntilExpiry returns time remaining until session expires
func (m *Manager) TimeUntilExpiry(playerName string) time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.byPlayer[playerName]
	if !ok || session.IsConnected {
		return 0
	}

	remaining := SessionDuration - time.Since(session.Disconnected)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// ActiveCount returns count of connected sessions
func (m *Manager) ActiveCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, s := range m.sessions {
		if s.IsConnected {
			count++
		}
	}
	return count
}

// cleanupLoop periodically removes expired sessions
func (m *Manager) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		m.cleanup()
	}
}

// cleanup removes expired sessions
func (m *Manager) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for token, session := range m.sessions {
		if !session.IsConnected && time.Since(session.Disconnected) > SessionDuration {
			delete(m.sessions, token)
			delete(m.byPlayer, session.PlayerName)
		}
	}
}
