// Package main implements the administrative server for Matrix MUD.
// This file provides monitoring, management, and debugging endpoints with HTTP Basic Auth.
package main

import (
	"fmt"
	"log"
	"net/http"
)

// adminWorld is a global reference to the world state for admin panel access.
// This allows admin handlers to access player data and perform management operations.
var adminWorld *World

// startAdminServer initializes the admin HTTP server.
// By default, binds to 127.0.0.1 (localhost only) for security.
// Set ADMIN_BIND_ADDR environment variable to change (e.g., "0.0.0.0:9090" for public access).
//
// Provides endpoints:
//
//	GET /        - Admin dashboard showing connected players and stats
//	GET /kick    - Forcibly disconnect a player by name
//
// All endpoints require HTTP Basic Auth with credentials from Config.
func startAdminServer(w *World) {
	adminWorld = w

	// Create a private router for the Admin Interface
	mux := http.NewServeMux()
	mux.HandleFunc("/", adminDashboard)
	mux.HandleFunc("/kick", adminKick)

	// Use configured bind address (defaults to localhost only)
	bindAddr := Config.AdminBindAddr
	log.Printf(">>> Admin Panel active on http://%s", bindAddr)

	if bindAddr == "0.0.0.0:9090" || bindAddr == ":9090" {
		log.Printf("WARNING: Admin panel is exposed to all interfaces. Set ADMIN_BIND_ADDR=127.0.0.1:9090 for localhost only.")
	}

	go func() {
		if err := http.ListenAndServe(bindAddr, mux); err != nil {
			log.Printf("Admin server error: %v", err)
		}
	}()
}

// checkAdminAuth validates HTTP Basic Auth credentials against Config values.
// Returns true if authentication succeeds, false otherwise.
// Also handles setting the WWW-Authenticate header on failure.
func checkAdminAuth(w http.ResponseWriter, r *http.Request) bool {
	user, pass, ok := r.BasicAuth()
	if !ok || user != Config.AdminUser || pass != Config.AdminPass {
		w.Header().Set("WWW-Authenticate", `Basic realm="Matrix Construct Admin"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}
	return true
}

// adminDashboard renders the main admin interface showing all connected players.
// Displays player name, current room, HP status, and provides kick buttons.
// Requires HTTP Basic Auth with credentials from environment variables.
func adminDashboard(w http.ResponseWriter, r *http.Request) {
	if !checkAdminAuth(w, r) {
		return
	}

	html := `<html><head><title>Construct Monitor</title>
	<style>
		body { background: #111; color: #0f0; font-family: monospace; padding: 20px; }
		table { border-collapse: collapse; width: 100%; }
		th, td { border: 1px solid #333; padding: 8px; text-align: left; }
		th { background: #222; }
		.btn { background: #300; color: #fff; text-decoration: none; padding: 5px; }
		.warning { color: #ff0; background: #330; padding: 10px; margin-bottom: 10px; }
	</style>
	</head><body>
	<h1>/// CONSTRUCT MONITOR ///</h1>`

	// Show warning if using generated password
	if Config.AdminPass != "" && len(Config.AdminPass) == 32 {
		html += `<div class="warning">⚠️ Using auto-generated admin password. Set ADMIN_PASS environment variable for production.</div>`
	}

	html += `<h3>Connected Signals</h3>
	<table>
		<tr><th>Name</th><th>Room</th><th>HP</th><th>Action</th></tr>`

	adminWorld.mutex.RLock()
	for client, p := range adminWorld.Players {
		html += fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%d/%d</td><td><a href='/kick?name=%s' class='btn'>EJECT</a></td></tr>",
			p.Name, p.RoomID, p.HP, p.MaxHP, p.Name)
		_ = client // unused in loop
	}
	adminWorld.mutex.RUnlock()

	html += `</table></body></html>`
	w.Write([]byte(html))
}

// adminKick forcibly disconnects a player from the server.
// Takes a "name" query parameter to identify the player to kick.
// Sends a warning message to the player before closing their connection.
// Requires HTTP Basic Auth with credentials from environment variables.
func adminKick(w http.ResponseWriter, r *http.Request) {
	if !checkAdminAuth(w, r) {
		return
	}

	targetName := r.URL.Query().Get("name")
	if targetName == "" {
		http.Error(w, "Missing 'name' parameter", http.StatusBadRequest)
		return
	}

	adminWorld.mutex.Lock()
	defer adminWorld.mutex.Unlock()

	for client, p := range adminWorld.Players {
		if p.Name == targetName {
			client.Write("\r\n\033[31m[OPERATOR EJECTION]\033[0m\r\n")
			client.conn.Close()
			delete(adminWorld.Players, client)
			log.Printf("Admin kicked player: %s", targetName)
			fmt.Fprintf(w, "Ejected %s", targetName)
			return
		}
	}
	http.Error(w, fmt.Sprintf("User %s not found", targetName), http.StatusNotFound)
}
