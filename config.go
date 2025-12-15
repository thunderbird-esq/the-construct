// Package main provides configuration management for Matrix MUD.
// All sensitive values should be set via environment variables in production.
package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"time"
)

// Game balance constants
const (
	// Version is the single source of truth for the application version
	Version = "1.57.0"

	// MaxInventorySize is the maximum number of items a player can carry
	MaxInventorySize = 20

	// MaxConnections is the maximum number of concurrent player connections
	MaxConnections = 100

	// ConnectionTimeout is how long to wait for initial connection handshake
	ConnectionTimeout = 30 * time.Second

	// IdleTimeout is how long a player can be idle before being disconnected
	IdleTimeout = 30 * time.Minute

	// DefaultNPCHP is the default HP for NPCs if not specified
	DefaultNPCHP = 50

	// DefaultNPCMaxHP is the default MaxHP for NPCs if not specified
	DefaultNPCMaxHP = 50
)

// Config holds all configuration values for the server.
// Values are loaded from environment variables with sensible defaults for development.
var Config = struct {
	// Server ports
	TelnetPort string
	WebPort    string
	AdminPort  string

	// Admin credentials - MUST be set via environment in production
	AdminUser string
	AdminPass string

	// Security settings
	AdminBindAddr  string // Default: localhost only
	AllowedOrigins string // Comma-separated list, or "*" for development

	// Logging settings
	LogLevel  string // debug, info, warn, error
	LogPretty bool   // true for console, false for JSON
}{
	TelnetPort:     getEnv("TELNET_PORT", "2323"),
	WebPort:        getEnv("WEB_PORT", "8080"),
	AdminPort:      getEnv("ADMIN_PORT", "9090"),
	AdminUser:      getEnv("ADMIN_USER", "admin"),
	AdminPass:      getEnvOrGenerate("ADMIN_PASS"),
	AdminBindAddr:  getEnv("ADMIN_BIND_ADDR", "127.0.0.1:9090"),
	AllowedOrigins: getEnv("ALLOWED_ORIGINS", "*"),
	LogLevel:       getEnv("LOG_LEVEL", "info"),
	LogPretty:      getEnv("LOG_PRETTY", "true") == "true",
}

// getEnv retrieves an environment variable or returns the fallback value.
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvOrGenerate retrieves an environment variable or generates a secure random value.
// Used for secrets that must not have predictable defaults.
func getEnvOrGenerate(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	// Generate a secure random password
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatalf("Failed to generate secure random value for %s: %v", key, err)
	}
	generated := hex.EncodeToString(bytes)

	log.Printf("WARNING: %s not set. Generated temporary value: %s", key, generated)
	log.Printf("WARNING: Set %s environment variable for production use!", key)

	return generated
}
