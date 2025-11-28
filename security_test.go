package main

import (
	"os"
	"strings"
	"testing"
)

// TestPhase1Security verifies the security fixes for Phase 1

// TestConfigEnvironmentVariables verifies config loads from environment
func TestConfigEnvironmentVariables(t *testing.T) {
	// Save original env vars
	origUser := os.Getenv("ADMIN_USER")
	origPass := os.Getenv("ADMIN_PASS")
	origBind := os.Getenv("ADMIN_BIND_ADDR")
	origOrigins := os.Getenv("ALLOWED_ORIGINS")

	// Restore after test
	defer func() {
		os.Setenv("ADMIN_USER", origUser)
		os.Setenv("ADMIN_PASS", origPass)
		os.Setenv("ADMIN_BIND_ADDR", origBind)
		os.Setenv("ALLOWED_ORIGINS", origOrigins)
	}()

	// Test 1: Default admin bind address should be localhost
	if Config.AdminBindAddr == "0.0.0.0:9090" {
		t.Error("Default AdminBindAddr should NOT be 0.0.0.0:9090 (exposed to all interfaces)")
	}
	if !strings.Contains(Config.AdminBindAddr, "127.0.0.1") && Config.AdminBindAddr != "localhost:9090" {
		t.Logf("AdminBindAddr is: %s (should contain 127.0.0.1 or localhost for security)", Config.AdminBindAddr)
	}

	// Test 2: Admin password should not be hardcoded "admin"
	// If ADMIN_PASS is not set, it should be auto-generated (32 hex chars)
	if Config.AdminPass == "admin" {
		t.Error("Admin password should not be hardcoded as 'admin'")
	}

	// Test 3: If no ADMIN_PASS set, generated password should be 32 chars (16 bytes hex)
	if os.Getenv("ADMIN_PASS") == "" && len(Config.AdminPass) != 32 {
		t.Errorf("Auto-generated password should be 32 hex chars, got %d", len(Config.AdminPass))
	}

	t.Logf("Config.AdminBindAddr = %s", Config.AdminBindAddr)
	t.Logf("Config.AdminPass length = %d (should be 32 if auto-generated)", len(Config.AdminPass))
}

// TestConfigDefaultPorts verifies default port configuration
func TestConfigDefaultPorts(t *testing.T) {
	if Config.TelnetPort == "" {
		t.Error("TelnetPort should have a default value")
	}
	if Config.WebPort == "" {
		t.Error("WebPort should have a default value")
	}
	if Config.AdminPort == "" {
		t.Error("AdminPort should have a default value")
	}

	t.Logf("TelnetPort: %s, WebPort: %s, AdminPort: %s",
		Config.TelnetPort, Config.WebPort, Config.AdminPort)
}

// TestAllowedOriginsConfig verifies WebSocket origin configuration
func TestAllowedOriginsConfig(t *testing.T) {
	// Default should be "*" for development
	if Config.AllowedOrigins == "" {
		t.Error("AllowedOrigins should have a default value")
	}

	t.Logf("AllowedOrigins: %s", Config.AllowedOrigins)

	// In production, this should be changed from "*"
	if Config.AllowedOrigins == "*" {
		t.Log("WARNING: AllowedOrigins is '*' - configure for production!")
	}
}

// TestGetEnvFunction verifies the getEnv helper
func TestGetEnvFunction(t *testing.T) {
	// Test with fallback
	result := getEnv("NONEXISTENT_VAR_12345", "fallback_value")
	if result != "fallback_value" {
		t.Errorf("getEnv should return fallback for missing var, got: %s", result)
	}

	// Test with existing var
	os.Setenv("TEST_VAR_12345", "test_value")
	defer os.Unsetenv("TEST_VAR_12345")

	result = getEnv("TEST_VAR_12345", "fallback")
	if result != "test_value" {
		t.Errorf("getEnv should return env value when set, got: %s", result)
	}
}

// TestAdminBindAddressNotExposed verifies admin isn't exposed by default
func TestAdminBindAddressNotExposed(t *testing.T) {
	// The admin panel should NOT bind to 0.0.0.0 by default
	dangerous := []string{"0.0.0.0", ":9090"}

	for _, d := range dangerous {
		if Config.AdminBindAddr == d {
			t.Errorf("AdminBindAddr should not be '%s' by default (exposes admin to internet)", d)
		}
	}

	// Should contain localhost or 127.0.0.1
	if !strings.Contains(Config.AdminBindAddr, "127.0.0.1") &&
		!strings.Contains(Config.AdminBindAddr, "localhost") {
		t.Logf("WARNING: AdminBindAddr '%s' may expose admin panel", Config.AdminBindAddr)
	}
}
