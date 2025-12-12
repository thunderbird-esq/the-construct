// awakening.go - The Awakening System and Agent Hunting for The Construct
// Phase 1 Matrix Identity features

package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// Heat thresholds for Agent spawning
const (
	HeatPerCombat    = 5   // Heat gained per combat action
	HeatPerKill      = 15  // Heat gained per NPC kill
	HeatDecayRate    = 1   // Heat lost per minute
	HeatAgentSpawn   = 50  // Heat level that triggers Agent spawn
	HeatMaximum      = 100 // Maximum heat level
	AgentRespawnTime = 5 * time.Minute
)

// TakePill handles the iconic red pill / blue pill choice
func (w *World) TakePill(p *Player, pillColor string) string {
	pillColor = strings.ToLower(pillColor)
	
	// Check if already awakened
	if p.Awakened {
		return "You have already seen the truth. There is no going back.\r\n"
	}
	
	// Must be in a room with Morpheus or the pills
	room := w.Rooms[p.RoomID]
	if room == nil {
		return "You cannot do that here.\r\n"
	}
	
	// Check for red or blue pill in room or inventory
	hasPill := false
	var pillItem *Item
	
	// Check room items
	for _, item := range room.Items {
		if (pillColor == "red" && item.ID == "red_pill") || 
		   (pillColor == "blue" && item.ID == "blue_pill") {
			hasPill = true
			pillItem = item
			break
		}
	}
	
	// Check inventory
	if !hasPill {
		for _, item := range p.Inventory {
			if (pillColor == "red" && item.ID == "red_pill") ||
			   (pillColor == "blue" && item.ID == "blue_pill") {
				hasPill = true
				pillItem = item
				break
			}
		}
	}
	
	if !hasPill {
		return fmt.Sprintf("You don't see a %s pill here.\r\n", pillColor)
	}
	
	switch pillColor {
	case "red":
		// Remove the pill
		w.removeItemFromRoom(room, pillItem)
		w.removeItemFromInventory(p, pillItem)
		
		p.Awakened = true
		
		// Grant awakening bonuses
		p.MaxHP += 10
		p.HP = p.MaxHP
		p.Strength += 2
		
		return fmt.Sprintf("%s%s%s\r\n\r\n%s%s%s\r\n\r\n%s%s%s\r\n",
			Red, "You swallow the red pill.", Reset,
			White, "The world dissolves around you...", Reset,
			Green, `
    ██╗    ██╗ █████╗ ██╗  ██╗███████╗    ██╗   ██╗██████╗ 
    ██║    ██║██╔══██╗██║ ██╔╝██╔════╝    ██║   ██║██╔══██╗
    ██║ █╗ ██║███████║█████╔╝ █████╗      ██║   ██║██████╔╝
    ██║███╗██║██╔══██║██╔═██╗ ██╔══╝      ██║   ██║██╔═══╝ 
    ╚███╔███╔╝██║  ██║██║  ██╗███████╗    ╚██████╔╝██║     
     ╚══╝╚══╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝     ╚═════╝ ╚═╝     

You can now see the code. The Matrix reveals itself to you.
Your eyes have been opened. (+10 MaxHP, +2 Strength)

Type 'abilities' to see your new awakened powers.`, Reset)
		
	case "blue":
		// Remove the pill
		w.removeItemFromRoom(room, pillItem)
		w.removeItemFromInventory(p, pillItem)
		
		return fmt.Sprintf("%s%s%s\r\n\r\n%s%s%s\r\n",
			Cyan, "You swallow the blue pill.", Reset,
			White, "You wake up in your bed and believe whatever you want to believe.\r\nThe story continues... but you remain blind to the truth.", Reset)
		
	default:
		return "Take what? (red or blue)\r\n"
	}
}

// removeItemFromRoom removes an item from a room's item list
func (w *World) removeItemFromRoom(room *Room, item *Item) {
	for i, roomItem := range room.Items {
		if roomItem == item || roomItem.ID == item.ID {
			room.Items = append(room.Items[:i], room.Items[i+1:]...)
			if room.ItemMap != nil {
				delete(room.ItemMap, item.ID)
			}
			return
		}
	}
}

// removeItemFromInventory removes an item from player inventory
func (w *World) removeItemFromInventory(p *Player, item *Item) {
	for i, invItem := range p.Inventory {
		if invItem == item || invItem.ID == item.ID {
			p.Inventory = append(p.Inventory[:i], p.Inventory[i+1:]...)
			return
		}
	}
}

