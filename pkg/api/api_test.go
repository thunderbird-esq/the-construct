package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func testConfig() *Config {
	return &Config{
		BindAddr:     ":8081",
		RateLimitRPS: 100,
		CORSOrigins:  []string{"*"},
		APIKeys: map[string]*APIKey{
			"test-key": {
				Key:         "test-key",
				Name:        "Test Key",
				Permissions: []string{"read", "write"},
				CreatedAt:   time.Now(),
				Enabled:     true,
			},
		},
	}
}

func TestNewServer(t *testing.T) {
	s := NewServer(testConfig(), "1.0.0")
	if s == nil {
		t.Fatal("NewServer returned nil")
	}
}

func TestHealthEndpoint(t *testing.T) {
	s := NewServer(testConfig(), "1.0.0")

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want 200", w.Code)
	}

	var resp Response
	json.NewDecoder(w.Body).Decode(&resp)

	if !resp.Success {
		t.Error("Response should be successful")
	}
}

func TestStatusEndpoint(t *testing.T) {
	s := NewServer(testConfig(), "1.0.0")
	s.GetServerStatus = func() *ServerStatus {
		return &ServerStatus{
			Status:        "online",
			Version:       "1.0.0",
			PlayersOnline: 5,
		}
	}

	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want 200", w.Code)
	}
}

func TestPlayersEndpointNoAuth(t *testing.T) {
	s := NewServer(testConfig(), "1.0.0")

	req := httptest.NewRequest("GET", "/api/players", nil)
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want 401", w.Code)
	}
}

func TestPlayersEndpointWithAuth(t *testing.T) {
	s := NewServer(testConfig(), "1.0.0")
	s.GetOnlinePlayers = func() []PlayerInfo {
		return []PlayerInfo{
			{Name: "Neo", Level: 10, Class: "Hacker"},
			{Name: "Trinity", Level: 8, Class: "Runner"},
		}
	}

	req := httptest.NewRequest("GET", "/api/players", nil)
	req.Header.Set("X-API-Key", "test-key")
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want 200", w.Code)
	}

	var resp Response
	json.NewDecoder(w.Body).Decode(&resp)

	if !resp.Success {
		t.Error("Response should be successful")
	}
	if resp.Meta == nil || resp.Meta.Total != 2 {
		t.Error("Should have 2 players")
	}
}

func TestPlayersEndpointQueryAuth(t *testing.T) {
	s := NewServer(testConfig(), "1.0.0")
	s.GetOnlinePlayers = func() []PlayerInfo {
		return []PlayerInfo{}
	}

	req := httptest.NewRequest("GET", "/api/players?api_key=test-key", nil)
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want 200", w.Code)
	}
}

func TestPlayerByNameEndpoint(t *testing.T) {
	s := NewServer(testConfig(), "1.0.0")
	s.GetPlayerByName = func(name string) *PlayerInfo {
		if name == "Neo" {
			return &PlayerInfo{Name: "Neo", Level: 10, Class: "Hacker"}
		}
		return nil
	}

	req := httptest.NewRequest("GET", "/api/players/Neo", nil)
	req.Header.Set("X-API-Key", "test-key")
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want 200", w.Code)
	}
}

func TestPlayerByNameNotFound(t *testing.T) {
	s := NewServer(testConfig(), "1.0.0")
	s.GetPlayerByName = func(name string) *PlayerInfo {
		return nil
	}

	req := httptest.NewRequest("GET", "/api/players/Unknown", nil)
	req.Header.Set("X-API-Key", "test-key")
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want 404", w.Code)
	}
}

func TestPlayerStatsSubresource(t *testing.T) {
	s := NewServer(testConfig(), "1.0.0")
	s.GetPlayerByName = func(name string) *PlayerInfo {
		return &PlayerInfo{Name: "Neo", Level: 10, Class: "Hacker"}
	}

	req := httptest.NewRequest("GET", "/api/players/Neo/stats", nil)
	req.Header.Set("X-API-Key", "test-key")
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want 200", w.Code)
	}
}

