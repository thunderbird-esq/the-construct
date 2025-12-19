// Package chat implements global chat channels for Matrix MUD.
// Supports global, faction, trade, help, and party channels with moderation.
package chat

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// ChannelType defines the type of chat channel
type ChannelType string

const (
	ChannelGlobal  ChannelType = "global"
	ChannelTrade   ChannelType = "trade"
	ChannelHelp    ChannelType = "help"
	ChannelFaction ChannelType = "faction"
	ChannelParty   ChannelType = "party"
)

// Message represents a chat message
type Message struct {
	ID        int64
	Channel   string
	Sender    string
	Content   string
	Timestamp time.Time
}

// Channel represents a chat channel
type Channel struct {
	ID          string
	Name        string
	Type        ChannelType
	Description string
	Members     map[string]bool      // player name -> is member
	Moderators  map[string]bool      // player name -> is moderator
	Muted       map[string]time.Time // player name -> mute expires
	FactionID   string               // for faction channels
	mu          sync.RWMutex
}

// Manager handles all chat operations
type Manager struct {
	mu             sync.RWMutex
	Channels       map[string]*Channel      // channel ID -> channel
	PlayerChannels map[string][]string      // player name -> channel IDs
	Ignored        map[string]map[string]bool // player -> ignored players
	MessageHistory map[string][]Message     // channel ID -> recent messages
	messageID      int64
	rateLimits     map[string][]time.Time   // player -> message timestamps
}