// ShowAbilities shows awakened abilities or class skills
func (w *World) ShowAbilities(p *Player) string {
	var sb strings.Builder
	
	sb.WriteString(Green + "=== YOUR ABILITIES ===" + Reset + "\r\n\r\n")
	
	// Class abilities
	sb.WriteString(White + "Class: " + p.Class + Reset + "\r\n")
	switch p.Class {
	case "Hacker":
		sb.WriteString("  glitch <target> - Disrupt an enemy's code (5s cooldown)\r\n")
		sb.WriteString("  patch - Heal yourself by rewriting your code (15s cooldown)\r\n")
		sb.WriteString("  overflow <target> - Massive damage attack (30s cooldown)\r\n")
	case "Rebel":
		sb.WriteString("  smash <target> - Powerful physical attack (3s cooldown)\r\n")
		sb.WriteString("  fortify - Increase your defenses (20s cooldown)\r\n")
		sb.WriteString("  rampage - Attack all enemies in room (45s cooldown)\r\n")
	case "Operator":
		sb.WriteString("  strike <target> - Quick precision attack (4s cooldown)\r\n")
		sb.WriteString("  vanish - Become temporarily invisible (30s cooldown)\r\n")
		sb.WriteString("  assassinate <target> - High damage sneak attack (60s cooldown)\r\n")
	}
	
	// Awakened abilities
	if p.Awakened {
		sb.WriteString("\r\n" + Red + "Awakened Powers:" + Reset + "\r\n")
		sb.WriteString("  " + Green + "see_code" + Reset + " - Reveal hidden information about your surroundings\r\n")
		sb.WriteString("  " + Green + "dodge" + Reset + " - Automatically evade some attacks (passive)\r\n")
		sb.WriteString("  " + Green + "focus" + Reset + " - Enter bullet-time, increasing your next attack (30s cooldown)\r\n")
	} else {
		sb.WriteString("\r\n" + Yellow + "Take the red pill to unlock awakened powers..." + Reset + "\r\n")
	}
	
	return sb.String()
}

// SeeCode reveals hidden information for awakened players
func (w *World) SeeCode(p *Player) string {
	if !p.Awakened {
		return "You stare intently but see only the surface of things.\r\n"
	}
	
	room := w.Rooms[p.RoomID]
	if room == nil {
		return "Error: Location not found in the Matrix.\r\n"
	}
	
	var sb strings.Builder
	sb.WriteString(Green + "<<< CODE VISION ACTIVATED >>>" + Reset + "\r\n\r\n")
	
	// Show room data
	sb.WriteString(fmt.Sprintf("Room ID: %s%s%s\r\n", Cyan, room.ID, Reset))
	sb.WriteString(fmt.Sprintf("Exits: %d connections\r\n", len(room.Exits)))
	
	// Reveal hidden items
	if len(room.Items) > 0 {
		sb.WriteString("\r\nItem signatures detected:\r\n")
		for _, item := range room.Items {
			rarity := "common"
			if item.Rarity == 1 {
				rarity = "uncommon"
			} else if item.Rarity == 2 {
				rarity = "rare"
			} else if item.Rarity == 3 {
				rarity = "legendary"
			}
			sb.WriteString(fmt.Sprintf("  - %s [%s] DMG:%d AC:%d\r\n", item.Name, rarity, item.Damage, item.AC))
		}
	}
	
	// Reveal NPC data
	if len(room.NPCs) > 0 {
		sb.WriteString("\r\nEntity analysis:\r\n")
		for _, npc := range room.NPCs {
			threat := "neutral"
			if npc.Aggro {
				threat = Yellow + "hostile" + Reset
			}
			if npc.IsAgent {
				threat = Red + "AGENT" + Reset
			}
			sb.WriteString(fmt.Sprintf("  - %s [HP:%d/%d ATK:%d DEF:%d] <%s>\r\n", 
				npc.Name, npc.HP, npc.MaxHP, npc.Damage, npc.AC, threat))
		}
	}
	
	// Show heat level
	sb.WriteString(fmt.Sprintf("\r\nYour heat signature: %s%d/100%s", Yellow, p.Heat, Reset))
	if p.Heat >= HeatAgentSpawn {
		sb.WriteString(Red + " [AGENTS ALERTED]" + Reset)
	}
	sb.WriteString("\r\n")
	
	return sb.String()
}

// Focus activates bullet-time for the next attack
func (w *World) Focus(p *Player) string {
	if !p.Awakened {
		return "You try to focus but your mind is still clouded.\r\n"
	}
	
	// Set a temporary state that boosts next attack
	p.State = "focused"
	
	return fmt.Sprintf("%s%s%s\r\n%s%s%s\r\n",
		Cyan, "Time seems to slow around you...", Reset,
		White, "Your next attack will deal double damage.", Reset)
}

// AddHeat increases player heat and potentially spawns Agents
func (w *World) AddHeat(p *Player, amount int) {
	if !p.Awakened {
		return // Bluepills don't attract Agent attention
	}
	
	p.Heat += amount
	if p.Heat > HeatMaximum {
		p.Heat = HeatMaximum
	}
	
	// Check for Agent spawn
	if p.Heat >= HeatAgentSpawn {
		w.maybeSpawnAgent(p)
	}
}

