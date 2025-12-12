// Package admin provides the admin dashboard for Matrix MUD.
package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/yourusername/matrix-mud/pkg/analytics"
	"github.com/yourusername/matrix-mud/pkg/metrics"
)

// DashboardData represents the admin dashboard response
type DashboardData struct {
	Server    ServerInfo    `json:"server"`
	Players   PlayerStats   `json:"players"`
	World     WorldStats    `json:"world"`
	Combat    CombatStats   `json:"combat"`
	Economy   EconomyStats  `json:"economy"`
	Analytics AnalyticsData `json:"analytics"`
}

// ServerInfo contains server status
type ServerInfo struct {
	Version       string  `json:"version"`
	Uptime        string  `json:"uptime"`
	UptimeSeconds float64 `json:"uptime_seconds"`
	GoVersion     string  `json:"go_version"`
	NumGoroutines int     `json:"num_goroutines"`
	MemoryMB      float64 `json:"memory_mb"`
	UpdateCycles  int64   `json:"update_cycles"`
	ErrorCount    int64   `json:"error_count"`
}

// PlayerStats contains player statistics
type PlayerStats struct {
	Online           int64 `json:"online"`
	TotalLogins      int64 `json:"total_logins"`
	FailedLogins     int64 `json:"failed_logins"`
	NewRegistrations int64 `json:"new_registrations"`
	TotalLogouts     int64 `json:"total_logouts"`
}

// WorldStats contains world statistics
type WorldStats struct {
	Rooms int64 `json:"rooms"`
	NPCs  int64 `json:"npcs"`
	Items int64 `json:"items"`
}

// CombatStats contains combat statistics
type CombatStats struct {
	CombatsStarted int64 `json:"combats_started"`
	CombatsEnded   int64 `json:"combats_ended"`
	NPCsKilled     int64 `json:"npcs_killed"`
	PlayerDeaths   int64 `json:"player_deaths"`
	DamageDealt    int64 `json:"damage_dealt"`
	DamageReceived int64 `json:"damage_received"`
}

// EconomyStats contains economy statistics
type EconomyStats struct {
	ItemsBought     int64 `json:"items_bought"`
	ItemsSold       int64 `json:"items_sold"`
	MoneyCirculated int64 `json:"money_circulated"`
}

// AnalyticsData contains analytics data
type AnalyticsData struct {
	UniqueVisitors int         `json:"unique_visitors"`
	PeakPlayers    int         `json:"peak_players"`
	TopRooms       []RoomVisit `json:"top_rooms"`
	TopCommands    []CmdUsage  `json:"top_commands"`
}

// RoomVisit for JSON
type RoomVisit struct {
	RoomID string `json:"room_id"`
	Visits int64  `json:"visits"`
}

// CmdUsage for JSON
type CmdUsage struct {
	Command string `json:"command"`
	Count   int64  `json:"count"`
}

// Handler returns the admin dashboard HTTP handler
func Handler(version string, startTime time.Time) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get memory stats
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		// Build dashboard data
		data := DashboardData{
			Server: ServerInfo{
				Version:       version,
				Uptime:        formatDuration(time.Since(startTime)),
				UptimeSeconds: time.Since(startTime).Seconds(),
				GoVersion:     runtime.Version(),
				NumGoroutines: runtime.NumGoroutine(),
				MemoryMB:      float64(memStats.Alloc) / 1024 / 1024,
				UpdateCycles:  metrics.M.UpdateCycles,
				ErrorCount:    metrics.M.ErrorCount,
			},
			Players: PlayerStats{
				Online:           metrics.M.PlayersOnline,
				TotalLogins:      metrics.M.TotalLogins,
				FailedLogins:     metrics.M.FailedLogins,
				NewRegistrations: metrics.M.NewRegistrations,
				TotalLogouts:     metrics.M.TotalLogouts,
			},
			World: WorldStats{
				Rooms: metrics.M.RoomsCount,
				NPCs:  metrics.M.NPCsCount,
				Items: metrics.M.ItemsCount,
			},
			Combat: CombatStats{
				CombatsStarted: metrics.M.CombatsStarted,
				CombatsEnded:   metrics.M.CombatsEnded,
				NPCsKilled:     metrics.M.NPCsKilled,
				PlayerDeaths:   metrics.M.PlayerDeaths,
				DamageDealt:    metrics.M.DamageDealt,
				DamageReceived: metrics.M.DamageReceived,
			},
			Economy: EconomyStats{
				ItemsBought:     metrics.M.ItemsBought,
				ItemsSold:       metrics.M.ItemsSold,
				MoneyCirculated: metrics.M.MoneyCirculated,
			},
		}

		// Get analytics data
		stats := analytics.GetStats()
		data.Analytics.UniqueVisitors = stats["unique_visitors"].(int)
		data.Analytics.PeakPlayers = stats["peak_players"].(int)

		// Top rooms
		topRooms := analytics.GetTopRooms(10)
		data.Analytics.TopRooms = make([]RoomVisit, len(topRooms))
		for i, r := range topRooms {
			data.Analytics.TopRooms[i] = RoomVisit{RoomID: r.RoomID, Visits: r.Visits}
		}

		// Top commands
		topCmds := analytics.GetTopCommands(10)
		data.Analytics.TopCommands = make([]CmdUsage, len(topCmds))
		for i, c := range topCmds {
			data.Analytics.TopCommands[i] = CmdUsage{Command: c.Command, Count: c.Count}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
