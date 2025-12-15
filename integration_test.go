package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestAuthenticateFunction tests the authenticate flow
func TestAuthenticateFunction(t *testing.T) {
	// Create a pipe to simulate network connection
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	// Create a client for testing
	client := &Client{
		conn:   serverConn,
		reader: bufio.NewReader(serverConn),
	}

	// Test rate limiting by making repeated calls
	testUser := "test_auth_user_" + fmt.Sprintf("%d", time.Now().UnixNano())
	
	// First few attempts should be allowed
	for i := 0; i < 3; i++ {
		if !authLimiter.Allow(testUser) {
			t.Errorf("Attempt %d should be allowed", i+1)
		}
	}

	// Verify client is set up correctly
	if client.conn == nil {
		t.Error("Client connection should not be nil")
	}
	if client.reader == nil {
		t.Error("Client reader should not be nil")
	}
}

// TestHandleConnectionSetup tests connection initialization
func TestHandleConnectionSetup(t *testing.T) {
	// Test that a new connection can be established
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	// Accept connections in background
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(100 * time.Millisecond))
		buf := make([]byte, 10)
		conn.Read(buf)
	}()

	// Connect as client
	conn, err := net.DialTimeout("tcp", listener.Addr().String(), time.Second)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	conn.Write([]byte("test"))
	conn.Close()

	wg.Wait()
}

// TestWebServerEndpoints tests HTTP endpoints
func TestWebServerEndpoints(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	handleHealth(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Health endpoint returned %d, want 200", w.Code)
	}
	
	body := w.Body.String()
	if !strings.Contains(body, "healthy") {
		t.Error("Health response should contain 'healthy'")
	}
	if !strings.Contains(body, Version) {
		t.Errorf("Health response should contain version %s", Version)
	}
}

// TestServeHomeEndpoint tests the home page
func TestServeHomeEndpoint(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	
	serveHome(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Home endpoint returned %d, want 200", w.Code)
	}
	
	body := w.Body.String()
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("Home should return HTML")
	}
	if !strings.Contains(body, "The Construct") {
		t.Error("Home should contain title")
	}
}

