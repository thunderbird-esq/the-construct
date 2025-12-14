package main

import (
	"strings"
	"testing"
)

// TestDiscoverPhone verifies phone booth discovery
func TestDiscoverPhone(t *testing.T) {
	world := NewWorld()
	
	player := &Player{
		Name:             "Operator",
		RoomID:           "loading_program",
		DiscoveredPhones: []string{},
	}
	
	// Find a room with a phone
	var phoneRoom string
	for roomID, room := range world.Rooms {
		if room.HasPhone {
			phoneRoom = roomID
			break
		}
	}
	
	if phoneRoom == "" {
		t.Skip("No phone booth rooms found")
	}
	
	// Discover the phone
	world.DiscoverPhone(player, phoneRoom)
	
	if len(player.DiscoveredPhones) != 1 {
		t.Errorf("Should have 1 discovered phone, got %d", len(player.DiscoveredPhones))
	}
	
	// Discovering again should not add duplicate
	world.DiscoverPhone(player, phoneRoom)
	
	if len(player.DiscoveredPhones) != 1 {
		t.Errorf("Should still have 1 discovered phone (no duplicates), got %d", len(player.DiscoveredPhones))
	}
}

// TestDiscoverPhoneInvalidRoom verifies no discovery for invalid rooms
func TestDiscoverPhoneInvalidRoom(t *testing.T) {
	world := NewWorld()
	
	player := &Player{
		Name:             "Operator",
		DiscoveredPhones: []string{},
	}
	
	// Try to discover non-existent room
	world.DiscoverPhone(player, "nonexistent_room")
	
	if len(player.DiscoveredPhones) != 0 {
		t.Errorf("Should not discover non-existent room, got %d phones", len(player.DiscoveredPhones))
	}
}

// TestDiscoverPhoneNoPhone verifies no discovery for rooms without phones
func TestDiscoverPhoneNoPhone(t *testing.T) {
	world := NewWorld()
	
	player := &Player{
		Name:             "Operator",
		DiscoveredPhones: []string{},
	}
	
	// Find a room without a phone
	var noPhoneRoom string
	for roomID, room := range world.Rooms {
		if !room.HasPhone {
			noPhoneRoom = roomID
			break
		}
	}
	
	if noPhoneRoom == "" {
		t.Skip("All rooms have phones")
	}
	
	world.DiscoverPhone(player, noPhoneRoom)
	
	if len(player.DiscoveredPhones) != 0 {
		t.Errorf("Should not discover room without phone, got %d phones", len(player.DiscoveredPhones))
	}
}

// TestCallPhoneNotAtPhone verifies cannot call from non-phone location
func TestCallPhoneNotAtPhone(t *testing.T) {
	world := NewWorld()
	
	// Find a room without a phone
	var noPhoneRoom string
	for roomID, room := range world.Rooms {
		if !room.HasPhone {
			noPhoneRoom = roomID
			break
		}
	}
	
	if noPhoneRoom == "" {
		t.Skip("All rooms have phones")
	}
	
	player := &Player{
		Name:             "Operator",
		RoomID:           noPhoneRoom,
		DiscoveredPhones: []string{},
	}
	
	result := world.CallPhone(player, "anywhere")
	
	if !strings.Contains(result, "need to be at a phone booth") {
		t.Errorf("Should indicate need to be at phone booth: %s", result)
	}
}

// TestCallPhoneEmpty verifies listing phones when no destination given
func TestCallPhoneEmpty(t *testing.T) {
	world := NewWorld()
	
	// Find a phone room
	var phoneRoom string
	for roomID, room := range world.Rooms {
		if room.HasPhone {
			phoneRoom = roomID
			break
		}
	}
	
	if phoneRoom == "" {
		t.Skip("No phone booth rooms found")
	}
	
	player := &Player{
		Name:             "Operator",
		RoomID:           phoneRoom,
		DiscoveredPhones: []string{phoneRoom},
	}
	
	result := world.CallPhone(player, "")
	
	// Should show phone list
	if !strings.Contains(result, "KNOWN PHONE BOOTHS") && !strings.Contains(result, "haven't discovered") {
		t.Errorf("Should show phone list: %s", result)
	}
}

// TestCallPhoneUnknownDestination verifies error for unknown destination
func TestCallPhoneUnknownDestination(t *testing.T) {
	world := NewWorld()
	
	// Find a phone room
	var phoneRoom string
	for roomID, room := range world.Rooms {
		if room.HasPhone {
			phoneRoom = roomID
			break
		}
	}
	
	if phoneRoom == "" {
		t.Skip("No phone booth rooms found")
	}
	
	player := &Player{
		Name:             "Operator",
		RoomID:           phoneRoom,
		DiscoveredPhones: []string{phoneRoom},
	}
	
	result := world.CallPhone(player, "nonexistent_place_xyz")
	
	if !strings.Contains(result, "Unknown destination") {
		t.Errorf("Should indicate unknown destination: %s", result)
	}
}

// TestCallPhoneSameLocation verifies error when already at destination
func TestCallPhoneSameLocation(t *testing.T) {
	world := NewWorld()
	
	// Find a phone room
	var phoneRoom string
	for roomID, room := range world.Rooms {
		if room.HasPhone {
			phoneRoom = roomID
			break
		}
	}
	
	if phoneRoom == "" {
		t.Skip("No phone booth rooms found")
	}
	
	player := &Player{
		Name:             "Operator",
		RoomID:           phoneRoom,
		DiscoveredPhones: []string{phoneRoom},
	}
	
	result := world.CallPhone(player, phoneRoom)
	
	if !strings.Contains(result, "already here") {
		t.Errorf("Should indicate already at destination: %s", result)
	}
}

