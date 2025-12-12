// Package world provides world simulation systems for Matrix MUD.
// This includes the day/night cycle that affects world descriptions
// and creates atmospheric time progression in the game.
package world

import (
	"fmt"
	"time"
)

// TimeOfDay represents the current period of the day
type TimeOfDay int

const (
	Dawn TimeOfDay = iota
	Morning
	Noon
	Afternoon
	Dusk
	Evening
	Night
	Midnight
)

// GameClock manages the in-game time cycle
type GameClock struct {
	// CycleLength is the real-time duration of a full day cycle
	CycleLength time.Duration
	// StartTime is when the clock was started (real time)
	StartTime time.Time
}

// NewGameClock creates a game clock with the specified cycle length
// Default: 1 real hour = 1 game day
func NewGameClock(cycleLength time.Duration) *GameClock {
	return &GameClock{
		CycleLength: cycleLength,
		StartTime:   time.Now(),
	}
}

// DefaultClock returns a clock with 1-hour day cycles
func DefaultClock() *GameClock {
	return NewGameClock(1 * time.Hour)
}

// CurrentTime returns the current time of day
func (c *GameClock) CurrentTime() TimeOfDay {
	elapsed := time.Since(c.StartTime)
	cycleProgress := float64(elapsed.Nanoseconds()%c.CycleLength.Nanoseconds()) / float64(c.CycleLength.Nanoseconds())

	// Divide the day into 8 periods
	switch {
	case cycleProgress < 0.0625: // 0-6.25%
		return Midnight
	case cycleProgress < 0.125: // 6.25-12.5%
		return Dawn
	case cycleProgress < 0.25: // 12.5-25%
		return Morning
	case cycleProgress < 0.375: // 25-37.5%
		return Noon
	case cycleProgress < 0.5: // 37.5-50%
		return Afternoon
	case cycleProgress < 0.625: // 50-62.5%
		return Dusk
	case cycleProgress < 0.75: // 62.5-75%
		return Evening
	case cycleProgress < 0.875: // 75-87.5%
		return Night
	default: // 87.5-100%
		return Midnight
	}
}

// TimeString returns a human-readable time description
func (c *GameClock) TimeString() string {
	switch c.CurrentTime() {
	case Dawn:
		return "The first light of dawn breaks through the digital sky."
	case Morning:
		return "Morning light filters through the simulated atmosphere."
	case Noon:
		return "The sun hangs directly overhead in the artificial sky."
	case Afternoon:
		return "The afternoon sun casts long digital shadows."
	case Dusk:
		return "The sky shifts to deep orange as dusk settles in."
	case Evening:
		return "Evening darkness creeps across the Matrix."
	case Night:
		return "Night has fallen. The city glows with neon light."
	case Midnight:
		return "It is the deepest hour of night. The Matrix hums quietly."
	default:
		return "Time flows strangely here."
	}
}

// TimeName returns the name of the current time period
func (c *GameClock) TimeName() string {
	names := []string{"Midnight", "Dawn", "Morning", "Noon", "Afternoon", "Dusk", "Evening", "Night"}
	return names[c.CurrentTime()]
}

// AmbientDescription returns atmospheric text based on time
func (c *GameClock) AmbientDescription() string {
	switch c.CurrentTime() {
	case Dawn:
		return "A cool digital breeze carries the scent of initialization."
	case Morning:
		return "Early programs begin their routines. The air feels fresh."
	case Noon:
		return "Heat shimmer rises from the pavement. The city bustles."
	case Afternoon:
		return "The golden light makes everything seem slightly unreal."
	case Dusk:
		return "Colors bleed across the horizon. Streetlights flicker on."
	case Evening:
		return "Neon signs illuminate the gathering darkness."
	case Night:
		return "Shadows pool in every corner. The city never truly sleeps."
	case Midnight:
		return "An eerie stillness hangs in the air. Something watches."
	default:
		return ""
	}
}

// NPCActivityModifier returns a modifier for NPC behavior based on time
// Higher values mean more active/dangerous NPCs
func (c *GameClock) NPCActivityModifier() float64 {
	switch c.CurrentTime() {
	case Dawn:
		return 0.7
	case Morning:
		return 0.9
	case Noon:
		return 1.0
	case Afternoon:
		return 1.0
	case Dusk:
		return 1.1
	case Evening:
		return 1.2
	case Night:
		return 1.4
	case Midnight:
		return 1.5
	default:
		return 1.0
	}
}

// LightLevel returns a value 0-100 representing ambient light
func (c *GameClock) LightLevel() int {
	switch c.CurrentTime() {
	case Midnight:
		return 10
	case Dawn:
		return 40
	case Morning:
		return 70
	case Noon:
		return 100
	case Afternoon:
		return 90
	case Dusk:
		return 50
	case Evening:
		return 30
	case Night:
		return 15
	default:
		return 50
	}
}

// TimeIcon returns an ASCII icon for the current time
func (c *GameClock) TimeIcon() string {
	switch c.CurrentTime() {
	case Dawn:
		return "🌅"
	case Morning:
		return "☀️"
	case Noon:
		return "🌞"
	case Afternoon:
		return "🌤️"
	case Dusk:
		return "🌆"
	case Evening:
		return "🌙"
	case Night:
		return "🌃"
	case Midnight:
		return "🌑"
	default:
		return "⏰"
	}
}

// FormatTimeDisplay returns a formatted time display for the UI
func (c *GameClock) FormatTimeDisplay() string {
	return fmt.Sprintf("[%s %s]", c.TimeIcon(), c.TimeName())
}
