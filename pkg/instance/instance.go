// Package instance implements dungeon/instance system for The Construct.
// Instances are temporary, isolated areas with specific challenges and rewards.
package instance

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// Difficulty levels for instances
type Difficulty int

const (
	DiffEasy   Difficulty = 1
	DiffNormal Difficulty = 2
	DiffHard   Difficulty = 3
	DiffBoss   Difficulty = 4
)

// RoomTemplate defines a room within an instance
type RoomTemplate struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	NPCIDs      []string `json:"npcs"`      // NPC IDs to spawn
	ItemIDs     []string `json:"items"`     // Item IDs to spawn
	Exits       map[string]string `json:"exits"` // direction -> room template ID
	IsBossRoom  bool     `json:"boss_room"`
	Objective   string   `json:"objective"` // Optional objective text
}

// RewardDef defines instance completion rewards
type RewardDef struct {
	XP     int      `json:"xp"`
	Money  int      `json:"money"`
	Items  []string `json:"items"`
	Title  string   `json:"title"`
}

// Template defines an instance template
type Template struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	Difficulty   Difficulty     `json:"difficulty"`
	MinLevel     int            `json:"min_level"`
	MaxPlayers   int            `json:"max_players"`
	TimeLimit    int            `json:"time_limit"`    // Minutes
	EntryRoom    string         `json:"entry_room"`    // Starting room template ID
	RoomTemplates []RoomTemplate `json:"rooms"`
	Rewards      RewardDef      `json:"rewards"`
}

