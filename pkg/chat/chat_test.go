package chat

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.Channels == nil {
		t.Error("Channels map is nil")
	}
	if m.PlayerChannels == nil {
		t.Error("PlayerChannels map is nil")
	}
}

func TestDefaultChannels(t *testing.T) {
	m := NewManager()

	expectedChannels := []string{"global", "trade", "help", "zion", "machine", "exile"}
	for _, id := range expectedChannels {
		if _, ok := m.Channels[id]; !ok {
			t.Errorf("Missing default channel: %s", id)
		}
	}
}

func TestJoinChannel(t *testing.T) {
	m := NewManager()

	err := m.JoinChannel("TestPlayer", "global")
	if err != nil {
		t.Fatalf("JoinChannel failed: %v", err)
	}

	if !m.IsInChannel("TestPlayer", "global") {
		t.Error("Player should be in channel")
	}
}

func TestJoinChannelInvalid(t *testing.T) {
	m := NewManager()

	err := m.JoinChannel("TestPlayer", "nonexistent")
	if err == nil {
		t.Error("Should fail for nonexistent channel")
	}
}

func TestJoinChannelTwice(t *testing.T) {
	m := NewManager()

	m.JoinChannel("TestPlayer", "global")
	err := m.JoinChannel("TestPlayer", "global")
	if err == nil {
		t.Error("Should fail when already in channel")
	}
}

func TestLeaveChannel(t *testing.T) {
	m := NewManager()

	m.JoinChannel("TestPlayer", "global")
	err := m.LeaveChannel("TestPlayer", "global")
	if err != nil {
		t.Fatalf("LeaveChannel failed: %v", err)
	}

	if m.IsInChannel("TestPlayer", "global") {
		t.Error("Player should not be in channel after leaving")
	}
}

func TestLeaveChannelNotIn(t *testing.T) {
	m := NewManager()

	err := m.LeaveChannel("TestPlayer", "global")
	if err == nil {
		t.Error("Should fail when not in channel")
	}
}

func TestSendMessage(t *testing.T) {
	m := NewManager()

	m.JoinChannel("Sender", "global")
	m.JoinChannel("Receiver", "global")

	recipients, err := m.SendMessage("Sender", "global", "Hello world!")
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}

	if len(recipients) != 1 {
		t.Errorf("Expected 1 recipient, got %d", len(recipients))
	}
	if recipients[0] != "receiver" {
		t.Errorf("Expected receiver, got %s", recipients[0])
	}
}

func TestSendMessageNotInChannel(t *testing.T) {
	m := NewManager()

	_, err := m.SendMessage("NotMember", "global", "Hello")
	if err == nil {
		t.Error("Should fail when not in channel")
	}
}

func TestProfanityFilter(t *testing.T) {
	m := NewManager()

	filtered := m.filterProfanity("This is a fuck test")
	if strings.Contains(filtered, "fuck") {
		t.Error("Profanity should be filtered")
	}
	if !strings.Contains(filtered, "****") {
		t.Error("Profanity should be replaced with asterisks")
	}
}

func TestProfanityFilterCleanMessage(t *testing.T) {
	m := NewManager()

	original := "This is a clean message"
	filtered := m.filterProfanity(original)
	if filtered != original {
		t.Error("Clean message should not be modified")
	}
}

func TestRateLimit(t *testing.T) {
	m := NewManager()
	m.JoinChannel("Spammer", "global")

	// First 5 should succeed (use different messages to avoid spam detection)
	for i := 0; i < 5; i++ {
		_, err := m.SendMessage("Spammer", "global", fmt.Sprintf("Message %d", i))
		if err != nil {
			t.Fatalf("Message %d should succeed: %v", i+1, err)
		}
	}

	// 6th should fail due to rate limit
	_, err := m.SendMessage("Spammer", "global", "Message 5")
	if err == nil {
		t.Error("6th message should be rate limited")
	}
}

func TestMutePlayer(t *testing.T) {
	m := NewManager()

	m.JoinChannel("Mod", "global")
	m.JoinChannel("User", "global")
	m.AddModerator("global", "Mod")

	err := m.MutePlayer("Mod", "User", "global", time.Minute)
	if err != nil {
		t.Fatalf("MutePlayer failed: %v", err)
	}

	// Muted user should not be able to send
	_, err = m.SendMessage("User", "global", "Hello")
	if err == nil {
		t.Error("Muted user should not be able to send")
	}
}

func TestMutePlayerNotModerator(t *testing.T) {
	m := NewManager()

	m.JoinChannel("User1", "global")
	m.JoinChannel("User2", "global")

	err := m.MutePlayer("User1", "User2", "global", time.Minute)
	if err == nil {
		t.Error("Non-moderator should not be able to mute")
	}
}

func TestUnmutePlayer(t *testing.T) {
	m := NewManager()

	m.JoinChannel("Mod", "global")
	m.JoinChannel("User", "global")
	m.AddModerator("global", "Mod")
	m.MutePlayer("Mod", "User", "global", time.Hour)

	err := m.UnmutePlayer("Mod", "User", "global")
	if err != nil {
		t.Fatalf("UnmutePlayer failed: %v", err)
	}

	// User should be able to send now
	_, err = m.SendMessage("User", "global", "Hello")
	if err != nil {
		t.Error("Unmuted user should be able to send")
	}
}

