package admin

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandler(t *testing.T) {
	startTime := time.Now().Add(-time.Hour) // 1 hour ago
	handler := Handler("1.0.0", startTime)

	req := httptest.NewRequest("GET", "/admin/dashboard", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("Status = %d, want 200", resp.StatusCode)
	}

	// Parse response
	var data DashboardData
	if err := json.NewDecoder(w.Body).Decode(&data); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check server info
	if data.Server.Version != "1.0.0" {
		t.Errorf("Version = %q, want 1.0.0", data.Server.Version)
	}
	if data.Server.UptimeSeconds < 3600 {
		t.Errorf("Uptime = %.0f, want >= 3600", data.Server.UptimeSeconds)
	}
	if data.Server.GoVersion == "" {
		t.Error("GoVersion should not be empty")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		contains string
	}{
		{30 * time.Second, "30s"},
		{5 * time.Minute, "5m"},
		{2 * time.Hour, "2h"},
		{26 * time.Hour, "1d"},
	}

	for _, tt := range tests {
		result := formatDuration(tt.duration)
		if result == "" {
			t.Errorf("formatDuration(%v) returned empty", tt.duration)
		}
		t.Logf("formatDuration(%v) = %s", tt.duration, result)
	}
}

func TestDashboardDataStructure(t *testing.T) {
	startTime := time.Now()
	handler := Handler("test", startTime)

	req := httptest.NewRequest("GET", "/admin/dashboard", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	var data DashboardData
	if err := json.NewDecoder(w.Body).Decode(&data); err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	// Verify all sections exist
	if data.Server.Version == "" {
		t.Error("Server section missing version")
	}

	// Players section should have valid values (>= 0)
	if data.Players.Online < 0 {
		t.Error("Players.Online should be >= 0")
	}

	// Analytics section
	if data.Analytics.TopRooms == nil {
		t.Error("Analytics.TopRooms should not be nil")
	}
	if data.Analytics.TopCommands == nil {
		t.Error("Analytics.TopCommands should not be nil")
	}
}