// InstanceRoom represents a room in an active instance
type InstanceRoom struct {
	TemplateID  string            `json:"template_id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	NPCs        []InstanceNPC     `json:"npcs"`
	Items       []string          `json:"items"`
	Exits       map[string]string `json:"exits"`
	Cleared     bool              `json:"cleared"`
	IsBossRoom  bool              `json:"boss_room"`
}

// InstanceNPC represents an NPC in an instance
type InstanceNPC struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	HP       int    `json:"hp"`
	MaxHP    int    `json:"max_hp"`
	Damage   int    `json:"damage"`
	IsAlive  bool   `json:"alive"`
}

// Instance represents an active instance
type Instance struct {
	ID          string                   `json:"id"`
	TemplateID  string                   `json:"template_id"`
	Name        string                   `json:"name"`
	Owner       string                   `json:"owner"`
	Players     []string                 `json:"players"`
	Rooms       map[string]*InstanceRoom `json:"rooms"`
	CurrentRoom string                   `json:"current_room"`
	State       string                   `json:"state"` // "active", "completed", "failed"
	StartedAt   time.Time                `json:"started_at"`
	ExpiresAt   time.Time                `json:"expires_at"`
	KillCount   int                      `json:"kill_count"`
	TotalNPCs   int                      `json:"total_npcs"`
}

// Manager handles instance operations
type Manager struct {
	mu        sync.RWMutex
	Templates map[string]*Template     // Template ID -> Template
	Instances map[string]*Instance     // Instance ID -> Instance
	PlayerMap map[string]string        // Player name -> Instance ID
	nextID    int
}

// NewManager creates a new instance manager
func NewManager() *Manager {
	m := &Manager{
		Templates: make(map[string]*Template),
		Instances: make(map[string]*Instance),
		PlayerMap: make(map[string]string),
		nextID:    1,
	}
	m.loadTemplates()
	return m
}

// loadTemplates loads instance templates from data/instances.json
func (m *Manager) loadTemplates() {
	data, err := os.ReadFile("data/instances.json")
	if err != nil {
		m.loadDefaultTemplates()
		return
	}

	var templateData struct {
		Templates map[string]*Template `json:"templates"`
	}
	if err := json.Unmarshal(data, &templateData); err != nil {
		m.loadDefaultTemplates()
		return
	}

	m.Templates = templateData.Templates
}

// loadDefaultTemplates creates default instance templates
func (m *Manager) loadDefaultTemplates() {
	m.Templates = map[string]*Template{
		"gov_raid": {
			ID:          "gov_raid",
			Name:        "Government Building Raid",
			Description: "Infiltrate the government building and extract critical data.",
			Difficulty:  DiffNormal,
			MinLevel:    3,
			MaxPlayers:  4,
			TimeLimit:   30,
			EntryRoom:   "lobby",
			RoomTemplates: []RoomTemplate{
				{
					ID:          "lobby",
					Name:        "Government Lobby",
					Description: "A stark white lobby with security checkpoints. Guards patrol the area.",
					NPCIDs:      []string{"security_guard", "security_guard"},
					Exits:       map[string]string{"north": "stairwell"},
					Objective:   "Clear the guards and proceed north.",
				},
				{
					ID:          "stairwell",
					Name:        "Emergency Stairwell",
					Description: "A concrete stairwell echoing with distant footsteps.",
					NPCIDs:      []string{"security_guard", "riot_cop"},
					Exits:       map[string]string{"south": "lobby", "up": "offices"},
					Objective:   "Fight your way up.",
				},
				{
					ID:          "offices",
					Name:        "Office Floor",
					Description: "Rows of cubicles filled with oblivious workers. More guards patrol here.",
					NPCIDs:      []string{"riot_cop", "riot_cop", "security_guard"},
					Exits:       map[string]string{"down": "stairwell", "north": "server_room"},
					Objective:   "Find the server room.",
				},
				{
					ID:          "server_room",
					Name:        "Server Room",
					Description: "Banks of humming servers. An Agent stands guard over critical systems.",
					NPCIDs:      []string{"agent"},
					IsBossRoom:  true,
					Exits:       map[string]string{"south": "offices"},
					Objective:   "Defeat the Agent and extract the data.",
				},
			},
			Rewards: RewardDef{
				XP:    200,
				Money: 150,
				Items: []string{"deck"},
				Title: "Infiltrator",
			},
		},
		"club_depths": {
			ID:          "club_depths",
			Name:        "Club Hel Depths",
			Description: "Descend into the depths beneath Club Hel to free the Keymaker.",
			Difficulty:  DiffHard,
			MinLevel:    5,
			MaxPlayers:  4,
			TimeLimit:   45,
			EntryRoom:   "back_room",
			RoomTemplates: []RoomTemplate{
				{
					ID:          "back_room",
					Name:        "Back Room",
					Description: "Behind the club's VIP area. Exile programs linger in the shadows.",
					NPCIDs:      []string{"exile", "exile"},
					Exits:       map[string]string{"down": "wine_cellar"},
					Objective:   "Find the way down.",
				},
				{
					ID:          "wine_cellar",
					Name:        "Wine Cellar",
					Description: "Dusty bottles line the walls. Something moves in the darkness.",
					NPCIDs:      []string{"exile", "exile", "exile"},
					Exits:       map[string]string{"up": "back_room", "east": "dungeon"},
					Objective:   "Clear the cellar and find the dungeon.",
				},
				{
					ID:          "dungeon",
					Name:        "The Dungeon",
					Description: "Cold stone cells line the corridor. The Keymaker must be here somewhere.",
					NPCIDs:      []string{"exile"},
					ItemIDs:     []string{"health_vial"},
					Exits:       map[string]string{"west": "wine_cellar", "north": "twins_chamber"},
					Objective:   "The Twins guard the way forward.",
				},
				{
					ID:          "twins_chamber",
					Name:        "The Twins' Chamber",
					Description: "A grand chamber where the Twins await. They phase in and out of existence.",
					NPCIDs:      []string{"twin_1", "twin_2"},
					IsBossRoom:  true,
					Exits:       map[string]string{"south": "dungeon"},
					Objective:   "Defeat the Twins to free the Keymaker.",
				},
			},
			Rewards: RewardDef{
				XP:    350,
				Money: 250,
				Items: []string{"code_blade"},
				Title: "Keymaker's Savior",
			},
		},
		"training_gauntlet": {
			ID:          "training_gauntlet",
			Name:        "Training Gauntlet",
			Description: "A series of combat trials to test your skills.",
			Difficulty:  DiffEasy,
			MinLevel:    1,
			MaxPlayers:  1,
			TimeLimit:   15,
			EntryRoom:   "trial_1",
			RoomTemplates: []RoomTemplate{
				{
					ID:          "trial_1",
					Name:        "First Trial",
					Description: "A simple sparring room. Your opponent awaits.",
					NPCIDs:      []string{"training_bot"},
					Exits:       map[string]string{"north": "trial_2"},
					Objective:   "Defeat the training bot.",
				},
				{
					ID:          "trial_2",
					Name:        "Second Trial",
					Description: "The difficulty increases. Two opponents this time.",
					NPCIDs:      []string{"training_bot", "training_bot"},
					Exits:       map[string]string{"south": "trial_1", "north": "trial_3"},
					Objective:   "Defeat both opponents.",
				},
				{
					ID:          "trial_3",
					Name:        "Final Trial",
					Description: "Morpheus himself appears as your final test.",
					NPCIDs:      []string{"morpheus_trial"},
					IsBossRoom:  true,
					Exits:       map[string]string{"south": "trial_2"},
					Objective:   "Prove yourself to Morpheus.",
				},
			},
			Rewards: RewardDef{
				XP:    100,
				Money: 50,
				Items: []string{"katana"},
				Title: "Trained",
			},
		},
	}
}

// CreateInstance creates a new instance from a template
func (m *Manager) CreateInstance(templateID, ownerName string, playerLevel int) (*Instance, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	template, ok := m.Templates[templateID]
	if !ok {
		return nil, "Instance template not found."
	}

	if playerLevel < template.MinLevel {
		return nil, fmt.Sprintf("You must be level %d to enter this instance.", template.MinLevel)
	}

	// Check if player is already in an instance
	name := strings.ToLower(ownerName)
	if _, inInstance := m.PlayerMap[name]; inInstance {
		return nil, "You are already in an instance. Leave first with 'instance leave'."
	}

	// Generate unique instance ID
	instanceID := fmt.Sprintf("%s_%d", templateID, m.nextID)
	m.nextID++

	// Create rooms from template
	rooms := make(map[string]*InstanceRoom)
	totalNPCs := 0
	for _, rt := range template.RoomTemplates {
		room := &InstanceRoom{
			TemplateID:  rt.ID,
			Name:        rt.Name,
			Description: rt.Description,
			NPCs:        make([]InstanceNPC, 0),
			Items:       rt.ItemIDs,
			Exits:       rt.Exits,
			Cleared:     false,
			IsBossRoom:  rt.IsBossRoom,
		}

		// Create NPCs
		for i, npcID := range rt.NPCIDs {
			npc := InstanceNPC{
				ID:      fmt.Sprintf("%s_%d", npcID, i),
				Name:    formatNPCName(npcID),
				HP:      getNPCHP(npcID, template.Difficulty),
				MaxHP:   getNPCHP(npcID, template.Difficulty),
				Damage:  getNPCDamage(npcID, template.Difficulty),
				IsAlive: true,
			}
			room.NPCs = append(room.NPCs, npc)
			totalNPCs++
		}

		rooms[rt.ID] = room
	}

	instance := &Instance{
		ID:          instanceID,
		TemplateID:  templateID,
		Name:        template.Name,
		Owner:       ownerName,
		Players:     []string{ownerName},
		Rooms:       rooms,
		CurrentRoom: template.EntryRoom,
		State:       "active",
		StartedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(time.Duration(template.TimeLimit) * time.Minute),
		KillCount:   0,
		TotalNPCs:   totalNPCs,
	}

	m.Instances[instanceID] = instance
	m.PlayerMap[name] = instanceID

	return instance, ""
}

// GetPlayerInstance returns the instance a player is in
func (m *Manager) GetPlayerInstance(playerName string) *Instance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	name := strings.ToLower(playerName)
	instanceID, ok := m.PlayerMap[name]
	if !ok {
		return nil
	}
	return m.Instances[instanceID]
}

// GetCurrentRoom returns the current room in a player's instance
func (m *Manager) GetCurrentRoom(playerName string) *InstanceRoom {
	inst := m.GetPlayerInstance(playerName)
	if inst == nil {
		return nil
	}
	return inst.Rooms[inst.CurrentRoom]
}

// MoveInInstance moves a player within their instance
func (m *Manager) MoveInInstance(playerName, direction string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := strings.ToLower(playerName)
	instanceID, ok := m.PlayerMap[name]
	if !ok {
		return "You are not in an instance.", false
	}

	instance := m.Instances[instanceID]
	currentRoom := instance.Rooms[instance.CurrentRoom]

	// Check if current room is cleared (all NPCs dead)
	if !currentRoom.Cleared {
		for _, npc := range currentRoom.NPCs {
			if npc.IsAlive {
				return "You must defeat all enemies before moving on.", false
			}
		}
		currentRoom.Cleared = true
	}

	// Check for exit
	nextRoomID, hasExit := currentRoom.Exits[direction]
	if !hasExit {
		return "You cannot go that way.", false
	}

	instance.CurrentRoom = nextRoomID
	newRoom := instance.Rooms[nextRoomID]

	return fmt.Sprintf("You enter %s.\r\n\r\n%s", newRoom.Name, newRoom.Description), true
}

// AttackInInstance handles combat in an instance
func (m *Manager) AttackInInstance(playerName, targetName string, playerDamage int) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := strings.ToLower(playerName)
	instanceID, ok := m.PlayerMap[name]
	if !ok {
		return "You are not in an instance.", false
	}

	instance := m.Instances[instanceID]
	room := instance.Rooms[instance.CurrentRoom]

	// Find target NPC
	targetLower := strings.ToLower(targetName)
	for i := range room.NPCs {
		npc := &room.NPCs[i]
		if !npc.IsAlive {
			continue
		}
		if strings.Contains(strings.ToLower(npc.Name), targetLower) {
			// Deal damage
			npc.HP -= playerDamage
			if npc.HP <= 0 {
				npc.HP = 0
				npc.IsAlive = false
				instance.KillCount++

				// Check if room is now cleared
				allDead := true
				for _, n := range room.NPCs {
					if n.IsAlive {
						allDead = false
						break
					}
				}
				if allDead {
					room.Cleared = true
					if room.IsBossRoom {
						instance.State = "completed"
						return fmt.Sprintf("You defeat %s! The boss falls!\r\n\r\nINSTANCE COMPLETE! Use 'instance rewards' to claim your rewards.", npc.Name), true
					}
					return fmt.Sprintf("You defeat %s! The room is cleared. You may proceed.", npc.Name), true
				}
				return fmt.Sprintf("You defeat %s!", npc.Name), true
			}
			return fmt.Sprintf("You hit %s for %d damage. (%d/%d HP)", npc.Name, playerDamage, npc.HP, npc.MaxHP), true
		}
	}

	return "Target not found.", false
}

// LeaveInstance removes a player from their instance
func (m *Manager) LeaveInstance(playerName string) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := strings.ToLower(playerName)
	instanceID, ok := m.PlayerMap[name]
	if !ok {
		return "You are not in an instance."
	}

	delete(m.PlayerMap, name)

	// Check if instance is empty
	instance := m.Instances[instanceID]
	for i, p := range instance.Players {
		if strings.EqualFold(p, playerName) {
			instance.Players = append(instance.Players[:i], instance.Players[i+1:]...)
			break
		}
	}

	if len(instance.Players) == 0 {
		delete(m.Instances, instanceID)
	}

	return "You leave the instance."
}

// GetRewards returns the rewards for a completed instance
func (m *Manager) GetRewards(playerName string) (*RewardDef, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := strings.ToLower(playerName)
	instanceID, ok := m.PlayerMap[name]
	if !ok {
		return nil, "You are not in an instance."
	}

	instance := m.Instances[instanceID]
	if instance.State != "completed" {
		return nil, "Instance not completed yet."
	}

	template := m.Templates[instance.TemplateID]
	rewards := &template.Rewards

	// Remove player from instance after claiming rewards
	delete(m.PlayerMap, name)
	for i, p := range instance.Players {
		if strings.EqualFold(p, playerName) {
			instance.Players = append(instance.Players[:i], instance.Players[i+1:]...)
			break
		}
	}
	if len(instance.Players) == 0 {
		delete(m.Instances, instanceID)
	}

	return rewards, ""
}

// ListTemplates returns available instance templates
func (m *Manager) ListTemplates(playerLevel int) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString("=== AVAILABLE INSTANCES ===\r\n\r\n")

	for _, t := range m.Templates {
		status := "✓"
		if playerLevel < t.MinLevel {
			status = fmt.Sprintf("(Requires Level %d)", t.MinLevel)
		}
		sb.WriteString(fmt.Sprintf("%s - %s %s\r\n", t.ID, t.Name, status))
		sb.WriteString(fmt.Sprintf("  Difficulty: %s | Time: %d min | Players: 1-%d\r\n",
			difficultyName(t.Difficulty), t.TimeLimit, t.MaxPlayers))
		sb.WriteString(fmt.Sprintf("  %s\r\n\r\n", t.Description))
	}

	return sb.String()
}

// LookInInstance returns a description of the current room
func (m *Manager) LookInInstance(playerName string) string {
	inst := m.GetPlayerInstance(playerName)
	if inst == nil {
		return ""
	}

	room := inst.Rooms[inst.CurrentRoom]
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("=== %s ===\r\n", room.Name))
	sb.WriteString(room.Description + "\r\n\r\n")

	// NPCs
	aliveNPCs := 0
	for _, npc := range room.NPCs {
		if npc.IsAlive {
			sb.WriteString(fmt.Sprintf("  [ENEMY] %s (%d/%d HP)\r\n", npc.Name, npc.HP, npc.MaxHP))
			aliveNPCs++
		}
	}

	if aliveNPCs == 0 && len(room.NPCs) > 0 {
		sb.WriteString("  All enemies defeated!\r\n")
	}

	// Items
	for _, item := range room.Items {
		sb.WriteString(fmt.Sprintf("  [ITEM] %s\r\n", item))
	}

	// Exits
	sb.WriteString("\r\nExits: ")
	exits := make([]string, 0)
	for dir := range room.Exits {
		exits = append(exits, dir)
	}
	if len(exits) == 0 {
		sb.WriteString("none")
	} else {
		sb.WriteString(strings.Join(exits, ", "))
	}
	sb.WriteString("\r\n")

	// Progress
	sb.WriteString(fmt.Sprintf("\r\nProgress: %d/%d enemies defeated\r\n", inst.KillCount, inst.TotalNPCs))

	return sb.String()
}

// Helper functions

func formatNPCName(id string) string {
	parts := strings.Split(id, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}

func getNPCHP(npcID string, diff Difficulty) int {
	base := 20
	switch {
	case strings.Contains(npcID, "agent"):
		base = 80
	case strings.Contains(npcID, "twin"):
		base = 60
	case strings.Contains(npcID, "riot"):
		base = 30
	case strings.Contains(npcID, "guard"):
		base = 25
	case strings.Contains(npcID, "exile"):
		base = 35
	case strings.Contains(npcID, "morpheus"):
		base = 50
	case strings.Contains(npcID, "bot"):
		base = 15
	}
	return base * int(diff)
}

func getNPCDamage(npcID string, diff Difficulty) int {
	base := 5
	switch {
	case strings.Contains(npcID, "agent"):
		base = 15
	case strings.Contains(npcID, "twin"):
		base = 12
	case strings.Contains(npcID, "riot"):
		base = 8
	case strings.Contains(npcID, "guard"):
		base = 6
	case strings.Contains(npcID, "exile"):
		base = 10
	case strings.Contains(npcID, "morpheus"):
		base = 10
	case strings.Contains(npcID, "bot"):
		base = 4
	}
	return base * int(diff) / 2
}

func difficultyName(d Difficulty) string {
	switch d {
	case DiffEasy:
		return "Easy"
	case DiffNormal:
		return "Normal"
	case DiffHard:
		return "Hard"
	case DiffBoss:
		return "Boss"
	default:
		return "Unknown"
	}
}

// IsInInstance checks if a player is in an instance
func (m *Manager) IsInInstance(playerName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.PlayerMap[strings.ToLower(playerName)]
	return ok
}

// Global instance manager
var GlobalInstance = NewManager()
