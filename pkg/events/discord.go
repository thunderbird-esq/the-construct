// Package events provides Discord webhook integration.
package events

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// DiscordEmbed represents a Discord embed message
type DiscordEmbed struct {
	Title       string         `json:"title,omitempty"`
	Description string         `json:"description,omitempty"`
	URL         string         `json:"url,omitempty"`
	Color       int            `json:"color,omitempty"`
	Footer      *DiscordFooter `json:"footer,omitempty"`
	Thumbnail   *DiscordImage  `json:"thumbnail,omitempty"`
	Image       *DiscordImage  `json:"image,omitempty"`
	Author      *DiscordAuthor `json:"author,omitempty"`
	Fields      []DiscordField `json:"fields,omitempty"`
	Timestamp   string         `json:"timestamp,omitempty"`
}

// DiscordFooter represents embed footer
type DiscordFooter struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url,omitempty"`
}

// DiscordImage represents embed image
type DiscordImage struct {
	URL string `json:"url"`
}

// DiscordAuthor represents embed author
type DiscordAuthor struct {
	Name    string `json:"name"`
	URL     string `json:"url,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

// DiscordField represents embed field
type DiscordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// DiscordMessage represents a Discord webhook message
type DiscordMessage struct {
	Content   string         `json:"content,omitempty"`
	Username  string         `json:"username,omitempty"`
	AvatarURL string         `json:"avatar_url,omitempty"`
	Embeds    []DiscordEmbed `json:"embeds,omitempty"`
}

// Color constants for Discord embeds
const (
	ColorGreen  = 0x00FF00
	ColorRed    = 0xFF0000
	ColorBlue   = 0x0000FF
	ColorYellow = 0xFFFF00
	ColorPurple = 0x800080
	ColorOrange = 0xFFA500
	ColorCyan   = 0x00FFFF
	ColorGold   = 0xFFD700
	ColorMatrix = 0x00FF41 // Matrix green
)

// DiscordIntegration handles Discord webhook notifications
type DiscordIntegration struct {
	webhookURL string
	serverName string
	serverURL  string
	eventBus   *EventBus
	subID      string
	httpClient *http.Client
	events     []EventType
}

// NewDiscordIntegration creates a new Discord integration
func NewDiscordIntegration(webhookURL, serverName, serverURL string, eventBus *EventBus) *DiscordIntegration {
	return &DiscordIntegration{
		webhookURL: webhookURL,
		serverName: serverName,
		serverURL:  serverURL,
		eventBus:   eventBus,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		events:     defaultDiscordEvents(),
	}
}

// defaultDiscordEvents returns the default events to notify
func defaultDiscordEvents() []EventType {
	return []EventType{
		EventPlayerJoin,
		EventPlayerLeave,
		EventPlayerLevelUp,
		EventAchievement,
		EventNPCKill,
		EventPvPKill,
		EventQuestComplete,
		EventServerStart,
		EventServerStop,
	}
}

// SetEvents sets which events to notify
func (d *DiscordIntegration) SetEvents(events []EventType) {
	d.events = events
}

// Start begins listening for events
func (d *DiscordIntegration) Start() {
	d.subID = d.eventBus.SubscribeAllWithFilter(d.handleEvent, d.filterEvent)
}

// Stop stops listening for events
func (d *DiscordIntegration) Stop() {
	if d.subID != "" {
		d.eventBus.Unsubscribe(d.subID)
	}
}

// filterEvent checks if event should be sent to Discord
func (d *DiscordIntegration) filterEvent(event *Event) bool {
	for _, et := range d.events {
		if et == event.Type {
			return true
		}
	}
	return false
}

// handleEvent processes an event and sends to Discord
func (d *DiscordIntegration) handleEvent(event *Event) {
	message := d.formatEvent(event)
	if message != nil {
		d.sendMessage(message)
	}
}

// formatEvent formats an event as a Discord message
func (d *DiscordIntegration) formatEvent(event *Event) *DiscordMessage {
	switch event.Type {
	case EventPlayerJoin:
		return d.formatPlayerJoin(event)
	case EventPlayerLeave:
		return d.formatPlayerLeave(event)
	case EventPlayerLevelUp:
		return d.formatLevelUp(event)
	case EventPlayerDeath:
		return d.formatDeath(event)
	case EventAchievement:
		return d.formatAchievement(event)
	case EventNPCKill:
		return d.formatNPCKill(event)
	case EventPvPKill:
		return d.formatPvPKill(event)
	case EventQuestComplete:
		return d.formatQuestComplete(event)
	case EventServerStart:
		return d.formatServerStart(event)
	case EventServerStop:
		return d.formatServerStop(event)
	default:
		return nil
	}
}

func (d *DiscordIntegration) formatPlayerJoin(event *Event) *DiscordMessage {
	return &DiscordMessage{
		Username:  d.serverName,
		AvatarURL: "",
		Embeds: []DiscordEmbed{{
			Title:       "🟢 Player Joined",
			Description: fmt.Sprintf("**%s** has entered the Matrix", event.PlayerName),
			Color:       ColorGreen,
			Timestamp:   event.Timestamp.Format(time.RFC3339),
			Footer:      &DiscordFooter{Text: d.serverName},
		}},
	}
}

func (d *DiscordIntegration) formatPlayerLeave(event *Event) *DiscordMessage {
	return &DiscordMessage{
		Username: d.serverName,
		Embeds: []DiscordEmbed{{
			Title:       "🔴 Player Left",
			Description: fmt.Sprintf("**%s** has disconnected", event.PlayerName),
			Color:       ColorRed,
			Timestamp:   event.Timestamp.Format(time.RFC3339),
			Footer:      &DiscordFooter{Text: d.serverName},
		}},
	}
}

func (d *DiscordIntegration) formatLevelUp(event *Event) *DiscordMessage {
	level := 0
	if l, ok := event.Data["level"].(int); ok {
		level = l
	}
	if l, ok := event.Data["level"].(float64); ok {
		level = int(l)
	}

	return &DiscordMessage{
		Username: d.serverName,
		Embeds: []DiscordEmbed{{
			Title:       "⬆️ Level Up!",
			Description: fmt.Sprintf("**%s** has reached level **%d**!", event.PlayerName, level),
			Color:       ColorGold,
			Timestamp:   event.Timestamp.Format(time.RFC3339),
			Footer:      &DiscordFooter{Text: d.serverName},
		}},
	}
}

func (d *DiscordIntegration) formatDeath(event *Event) *DiscordMessage {
	cause := "unknown causes"
	if c, ok := event.Data["cause"].(string); ok {
		cause = c
	}

	return &DiscordMessage{
		Username: d.serverName,
		Embeds: []DiscordEmbed{{
			Title:       "💀 Player Death",
			Description: fmt.Sprintf("**%s** was killed by %s", event.PlayerName, cause),
			Color:       ColorRed,
			Timestamp:   event.Timestamp.Format(time.RFC3339),
			Footer:      &DiscordFooter{Text: d.serverName},
		}},
	}
}

func (d *DiscordIntegration) formatAchievement(event *Event) *DiscordMessage {
	achievement := "Unknown Achievement"
	if a, ok := event.Data["achievement"].(string); ok {
		achievement = a
	}

	return &DiscordMessage{
		Username: d.serverName,
		Embeds: []DiscordEmbed{{
			Title:       "🏆 Achievement Unlocked!",
			Description: fmt.Sprintf("**%s** earned: **%s**", event.PlayerName, achievement),
			Color:       ColorPurple,
			Timestamp:   event.Timestamp.Format(time.RFC3339),
			Footer:      &DiscordFooter{Text: d.serverName},
		}},
	}
}

func (d *DiscordIntegration) formatNPCKill(event *Event) *DiscordMessage {
	npc := "an enemy"
	if n, ok := event.Data["npc_name"].(string); ok {
		npc = n
	}

	return &DiscordMessage{
		Username: d.serverName,
		Embeds: []DiscordEmbed{{
			Title:       "⚔️ Enemy Defeated",
			Description: fmt.Sprintf("**%s** defeated **%s**", event.PlayerName, npc),
			Color:       ColorMatrix,
			Timestamp:   event.Timestamp.Format(time.RFC3339),
			Footer:      &DiscordFooter{Text: d.serverName},
		}},
	}
}

func (d *DiscordIntegration) formatPvPKill(event *Event) *DiscordMessage {
	victim := "another player"
	if v, ok := event.Data["victim"].(string); ok {
		victim = v
	}

	return &DiscordMessage{
		Username: d.serverName,
		Embeds: []DiscordEmbed{{
			Title:       "🗡️ PvP Kill",
			Description: fmt.Sprintf("**%s** defeated **%s** in combat!", event.PlayerName, victim),
			Color:       ColorOrange,
			Fields: []DiscordField{
				{Name: "Winner", Value: event.PlayerName, Inline: true},
				{Name: "Loser", Value: victim, Inline: true},
			},
			Timestamp: event.Timestamp.Format(time.RFC3339),
			Footer:    &DiscordFooter{Text: d.serverName},
		}},
	}
}

func (d *DiscordIntegration) formatQuestComplete(event *Event) *DiscordMessage {
	quest := "a quest"
	if q, ok := event.Data["quest_name"].(string); ok {
		quest = q
	}

	return &DiscordMessage{
		Username: d.serverName,
		Embeds: []DiscordEmbed{{
			Title:       "📜 Quest Complete!",
			Description: fmt.Sprintf("**%s** completed: **%s**", event.PlayerName, quest),
			Color:       ColorCyan,
			Timestamp:   event.Timestamp.Format(time.RFC3339),
			Footer:      &DiscordFooter{Text: d.serverName},
		}},
	}
}

func (d *DiscordIntegration) formatServerStart(event *Event) *DiscordMessage {
	version := "unknown"
	if v, ok := event.Data["version"].(string); ok {
		version = v
	}

	return &DiscordMessage{
		Username: d.serverName,
		Embeds: []DiscordEmbed{{
			Title:       "🚀 Server Online",
			Description: fmt.Sprintf("**%s** is now online!", d.serverName),
			Color:       ColorGreen,
			Fields: []DiscordField{
				{Name: "Version", Value: version, Inline: true},
			},
			Timestamp: event.Timestamp.Format(time.RFC3339),
			Footer:    &DiscordFooter{Text: d.serverName},
		}},
	}
}

func (d *DiscordIntegration) formatServerStop(event *Event) *DiscordMessage {
	return &DiscordMessage{
		Username: d.serverName,
		Embeds: []DiscordEmbed{{
			Title:       "🛑 Server Offline",
			Description: fmt.Sprintf("**%s** is shutting down", d.serverName),
			Color:       ColorRed,
			Timestamp:   event.Timestamp.Format(time.RFC3339),
			Footer:      &DiscordFooter{Text: d.serverName},
		}},
	}
}

// sendMessage sends a message to the Discord webhook
func (d *DiscordIntegration) sendMessage(message *DiscordMessage) error {
	if d.webhookURL == "" {
		return nil
	}

	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", d.webhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// SendCustomMessage sends a custom message to Discord
func (d *DiscordIntegration) SendCustomMessage(content string) error {
	return d.sendMessage(&DiscordMessage{
		Username: d.serverName,
		Content:  content,
	})
}

// SendCustomEmbed sends a custom embed to Discord
func (d *DiscordIntegration) SendCustomEmbed(embed DiscordEmbed) error {
	embed.Timestamp = time.Now().Format(time.RFC3339)
	if embed.Footer == nil {
		embed.Footer = &DiscordFooter{Text: d.serverName}
	}

	return d.sendMessage(&DiscordMessage{
		Username: d.serverName,
		Embeds:   []DiscordEmbed{embed},
	})
}
