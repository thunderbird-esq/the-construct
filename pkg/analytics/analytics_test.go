package analytics

import (
	"testing"
)

func TestStartEndSession(t *testing.T) {
	playerID := "test_player_session"

	StartSession(playerID)

	if T.CurrentPlayers != 1 {
		t.Errorf("CurrentPlayers = %d, want 1", T.CurrentPlayers)
	}
	if !T.UniqueVisitors[playerID] {
		t.Error("Player should be in unique visitors")
	}

	session := EndSession(playerID)
	if session == nil {
		t.Error("EndSession should return session")
	}
	if T.CurrentPlayers != 0 {
		t.Errorf("CurrentPlayers = %d, want 0", T.CurrentPlayers)
	}
}

func TestRecordRoomVisit(t *testing.T) {
	playerID := "test_player_room"
	StartSession(playerID)
	defer EndSession(playerID)

	RecordRoomVisit(playerID, "loading_program")
	RecordRoomVisit(playerID, "dojo")
	RecordRoomVisit(playerID, "loading_program")

	if T.RoomVisits["loading_program"] != 2 {
		t.Errorf("loading_program visits = %d, want 2", T.RoomVisits["loading_program"])
	}
	if T.RoomVisits["dojo"] != 1 {
		t.Errorf("dojo visits = %d, want 1", T.RoomVisits["dojo"])
	}
}

func TestRecordCommand(t *testing.T) {
	playerID := "test_player_cmd"
	StartSession(playerID)
	defer EndSession(playerID)

	RecordCommand(playerID, "look")
	RecordCommand(playerID, "look")
	RecordCommand(playerID, "north")

	if T.CommandUsage["look"] < 2 {
		t.Error("look should be recorded at least twice")
	}
}

func TestGetTopRooms(t *testing.T) {
	// Record some visits
	T.RoomVisits["popular_room"] = 100
	T.RoomVisits["less_popular"] = 50

	topRooms := GetTopRooms(2)
	if len(topRooms) < 1 {
		t.Error("Should return top rooms")
	}
}

func TestGetTopCommands(t *testing.T) {
	T.CommandUsage["frequent_cmd"] = 100

	topCmds := GetTopCommands(5)
	if len(topCmds) < 1 {
		t.Error("Should return top commands")
	}
}

func TestGetStats(t *testing.T) {
	stats := GetStats()

	if _, ok := stats["current_players"]; !ok {
		t.Error("Stats should include current_players")
	}
	if _, ok := stats["peak_players"]; !ok {
		t.Error("Stats should include peak_players")
	}
	if _, ok := stats["unique_visitors"]; !ok {
		t.Error("Stats should include unique_visitors")
	}
}

func TestPeakPlayers(t *testing.T) {
	// Reset
	T.PeakPlayers = 0
	T.CurrentPlayers = 0

	StartSession("p1")
	StartSession("p2")
	StartSession("p3")

	if T.PeakPlayers != 3 {
		t.Errorf("PeakPlayers = %d, want 3", T.PeakPlayers)
	}

	EndSession("p1")
	EndSession("p2")

	// Peak should still be 3
	if T.PeakPlayers != 3 {
		t.Errorf("PeakPlayers after logout = %d, want 3", T.PeakPlayers)
	}

	EndSession("p3")
}

func TestRoomTransitions(t *testing.T) {
	playerID := "test_transitions"
	StartSession(playerID)
	defer EndSession(playerID)

	RecordRoomVisit(playerID, "room_a")
	RecordRoomVisit(playerID, "room_b")
	RecordRoomVisit(playerID, "room_c")

	// Should have recorded transitions
	if T.RoomTransitions["room_a"] == nil {
		t.Error("Should have transition from room_a")
	}
	if T.RoomTransitions["room_a"]["room_b"] != 1 {
		t.Error("Should have transition room_a -> room_b")
	}
}
