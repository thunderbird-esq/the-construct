// Package metrics provides Prometheus metrics for Matrix MUD.
// It tracks game statistics like player counts, commands executed,
// combat events, and server performance.
package metrics

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Metrics holds all server metrics
type Metrics struct {
	mu sync.RWMutex

	// Connection metrics
	ActiveConnections   int64
	TotalConnections    int64
	ConnectionsAccepted int64
	ConnectionsRejected int64

	// Player metrics
	PlayersOnline    int64
	TotalLogins      int64
	TotalLogouts     int64
	FailedLogins     int64
	NewRegistrations int64

	// Command metrics
	CommandsExecuted int64
	CommandsByType   map[string]int64
	CommandLatencyMs float64
	RateLimitedCmds  int64

	// Combat metrics
	CombatsStarted int64
	CombatsEnded   int64
	NPCsKilled     int64
	PlayerDeaths   int64
	DamageDealt    int64
	DamageReceived int64

	// Economy metrics
	ItemsBought     int64
	ItemsSold       int64
	MoneyCirculated int64

	// World metrics
	RoomsCount int64
	NPCsCount  int64
	ItemsCount int64

	// Server metrics
	StartTime    time.Time
	LastUpdate   time.Time
	UpdateCycles int64
	ErrorCount   int64
}

// Global metrics instance
var M = &Metrics{
	CommandsByType: make(map[string]int64),
	StartTime:      time.Now(),
}

// IncrConnections increments connection counters
func IncrConnections() {
	M.mu.Lock()
	M.ActiveConnections++
	M.TotalConnections++
	M.ConnectionsAccepted++
	M.mu.Unlock()
}

// DecrConnections decrements active connections
func DecrConnections() {
	M.mu.Lock()
	if M.ActiveConnections > 0 {
		M.ActiveConnections--
	}
	M.mu.Unlock()
}

// IncrPlayers increments player count
func IncrPlayers() {
	M.mu.Lock()
	M.PlayersOnline++
	M.TotalLogins++
	M.mu.Unlock()
}

// DecrPlayers decrements player count
func DecrPlayers() {
	M.mu.Lock()
	if M.PlayersOnline > 0 {
		M.PlayersOnline--
	}
	M.TotalLogouts++
	M.mu.Unlock()
}

// RecordCommand records a command execution
func RecordCommand(cmdType string) {
	M.mu.Lock()
	M.CommandsExecuted++
	M.CommandsByType[cmdType]++
	M.mu.Unlock()
}

// RecordRateLimited records a rate-limited command
func RecordRateLimited() {
	M.mu.Lock()
	M.RateLimitedCmds++
	M.mu.Unlock()
}

// RecordCombatStart records combat initiation
func RecordCombatStart() {
	M.mu.Lock()
	M.CombatsStarted++
	M.mu.Unlock()
}

// RecordCombatEnd records combat end
func RecordCombatEnd() {
	M.mu.Lock()
	M.CombatsEnded++
	M.mu.Unlock()
}

// RecordNPCKill records an NPC death
func RecordNPCKill() {
	M.mu.Lock()
	M.NPCsKilled++
	M.mu.Unlock()
}

// RecordPlayerDeath records a player death
func RecordPlayerDeath() {
	M.mu.Lock()
	M.PlayerDeaths++
	M.mu.Unlock()
}

// RecordDamage records damage dealt/received
func RecordDamage(dealt, received int64) {
	M.mu.Lock()
	M.DamageDealt += dealt
	M.DamageReceived += received
	M.mu.Unlock()
}

// RecordPurchase records an item purchase
func RecordPurchase(amount int64) {
	M.mu.Lock()
	M.ItemsBought++
	M.MoneyCirculated += amount
	M.mu.Unlock()
}

// RecordSale records an item sale
func RecordSale(amount int64) {
	M.mu.Lock()
	M.ItemsSold++
	M.MoneyCirculated += amount
	M.mu.Unlock()
}

// RecordFailedLogin records a failed login attempt
func RecordFailedLogin() {
	M.mu.Lock()
	M.FailedLogins++
	M.mu.Unlock()
}

// RecordNewUser records a new user registration
func RecordNewUser() {
	M.mu.Lock()
	M.NewRegistrations++
	M.mu.Unlock()
}

// RecordError records an error
func RecordError() {
	M.mu.Lock()
	M.ErrorCount++
	M.mu.Unlock()
}

// RecordUpdateCycle records a game update cycle
func RecordUpdateCycle() {
	M.mu.Lock()
	M.UpdateCycles++
	M.LastUpdate = time.Now()
	M.mu.Unlock()
}

// SetWorldCounts sets world entity counts
func SetWorldCounts(rooms, npcs, items int64) {
	M.mu.Lock()
	M.RoomsCount = rooms
	M.NPCsCount = npcs
	M.ItemsCount = items
	M.mu.Unlock()
}

