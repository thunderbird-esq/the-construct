// Package party implements the group/party system for Matrix MUD.
// Parties allow players to team up for combat, share XP, and coordinate.
package party

import (
	"fmt"
	"sync"
	"time"
)

// MaxPartySize is the maximum number of members in a party
const MaxPartySize = 5

// Party represents a group of players
type Party struct {
	ID        string    // Unique party ID
	Leader    string    // Leader's player name (lowercase)
	Members   []string  // All members including leader
	Invites   map[string]time.Time // Pending invites (player -> invite time)
	CreatedAt time.Time
	mu        sync.RWMutex
}

// Manager handles all active parties
type Manager struct {
	mu       sync.RWMutex
	parties  map[string]*Party // party ID -> party
	byPlayer map[string]string // player name -> party ID
}

// NewManager creates a new party manager
func NewManager() *Manager {
	return &Manager{
		parties:  make(map[string]*Party),
		byPlayer: make(map[string]string),
	}
}

// generateID creates a unique party ID
func generateID() string {
	return fmt.Sprintf("party_%d", time.Now().UnixNano())
}

// Create creates a new party with the given player as leader
func (m *Manager) Create(leaderName string) (*Party, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already in a party
	if _, ok := m.byPlayer[leaderName]; ok {
		return nil, fmt.Errorf("you are already in a party")
	}

	party := &Party{
		ID:        generateID(),
		Leader:    leaderName,
		Members:   []string{leaderName},
		Invites:   make(map[string]time.Time),
		CreatedAt: time.Now(),
	}

	m.parties[party.ID] = party
	m.byPlayer[leaderName] = party.ID

	return party, nil
}

// Invite sends a party invite to a player
func (m *Manager) Invite(inviterName, inviteeName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get inviter's party
	partyID, ok := m.byPlayer[inviterName]
	if !ok {
		return fmt.Errorf("you are not in a party")
	}

	party := m.parties[partyID]
	party.mu.Lock()
	defer party.mu.Unlock()

	// Only leader can invite
	if party.Leader != inviterName {
		return fmt.Errorf("only the party leader can invite")
	}

	// Check if party is full
	if len(party.Members) >= MaxPartySize {
		return fmt.Errorf("party is full (max %d members)", MaxPartySize)
	}

	// Check if target is already in a party
	if _, ok := m.byPlayer[inviteeName]; ok {
		return fmt.Errorf("%s is already in a party", inviteeName)
	}

	// Check if already invited
	if _, ok := party.Invites[inviteeName]; ok {
		return fmt.Errorf("%s already has a pending invite", inviteeName)
	}

	party.Invites[inviteeName] = time.Now()
	return nil
}

// Accept accepts a party invite
func (m *Manager) Accept(playerName, leaderName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already in a party
	if _, ok := m.byPlayer[playerName]; ok {
		return fmt.Errorf("you are already in a party")
	}

	// Find party by leader
	var party *Party
	for _, p := range m.parties {
		if p.Leader == leaderName {
			party = p
			break
		}
	}

	if party == nil {
		return fmt.Errorf("no party found with leader %s", leaderName)
	}

	party.mu.Lock()
	defer party.mu.Unlock()

	// Check for valid invite
	inviteTime, ok := party.Invites[playerName]
	if !ok {
		return fmt.Errorf("you don't have an invite from %s", leaderName)
	}

	// Check invite expiry (5 minutes)
	if time.Since(inviteTime) > 5*time.Minute {
		delete(party.Invites, playerName)
		return fmt.Errorf("invite has expired")
	}

	// Check party size
	if len(party.Members) >= MaxPartySize {
		return fmt.Errorf("party is full")
	}

	// Add to party
	party.Members = append(party.Members, playerName)
	delete(party.Invites, playerName)
	m.byPlayer[playerName] = party.ID

	return nil
}

// Decline declines a party invite
func (m *Manager) Decline(playerName, leaderName string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Find party by leader
	for _, party := range m.parties {
		if party.Leader == leaderName {
			party.mu.Lock()
			if _, ok := party.Invites[playerName]; ok {
				delete(party.Invites, playerName)
				party.mu.Unlock()
				return nil
			}
			party.mu.Unlock()
			return fmt.Errorf("no pending invite from %s", leaderName)
		}
	}

	return fmt.Errorf("no party found with leader %s", leaderName)
}

// Leave removes a player from their party
func (m *Manager) Leave(playerName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	partyID, ok := m.byPlayer[playerName]
	if !ok {
		return fmt.Errorf("you are not in a party")
	}

	party := m.parties[partyID]
	party.mu.Lock()
	defer party.mu.Unlock()

	// Remove from members
	for i, member := range party.Members {
		if member == playerName {
			party.Members = append(party.Members[:i], party.Members[i+1:]...)
			break
		}
	}
	delete(m.byPlayer, playerName)

	// If leader left, promote or disband
	if party.Leader == playerName {
		if len(party.Members) > 0 {
			party.Leader = party.Members[0]
		} else {
			// Disband empty party
			delete(m.parties, partyID)
		}
	}

	// Disband if empty
	if len(party.Members) == 0 {
		delete(m.parties, partyID)
	}

	return nil
}

