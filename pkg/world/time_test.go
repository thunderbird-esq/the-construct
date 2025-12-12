package world

import (
	"testing"
	"time"
)

func TestNewGameClock(t *testing.T) {
	clock := NewGameClock(1 * time.Hour)

	if clock == nil {
		t.Fatal("NewGameClock returned nil")
	}
	if clock.CycleLength != 1*time.Hour {
		t.Errorf("CycleLength = %v, want 1h", clock.CycleLength)
	}
	if clock.StartTime.IsZero() {
		t.Error("StartTime should be set")
	}
}

func TestDefaultClock(t *testing.T) {
	clock := DefaultClock()

	if clock.CycleLength != 1*time.Hour {
		t.Errorf("Default CycleLength = %v, want 1h", clock.CycleLength)
	}
}

func TestCurrentTimeAtDifferentCyclePositions(t *testing.T) {
	tests := []struct {
		position float64
		expected TimeOfDay
	}{
		{0.03, Midnight},
		{0.09, Dawn},
		{0.18, Morning},
		{0.30, Noon},
		{0.43, Afternoon},
		{0.56, Dusk},
		{0.68, Evening},
		{0.80, Night},
		{0.92, Midnight},
	}

	for _, tt := range tests {
		clock := &GameClock{
			CycleLength: 100 * time.Second,
			StartTime:   time.Now().Add(-time.Duration(tt.position*100) * time.Second),
		}

		got := clock.CurrentTime()
		if got != tt.expected {
			t.Errorf("At position %.2f: got %v, want %v", tt.position, got, tt.expected)
		}
	}
}

func TestTimeStringAllPeriods(t *testing.T) {
	// Verify TimeName returns valid values
	clock := &GameClock{CycleLength: time.Hour, StartTime: time.Now()}
	name := clock.TimeName()
	if name == "" {
		t.Error("TimeName should not be empty")
	}
}

func TestTimeNameAllPeriods(t *testing.T) {
	validNames := map[string]bool{
		"Midnight": true, "Dawn": true, "Morning": true, "Noon": true,
		"Afternoon": true, "Dusk": true, "Evening": true, "Night": true,
	}

	clock := NewGameClock(1 * time.Hour)
	name := clock.TimeName()
	if !validNames[name] {
		t.Errorf("TimeName = %q is not valid", name)
	}
}

func TestAmbientDescriptionNotEmpty(t *testing.T) {
	clock := NewGameClock(1 * time.Hour)
	desc := clock.AmbientDescription()
	if desc == "" {
		t.Error("AmbientDescription should not be empty")
	}
}

func TestNPCActivityModifierRange(t *testing.T) {
	clock := NewGameClock(1 * time.Hour)
	modifier := clock.NPCActivityModifier()

	if modifier < 0.5 || modifier > 2.0 {
		t.Errorf("NPCActivityModifier = %f, should be 0.5-2.0", modifier)
	}
}

func TestLightLevelRange(t *testing.T) {
	clock := NewGameClock(1 * time.Hour)
	level := clock.LightLevel()

	if level < 0 || level > 100 {
		t.Errorf("LightLevel = %d, should be 0-100", level)
	}
}

func TestTimeIconNotEmpty(t *testing.T) {
	clock := NewGameClock(1 * time.Hour)
	icon := clock.TimeIcon()
	if icon == "" {
		t.Error("TimeIcon should not be empty")
	}
}

func TestFormatTimeDisplayFormat(t *testing.T) {
	clock := NewGameClock(1 * time.Hour)
	display := clock.FormatTimeDisplay()

	if len(display) < 3 {
		t.Error("FormatTimeDisplay too short")
	}
	if display[0] != '[' {
		t.Error("FormatTimeDisplay should start with [")
	}
	if display[len(display)-1] != ']' {
		t.Error("FormatTimeDisplay should end with ]")
	}
}

func TestTimeOfDayConstants(t *testing.T) {
	// Verify all constants are unique
	times := []TimeOfDay{Dawn, Morning, Noon, Afternoon, Dusk, Evening, Night, Midnight}
	seen := make(map[TimeOfDay]bool)

	for _, tod := range times {
		if seen[tod] {
			t.Errorf("Duplicate TimeOfDay value: %d", tod)
		}
		seen[tod] = true
	}
}

func TestTimeProgressionWraps(t *testing.T) {
	// Create a clock that has gone through multiple cycles
	clock := &GameClock{
		CycleLength: 1 * time.Minute,
		StartTime:   time.Now().Add(-5 * time.Minute), // 5 cycles ago
	}

	// Should still return a valid time
	tod := clock.CurrentTime()
	if tod < Dawn || tod > Midnight {
		t.Errorf("Invalid TimeOfDay after multiple cycles: %d", tod)
	}
}

func TestAllTimePeriodsHaveDescriptions(t *testing.T) {
	// Test each time period has a non-empty description
	cycleLength := 80 * time.Millisecond
	
	// Sample at each 1/8th of the cycle
	for i := 0; i < 8; i++ {
		offset := time.Duration(float64(cycleLength) * (float64(i) + 0.5) / 8)
		clock := &GameClock{
			CycleLength: cycleLength,
			StartTime:   time.Now().Add(-offset),
		}

		if clock.TimeString() == "" {
			t.Errorf("Empty TimeString at offset %v", offset)
		}
		if clock.AmbientDescription() == "" {
			t.Errorf("Empty AmbientDescription at offset %v", offset)
		}
	}
}
