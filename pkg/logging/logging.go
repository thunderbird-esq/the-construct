// Package logging provides structured logging with zerolog for Matrix MUD.
// It supports multiple output formats (JSON for production, console for development)
// and provides convenience functions for logging with context.
package logging

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Logger is the global logger instance
var Logger zerolog.Logger

// Init initializes the global logger with the specified configuration.
// If pretty is true, logs are formatted for human readability.
// If pretty is false, logs are output as JSON for machine parsing.
func Init(pretty bool, level string) {
	var output io.Writer = os.Stdout

	if pretty {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(lvl)
	Logger = zerolog.New(output).With().Timestamp().Caller().Logger()
}

// Info logs an info-level message
func Info() *zerolog.Event {
	return Logger.Info()
}

// Debug logs a debug-level message
func Debug() *zerolog.Event {
	return Logger.Debug()
}

// Warn logs a warning-level message
func Warn() *zerolog.Event {
	return Logger.Warn()
}

// Error logs an error-level message
func Error() *zerolog.Event {
	return Logger.Error()
}

// Fatal logs a fatal-level message and exits
func Fatal() *zerolog.Event {
	return Logger.Fatal()
}

// WithPlayer returns a logger with player context
func WithPlayer(name string) zerolog.Logger {
	return Logger.With().Str("player", name).Logger()
}

// WithRoom returns a logger with room context
func WithRoom(roomID string) zerolog.Logger {
	return Logger.With().Str("room", roomID).Logger()
}

// WithNPC returns a logger with NPC context
func WithNPC(npcID string) zerolog.Logger {
	return Logger.With().Str("npc", npcID).Logger()
}

// WithConnection returns a logger with connection context
func WithConnection(addr string) zerolog.Logger {
	return Logger.With().Str("remote_addr", addr).Logger()
}