// Kick removes a member from the party (leader only)
func (m *Manager) Kick(leaderName, targetName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	partyID, ok := m.byPlayer[leaderName]
	if !ok {
		return fmt.Errorf("you are not in a party")
	}

	party := m.parties[partyID]
	party.mu.Lock()
	defer party.mu.Unlock()

	if party.Leader != leaderName {
		return fmt.Errorf("only the party leader can kick members")
	}

	if targetName == leaderName {
		return fmt.Errorf("you cannot kick yourself")
	}

	// Find and remove target
	found := false
	for i, member := range party.Members {
		if member == targetName {
			party.Members = append(party.Members[:i], party.Members[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("%s is not in your party", targetName)
	}

	delete(m.byPlayer, targetName)
	return nil
}

// Promote makes another member the party leader
func (m *Manager) Promote(leaderName, targetName string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	partyID, ok := m.byPlayer[leaderName]
	if !ok {
		return fmt.Errorf("you are not in a party")
	}

	party := m.parties[partyID]
	party.mu.Lock()
	defer party.mu.Unlock()

	if party.Leader != leaderName {
		return fmt.Errorf("only the party leader can promote")
	}

	// Verify target is in party
	found := false
	for _, member := range party.Members {
		if member == targetName {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("%s is not in your party", targetName)
	}

	party.Leader = targetName
	return nil
}

// Disband removes all members and deletes the party
func (m *Manager) Disband(leaderName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	partyID, ok := m.byPlayer[leaderName]
	if !ok {
		return fmt.Errorf("you are not in a party")
	}

	party := m.parties[partyID]
	party.mu.Lock()

	if party.Leader != leaderName {
		party.mu.Unlock()
		return fmt.Errorf("only the party leader can disband")
	}

	// Remove all members
	for _, member := range party.Members {
		delete(m.byPlayer, member)
	}

	party.mu.Unlock()
	delete(m.parties, partyID)

	return nil
}

// GetParty returns the party for a player, or nil if not in one
func (m *Manager) GetParty(playerName string) *Party {
	m.mu.RLock()
	defer m.mu.RUnlock()

	partyID, ok := m.byPlayer[playerName]
	if !ok {
		return nil
	}

	return m.parties[partyID]
}

// GetMembers returns all members of a player's party
func (m *Manager) GetMembers(playerName string) []string {
	party := m.GetParty(playerName)
	if party == nil {
		return nil
	}

	party.mu.RLock()
	defer party.mu.RUnlock()

	members := make([]string, len(party.Members))
	copy(members, party.Members)
	return members
}

// IsLeader checks if a player is a party leader
func (m *Manager) IsLeader(playerName string) bool {
	party := m.GetParty(playerName)
	if party == nil {
		return false
	}

	party.mu.RLock()
	defer party.mu.RUnlock()

	return party.Leader == playerName
}

// IsInParty checks if a player is in any party
func (m *Manager) IsInParty(playerName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.byPlayer[playerName]
	return ok
}

// AreInSameParty checks if two players are in the same party
func (m *Manager) AreInSameParty(player1, player2 string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	party1, ok1 := m.byPlayer[player1]
	party2, ok2 := m.byPlayer[player2]

	return ok1 && ok2 && party1 == party2
}

// GetPendingInvites returns any pending invites for a player
func (m *Manager) GetPendingInvites(playerName string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var invites []string
	for _, party := range m.parties {
		party.mu.RLock()
		if _, ok := party.Invites[playerName]; ok {
			invites = append(invites, party.Leader)
		}
		party.mu.RUnlock()
	}

	return invites
}

// CalculateXPShare calculates XP distribution for party members
// Returns a map of player -> XP amount
func (m *Manager) CalculateXPShare(playerName string, baseXP int) map[string]int {
	party := m.GetParty(playerName)
	if party == nil {
		return map[string]int{playerName: baseXP}
	}

	party.mu.RLock()
	defer party.mu.RUnlock()

	// Party bonus: 10% per additional member
	memberCount := len(party.Members)
	bonusMultiplier := 1.0 + float64(memberCount-1)*0.1

	totalXP := int(float64(baseXP) * bonusMultiplier)
	shareXP := totalXP / memberCount

	result := make(map[string]int)
	for _, member := range party.Members {
		result[member] = shareXP
	}

	return result
}

// GlobalParty is a global party manager instance
var GlobalParty = NewManager()