// DecayHeat reduces heat over time (called from world update loop)
func (w *World) DecayHeat() {
	for _, p := range w.Players {
		if p != nil && p.Heat > 0 {
			p.Heat -= HeatDecayRate
			if p.Heat < 0 {
				p.Heat = 0
			}
		}
	}
}

// maybeSpawnAgent potentially spawns an Agent to hunt a high-heat player
func (w *World) maybeSpawnAgent(p *Player) {
	// 20% chance per check when heat is high
	if rand.Intn(100) > 20 {
		return
	}
	
	room := w.Rooms[p.RoomID]
	if room == nil {
		return
	}
	
	// Check if there's already an Agent in the room
	for _, npc := range room.NPCs {
		if npc.IsAgent && !npc.IsDead {
			return // Already an Agent here
		}
	}
	
	// Find an adjacent room to spawn the Agent
	var spawnRoom *Room
	for _, exitID := range room.Exits {
		if r := w.Rooms[exitID]; r != nil {
			spawnRoom = r
			break
		}
	}
	
	if spawnRoom == nil {
		spawnRoom = room // Spawn in same room if no exits
	}
	
	// Create the Agent
	agent := &NPC{
		ID:           fmt.Sprintf("agent_%d", time.Now().UnixNano()),
		Name:         "Agent Smith",
		Description:  "A man in a dark suit and sunglasses. His movements are unnaturally precise.",
		RoomID:       spawnRoom.ID,
		HP:           100,
		MaxHP:        100,
		Damage:       15,
		AC:           15,
		XP:           200,
		Aggro:        true,
		IsAgent:      true,
		TargetPlayer: p.Name,
		OriginalRoom: spawnRoom.ID,
	}
	
	spawnRoom.NPCs = append(spawnRoom.NPCs, agent)
	if spawnRoom.NPCMap == nil {
		spawnRoom.NPCMap = make(map[string]*NPC)
	}
	spawnRoom.NPCMap[agent.ID] = agent
	
	// Warn the player
	if p.Conn != nil {
		p.Conn.Write(fmt.Sprintf("\r\n%s>>> You feel a glitch in the Matrix... something is coming. <<<%s\r\n> ",
			Red, Reset))
	}
	
	// Warn players in the spawn room
	w.Broadcast(spawnRoom.ID, nil, fmt.Sprintf("%sAn Agent materializes from the crowd!%s\r\n", Red, Reset))
}

// AgentAI handles Agent pursuit behavior (called from world update loop)
func (w *World) AgentAI() {
	for _, room := range w.Rooms {
		for _, npc := range room.NPCs {
			if !npc.IsAgent || npc.IsDead || npc.TargetPlayer == "" {
				continue
			}
			
			// Find target player
			var target *Player
			for _, p := range w.Players {
				if p != nil && strings.ToLower(p.Name) == strings.ToLower(npc.TargetPlayer) {
					target = p
					break
				}
			}
			
			if target == nil {
				npc.TargetPlayer = "" // Target logged off
				continue
			}
			
			// If in same room, attack
			if npc.RoomID == target.RoomID {
				if npc.State != "combat" {
					npc.State = "combat"
					w.Broadcast(target.RoomID, nil, fmt.Sprintf("%s%s turns to face %s. \"Mr. %s...\"%s\r\n",
						Red, npc.Name, target.Name, target.Name, Reset))
				}
				continue
			}
			
			// Move toward target (pathfinding - simple: check if adjacent)
			currentRoom := w.Rooms[npc.RoomID]
			if currentRoom == nil {
				continue
			}
			
			for dir, exitID := range currentRoom.Exits {
				if exitID == target.RoomID {
					// Move to target's room
					w.moveNPC(npc, currentRoom, w.Rooms[exitID], dir)
					break
				}
			}
		}
	}
}

// moveNPC moves an NPC from one room to another
func (w *World) moveNPC(npc *NPC, from, to *Room, direction string) {
	if from == nil || to == nil {
		return
	}
	
	// Remove from old room
	for i, n := range from.NPCs {
		if n == npc {
			from.NPCs = append(from.NPCs[:i], from.NPCs[i+1:]...)
			break
		}
	}
	if from.NPCMap != nil {
		delete(from.NPCMap, npc.ID)
	}
	
	// Add to new room
	npc.RoomID = to.ID
	to.NPCs = append(to.NPCs, npc)
	if to.NPCMap == nil {
		to.NPCMap = make(map[string]*NPC)
	}
	to.NPCMap[npc.ID] = npc
	
	// Announce movement
	w.Broadcast(from.ID, nil, fmt.Sprintf("%s%s walks %s.%s\r\n", Yellow, npc.Name, direction, Reset))
	w.Broadcast(to.ID, nil, fmt.Sprintf("%s%s arrives.%s\r\n", Yellow, npc.Name, Reset))
}