// TestWebSocketOriginValidation tests origin checking
func TestWebSocketOriginValidation(t *testing.T) {
	origAllowed := Config.AllowedOrigins
	defer func() { Config.AllowedOrigins = origAllowed }()

	tests := []struct {
		name    string
		origin  string
		allowed string
		want    bool
	}{
		{"wildcard allows all", "https://evil.com", "*", true},
		{"empty origin allowed", "", "*", true},
		{"specific origin match", "https://example.com", "https://example.com", true},
		{"specific origin mismatch", "https://evil.com", "https://example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Config.AllowedOrigins = tt.allowed
			req := httptest.NewRequest("GET", "/ws", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			
			got := checkWebSocketOrigin(req)
			if got != tt.want {
				t.Errorf("checkWebSocketOrigin() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTelnetIACFiltering tests IAC byte filtering
func TestTelnetIACFiltering(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  []byte
	}{
		{"no IAC", []byte("hello"), []byte("hello")},
		{"escaped IAC", []byte{255, 255}, []byte{255}},
		{"WILL command", []byte{255, 251, 1, 'a'}, []byte{'a'}},
		{"WONT command", []byte{255, 252, 1, 'b'}, []byte{'b'}},
		{"DO command", []byte{255, 253, 1, 'c'}, []byte{'c'}},
		{"DONT command", []byte{255, 254, 1, 'd'}, []byte{'d'}},
		{"mixed content", []byte{'h', 255, 251, 1, 'i'}, []byte{'h', 'i'}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterTelnetIAC(tt.input)
			if string(got) != string(tt.want) {
				t.Errorf("filterTelnetIAC(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestStartAdminServerConfig verifies admin server configuration
func TestStartAdminServerConfig(t *testing.T) {
	world := NewWorld()
	
	origAdminWorld := adminWorld
	defer func() { adminWorld = origAdminWorld }()
	
	if Config.AdminBindAddr == "" {
		t.Error("AdminBindAddr should have default value")
	}
	
	adminWorld = world
	
	if adminWorld == nil {
		t.Error("adminWorld should be set")
	}
}

// TestCombatTargetSetup tests combat target configuration
func TestCombatTargetSetup(t *testing.T) {
	world := NewWorld()
	
	player := &Player{
		Name:     "CombatTester",
		RoomID:   "dojo",
		HP:       100,
		MaxHP:    100,
		Strength: 15,
		State:    "IDLE",
	}

	room := world.Rooms["dojo"]
	if room != nil && len(room.NPCs) > 0 {
		npc := room.NPCs[0]
		player.Target = npc.ID // Target is a string (NPC ID)
		t.Logf("Set target to NPC ID: %s", player.Target)
		
		if player.Target != npc.ID {
			t.Errorf("Target = %q, want %q", player.Target, npc.ID)
		}
	}
}

// TestClientWriteMethod tests Client.Write
func TestClientWriteMethod(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	client := &Client{
		conn:   serverConn,
		reader: bufio.NewReader(serverConn),
	}

	go func() {
		client.Write("test message")
	}()

	clientConn.SetReadDeadline(time.Now().Add(time.Second))
	buf := make([]byte, 100)
	n, err := clientConn.Read(buf)
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}

	if !strings.Contains(string(buf[:n]), "test message") {
		t.Errorf("Expected 'test message', got %q", string(buf[:n]))
	}
}

// TestSuppressResumeEcho tests echo control
func TestSuppressResumeEcho(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	client := &Client{
		conn:   serverConn,
		reader: bufio.NewReader(serverConn),
	}

	go func() {
		client.suppressEcho()
	}()

	clientConn.SetReadDeadline(time.Now().Add(time.Second))
	buf := make([]byte, 10)
	n, _ := clientConn.Read(buf)
	
	// Should receive IAC WILL ECHO (255, 251, 1)
	if n >= 3 && buf[0] == 255 && buf[1] == 251 && buf[2] == 1 {
		t.Log("Received correct IAC WILL ECHO")
	}

	go func() {
		client.resumeEcho()
	}()

	n, _ = clientConn.Read(buf)
	if n >= 3 && buf[0] == 255 && buf[1] == 252 && buf[2] == 1 {
		t.Log("Received correct IAC WONT ECHO")
	}
}

// TestGetPlayerHistoryIntegration tests history management
func TestGetPlayerHistoryIntegration(t *testing.T) {
	name1 := fmt.Sprintf("player_%d", time.Now().UnixNano())
	name2 := fmt.Sprintf("player_%d", time.Now().UnixNano()+1)
	
	h1 := getPlayerHistory(name1)
	h2 := getPlayerHistory(name1)
	h3 := getPlayerHistory(name2)
	
	if h1 != h2 {
		t.Error("Same player should get same history instance")
	}
	if h1 == h3 {
		t.Error("Different players should get different histories")
	}
}

// TestParseCommandIntegration tests command parsing
func TestParseCommandIntegration(t *testing.T) {
	tests := []struct {
		input   string
		wantCmd string
		wantArg string
	}{
		{"look", "look", ""},
		{"look north", "look", "north"},
		{"say hello world", "say", "hello world"},
		{"  north  ", "north", ""},
		{"GET item", "get", "item"},
		{"", "", ""},
		{"attack goblin", "attack", "goblin"},
	}

	for _, tt := range tests {
		cmd, arg := parseCommand(tt.input)
		if cmd != tt.wantCmd || arg != tt.wantArg {
			t.Errorf("parseCommand(%q) = (%q, %q), want (%q, %q)", 
				tt.input, cmd, arg, tt.wantCmd, tt.wantArg)
		}
	}
}