func TestRoomsEndpoint(t *testing.T) {
	s := NewServer(testConfig(), "1.0.0")
	s.GetRooms = func() []RoomInfo {
		return []RoomInfo{
			{ID: "dojo", Name: "The Dojo"},
			{ID: "construct", Name: "The Construct"},
		}
	}

	req := httptest.NewRequest("GET", "/api/world/rooms", nil)
	req.Header.Set("X-API-Key", "test-key")
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want 200", w.Code)
	}
}

func TestRoomByIDEndpoint(t *testing.T) {
	s := NewServer(testConfig(), "1.0.0")
	s.GetRoom = func(id string) *RoomInfo {
		if id == "dojo" {
			return &RoomInfo{ID: "dojo", Name: "The Dojo"}
		}
		return nil
	}

	req := httptest.NewRequest("GET", "/api/world/rooms/dojo", nil)
	req.Header.Set("X-API-Key", "test-key")
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want 200", w.Code)
	}
}

func TestNPCsEndpoint(t *testing.T) {
	s := NewServer(testConfig(), "1.0.0")
	s.GetNPCs = func() []NPCInfo {
		return []NPCInfo{
			{ID: "morpheus", Name: "Morpheus", Level: 20},
		}
	}

	req := httptest.NewRequest("GET", "/api/world/npcs", nil)
	req.Header.Set("X-API-Key", "test-key")
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want 200", w.Code)
	}
}

func TestItemsEndpoint(t *testing.T) {
	s := NewServer(testConfig(), "1.0.0")
	s.GetItems = func() []ItemInfo {
		return []ItemInfo{
			{ID: "katana", Name: "Katana", Type: "weapon"},
		}
	}

	req := httptest.NewRequest("GET", "/api/world/items", nil)
	req.Header.Set("X-API-Key", "test-key")
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want 200", w.Code)
	}
}

func TestLeaderboardCategoriesEndpoint(t *testing.T) {
	s := NewServer(testConfig(), "1.0.0")

	req := httptest.NewRequest("GET", "/api/leaderboards", nil)
	req.Header.Set("X-API-Key", "test-key")
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want 200", w.Code)
	}
}

func TestLeaderboardEndpoint(t *testing.T) {
	s := NewServer(testConfig(), "1.0.0")
	s.GetLeaderboard = func(category string, limit int) []LeaderboardEntry {
		return []LeaderboardEntry{
			{Rank: 1, Name: "Neo", Value: 100},
			{Rank: 2, Name: "Trinity", Value: 80},
		}
	}

	req := httptest.NewRequest("GET", "/api/leaderboards/level?limit=10", nil)
	req.Header.Set("X-API-Key", "test-key")
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want 200", w.Code)
	}
}

func TestSendMessageEndpoint(t *testing.T) {
	s := NewServer(testConfig(), "1.0.0")
	var sentTo, sentMsg string
	s.SendMessageToPlayer = func(name, message string) error {
		sentTo = name
		sentMsg = message
		return nil
	}

	body := bytes.NewBufferString(`{"player":"Neo","message":"Hello"}`)
	req := httptest.NewRequest("POST", "/api/messages", body)
	req.Header.Set("X-API-Key", "test-key")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want 200", w.Code)
	}
	if sentTo != "Neo" || sentMsg != "Hello" {
		t.Error("Message not sent correctly")
	}
}

func TestSendMessageInvalidBody(t *testing.T) {
	s := NewServer(testConfig(), "1.0.0")

	body := bytes.NewBufferString(`invalid json`)
	req := httptest.NewRequest("POST", "/api/messages", body)
	req.Header.Set("X-API-Key", "test-key")
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want 400", w.Code)
	}
}

func TestSendMessageMissingFields(t *testing.T) {
	s := NewServer(testConfig(), "1.0.0")

	body := bytes.NewBufferString(`{"player":"Neo"}`)
	req := httptest.NewRequest("POST", "/api/messages", body)
	req.Header.Set("X-API-Key", "test-key")
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want 400", w.Code)
	}
}

func TestInvalidAPIKey(t *testing.T) {
	s := NewServer(testConfig(), "1.0.0")

	req := httptest.NewRequest("GET", "/api/players", nil)
	req.Header.Set("X-API-Key", "invalid-key")
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want 401", w.Code)
	}
}