// Profanity filter patterns
var profanityPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bf+u+c+k+\b`),
	regexp.MustCompile(`(?i)\bs+h+i+t+\b`),
	regexp.MustCompile(`(?i)\ba+s+s+h+o+l+e+\b`),
	regexp.MustCompile(`(?i)\bb+i+t+c+h+\b`),
	regexp.MustCompile(`(?i)\bc+u+n+t+\b`),
	regexp.MustCompile(`(?i)\bn+i+g+g+\b`),
	regexp.MustCompile(`(?i)\bf+a+g+g*o*t*\b`),
}

// NewManager creates a new chat manager with default channels
func NewManager() *Manager {
	m := &Manager{
		Channels:       make(map[string]*Channel),
		PlayerChannels: make(map[string][]string),
		Ignored:        make(map[string]map[string]bool),
		MessageHistory: make(map[string][]Message),
		rateLimits:     make(map[string][]time.Time),
	}
	m.createDefaultChannels()
	return m
}

// createDefaultChannels sets up the standard chat channels
func (m *Manager) createDefaultChannels() {
	// Global channel
	m.Channels["global"] = &Channel{
		ID:          "global",
		Name:        "Global",
		Type:        ChannelGlobal,
		Description: "Server-wide chat for all players",
		Members:     make(map[string]bool),
		Moderators:  make(map[string]bool),
		Muted:       make(map[string]time.Time),
	}

	// Trade channel
	m.Channels["trade"] = &Channel{
		ID:          "trade",
		Name:        "Trade",
		Type:        ChannelTrade,
		Description: "Buying, selling, and trading items",
		Members:     make(map[string]bool),
		Moderators:  make(map[string]bool),
		Muted:       make(map[string]time.Time),
	}

	// Help channel
	m.Channels["help"] = &Channel{
		ID:          "help",
		Name:        "Help",
		Type:        ChannelHelp,
		Description: "Ask questions and help new players",
		Members:     make(map[string]bool),
		Moderators:  make(map[string]bool),
		Muted:       make(map[string]time.Time),
	}

	// Faction channels
	m.Channels["zion"] = &Channel{
		ID:          "zion",
		Name:        "Zion",
		Type:        ChannelFaction,
		Description: "Resistance faction channel",
		Members:     make(map[string]bool),
		Moderators:  make(map[string]bool),
		Muted:       make(map[string]time.Time),
		FactionID:   "zion",
	}

	m.Channels["machine"] = &Channel{
		ID:          "machine",
		Name:        "Machine",
		Type:        ChannelFaction,
		Description: "Machine faction channel",
		Members:     make(map[string]bool),
		Moderators:  make(map[string]bool),
		Muted:       make(map[string]time.Time),
		FactionID:   "machine",
	}

	m.Channels["exile"] = &Channel{
		ID:          "exile",
		Name:        "Exile",
		Type:        ChannelFaction,
		Description: "Exile faction channel",
		Members:     make(map[string]bool),
		Moderators:  make(map[string]bool),
		Muted:       make(map[string]time.Time),
		FactionID:   "exile",
	}

	// Initialize message history for each channel
	for id := range m.Channels {
		m.MessageHistory[id] = make([]Message, 0, 100)
	}
}

// JoinChannel adds a player to a channel
func (m *Manager) JoinChannel(playerName, channelID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := strings.ToLower(playerName)
	channel, ok := m.Channels[channelID]
	if !ok {
		return fmt.Errorf("channel '%s' not found", channelID)
	}

	channel.mu.Lock()
	defer channel.mu.Unlock()

	if channel.Members[name] {
		return fmt.Errorf("you are already in channel '%s'", channel.Name)
	}

	channel.Members[name] = true

	// Update player's channel list
	if m.PlayerChannels[name] == nil {
		m.PlayerChannels[name] = make([]string, 0)
	}
	m.PlayerChannels[name] = append(m.PlayerChannels[name], channelID)

	return nil
}

// LeaveChannel removes a player from a channel
func (m *Manager) LeaveChannel(playerName, channelID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := strings.ToLower(playerName)
	channel, ok := m.Channels[channelID]
	if !ok {
		return fmt.Errorf("channel '%s' not found", channelID)
	}

	channel.mu.Lock()
	defer channel.mu.Unlock()

	if !channel.Members[name] {
		return fmt.Errorf("you are not in channel '%s'", channel.Name)
	}

	delete(channel.Members, name)

	// Update player's channel list
	for i, ch := range m.PlayerChannels[name] {
		if ch == channelID {
			m.PlayerChannels[name] = append(m.PlayerChannels[name][:i], m.PlayerChannels[name][i+1:]...)
			break
		}
	}

	return nil
}

// SendMessage sends a message to a channel
func (m *Manager) SendMessage(playerName, channelID, content string) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := strings.ToLower(playerName)
	channel, ok := m.Channels[channelID]
	if !ok {
		return nil, fmt.Errorf("channel '%s' not found", channelID)
	}

	channel.mu.Lock()
	defer channel.mu.Unlock()

	// Check membership
	if !channel.Members[name] {
		return nil, fmt.Errorf("you are not in channel '%s'", channel.Name)
	}

	// Check mute status
	if muteExpires, muted := channel.Muted[name]; muted {
		if time.Now().Before(muteExpires) {
			remaining := time.Until(muteExpires).Round(time.Second)
			return nil, fmt.Errorf("you are muted for %s", remaining)
		}
		delete(channel.Muted, name)
	}

	// Rate limiting (5 messages per 10 seconds)
	if !m.checkRateLimit(name) {
		return nil, fmt.Errorf("slow down! You're sending messages too quickly")
	}

	// Filter profanity
	filteredContent := m.filterProfanity(content)

	// Check for spam (repeated messages)
	if m.isSpam(name, channelID, filteredContent) {
		return nil, fmt.Errorf("please don't spam")
	}

	// Create message
	m.messageID++
	msg := Message{
		ID:        m.messageID,
		Channel:   channelID,
		Sender:    playerName,
		Content:   filteredContent,
		Timestamp: time.Now(),
	}

	// Store in history (keep last 100)
	history := m.MessageHistory[channelID]
	if len(history) >= 100 {
		history = history[1:]
	}
	m.MessageHistory[channelID] = append(history, msg)

	// Get recipients (members not ignoring sender)
	recipients := make([]string, 0)
	for member := range channel.Members {
		if member == name {
			continue
		}
		// Check if recipient is ignoring sender
		if ignored, ok := m.Ignored[member]; ok && ignored[name] {
			continue
		}
		recipients = append(recipients, member)
	}

	return recipients, nil
}

// checkRateLimit checks if a player can send a message
func (m *Manager) checkRateLimit(playerName string) bool {
	now := time.Now()
	window := 10 * time.Second
	maxMessages := 5

	// Clean old timestamps
	timestamps := m.rateLimits[playerName]
	cutoff := now.Add(-window)
	newTimestamps := make([]time.Time, 0)
	for _, t := range timestamps {
		if t.After(cutoff) {
			newTimestamps = append(newTimestamps, t)
		}
	}

	if len(newTimestamps) >= maxMessages {
		return false
	}

	newTimestamps = append(newTimestamps, now)
	m.rateLimits[playerName] = newTimestamps
	return true
}

// filterProfanity replaces profanity with asterisks
func (m *Manager) filterProfanity(content string) string {
	result := content
	for _, pattern := range profanityPatterns {
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			return strings.Repeat("*", len(match))
		})
	}
	return result
}

// isSpam checks for repeated messages (3+ identical messages in a row)
func (m *Manager) isSpam(playerName, channelID, content string) bool {
	history := m.MessageHistory[channelID]
	if len(history) < 2 {
		return false
	}

	// Check if last 2 messages from this sender are identical to new content
	sameCount := 0
	for i := len(history) - 1; i >= 0 && sameCount < 2; i-- {
		msg := history[i]
		if strings.ToLower(msg.Sender) != playerName {
			continue
		}
		if strings.EqualFold(msg.Content, content) {
			sameCount++
		} else {
			break // Different message, not consecutive spam
		}
	}
	return sameCount >= 2 // 2 previous + this one = 3 identical
}

// MutePlayer mutes a player in a channel
func (m *Manager) MutePlayer(moderatorName, targetName, channelID string, duration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	modName := strings.ToLower(moderatorName)
	targName := strings.ToLower(targetName)

	channel, ok := m.Channels[channelID]
	if !ok {
		return fmt.Errorf("channel '%s' not found", channelID)
	}

	channel.mu.Lock()
	defer channel.mu.Unlock()

	// Check moderator status
	if !channel.Moderators[modName] {
		return fmt.Errorf("you are not a moderator of '%s'", channel.Name)
	}

	// Check target is in channel
	if !channel.Members[targName] {
		return fmt.Errorf("%s is not in this channel", targetName)
	}

	// Can't mute moderators
	if channel.Moderators[targName] {
		return fmt.Errorf("cannot mute a moderator")
	}

	channel.Muted[targName] = time.Now().Add(duration)
	return nil
}

// UnmutePlayer removes a mute from a player
func (m *Manager) UnmutePlayer(moderatorName, targetName, channelID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	modName := strings.ToLower(moderatorName)
	targName := strings.ToLower(targetName)

	channel, ok := m.Channels[channelID]
	if !ok {
		return fmt.Errorf("channel '%s' not found", channelID)
	}

	channel.mu.Lock()
	defer channel.mu.Unlock()

	if !channel.Moderators[modName] {
		return fmt.Errorf("you are not a moderator of '%s'", channel.Name)
	}

	delete(channel.Muted, targName)
	return nil
}

// AddModerator adds a moderator to a channel
func (m *Manager) AddModerator(channelID, playerName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := strings.ToLower(playerName)
	channel, ok := m.Channels[channelID]
	if !ok {
		return fmt.Errorf("channel '%s' not found", channelID)
	}

	channel.mu.Lock()
	defer channel.mu.Unlock()

	channel.Moderators[name] = true
	return nil
}

// RemoveModerator removes a moderator from a channel
func (m *Manager) RemoveModerator(channelID, playerName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := strings.ToLower(playerName)
	channel, ok := m.Channels[channelID]
	if !ok {
		return fmt.Errorf("channel '%s' not found", channelID)
	}

	channel.mu.Lock()
	defer channel.mu.Unlock()

	delete(channel.Moderators, name)
	return nil
}

// IgnorePlayer adds a player to the ignore list
func (m *Manager) IgnorePlayer(playerName, targetName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := strings.ToLower(playerName)
	target := strings.ToLower(targetName)

	if name == target {
		return fmt.Errorf("you cannot ignore yourself")
	}

	if m.Ignored[name] == nil {
		m.Ignored[name] = make(map[string]bool)
	}

	if m.Ignored[name][target] {
		return fmt.Errorf("you are already ignoring %s", targetName)
	}

	m.Ignored[name][target] = true
	return nil
}

// UnignorePlayer removes a player from the ignore list
func (m *Manager) UnignorePlayer(playerName, targetName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := strings.ToLower(playerName)
	target := strings.ToLower(targetName)

	if m.Ignored[name] == nil || !m.Ignored[name][target] {
		return fmt.Errorf("you are not ignoring %s", targetName)
	}

	delete(m.Ignored[name], target)
	return nil
}

// GetIgnoredPlayers returns the list of ignored players
func (m *Manager) GetIgnoredPlayers(playerName string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	name := strings.ToLower(playerName)
	if m.Ignored[name] == nil {
		return nil
	}

	result := make([]string, 0, len(m.Ignored[name]))
	for ignored := range m.Ignored[name] {
		result = append(result, ignored)
	}
	return result
}

// ListChannels returns available channels for a player
func (m *Manager) ListChannels(playerName string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	name := strings.ToLower(playerName)
	var sb strings.Builder
	sb.WriteString("=== CHAT CHANNELS ===\r\n\r\n")

	for _, channel := range m.Channels {
		channel.mu.RLock()
		status := "[ ]"
		if channel.Members[name] {
			status = "[X]"
		}
		sb.WriteString(fmt.Sprintf("%s %s - %s\r\n", status, channel.Name, channel.Description))
		channel.mu.RUnlock()
	}

	sb.WriteString("\r\nCommands: /join <channel>, /leave <channel>, /chat <channel> <message>\r\n")
	sb.WriteString("Shortcuts: /g (global), /t (trade), /h (help)\r\n")
	return sb.String()
}

// GetChannelMembers returns members of a channel
func (m *Manager) GetChannelMembers(channelID string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	channel, ok := m.Channels[channelID]
	if !ok {
		return nil
	}

	channel.mu.RLock()
	defer channel.mu.RUnlock()

	members := make([]string, 0, len(channel.Members))
	for member := range channel.Members {
		members = append(members, member)
	}
	return members
}

// GetPlayerChannels returns channels a player has joined
func (m *Manager) GetPlayerChannels(playerName string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	name := strings.ToLower(playerName)
	channels := m.PlayerChannels[name]
	if channels == nil {
		return []string{}
	}

	result := make([]string, len(channels))
	copy(result, channels)
	return result
}

// AutoJoinDefaultChannels joins a player to default channels
func (m *Manager) AutoJoinDefaultChannels(playerName string) {
	m.JoinChannel(playerName, "global")
	m.JoinChannel(playerName, "help")
}

// JoinFactionChannel joins a player to their faction's channel
func (m *Manager) JoinFactionChannel(playerName, factionID string) error {
	channelID := strings.ToLower(factionID)
	if _, ok := m.Channels[channelID]; !ok {
		return fmt.Errorf("faction channel '%s' not found", factionID)
	}
	return m.JoinChannel(playerName, channelID)
}

// LeaveFactionChannel removes a player from a faction channel
func (m *Manager) LeaveFactionChannel(playerName, factionID string) error {
	channelID := strings.ToLower(factionID)
	return m.LeaveChannel(playerName, channelID)
}

// FormatMessage formats a chat message for display
func FormatMessage(msg Message, channelName string) string {
	timestamp := msg.Timestamp.Format("15:04")
	return fmt.Sprintf("[%s][%s] %s: %s\r\n", timestamp, channelName, msg.Sender, msg.Content)
}

// IsInChannel checks if a player is in a channel
func (m *Manager) IsInChannel(playerName, channelID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	name := strings.ToLower(playerName)
	channel, ok := m.Channels[channelID]
	if !ok {
		return false
	}

	channel.mu.RLock()
	defer channel.mu.RUnlock()

	return channel.Members[name]
}

// GetChannel returns a channel by ID
func (m *Manager) GetChannel(channelID string) *Channel {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Channels[channelID]
}

// Global chat manager
var GlobalChat = NewManager()
