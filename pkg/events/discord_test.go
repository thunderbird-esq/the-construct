package events

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewDiscordIntegration(t *testing.T) {
	eb := NewEventBus(2)
	discord := NewDiscordIntegration("http://discord.webhook", "TestServer", "http://game.url", eb)

	if discord == nil {
		t.Fatal("NewDiscordIntegration returned nil")
	}
}

func TestDiscordSetEvents(t *testing.T) {
	eb := NewEventBus(2)
	discord := NewDiscordIntegration("", "Test", "", eb)

	// Default events should be set
	if len(discord.events) == 0 {
		t.Error("Default events should be set")
	}

	// Set custom events
	discord.SetEvents([]EventType{EventPlayerJoin, EventPlayerLeave})
	if len(discord.events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(discord.events))
	}
}

func TestDiscordStartStop(t *testing.T) {
	eb := NewEventBus(2)
	discord := NewDiscordIntegration("", "Test", "", eb)

	discord.Start()
	if discord.subID == "" {
		t.Error("subID should be set after Start")
	}

	discord.Stop()
}

func TestDiscordEventDelivery(t *testing.T) {
	var receivedMessage *DiscordMessage
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var msg DiscordMessage
		json.NewDecoder(r.Body).Decode(&msg)
		receivedMessage = &msg
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	eb := NewEventBus(2)
	eb.Start()
	defer eb.Stop()

	discord := NewDiscordIntegration(server.URL, "TestServer", "", eb)
	discord.Start()
	defer discord.Stop()

	eb.Publish(NewEvent(EventPlayerJoin).WithPlayer("TestPlayer", 1))
	time.Sleep(100 * time.Millisecond)

	if receivedMessage == nil {
		t.Fatal("Message should be received")
	}
	if receivedMessage.Username != "TestServer" {
		t.Errorf("Username = %s, want TestServer", receivedMessage.Username)
	}
	if len(receivedMessage.Embeds) == 0 {
		t.Error("Should have embeds")
	}
}

func TestDiscordFormatPlayerJoin(t *testing.T) {
	eb := NewEventBus(2)
	discord := NewDiscordIntegration("", "TestServer", "", eb)

	event := NewEvent(EventPlayerJoin).WithPlayer("Neo", 1)
	msg := discord.formatPlayerJoin(event)

	if msg == nil {
		t.Fatal("Message should not be nil")
	}
	if len(msg.Embeds) == 0 {
		t.Fatal("Should have embeds")
	}
	if msg.Embeds[0].Title != "🟢 Player Joined" {
		t.Errorf("Title = %s", msg.Embeds[0].Title)
	}
	if msg.Embeds[0].Color != ColorGreen {
		t.Errorf("Color = %d, want %d", msg.Embeds[0].Color, ColorGreen)
	}
}

func TestDiscordFormatPlayerLeave(t *testing.T) {
	eb := NewEventBus(2)
	discord := NewDiscordIntegration("", "TestServer", "", eb)

	event := NewEvent(EventPlayerLeave).WithPlayer("Neo", 1)
	msg := discord.formatPlayerLeave(event)

	if msg.Embeds[0].Title != "🔴 Player Left" {
		t.Errorf("Title = %s", msg.Embeds[0].Title)
	}
	if msg.Embeds[0].Color != ColorRed {
		t.Errorf("Color = %d, want %d", msg.Embeds[0].Color, ColorRed)
	}
}

func TestDiscordFormatLevelUp(t *testing.T) {
	eb := NewEventBus(2)
	discord := NewDiscordIntegration("", "TestServer", "", eb)

	event := NewEvent(EventPlayerLevelUp).WithPlayer("Neo", 1).WithData("level", 10)
	msg := discord.formatLevelUp(event)

	if msg.Embeds[0].Title != "⬆️ Level Up!" {
		t.Errorf("Title = %s", msg.Embeds[0].Title)
	}
	if msg.Embeds[0].Color != ColorGold {
		t.Errorf("Color = %d, want %d", msg.Embeds[0].Color, ColorGold)
	}
}

