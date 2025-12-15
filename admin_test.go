package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestCheckAdminAuth verifies admin authentication
func TestCheckAdminAuth(t *testing.T) {
	// Save original config
	origUser := Config.AdminUser
	origPass := Config.AdminPass
	defer func() {
		Config.AdminUser = origUser
		Config.AdminPass = origPass
	}()

	// Set test credentials
	Config.AdminUser = "testadmin"
	Config.AdminPass = "testpass"

	tests := []struct {
		name     string
		user     string
		pass     string
		wantAuth bool
	}{
		{"valid credentials", "testadmin", "testpass", true},
		{"wrong user", "wronguser", "testpass", false},
		{"wrong pass", "testadmin", "wrongpass", false},
		{"empty credentials", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.user != "" || tt.pass != "" {
				req.SetBasicAuth(tt.user, tt.pass)
			}
			w := httptest.NewRecorder()

			got := checkAdminAuth(w, req)
			if got != tt.wantAuth {
				t.Errorf("checkAdminAuth() = %v, want %v", got, tt.wantAuth)
			}

			if !tt.wantAuth && w.Code != http.StatusUnauthorized {
				t.Errorf("Expected 401 Unauthorized, got %d", w.Code)
			}
		})
	}
}

// TestAdminDashboard verifies dashboard rendering
func TestAdminDashboard(t *testing.T) {
	// Save original config
	origUser := Config.AdminUser
	origPass := Config.AdminPass
	defer func() {
		Config.AdminUser = origUser
		Config.AdminPass = origPass
	}()

	Config.AdminUser = "admin"
	Config.AdminPass = "pass"

	// Set up adminWorld
	adminWorld = NewWorld()

	// Test without auth
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	adminDashboard(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 without auth, got %d", w.Code)
	}

	// Test with auth
	req = httptest.NewRequest("GET", "/", nil)
	req.SetBasicAuth("admin", "pass")
	w = httptest.NewRecorder()
	adminDashboard(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 with auth, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "CONSTRUCT MONITOR") {
		t.Error("Dashboard should contain title")
	}
	if !strings.Contains(body, "Connected Signals") {
		t.Error("Dashboard should contain player section")
	}
}

// TestAdminKick verifies kick functionality
func TestAdminKick(t *testing.T) {
	// Save original config
	origUser := Config.AdminUser
	origPass := Config.AdminPass
	defer func() {
		Config.AdminUser = origUser
		Config.AdminPass = origPass
	}()

	Config.AdminUser = "admin"
	Config.AdminPass = "pass"

	// Set up adminWorld
	adminWorld = NewWorld()

	// Test without auth
	req := httptest.NewRequest("GET", "/kick?name=test", nil)
	w := httptest.NewRecorder()
	adminKick(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 without auth, got %d", w.Code)
	}

	// Test with auth but missing name
	req = httptest.NewRequest("GET", "/kick", nil)
	req.SetBasicAuth("admin", "pass")
	w = httptest.NewRecorder()
	adminKick(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for missing name, got %d", w.Code)
	}

	// Test with auth but nonexistent player
	req = httptest.NewRequest("GET", "/kick?name=nonexistent", nil)
	req.SetBasicAuth("admin", "pass")
	w = httptest.NewRecorder()
	adminKick(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for nonexistent player, got %d", w.Code)
	}
}