func TestIgnorePlayer(t *testing.T) {
	m := NewManager()

	err := m.IgnorePlayer("Player1", "Player2")
	if err != nil {
		t.Fatalf("IgnorePlayer failed: %v", err)
	}

	ignored := m.GetIgnoredPlayers("Player1")
	if len(ignored) != 1 {
		t.Errorf("Expected 1 ignored, got %d", len(ignored))
	}
}

func TestIgnoreSelf(t *testing.T) {
	m := NewManager()

	err := m.IgnorePlayer("Player1", "Player1")
	if err == nil {
		t.Error("Should not be able to ignore self")
	}
}

func TestUnignorePlayer(t *testing.T) {
	m := NewManager()

	m.IgnorePlayer("Player1", "Player2")
	err := m.UnignorePlayer("Player1", "Player2")
	if err != nil {
		t.Fatalf("UnignorePlayer failed: %v", err)
	}

	ignored := m.GetIgnoredPlayers("Player1")
	if len(ignored) != 0 {
		t.Error("Should have no ignored players")
	}
}

func TestIgnoredPlayersNotReceive(t *testing.T) {
	m := NewManager()

	m.JoinChannel("Sender", "global")
	m.JoinChannel("Receiver", "global")
	m.IgnorePlayer("Receiver", "Sender")

	recipients, _ := m.SendMessage("Sender", "global", "Hello")
	for _, r := range recipients {
		if r == "receiver" {
			t.Error("Ignoring player should not receive message")
		}
	}
}

func TestListChannels(t *testing.T) {
	m := NewManager()

	output := m.ListChannels("TestPlayer")
	if !strings.Contains(output, "Global") {
		t.Error("Should list Global channel")
	}
	if !strings.Contains(output, "Trade") {
		t.Error("Should list Trade channel")
	}
}

func TestGetChannelMembers(t *testing.T) {
	m := NewManager()

	m.JoinChannel("Player1", "global")
	m.JoinChannel("Player2", "global")

	members := m.GetChannelMembers("global")
	if len(members) != 2 {
		t.Errorf("Expected 2 members, got %d", len(members))
	}
}

func TestGetPlayerChannels(t *testing.T) {
	m := NewManager()

	m.JoinChannel("TestPlayer", "global")
	m.JoinChannel("TestPlayer", "trade")

	channels := m.GetPlayerChannels("TestPlayer")
	if len(channels) != 2 {
		t.Errorf("Expected 2 channels, got %d", len(channels))
	}
}

func TestAutoJoinDefaultChannels(t *testing.T) {
	m := NewManager()

	m.AutoJoinDefaultChannels("NewPlayer")

	if !m.IsInChannel("NewPlayer", "global") {
		t.Error("Should auto-join global")
	}
	if !m.IsInChannel("NewPlayer", "help") {
		t.Error("Should auto-join help")
	}
}

func TestJoinFactionChannel(t *testing.T) {
	m := NewManager()

	err := m.JoinFactionChannel("ZionPlayer", "zion")
	if err != nil {
		t.Fatalf("JoinFactionChannel failed: %v", err)
	}

	if !m.IsInChannel("ZionPlayer", "zion") {
		t.Error("Should be in faction channel")
	}
}

func TestFormatMessage(t *testing.T) {
	msg := Message{
		ID:        1,
		Channel:   "global",
		Sender:    "TestPlayer",
		Content:   "Hello world!",
		Timestamp: time.Now(),
	}

	formatted := FormatMessage(msg, "Global")
	if !strings.Contains(formatted, "TestPlayer") {
		t.Error("Should contain sender name")
	}
	if !strings.Contains(formatted, "Hello world!") {
		t.Error("Should contain message content")
	}
	if !strings.Contains(formatted, "Global") {
		t.Error("Should contain channel name")
	}
}

func TestAddRemoveModerator(t *testing.T) {
	m := NewManager()

	err := m.AddModerator("global", "ModPlayer")
	if err != nil {
		t.Fatalf("AddModerator failed: %v", err)
	}

	err = m.RemoveModerator("global", "ModPlayer")
	if err != nil {
		t.Fatalf("RemoveModerator failed: %v", err)
	}
}

func TestGetChannel(t *testing.T) {
	m := NewManager()

	channel := m.GetChannel("global")
	if channel == nil {
		t.Error("Should return global channel")
	}

	channel = m.GetChannel("nonexistent")
	if channel != nil {
		t.Error("Should return nil for nonexistent channel")
	}
}

func TestSpamDetection(t *testing.T) {
	m := NewManager()
	m.JoinChannel("Player", "global")

	// Add same message to history twice (simulating 2 previous identical messages)
	for i := 0; i < 2; i++ {
		m.MessageHistory["global"] = append(m.MessageHistory["global"], Message{
			Sender:  "Player",
			Content: "spam message",
		})
	}

	// Third identical message should be detected as spam
	if !m.isSpam("player", "global", "spam message") {
		t.Error("Should detect 3rd identical message as spam")
	}

	if m.isSpam("player", "global", "different message") {
		t.Error("Should not detect different message as spam")
	}
}

func TestCaseInsensitivePlayerNames(t *testing.T) {
	m := NewManager()

	m.JoinChannel("TestPlayer", "global")

	// Check with different cases
	if !m.IsInChannel("testplayer", "global") {
		t.Error("Should be case insensitive (lowercase)")
	}
	if !m.IsInChannel("TESTPLAYER", "global") {
		t.Error("Should be case insensitive (uppercase)")
	}
}