func TestDiscordFormatAchievement(t *testing.T) {
	eb := NewEventBus(2)
	discord := NewDiscordIntegration("", "TestServer", "", eb)

	event := NewEvent(EventAchievement).
		WithPlayer("Neo", 1).
		WithData("achievement", "The One")
	msg := discord.formatAchievement(event)

	if msg.Embeds[0].Title != "🏆 Achievement Unlocked!" {
		t.Errorf("Title = %s", msg.Embeds[0].Title)
	}
	if msg.Embeds[0].Color != ColorPurple {
		t.Errorf("Color = %d, want %d", msg.Embeds[0].Color, ColorPurple)
	}
}

func TestDiscordFormatNPCKill(t *testing.T) {
	eb := NewEventBus(2)
	discord := NewDiscordIntegration("", "TestServer", "", eb)

	event := NewEvent(EventNPCKill).
		WithPlayer("Neo", 1).
		WithData("npc_name", "Agent Smith")
	msg := discord.formatNPCKill(event)

	if msg.Embeds[0].Title != "⚔️ Enemy Defeated" {
		t.Errorf("Title = %s", msg.Embeds[0].Title)
	}
}

func TestDiscordFormatPvPKill(t *testing.T) {
	eb := NewEventBus(2)
	discord := NewDiscordIntegration("", "TestServer", "", eb)

	event := NewEvent(EventPvPKill).
		WithPlayer("Neo", 1).
		WithData("victim", "AgentSmith")
	msg := discord.formatPvPKill(event)

	if msg.Embeds[0].Title != "🗡️ PvP Kill" {
		t.Errorf("Title = %s", msg.Embeds[0].Title)
	}
	if len(msg.Embeds[0].Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(msg.Embeds[0].Fields))
	}
}

func TestDiscordFormatQuestComplete(t *testing.T) {
	eb := NewEventBus(2)
	discord := NewDiscordIntegration("", "TestServer", "", eb)

	event := NewEvent(EventQuestComplete).
		WithPlayer("Neo", 1).
		WithData("quest_name", "The One")
	msg := discord.formatQuestComplete(event)

	if msg.Embeds[0].Title != "📜 Quest Complete!" {
		t.Errorf("Title = %s", msg.Embeds[0].Title)
	}
}

func TestDiscordFormatServerStart(t *testing.T) {
	eb := NewEventBus(2)
	discord := NewDiscordIntegration("", "TestServer", "", eb)

	event := NewEvent(EventServerStart).WithData("version", "1.0.0")
	msg := discord.formatServerStart(event)

	if msg.Embeds[0].Title != "🚀 Server Online" {
		t.Errorf("Title = %s", msg.Embeds[0].Title)
	}
	if len(msg.Embeds[0].Fields) == 0 {
		t.Error("Should have version field")
	}
}

func TestDiscordFormatServerStop(t *testing.T) {
	eb := NewEventBus(2)
	discord := NewDiscordIntegration("", "TestServer", "", eb)

	event := NewEvent(EventServerStop)
	msg := discord.formatServerStop(event)

	if msg.Embeds[0].Title != "🛑 Server Offline" {
		t.Errorf("Title = %s", msg.Embeds[0].Title)
	}
}

func TestDiscordSendCustomMessage(t *testing.T) {
	var received string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var msg DiscordMessage
		json.NewDecoder(r.Body).Decode(&msg)
		received = msg.Content
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	eb := NewEventBus(2)
	discord := NewDiscordIntegration(server.URL, "TestServer", "", eb)

	err := discord.SendCustomMessage("Test message")
	if err != nil {
		t.Fatalf("SendCustomMessage failed: %v", err)
	}

	if received != "Test message" {
		t.Errorf("Content = %s, want 'Test message'", received)
	}
}

