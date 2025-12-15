package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestCheckWebSocketOriginWildcard verifies wildcard origin
func TestCheckWebSocketOriginWildcard(t *testing.T) {
	oldAllowed := Config.AllowedOrigins
	Config.AllowedOrigins = "*"
	defer func() { Config.AllowedOrigins = oldAllowed }()

	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Origin", "https://evil.com")

	result := checkWebSocketOrigin(req)

	if !result {
		t.Error("Wildcard should allow any origin")
	}
}

// TestCheckWebSocketOriginEmpty verifies empty origin
func TestCheckWebSocketOriginEmpty(t *testing.T) {
	req := httptest.NewRequest("GET", "/ws", nil)
	// No origin header

	result := checkWebSocketOrigin(req)

	if !result {
		t.Error("Empty origin should be allowed")
	}
}

// TestCheckWebSocketOriginAllowed verifies allowed origin
func TestCheckWebSocketOriginAllowed(t *testing.T) {
	oldAllowed := Config.AllowedOrigins
	Config.AllowedOrigins = "https://trusted.com,https://other.com"
	defer func() { Config.AllowedOrigins = oldAllowed }()

	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Origin", "https://trusted.com")

	result := checkWebSocketOrigin(req)

	if !result {
		t.Error("Allowed origin should pass")
	}
}

// TestCheckWebSocketOriginDenied verifies denied origin
func TestCheckWebSocketOriginDenied(t *testing.T) {
	oldAllowed := Config.AllowedOrigins
	Config.AllowedOrigins = "https://trusted.com"
	defer func() { Config.AllowedOrigins = oldAllowed }()

	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Origin", "https://evil.com")

	result := checkWebSocketOrigin(req)

	if result {
		t.Error("Non-allowed origin should be denied")
	}
}

// TestFilterTelnetIACEmpty verifies empty input
func TestFilterTelnetIACEmpty(t *testing.T) {
	result := filterTelnetIAC([]byte{})

	if len(result) != 0 {
		t.Error("Empty input should return empty output")
	}
}

// TestFilterTelnetIACNoCommands verifies passthrough of normal data
func TestFilterTelnetIACNoCommands(t *testing.T) {
	input := []byte("Hello, World!")
	result := filterTelnetIAC(input)

	if string(result) != "Hello, World!" {
		t.Errorf("Normal data should pass through: got %q", string(result))
	}
}

// TestFilterTelnetIACDoubleIAC verifies escaped IAC
func TestFilterTelnetIACDoubleIAC(t *testing.T) {
	// IAC IAC (255 255) should become single 255
	input := []byte{255, 255}
	result := filterTelnetIAC(input)

	if len(result) != 1 || result[0] != 255 {
		t.Errorf("IAC IAC should become single 255: got %v", result)
	}
}

// TestFilterTelnetIACWill verifies WILL command removal
func TestFilterTelnetIACWill(t *testing.T) {
	// IAC WILL ECHO (255 251 1) should be removed
	input := []byte{65, 255, 251, 1, 66} // A, IAC WILL ECHO, B
	result := filterTelnetIAC(input)

	if string(result) != "AB" {
		t.Errorf("WILL command should be removed: got %q", string(result))
	}
}

// TestFilterTelnetIACWont verifies WONT command removal
func TestFilterTelnetIACWont(t *testing.T) {
	// IAC WONT option (255 252 x) should be removed
	input := []byte{65, 255, 252, 3, 66}
	result := filterTelnetIAC(input)

	if string(result) != "AB" {
		t.Errorf("WONT command should be removed: got %q", string(result))
	}
}

// TestFilterTelnetIACDo verifies DO command removal
func TestFilterTelnetIACDo(t *testing.T) {
	// IAC DO option (255 253 x) should be removed
	input := []byte{65, 255, 253, 24, 66}
	result := filterTelnetIAC(input)

	if string(result) != "AB" {
		t.Errorf("DO command should be removed: got %q", string(result))
	}
}

// TestFilterTelnetIACDont verifies DONT command removal
func TestFilterTelnetIACDont(t *testing.T) {
	// IAC DONT option (255 254 x) should be removed
	input := []byte{65, 255, 254, 24, 66}
	result := filterTelnetIAC(input)

	if string(result) != "AB" {
		t.Errorf("DONT command should be removed: got %q", string(result))
	}
}

// TestFilterTelnetIACSB verifies subnegotiation removal
func TestFilterTelnetIACSB(t *testing.T) {
	// IAC SB ... IAC SE (255 250 ... 255 240) should be removed
	input := []byte{65, 255, 250, 1, 2, 3, 255, 240, 66}
	result := filterTelnetIAC(input)

	if string(result) != "AB" {
		t.Errorf("Subnegotiation should be removed: got %q", string(result))
	}
}

// TestFilterTelnetIACMixed verifies mixed content
func TestFilterTelnetIACMixed(t *testing.T) {
	// Mix of normal text and telnet commands
	input := []byte{72, 105, 255, 251, 1, 33, 255, 255} // "Hi" + IAC WILL 1 + "!" + IAC IAC
	result := filterTelnetIAC(input)

	expected := []byte{72, 105, 33, 255} // "Hi!" + 255
	if string(result) != string(expected) {
		t.Errorf("Mixed content filtering failed: got %v, want %v", result, expected)
	}
}

// TestServeHome verifies HTML client is served
func TestServeHome(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	serveHome(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", contentType)
	}

	body := w.Body.String()
	if !strings.Contains(body, "The Construct") {
		t.Error("Body should contain 'The Construct'")
	}
	if !strings.Contains(body, "xterm") {
		t.Error("Body should contain xterm")
	}
}

// TestHandleHealthEndpoint verifies health check
func TestHandleHealthEndpoint(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handleHealth(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", contentType)
	}

	body := w.Body.String()
	if !strings.Contains(body, "healthy") {
		t.Error("Body should contain 'healthy'")
	}
	if !strings.Contains(body, Version) {
		t.Errorf("Body should contain version %s", Version)
	}
}