// TestCallPhoneSuccess verifies successful teleportation
func TestCallPhoneSuccess(t *testing.T) {
	world := NewWorld()
	
	// Find two phone rooms
	var phoneRooms []string
	for roomID, room := range world.Rooms {
		if room.HasPhone {
			phoneRooms = append(phoneRooms, roomID)
			if len(phoneRooms) >= 2 {
				break
			}
		}
	}
	
	if len(phoneRooms) < 2 {
		t.Skip("Need at least 2 phone booth rooms")
	}
	
	player := &Player{
		Name:             "Operator",
		RoomID:           phoneRooms[0],
		DiscoveredPhones: phoneRooms,
	}
	
	originalRoom := player.RoomID
	result := world.CallPhone(player, phoneRooms[1])
	
	if player.RoomID == originalRoom {
		t.Error("Player should have moved to new room")
	}
	if !strings.Contains(result, "phone rings") {
		t.Errorf("Should have phone ring message: %s", result)
	}
}

// TestListPhonesEmpty verifies message when no phones discovered
func TestListPhonesEmpty(t *testing.T) {
	world := NewWorld()
	
	player := &Player{
		Name:             "Newbie",
		DiscoveredPhones: []string{},
	}
	
	result := world.ListPhones(player)
	
	if !strings.Contains(result, "haven't discovered") {
		t.Errorf("Should indicate no phones discovered: %s", result)
	}
}

// TestListPhonesWithPhones verifies phone listing
func TestListPhonesWithPhones(t *testing.T) {
	world := NewWorld()
	
	// Find phone rooms
	var phoneRooms []string
	for roomID, room := range world.Rooms {
		if room.HasPhone {
			phoneRooms = append(phoneRooms, roomID)
			if len(phoneRooms) >= 2 {
				break
			}
		}
	}
	
	if len(phoneRooms) == 0 {
		t.Skip("No phone booth rooms found")
	}
	
	player := &Player{
		Name:             "Operator",
		RoomID:           phoneRooms[0],
		DiscoveredPhones: phoneRooms,
	}
	
	result := world.ListPhones(player)
	
	if !strings.Contains(result, "KNOWN PHONE BOOTHS") {
		t.Errorf("Should have phone booth header: %s", result)
	}
	if !strings.Contains(result, "call") {
		t.Errorf("Should have usage instructions: %s", result)
	}
}

// TestJackOutAtPhone verifies safe logout at phone
func TestJackOutAtPhone(t *testing.T) {
	world := NewWorld()
	
	// Find a phone room
	var phoneRoom string
	for roomID, room := range world.Rooms {
		if room.HasPhone {
			phoneRoom = roomID
			break
		}
	}
	
	if phoneRoom == "" {
		t.Skip("No phone booth rooms found")
	}
	
	player := &Player{
		Name:   "Operator",
		RoomID: phoneRoom,
	}
	
	result := world.JackOut(player)
	
	if !strings.Contains(result, "phone rings") {
		t.Errorf("Should have phone ring message: %s", result)
	}
	if !strings.Contains(result, "Operator") || !strings.Contains(result, "saved") {
		t.Errorf("Should indicate progress saved: %s", result)
	}
}

// TestJackOutNotAtPhone verifies warning when not at phone
func TestJackOutNotAtPhone(t *testing.T) {
	world := NewWorld()
	
	// Find a room without a phone
	var noPhoneRoom string
	for roomID, room := range world.Rooms {
		if !room.HasPhone {
			noPhoneRoom = roomID
			break
		}
	}
	
	if noPhoneRoom == "" {
		t.Skip("All rooms have phones")
	}
	
	player := &Player{
		Name:   "Operator",
		RoomID: noPhoneRoom,
	}
	
	result := world.JackOut(player)
	
	if !strings.Contains(result, "need to be at a phone booth") {
		t.Errorf("Should indicate need phone booth: %s", result)
	}
}

// TestCheckPhoneDiscovery verifies auto-discovery on room entry
func TestCheckPhoneDiscovery(t *testing.T) {
	world := NewWorld()
	
	// Find a phone room
	var phoneRoom string
	for roomID, room := range world.Rooms {
		if room.HasPhone {
			phoneRoom = roomID
			break
		}
	}
	
	if phoneRoom == "" {
		t.Skip("No phone booth rooms found")
	}
	
	player := &Player{
		Name:             "Explorer",
		RoomID:           phoneRoom,
		DiscoveredPhones: []string{},
	}
	
	world.CheckPhoneDiscovery(player)
	
	if len(player.DiscoveredPhones) != 1 {
		t.Errorf("Should have discovered phone, got %d", len(player.DiscoveredPhones))
	}
}

// TestMinFunction verifies min helper
func TestMinFunction(t *testing.T) {
	if min(5, 10) != 5 {
		t.Error("min(5, 10) should be 5")
	}
	if min(10, 5) != 5 {
		t.Error("min(10, 5) should be 5")
	}
	if min(5, 5) != 5 {
		t.Error("min(5, 5) should be 5")
	}
}