func TestDisabledAPIKey(t *testing.T) {
	config := testConfig()
	config.APIKeys["test-key"].Enabled = false
	s := NewServer(config, "1.0.0")

	req := httptest.NewRequest("GET", "/api/players", nil)
	req.Header.Set("X-API-Key", "test-key")
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want 401", w.Code)
	}
}

func TestRateLimiting(t *testing.T) {
	config := testConfig()
	config.RateLimitRPS = 2
	s := NewServer(config, "1.0.0")
	s.GetOnlinePlayers = func() []PlayerInfo { return []PlayerInfo{} }

	// Make requests until rate limited
	var lastCode int
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/api/players", nil)
		req.Header.Set("X-API-Key", "test-key")
		w := httptest.NewRecorder()
		s.mux.ServeHTTP(w, req)
		lastCode = w.Code
		if w.Code == http.StatusTooManyRequests {
			break
		}
	}

	if lastCode != http.StatusTooManyRequests {
		t.Error("Should be rate limited")
	}
}

func TestCORSHeaders(t *testing.T) {
	s := NewServer(testConfig(), "1.0.0")

	req := httptest.NewRequest("GET", "/api/health", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("CORS header should be set")
	}
}

func TestCORSPreflight(t *testing.T) {
	s := NewServer(testConfig(), "1.0.0")

	req := httptest.NewRequest("OPTIONS", "/api/players", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("X-API-Key", "test-key")
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want 200 for OPTIONS", w.Code)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	s := NewServer(testConfig(), "1.0.0")

	req := httptest.NewRequest("POST", "/api/players", nil)
	req.Header.Set("X-API-Key", "test-key")
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status = %d, want 405", w.Code)
	}
}

func TestAddRemoveAPIKey(t *testing.T) {
	s := NewServer(testConfig(), "1.0.0")

	newKey := &APIKey{
		Key:       "new-key",
		Name:      "New Key",
		Enabled:   true,
		CreatedAt: time.Now(),
	}

	s.AddAPIKey(newKey)

	// Test with new key
	s.GetOnlinePlayers = func() []PlayerInfo { return []PlayerInfo{} }
	req := httptest.NewRequest("GET", "/api/players", nil)
	req.Header.Set("X-API-Key", "new-key")
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Error("New key should work")
	}

	// Remove key
	s.RemoveAPIKey("new-key")

	req = httptest.NewRequest("GET", "/api/players", nil)
	req.Header.Set("X-API-Key", "new-key")
	w = httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Error("Removed key should not work")
	}
}

func TestListAPIKeys(t *testing.T) {
	s := NewServer(testConfig(), "1.0.0")

	keys := s.ListAPIKeys()
	if len(keys) != 1 {
		t.Errorf("Expected 1 key, got %d", len(keys))
	}
	if !strings.HasSuffix(keys[0].Key, "...") {
		t.Error("Key should be truncated")
	}
}

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(2)

	// First two should pass
	if !rl.Allow("key1") {
		t.Error("First request should be allowed")
	}
	if !rl.Allow("key1") {
		t.Error("Second request should be allowed")
	}

	// Third should be denied
	if rl.Allow("key1") {
		t.Error("Third request should be denied")
	}

	// Different key should pass
	if !rl.Allow("key2") {
		t.Error("Different key should be allowed")
	}
}

func TestResponseHelpers(t *testing.T) {
	// Error response
	err := ErrorResponse("test error", 400)
	if err.Success {
		t.Error("Error response should have Success=false")
	}
	if err.Error != "test error" {
		t.Error("Error message incorrect")
	}

	// Success response
	success := SuccessResponse(map[string]string{"key": "value"})
	if !success.Success {
		t.Error("Success response should have Success=true")
	}

	// Success with meta
	meta := &Meta{Total: 10, Page: 1}
	withMeta := SuccessResponseWithMeta([]string{}, meta)
	if withMeta.Meta == nil || withMeta.Meta.Total != 10 {
		t.Error("Meta should be set")
	}
}
