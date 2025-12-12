// Package analytics provides player behavior tracking for Matrix MUD.
package analytics

import (
	"sync"
	"time"
)

// RoomVisit records a room visit event
type RoomVisit struct {
	RoomID    string
	PlayerID  string
	Timestamp time.Time
}

// PlayerSession tracks a player's session
type PlayerSession struct {
	PlayerID   string
	StartTime  time.Time
	EndTime    time.Time
	RoomsVisited []string
	CommandsUsed map[string]int
	NPCsKilled   int
	ItemsLooted  int
}

// Tracker manages analytics data
type Tracker struct {
	mu sync.RWMutex

	// Room visit counts
	RoomVisits map[string]int64

	// Popular routes (from -> to -> count)
	RoomTransitions map[string]map[string]int64

	// Command usage by player
	CommandUsage map[string]int64

	// Active sessions
	Sessions map[string]*PlayerSession

	// Peak concurrent players
	PeakPlayers     int
	CurrentPlayers  int

	// Total unique visitors
	UniqueVisitors map[string]bool
}

// Global tracker instance
var T = &Tracker{
	RoomVisits:      make(map[string]int64),
	RoomTransitions: make(map[string]map[string]int64),
	CommandUsage:    make(map[string]int64),
	Sessions:        make(map[string]*PlayerSession),
	UniqueVisitors:  make(map[string]bool),
}

// RecordRoomVisit tracks a room visit
func RecordRoomVisit(playerID, roomID string) {
	T.mu.Lock()
	defer T.mu.Unlock()

	T.RoomVisits[roomID]++

	// Track in session
	if session, ok := T.Sessions[playerID]; ok {
		// Record transition
		if len(session.RoomsVisited) > 0 {
			fromRoom := session.RoomsVisited[len(session.RoomsVisited)-1]
			if T.RoomTransitions[fromRoom] == nil {
				T.RoomTransitions[fromRoom] = make(map[string]int64)
			}
			T.RoomTransitions[fromRoom][roomID]++
		}
		session.RoomsVisited = append(session.RoomsVisited, roomID)
	}
}

// RecordCommand tracks command usage
func RecordCommand(playerID, command string) {
	T.mu.Lock()
	defer T.mu.Unlock()

	T.CommandUsage[command]++

	if session, ok := T.Sessions[playerID]; ok {
		if session.CommandsUsed == nil {
			session.CommandsUsed = make(map[string]int)
		}
		session.CommandsUsed[command]++
	}
}

// StartSession begins tracking a player session
func StartSession(playerID string) {
	T.mu.Lock()
	defer T.mu.Unlock()

	T.Sessions[playerID] = &PlayerSession{
		PlayerID:     playerID,
		StartTime:    time.Now(),
		RoomsVisited: make([]string, 0),
		CommandsUsed: make(map[string]int),
	}

	T.UniqueVisitors[playerID] = true
	T.CurrentPlayers++

	if T.CurrentPlayers > T.PeakPlayers {
		T.PeakPlayers = T.CurrentPlayers
	}
}

// EndSession ends tracking a player session
func EndSession(playerID string) *PlayerSession {
	T.mu.Lock()
	defer T.mu.Unlock()

	session, ok := T.Sessions[playerID]
	if !ok {
		return nil
	}

	session.EndTime = time.Now()
	delete(T.Sessions, playerID)

	if T.CurrentPlayers > 0 {
		T.CurrentPlayers--
	}

	return session
}

// GetTopRooms returns the most visited rooms
func GetTopRooms(limit int) []struct {
	RoomID string
	Visits int64
} {
	T.mu.RLock()
	defer T.mu.RUnlock()

	type roomCount struct {
		RoomID string
		Visits int64
	}

	rooms := make([]roomCount, 0, len(T.RoomVisits))
	for id, count := range T.RoomVisits {
		rooms = append(rooms, roomCount{id, count})
	}

	// Simple bubble sort for small datasets
	for i := 0; i < len(rooms)-1; i++ {
		for j := 0; j < len(rooms)-i-1; j++ {
			if rooms[j].Visits < rooms[j+1].Visits {
				rooms[j], rooms[j+1] = rooms[j+1], rooms[j]
			}
		}
	}

	if limit > len(rooms) {
		limit = len(rooms)
	}

	result := make([]struct {
		RoomID string
		Visits int64
	}, limit)
	for i := 0; i < limit; i++ {
		result[i].RoomID = rooms[i].RoomID
		result[i].Visits = rooms[i].Visits
	}

	return result
}

// GetTopCommands returns the most used commands
func GetTopCommands(limit int) []struct {
	Command string
	Count   int64
} {
	T.mu.RLock()
	defer T.mu.RUnlock()

	type cmdCount struct {
		Command string
		Count   int64
	}

	cmds := make([]cmdCount, 0, len(T.CommandUsage))
	for cmd, count := range T.CommandUsage {
		cmds = append(cmds, cmdCount{cmd, count})
	}

	// Simple bubble sort
	for i := 0; i < len(cmds)-1; i++ {
		for j := 0; j < len(cmds)-i-1; j++ {
			if cmds[j].Count < cmds[j+1].Count {
				cmds[j], cmds[j+1] = cmds[j+1], cmds[j]
			}
		}
	}

	if limit > len(cmds) {
		limit = len(cmds)
	}

	result := make([]struct {
		Command string
		Count   int64
	}, limit)
	for i := 0; i < limit; i++ {
		result[i].Command = cmds[i].Command
		result[i].Count = cmds[i].Count
	}

	return result
}

// GetStats returns current analytics statistics
func GetStats() map[string]interface{} {
	T.mu.RLock()
	defer T.mu.RUnlock()

	return map[string]interface{}{
		"current_players":  T.CurrentPlayers,
		"peak_players":     T.PeakPlayers,
		"unique_visitors":  len(T.UniqueVisitors),
		"rooms_discovered": len(T.RoomVisits),
		"total_room_visits": func() int64 {
			var total int64
			for _, v := range T.RoomVisits {
				total += v
			}
			return total
		}(),
	}
}