// Handler returns an HTTP handler for /metrics endpoint
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		M.mu.RLock()
		defer M.mu.RUnlock()

		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

		// Connection metrics
		fmt.Fprintf(w, "# HELP matrix_connections_active Current active connections\n")
		fmt.Fprintf(w, "# TYPE matrix_connections_active gauge\n")
		fmt.Fprintf(w, "matrix_connections_active %d\n\n", M.ActiveConnections)

		fmt.Fprintf(w, "# HELP matrix_connections_total Total connections since start\n")
		fmt.Fprintf(w, "# TYPE matrix_connections_total counter\n")
		fmt.Fprintf(w, "matrix_connections_total %d\n\n", M.TotalConnections)

		// Player metrics
		fmt.Fprintf(w, "# HELP matrix_players_online Current players online\n")
		fmt.Fprintf(w, "# TYPE matrix_players_online gauge\n")
		fmt.Fprintf(w, "matrix_players_online %d\n\n", M.PlayersOnline)

		fmt.Fprintf(w, "# HELP matrix_logins_total Total login attempts\n")
		fmt.Fprintf(w, "# TYPE matrix_logins_total counter\n")
		fmt.Fprintf(w, "matrix_logins_total{status=\"success\"} %d\n", M.TotalLogins)
		fmt.Fprintf(w, "matrix_logins_total{status=\"failed\"} %d\n\n", M.FailedLogins)

		fmt.Fprintf(w, "# HELP matrix_registrations_total Total new registrations\n")
		fmt.Fprintf(w, "# TYPE matrix_registrations_total counter\n")
		fmt.Fprintf(w, "matrix_registrations_total %d\n\n", M.NewRegistrations)

		// Command metrics
		fmt.Fprintf(w, "# HELP matrix_commands_total Total commands executed\n")
		fmt.Fprintf(w, "# TYPE matrix_commands_total counter\n")
		fmt.Fprintf(w, "matrix_commands_total %d\n\n", M.CommandsExecuted)

		fmt.Fprintf(w, "# HELP matrix_commands_rate_limited Commands rate limited\n")
		fmt.Fprintf(w, "# TYPE matrix_commands_rate_limited counter\n")
		fmt.Fprintf(w, "matrix_commands_rate_limited %d\n\n", M.RateLimitedCmds)

		// Combat metrics
		fmt.Fprintf(w, "# HELP matrix_combat_started Total combats started\n")
		fmt.Fprintf(w, "# TYPE matrix_combat_started counter\n")
		fmt.Fprintf(w, "matrix_combat_started %d\n\n", M.CombatsStarted)

		fmt.Fprintf(w, "# HELP matrix_npcs_killed Total NPCs killed\n")
		fmt.Fprintf(w, "# TYPE matrix_npcs_killed counter\n")
		fmt.Fprintf(w, "matrix_npcs_killed %d\n\n", M.NPCsKilled)

		fmt.Fprintf(w, "# HELP matrix_player_deaths Total player deaths\n")
		fmt.Fprintf(w, "# TYPE matrix_player_deaths counter\n")
		fmt.Fprintf(w, "matrix_player_deaths %d\n\n", M.PlayerDeaths)

		fmt.Fprintf(w, "# HELP matrix_damage_total Total damage dealt/received\n")
		fmt.Fprintf(w, "# TYPE matrix_damage_total counter\n")
		fmt.Fprintf(w, "matrix_damage_total{type=\"dealt\"} %d\n", M.DamageDealt)
		fmt.Fprintf(w, "matrix_damage_total{type=\"received\"} %d\n\n", M.DamageReceived)

		// Economy metrics
		fmt.Fprintf(w, "# HELP matrix_transactions_total Total buy/sell transactions\n")
		fmt.Fprintf(w, "# TYPE matrix_transactions_total counter\n")
		fmt.Fprintf(w, "matrix_transactions_total{type=\"buy\"} %d\n", M.ItemsBought)
		fmt.Fprintf(w, "matrix_transactions_total{type=\"sell\"} %d\n\n", M.ItemsSold)

		fmt.Fprintf(w, "# HELP matrix_money_circulated Total money in transactions\n")
		fmt.Fprintf(w, "# TYPE matrix_money_circulated counter\n")
		fmt.Fprintf(w, "matrix_money_circulated %d\n\n", M.MoneyCirculated)

		// World metrics
		fmt.Fprintf(w, "# HELP matrix_world_entities World entity counts\n")
		fmt.Fprintf(w, "# TYPE matrix_world_entities gauge\n")
		fmt.Fprintf(w, "matrix_world_entities{type=\"rooms\"} %d\n", M.RoomsCount)
		fmt.Fprintf(w, "matrix_world_entities{type=\"npcs\"} %d\n", M.NPCsCount)
		fmt.Fprintf(w, "matrix_world_entities{type=\"items\"} %d\n\n", M.ItemsCount)

		// Server metrics
		fmt.Fprintf(w, "# HELP matrix_uptime_seconds Server uptime in seconds\n")
		fmt.Fprintf(w, "# TYPE matrix_uptime_seconds gauge\n")
		fmt.Fprintf(w, "matrix_uptime_seconds %.0f\n\n", time.Since(M.StartTime).Seconds())

		fmt.Fprintf(w, "# HELP matrix_update_cycles Total game update cycles\n")
		fmt.Fprintf(w, "# TYPE matrix_update_cycles counter\n")
		fmt.Fprintf(w, "matrix_update_cycles %d\n\n", M.UpdateCycles)

		fmt.Fprintf(w, "# HELP matrix_errors_total Total errors\n")
		fmt.Fprintf(w, "# TYPE matrix_errors_total counter\n")
		fmt.Fprintf(w, "matrix_errors_total %d\n", M.ErrorCount)
	})
}