func TestDiscordSendCustomEmbed(t *testing.T) {
	var received DiscordMessage
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	eb := NewEventBus(2)
	discord := NewDiscordIntegration(server.URL, "TestServer", "", eb)

	err := discord.SendCustomEmbed(DiscordEmbed{
		Title:       "Custom Title",
		Description: "Custom Description",
		Color:       ColorMatrix,
	})
	if err != nil {
		t.Fatalf("SendCustomEmbed failed: %v", err)
	}

	if len(received.Embeds) == 0 {
		t.Fatal("Should have embeds")
	}
	if received.Embeds[0].Title != "Custom Title" {
		t.Errorf("Title = %s", received.Embeds[0].Title)
	}
	if received.Embeds[0].Footer == nil {
		t.Error("Footer should be auto-added")
	}
	if received.Embeds[0].Timestamp == "" {
		t.Error("Timestamp should be auto-added")
	}
}

func TestDiscordNoWebhookURL(t *testing.T) {
	eb := NewEventBus(2)
	discord := NewDiscordIntegration("", "TestServer", "", eb) // Empty URL

	// Should not error, just do nothing
	err := discord.SendCustomMessage("Test")
	if err != nil {
		t.Errorf("Should not error with empty URL: %v", err)
	}
}

func TestDiscordFilterEvent(t *testing.T) {
	eb := NewEventBus(2)
	discord := NewDiscordIntegration("", "Test", "", eb)
	discord.SetEvents([]EventType{EventPlayerJoin, EventPlayerLeave})

	if !discord.filterEvent(NewEvent(EventPlayerJoin)) {
		t.Error("Should pass filter for PlayerJoin")
	}
	if !discord.filterEvent(NewEvent(EventPlayerLeave)) {
		t.Error("Should pass filter for PlayerLeave")
	}
	if discord.filterEvent(NewEvent(EventNPCKill)) {
		t.Error("Should not pass filter for NPCKill")
	}
}

func TestDiscordColors(t *testing.T) {
	// Verify color constants are defined correctly
	colors := map[string]int{
		"Green":  ColorGreen,
		"Red":    ColorRed,
		"Blue":   ColorBlue,
		"Yellow": ColorYellow,
		"Purple": ColorPurple,
		"Orange": ColorOrange,
		"Cyan":   ColorCyan,
		"Gold":   ColorGold,
		"Matrix": ColorMatrix,
	}

	for name, color := range colors {
		if color == 0 {
			t.Errorf("Color %s should not be 0", name)
		}
	}
}

func TestDiscordFormatDeath(t *testing.T) {
	eb := NewEventBus(2)
	discord := NewDiscordIntegration("", "TestServer", "", eb)

	event := NewEvent(EventPlayerDeath).
		WithPlayer("Neo", 1).
		WithData("cause", "Agent Smith")
	msg := discord.formatDeath(event)

	if msg.Embeds[0].Title != "💀 Player Death" {
		t.Errorf("Title = %s", msg.Embeds[0].Title)
	}
}

func TestDiscordFormatUnknownEvent(t *testing.T) {
	eb := NewEventBus(2)
	discord := NewDiscordIntegration("", "TestServer", "", eb)

	event := NewEvent(EventItemPickup) // Not in default formatters
	msg := discord.formatEvent(event)

	if msg != nil {
		t.Error("Unknown event should return nil")
	}
}

func TestDiscordLevelUpWithFloatLevel(t *testing.T) {
	eb := NewEventBus(2)
	discord := NewDiscordIntegration("", "TestServer", "", eb)

	// Test with float64 (common from JSON)
	event := NewEvent(EventPlayerLevelUp).WithPlayer("Neo", 1).WithData("level", float64(15))
	msg := discord.formatLevelUp(event)

	if msg == nil {
		t.Fatal("Message should not be nil")
	}
}
