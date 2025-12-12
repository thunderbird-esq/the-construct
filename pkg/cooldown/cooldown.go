// Package cooldown manages ability and spell cooldowns for Matrix MUD.
// It tracks when abilities were last used and enforces cooldown periods
// before they can be used again.
package cooldown

import (
	"sync"
	"time"
)

// Manager tracks cooldowns for players
type Manager struct {
	mu        sync.RWMutex
	cooldowns map[string]map[string]time.Time // player -> ability -> last used
}

// NewManager creates a new cooldown manager
func NewManager() *Manager {
	return &Manager{
		cooldowns: make(map[string]map[string]time.Time),
	}
}

// AbilityCooldowns defines the cooldown duration for each ability
var AbilityCooldowns = map[string]time.Duration{
	// Hacker skills
	"glitch":    5 * time.Second,
	"patch":     15 * time.Second,
	"overflow":  30 * time.Second,
	"backdoor":  60 * time.Second,
	
	// Enforcer skills
	"smash":     3 * time.Second,
	"fortify":   20 * time.Second,
	"rampage":   45 * time.Second,
	"ironwall":  60 * time.Second,
	
	// Operative skills
	"strike":    4 * time.Second,
	"vanish":    30 * time.Second,
	"assassinate": 60 * time.Second,
	"shadowstep": 20 * time.Second,
	
	// Common abilities
	"flee":      10 * time.Second,
	"use":       2 * time.Second,
}

// Use marks an ability as used, starting its cooldown
func (m *Manager) Use(playerName, ability string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.cooldowns[playerName] == nil {
		m.cooldowns[playerName] = make(map[string]time.Time)
	}
	m.cooldowns[playerName][ability] = time.Now()
}

// IsReady checks if an ability is off cooldown
func (m *Manager) IsReady(playerName, ability string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	playerCDs, ok := m.cooldowns[playerName]
	if !ok {
		return true
	}
	
	lastUsed, ok := playerCDs[ability]
	if !ok {
		return true
	}
	
	cooldown, ok := AbilityCooldowns[ability]
	if !ok {
		cooldown = 5 * time.Second // Default cooldown
	}
	
	return time.Since(lastUsed) >= cooldown
}

// TimeRemaining returns how long until the ability is ready
func (m *Manager) TimeRemaining(playerName, ability string) time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	playerCDs, ok := m.cooldowns[playerName]
	if !ok {
		return 0
	}
	
	lastUsed, ok := playerCDs[ability]
	if !ok {
		return 0
	}
	
	cooldown, ok := AbilityCooldowns[ability]
	if !ok {
		cooldown = 5 * time.Second
	}
	
	remaining := cooldown - time.Since(lastUsed)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetAllCooldowns returns all active cooldowns for a player
func (m *Manager) GetAllCooldowns(playerName string) map[string]time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make(map[string]time.Duration)
	
	playerCDs, ok := m.cooldowns[playerName]
	if !ok {
		return result
	}
	
	for ability, lastUsed := range playerCDs {
		cooldown, ok := AbilityCooldowns[ability]
		if !ok {
			cooldown = 5 * time.Second
		}
		
		remaining := cooldown - time.Since(lastUsed)
		if remaining > 0 {
			result[ability] = remaining
		}
	}
	
	return result
}

// Reset clears all cooldowns for a player (e.g., on death or rest)
func (m *Manager) Reset(playerName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.cooldowns, playerName)
}

// ResetAbility clears a specific ability's cooldown
func (m *Manager) ResetAbility(playerName, ability string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if playerCDs, ok := m.cooldowns[playerName]; ok {
		delete(playerCDs, ability)
	}
}

// Cleanup removes stale cooldown entries (call periodically)
func (m *Manager) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	maxCooldown := 2 * time.Minute // Longest possible cooldown
	
	for player, abilities := range m.cooldowns {
		for ability, lastUsed := range abilities {
			if time.Since(lastUsed) > maxCooldown {
				delete(abilities, ability)
			}
		}
		if len(abilities) == 0 {
			delete(m.cooldowns, player)
		}
	}
}

// GlobalCD is a global cooldown manager instance
var GlobalCD = NewManager()
