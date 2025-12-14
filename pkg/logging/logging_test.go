package logging

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestInit(t *testing.T) {
	// Test pretty mode initialization
	Init(true, "info")
	// Logger is initialized - test that it doesn't panic
	Info().Msg("test info")

	// Test JSON mode initialization
	Init(false, "debug")
	Debug().Msg("test debug")

	// Test invalid level falls back to info
	Init(false, "invalid")
	Info().Msg("test invalid level fallback")
	
	// All initializations should complete without panic
}

func TestLogFunctions(t *testing.T) {
	var buf bytes.Buffer
	Logger = zerolog.New(&buf)

	// Test Info
	Info().Msg("info message")
	if !strings.Contains(buf.String(), "info message") {
		t.Error("Info log should contain message")
	}
	buf.Reset()

	// Test Warn
	Warn().Msg("warn message")
	if !strings.Contains(buf.String(), "warn message") {
		t.Error("Warn log should contain message")
	}
	buf.Reset()

	// Test Error
	Error().Msg("error message")
	if !strings.Contains(buf.String(), "error message") {
		t.Error("Error log should contain message")
	}
	buf.Reset()

	// Test Debug
	Debug().Msg("debug message")
	// Debug level might be filtered
}

func TestWithPlayer(t *testing.T) {
	var buf bytes.Buffer
	Logger = zerolog.New(&buf)

	playerLogger := WithPlayer("TestPlayer")
	playerLogger.Info().Msg("player action")

	output := buf.String()
	if !strings.Contains(output, "TestPlayer") {
		t.Error("WithPlayer should include player name in log")
	}
	if !strings.Contains(output, "player action") {
		t.Error("WithPlayer should include message")
	}
}

func TestWithRoom(t *testing.T) {
	var buf bytes.Buffer
	Logger = zerolog.New(&buf)

	roomLogger := WithRoom("dojo")
	roomLogger.Info().Msg("room event")

	output := buf.String()
	if !strings.Contains(output, "dojo") {
		t.Error("WithRoom should include room ID in log")
	}
}

func TestWithNPC(t *testing.T) {
	var buf bytes.Buffer
	Logger = zerolog.New(&buf)

	npcLogger := WithNPC("agent_smith")
	npcLogger.Info().Msg("npc action")

	output := buf.String()
	if !strings.Contains(output, "agent_smith") {
		t.Error("WithNPC should include NPC ID in log")
	}
}

func TestWithConnection(t *testing.T) {
	var buf bytes.Buffer
	Logger = zerolog.New(&buf)

	connLogger := WithConnection("192.168.1.1:12345")
	connLogger.Info().Msg("connection event")

	output := buf.String()
	if !strings.Contains(output, "192.168.1.1:12345") {
		t.Error("WithConnection should include remote address in log")
	}
}
