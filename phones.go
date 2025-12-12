// phones.go - Phone Booth Fast Travel System for The Construct
// Phase 1 Matrix Identity features

package main

import (
	"fmt"
	"strings"
)

// DiscoverPhone adds a phone booth to a player's known locations
func (w *World) DiscoverPhone(p *Player, roomID string) {
	room := w.Rooms[roomID]
	if room == nil || !room.HasPhone {
		return
	}
	
	// Check if already discovered
	for _, id := range p.DiscoveredPhones {
		if id == roomID {
			return
		}
	}
	
	p.DiscoveredPhones = append(p.DiscoveredPhones, roomID)
	
	if p.Conn != nil {
		p.Conn.Write(fmt.Sprintf("\r\n%s[Phone booth discovered: %s]%s\r\n",
			Cyan, room.Description[:min(40, len(room.Description))], Reset))
	}
}

// min returns the minimum of two ints
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CallPhone teleports a player to a discovered phone booth
func (w *World) CallPhone(p *Player, destination string) string {
	// Must be at a phone booth to call
	currentRoom := w.Rooms[p.RoomID]
	if currentRoom == nil || !currentRoom.HasPhone {
		return "You need to be at a phone booth to make a call.\r\n"
	}
	
	if destination == "" {
		return w.ListPhones(p)
	}
	
	destination = strings.ToLower(destination)
	
	// Find matching destination
	var targetRoom *Room
	var targetID string
	
	for _, phoneID := range p.DiscoveredPhones {
		room := w.Rooms[phoneID]
		if room == nil {
			continue
		}
		
		// Match by room ID or description keyword
		if strings.Contains(strings.ToLower(phoneID), destination) ||
		   strings.Contains(strings.ToLower(room.Description), destination) {
			targetRoom = room
			targetID = phoneID
			break
		}
	}
	
	if targetRoom == nil {
		return fmt.Sprintf("Unknown destination '%s'. Type 'phones' to see known locations.\r\n", destination)
	}
	
	if targetID == p.RoomID {
		return "You're already here.\r\n"
	}
	
	// Teleport
	w.mutex.Lock()
	oldRoom := w.Rooms[p.RoomID]
	p.RoomID = targetID
	w.mutex.Unlock()
	
	// Announce departure and arrival
	if oldRoom != nil {
		w.Broadcast(oldRoom.ID, p, fmt.Sprintf("%s%s picks up the phone and disappears.%s\r\n", 
			Cyan, p.Name, Reset))
	}
	w.Broadcast(targetID, p, fmt.Sprintf("%s%s materializes near the phone booth.%s\r\n",
		Cyan, p.Name, Reset))
	
	return fmt.Sprintf("%sThe phone rings. You answer it...%s\r\n\r\n%sYou find yourself elsewhere.%s\r\n\r\n%s",
		Cyan, Reset, White, Reset, w.Look(p, ""))
}

// ListPhones shows all discovered phone booths
func (w *World) ListPhones(p *Player) string {
	if len(p.DiscoveredPhones) == 0 {
		return "You haven't discovered any phone booths yet.\r\nExplore the city to find them.\r\n"
	}
	
	var sb strings.Builder
	sb.WriteString(Cyan + "=== KNOWN PHONE BOOTHS ===" + Reset + "\r\n\r\n")
	
	for i, phoneID := range p.DiscoveredPhones {
		room := w.Rooms[phoneID]
		if room == nil {
			continue
		}
		
		// Mark current location
		marker := "  "
		if phoneID == p.RoomID {
			marker = "> "
		}
		
		// Get a short description
		desc := room.Description
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}
		
		sb.WriteString(fmt.Sprintf("%s%d. %s%s%s\r\n", marker, i+1, White, desc, Reset))
		sb.WriteString(fmt.Sprintf("   ID: %s\r\n", phoneID))
	}
	
	sb.WriteString("\r\nUse 'call <name>' to travel (e.g., 'call subway')\r\n")
	
	return sb.String()
}

// JackOut safely logs the player out at a phone booth
func (w *World) JackOut(p *Player) string {
	room := w.Rooms[p.RoomID]
	if room == nil || !room.HasPhone {
		return "You need to be at a phone booth to jack out safely.\r\n" +
		       "Disconnecting elsewhere may have... consequences.\r\n"
	}
	
	// Save player state
	w.SavePlayer(p)
	
	return fmt.Sprintf("%sThe phone rings. You answer it.%s\r\n\r\n"+
		"%s\"Operator.\" Tank's voice crackles through.%s\r\n"+
		"%s\"Get me out of here.\" You close your eyes...%s\r\n\r\n"+
		"%sYour progress has been saved. Until next time.%s\r\n"+
		"%s[You may now safely disconnect]%s\r\n",
		Cyan, Reset,
		White, Reset,
		Green, Reset,
		Yellow, Reset,
		Cyan, Reset)
}

// CheckPhoneDiscovery is called when a player enters a room to auto-discover phones
func (w *World) CheckPhoneDiscovery(p *Player) {
	room := w.Rooms[p.RoomID]
	if room != nil && room.HasPhone {
		w.DiscoverPhone(p, p.RoomID)
	}
}
