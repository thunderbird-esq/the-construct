package main

import (
	"fmt"
	"net/http"
)

// Global reference to world for the admin panel
var adminWorld *World

func startAdminServer(w *World) {
	adminWorld = w

	// Create a private router for the Admin Interface
	mux := http.NewServeMux()
	mux.HandleFunc("/", adminDashboard)
	mux.HandleFunc("/kick", adminKick)

	fmt.Println(">>> Admin Panel active on http://0.0.0.0:9090")
	// Pass 'mux' instead of 'nil'
	go http.ListenAndServe("0.0.0.0:9090", mux)
}

func adminDashboard(w http.ResponseWriter, r *http.Request) {
	// Basic Auth (admin / admin)
	user, pass, ok := r.BasicAuth()
	if !ok || user != "admin" || pass != "admin" {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	html := `<html><head><title>Construct Monitor</title>
	<style>
		body { background: #111; color: #0f0; font-family: monospace; padding: 20px; }
		table { border-collapse: collapse; width: 100%; }
		th, td { border: 1px solid #333; padding: 8px; text-align: left; }
		th { background: #222; }
		.btn { background: #300; color: #fff; text-decoration: none; padding: 5px; }
	</style>
	</head><body>
	<h1>/// CONSTRUCT MONITOR ///</h1>
	<h3>Connected Signals</h3>
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

func adminKick(w http.ResponseWriter, r *http.Request) {
	user, pass, ok := r.BasicAuth()
	if !ok || user != "admin" || pass != "admin" {
		return
	}

	targetName := r.URL.Query().Get("name")

	adminWorld.mutex.Lock()
	defer adminWorld.mutex.Unlock()

	for client, p := range adminWorld.Players {
		if p.Name == targetName {
			client.Write("\r\n\033[31m[OPERATOR EJECTION]\033[0m\r\n")
			client.conn.Close()
			delete(adminWorld.Players, client)
			fmt.Fprintf(w, "Ejected %s", targetName)
			return
		}
	}
	fmt.Fprintf(w, "User %s not found", targetName)
}
