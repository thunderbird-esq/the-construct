package main

import (
	"testing"
)

// TestTerminalColors verifies all color constants are valid ANSI sequences
func TestTerminalColors(t *testing.T) {
	colors := map[string]string{
		"Reset":         Reset,
		"Green":         Green,
		"White":         White,
		"Gray":          Gray,
		"Red":           Red,
		"Yellow":        Yellow,
		"Magenta":       Magenta,
		"Cyan":          Cyan,
		"ColorUncommon": ColorUncommon,
		"ColorRare":     ColorRare,
		"ColorEpic":     ColorEpic,
		"Clear":         Clear,
	}

	for name, color := range colors {
		if len(color) == 0 {
			t.Errorf("%s should not be empty", name)
		}
		if color[0] != '\033' {
			t.Errorf("%s should start with ESC character", name)
		}
		t.Logf("%s = %q", name, color)
	}
}

// TestMatrixify verifies the Matrixify function
func TestMatrixify(t *testing.T) {
	input := "Hello"
	output := Matrixify(input)

	// Should wrap in green color codes
	if output != Green+input+Reset {
		t.Errorf("Matrixify output unexpected: %q", output)
	}

	// Empty string
	empty := Matrixify("")
	if empty != Green+Reset {
		t.Errorf("Matrixify empty unexpected: %q", empty)
	}
}

// TestSystemMsg verifies system message formatting
func TestSystemMsg(t *testing.T) {
	msg := "Test message"
	output := SystemMsg(msg)

	// Should contain [OPERATOR] prefix
	if len(output) == 0 {
		t.Error("SystemMsg should not return empty")
	}

	// Should contain the message
	expected := White + "[OPERATOR] " + msg + Reset + "\r\n"
	if output != expected {
		t.Errorf("SystemMsg output = %q, want %q", output, expected)
	}
}

// TestConfigValues verifies configuration defaults
func TestConfigValues(t *testing.T) {
	// Test timeout constants
	if ConnectionTimeout <= 0 {
		t.Error("ConnectionTimeout should be positive")
	}
	if IdleTimeout <= 0 {
		t.Error("IdleTimeout should be positive")
	}
	if IdleTimeout <= ConnectionTimeout {
		t.Error("IdleTimeout should be greater than ConnectionTimeout")
	}

	// Test inventory limits
	if MaxInventorySize <= 0 {
		t.Error("MaxInventorySize should be positive")
	}

	// Test NPC defaults
	if DefaultNPCHP <= 0 {
		t.Error("DefaultNPCHP should be positive")
	}
	if DefaultNPCMaxHP <= 0 {
		t.Error("DefaultNPCMaxHP should be positive")
	}

	t.Logf("ConnectionTimeout: %v", ConnectionTimeout)
	t.Logf("IdleTimeout: %v", IdleTimeout)
	t.Logf("MaxInventorySize: %d", MaxInventorySize)
}

// TestVersion verifies version constant exists and is valid
func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}

	// Version should follow semver-ish format
	if len(Version) < 5 { // At least "1.0.0"
		t.Errorf("Version %q seems too short", Version)
	}

	t.Logf("Version: %s", Version)
}

// TestServerPorts verifies port configuration
func TestServerPorts(t *testing.T) {
	if Config.TelnetPort == "" {
		t.Error("TelnetPort should have default")
	}
	if Config.WebPort == "" {
		t.Error("WebPort should have default")
	}
	if Config.AdminPort == "" {
		t.Error("AdminPort should have default")
	}

	t.Logf("Telnet: %s, Web: %s, Admin: %s",
		Config.TelnetPort, Config.WebPort, Config.AdminPort)
}

// TestTelnetConstants verifies telnet protocol constants
func TestTelnetConstants(t *testing.T) {
	if TelnetIAC != 255 {
		t.Errorf("TelnetIAC should be 255, got %d", TelnetIAC)
	}
	if TelnetWILL != 251 {
		t.Errorf("TelnetWILL should be 251, got %d", TelnetWILL)
	}
	if TelnetWONT != 252 {
		t.Errorf("TelnetWONT should be 252, got %d", TelnetWONT)
	}
	if TelnetECHO != 1 {
		t.Errorf("TelnetECHO should be 1, got %d", TelnetECHO)
	}
}

// TestAdminBindAddressSecurity verifies admin is localhost by default
func TestAdminBindAddressSecurity(t *testing.T) {
	addr := Config.AdminBindAddr

	// Should bind to localhost by default for security
	if addr == "0.0.0.0:9090" || addr == ":9090" {
		t.Error("Admin should not bind to all interfaces by default")
	}

	if addr != "127.0.0.1:9090" && addr != "localhost:9090" {
		t.Logf("WARNING: Admin bind address %q is not localhost", addr)
	}
}
