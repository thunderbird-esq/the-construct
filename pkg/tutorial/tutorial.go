// Package tutorial provides a guided onboarding experience for new players.
// Implements interactive tutorials that teach game mechanics progressively.
package tutorial

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// TutorialStep represents a single step in a tutorial
type TutorialStep struct {
	ID           string                 `json:"id"`
	Title        string                 `json:"title"`
	Instructions string                 `json:"instructions"`
	Hint         string                 `json:"hint"`
	Trigger      TriggerType            `json:"trigger"`
	TriggerData  map[string]interface{} `json:"trigger_data"`
	OnComplete   func(interface{})      `json:"-"`
}

// TriggerType defines what triggers step completion
type TriggerType string

const (
	TriggerCommand   TriggerType = "command"   // Player executes specific command
	TriggerMove      TriggerType = "move"      // Player moves to any/specific room
	TriggerPickup    TriggerType = "pickup"    // Player picks up item
	TriggerEquip     TriggerType = "equip"     // Player equips item
	TriggerAttack    TriggerType = "attack"    // Player attacks
	TriggerKill      TriggerType = "kill"      // Player kills NPC
	TriggerTalk      TriggerType = "talk"      // Player talks to NPC
	TriggerQuest     TriggerType = "quest"     // Player accepts/completes quest
	TriggerManual    TriggerType = "manual"    // Triggered by code
	TriggerAny       TriggerType = "any"       // Any action completes
)

// TutorialRewards defines rewards for completing a tutorial
type TutorialRewards struct {
	XP       int               `json:"xp"`
	Money    int               `json:"money"`
	Items    []string          `json:"items"`
	Title    string            `json:"title"`
	Unlocks  []string          `json:"unlocks"` // Tutorial IDs to unlock
}

// Tutorial represents a complete tutorial sequence
type Tutorial struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Steps        []TutorialStep  `json:"steps"`
	Rewards      TutorialRewards `json:"rewards"`
	Prerequisites []string       `json:"prerequisites"`
	AutoStart    bool            `json:"auto_start"`
	Skippable    bool            `json:"skippable"`
}

// PlayerProgress tracks a player's tutorial progress
type PlayerProgress struct {
	TutorialID    string    `json:"tutorial_id"`
	CurrentStep   int       `json:"current_step"`
	StartedAt     time.Time `json:"started_at"`
	CompletedAt   time.Time `json:"completed_at,omitempty"`
	StepsComplete []bool    `json:"steps_complete"`
}

// Manager manages tutorials and player progress
type Manager struct {
	mu              sync.RWMutex
	tutorials       map[string]*Tutorial
	playerProgress  map[string]map[string]*PlayerProgress // player -> tutorial -> progress
	completedTutorials map[string][]string                 // player -> completed tutorial IDs
	onReward        func(player interface{}, rewards TutorialRewards)
}

// NewManager creates a new tutorial manager
func NewManager() *Manager {
	m := &Manager{
		tutorials:          make(map[string]*Tutorial),
		playerProgress:     make(map[string]map[string]*PlayerProgress),
		completedTutorials: make(map[string][]string),
	}
	m.registerDefaultTutorials()
	return m
}

// SetRewardHandler sets the callback for granting rewards
func (m *Manager) SetRewardHandler(handler func(player interface{}, rewards TutorialRewards)) {
	m.onReward = handler
}

// RegisterTutorial registers a tutorial
func (m *Manager) RegisterTutorial(t *Tutorial) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tutorials[t.ID] = t
}

// GetTutorial returns a tutorial by ID
func (m *Manager) GetTutorial(id string) *Tutorial {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tutorials[id]
}

// ListTutorials returns all available tutorials
func (m *Manager) ListTutorials() []*Tutorial {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make([]*Tutorial, 0, len(m.tutorials))
	for _, t := range m.tutorials {
		result = append(result, t)
	}
	return result
}

// GetAvailableTutorials returns tutorials available to a player
func (m *Manager) GetAvailableTutorials(playerName string) []*Tutorial {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	completed := m.completedTutorials[playerName]
	completedMap := make(map[string]bool)
	for _, id := range completed {
		completedMap[id] = true
	}
	
	result := make([]*Tutorial, 0)
	for _, t := range m.tutorials {
		// Skip if already completed
		if completedMap[t.ID] {
			continue
		}
		
		// Check prerequisites
		prereqsMet := true
		for _, prereq := range t.Prerequisites {
			if !completedMap[prereq] {
				prereqsMet = false
				break
			}
		}
		
		if prereqsMet {
			result = append(result, t)
		}
	}
	return result
}

// StartTutorial begins a tutorial for a player
func (m *Manager) StartTutorial(playerName, tutorialID string) (*Tutorial, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	tutorial, ok := m.tutorials[tutorialID]
	if !ok {
		return nil, fmt.Errorf("tutorial not found: %s", tutorialID)
	}
	
	// Check prerequisites
	completed := m.completedTutorials[playerName]
	completedMap := make(map[string]bool)
	for _, id := range completed {
		completedMap[id] = true
	}
	
	for _, prereq := range tutorial.Prerequisites {
		if !completedMap[prereq] {
			return nil, fmt.Errorf("prerequisite not met: %s", prereq)
		}
	}
	
	// Initialize progress
	if m.playerProgress[playerName] == nil {
		m.playerProgress[playerName] = make(map[string]*PlayerProgress)
	}
	
	m.playerProgress[playerName][tutorialID] = &PlayerProgress{
		TutorialID:    tutorialID,
		CurrentStep:   0,
		StartedAt:     time.Now(),
		StepsComplete: make([]bool, len(tutorial.Steps)),
	}
	
	return tutorial, nil
}

// GetCurrentStep returns the current tutorial step for a player
func (m *Manager) GetCurrentStep(playerName string) (*Tutorial, *TutorialStep, int) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	progress := m.playerProgress[playerName]
	if progress == nil {
		return nil, nil, -1
	}
	
	// Find active tutorial
	for tutorialID, p := range progress {
		if p.CompletedAt.IsZero() && p.CurrentStep < len(m.tutorials[tutorialID].Steps) {
			tutorial := m.tutorials[tutorialID]
			return tutorial, &tutorial.Steps[p.CurrentStep], p.CurrentStep
		}
	}
	
	return nil, nil, -1
}

// CheckProgress checks if a trigger completes the current step
func (m *Manager) CheckProgress(playerName string, trigger TriggerType, data map[string]interface{}) (stepCompleted bool, tutorialCompleted bool, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	progress := m.playerProgress[playerName]
	if progress == nil {
		return false, false, ""
	}
	
	for tutorialID, p := range progress {
		if !p.CompletedAt.IsZero() {
			continue
		}
		
		tutorial := m.tutorials[tutorialID]
		if p.CurrentStep >= len(tutorial.Steps) {
			continue
		}
		
		step := &tutorial.Steps[p.CurrentStep]
		
		// Check if trigger matches
		if step.Trigger != trigger && step.Trigger != TriggerAny {
			continue
		}
		
		// Check trigger data if specified
		if !m.matchesTriggerData(step.TriggerData, data) {
			continue
		}
		
		// Step completed!
		p.StepsComplete[p.CurrentStep] = true
		p.CurrentStep++
		
		if step.OnComplete != nil {
			step.OnComplete(data)
		}
		
		// Check if tutorial is complete
		if p.CurrentStep >= len(tutorial.Steps) {
			p.CompletedAt = time.Now()
			m.completedTutorials[playerName] = append(m.completedTutorials[playerName], tutorialID)
			
			return true, true, fmt.Sprintf("Tutorial complete: %s!", tutorial.Name)
		}
		
		nextStep := &tutorial.Steps[p.CurrentStep]
		return true, false, fmt.Sprintf("Step complete! Next: %s", nextStep.Title)
	}
	
	return false, false, ""
}

// matchesTriggerData checks if action data matches step requirements
func (m *Manager) matchesTriggerData(required, actual map[string]interface{}) bool {
	if required == nil || len(required) == 0 {
		return true
	}
	
	for key, reqVal := range required {
		actVal, ok := actual[key]
		if !ok {
			return false
		}
		
		// Handle string comparison case-insensitively
		reqStr, reqIsStr := reqVal.(string)
		actStr, actIsStr := actVal.(string)
		if reqIsStr && actIsStr {
			if !strings.EqualFold(reqStr, actStr) {
				return false
			}
			continue
		}
		
		if reqVal != actVal {
			return false
		}
	}
	
	return true
}

// SkipStep skips the current tutorial step
func (m *Manager) SkipStep(playerName string) (bool, string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	progress := m.playerProgress[playerName]
	if progress == nil {
		return false, "No active tutorial"
	}
	
	for tutorialID, p := range progress {
		if !p.CompletedAt.IsZero() {
			continue
		}
		
		tutorial := m.tutorials[tutorialID]
		if p.CurrentStep >= len(tutorial.Steps) {
			continue
		}
		
		// Skip current step
		p.StepsComplete[p.CurrentStep] = true
		p.CurrentStep++
		
		if p.CurrentStep >= len(tutorial.Steps) {
			p.CompletedAt = time.Now()
			m.completedTutorials[playerName] = append(m.completedTutorials[playerName], tutorialID)
			return true, "Tutorial complete (skipped remaining steps)"
		}
		
		nextStep := &tutorial.Steps[p.CurrentStep]
		return true, fmt.Sprintf("Skipped to: %s", nextStep.Title)
	}
	
	return false, "No active tutorial step"
}

// SkipTutorial skips an entire tutorial
func (m *Manager) SkipTutorial(playerName, tutorialID string) (bool, string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	tutorial, ok := m.tutorials[tutorialID]
	if !ok {
		return false, "Tutorial not found"
	}
	
	if !tutorial.Skippable {
		return false, "This tutorial cannot be skipped"
	}
	
	// Check if in progress
	if progress, ok := m.playerProgress[playerName]; ok {
		if p, ok := progress[tutorialID]; ok && p.CompletedAt.IsZero() {
			p.CompletedAt = time.Now()
			m.completedTutorials[playerName] = append(m.completedTutorials[playerName], tutorialID)
			return true, fmt.Sprintf("Skipped tutorial: %s", tutorial.Name)
		}
	}
	
	return false, "Tutorial not in progress"
}

// GetProgress returns a player's tutorial progress
func (m *Manager) GetProgress(playerName string) map[string]*PlayerProgress {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make(map[string]*PlayerProgress)
	if progress, ok := m.playerProgress[playerName]; ok {
		for k, v := range progress {
			result[k] = v
		}
	}
	return result
}

// GetCompletedTutorials returns completed tutorial IDs for a player
func (m *Manager) GetCompletedTutorials(playerName string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if completed, ok := m.completedTutorials[playerName]; ok {
		result := make([]string, len(completed))
		copy(result, completed)
		return result
	}
	return []string{}
}

// HasCompletedTutorial checks if a player completed a specific tutorial
func (m *Manager) HasCompletedTutorial(playerName, tutorialID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if completed, ok := m.completedTutorials[playerName]; ok {
		for _, id := range completed {
			if id == tutorialID {
				return true
			}
		}
	}
	return false
}

// GetHint returns a hint for the current step
func (m *Manager) GetHint(playerName string) string {
	tutorial, step, _ := m.GetCurrentStep(playerName)
	if step == nil {
		return "No active tutorial. Type 'tutorial list' to see available tutorials."
	}
	
	if step.Hint != "" {
		return step.Hint
	}
	
	return fmt.Sprintf("[%s] %s", tutorial.Name, step.Instructions)
}

// FormatStepDisplay formats a tutorial step for display
func (m *Manager) FormatStepDisplay(playerName string) string {
	tutorial, step, stepNum := m.GetCurrentStep(playerName)
	if step == nil {
		return ""
	}
	
	var sb strings.Builder
	sb.WriteString("\n\033[1;36m╔══════════════════════════════════════╗\033[0m\n")
	sb.WriteString(fmt.Sprintf("\033[1;36m║\033[0m \033[1;33mTutorial:\033[0m %-27s \033[1;36m║\033[0m\n", tutorial.Name))
	sb.WriteString(fmt.Sprintf("\033[1;36m║\033[0m \033[1;32mStep %d/%d:\033[0m %-25s \033[1;36m║\033[0m\n", 
		stepNum+1, len(tutorial.Steps), step.Title))
	sb.WriteString("\033[1;36m╠══════════════════════════════════════╣\033[0m\n")
	
	// Word wrap instructions
	lines := wrapText(step.Instructions, 36)
	for _, line := range lines {
		sb.WriteString(fmt.Sprintf("\033[1;36m║\033[0m %-36s \033[1;36m║\033[0m\n", line))
	}
	
	sb.WriteString("\033[1;36m╚══════════════════════════════════════╝\033[0m\n")
	
	return sb.String()
}

// wrapText wraps text to a maximum width
func wrapText(text string, maxWidth int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}
	
	var lines []string
	var currentLine strings.Builder
	
	for _, word := range words {
		if currentLine.Len() == 0 {
			currentLine.WriteString(word)
		} else if currentLine.Len()+1+len(word) <= maxWidth {
			currentLine.WriteString(" ")
			currentLine.WriteString(word)
		} else {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentLine.WriteString(word)
		}
	}
	
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}
	
	return lines
}

// registerDefaultTutorials creates the built-in tutorials
func (m *Manager) registerDefaultTutorials() {
	// New Player Tutorial
	m.tutorials["new_player"] = &Tutorial{
		ID:          "new_player",
		Name:        "Welcome to the Matrix",
		Description: "Learn the basics of navigating and surviving in the Matrix.",
		AutoStart:   true,
		Skippable:   true,
		Steps: []TutorialStep{
			{
				ID:           "welcome",
				Title:        "Wake Up, Neo",
				Instructions: "You've just been freed from the Matrix. Type 'look' to observe your surroundings.",
				Hint:         "Type: look",
				Trigger:      TriggerCommand,
				TriggerData:  map[string]interface{}{"command": "look"},
			},
			{
				ID:           "movement",
				Title:        "Learn to Move",
				Instructions: "Use directional commands to move: north, south, east, west (or n, s, e, w for short).",
				Hint:         "Try typing: north (or just 'n')",
				Trigger:      TriggerMove,
			},
			{
				ID:           "inventory",
				Title:        "Check Your Inventory",
				Instructions: "Type 'inventory' (or 'i') to see what you're carrying.",
				Hint:         "Type: inventory",
				Trigger:      TriggerCommand,
				TriggerData:  map[string]interface{}{"command": "inventory"},
			},
			{
				ID:           "pickup",
				Title:        "Pick Up Items",
				Instructions: "Find an item and pick it up with 'get <item>' or 'take <item>'.",
				Hint:         "Look around for items, then type: get <itemname>",
				Trigger:      TriggerPickup,
			},
			{
				ID:           "equip",
				Title:        "Equip Your Gear",
				Instructions: "Equip a weapon or armor with 'equip <item>'. Check your stats with 'stats'.",
				Hint:         "Type: equip <weapon or armor you picked up>",
				Trigger:      TriggerEquip,
			},
			{
				ID:           "combat",
				Title:        "Enter Combat",
				Instructions: "Find an enemy and attack with 'attack <target>' or 'kill <target>'.",
				Hint:         "Find an NPC and type: attack <npcname>",
				Trigger:      TriggerAttack,
			},
			{
				ID:           "talk",
				Title:        "Talk to NPCs",
				Instructions: "Some NPCs have information. Type 'talk <npc>' to start a conversation.",
				Hint:         "Find Morpheus or another friendly NPC and type: talk <name>",
				Trigger:      TriggerTalk,
			},
			{
				ID:           "help",
				Title:        "Getting Help",
				Instructions: "Type 'help' to see all available commands. Try 'help <command>' for details.",
				Hint:         "Type: help",
				Trigger:      TriggerCommand,
				TriggerData:  map[string]interface{}{"command": "help"},
			},
		},
		Rewards: TutorialRewards{
			XP:      50,
			Money:   100,
			Items:   []string{"medkit"},
			Title:   "Awakened",
			Unlocks: []string{"combat_mastery", "crafting_101"},
		},
	}
	
	// Combat Mastery Tutorial
	m.tutorials["combat_mastery"] = &Tutorial{
		ID:            "combat_mastery",
		Name:          "Combat Mastery",
		Description:   "Master the art of combat in the Matrix.",
		Prerequisites: []string{"new_player"},
		Skippable:     true,
		Steps: []TutorialStep{
			{
				ID:           "skills",
				Title:        "Learn Your Skills",
				Instructions: "Type 'skills' to see your available combat abilities.",
				Hint:         "Type: skills",
				Trigger:      TriggerCommand,
				TriggerData:  map[string]interface{}{"command": "skills"},
			},
			{
				ID:           "use_skill",
				Title:        "Use a Skill",
				Instructions: "In combat, use 'skill <name>' to activate a special ability.",
				Hint:         "During combat, type: skill <skillname>",
				Trigger:      TriggerCommand,
				TriggerData:  map[string]interface{}{"command": "skill"},
			},
			{
				ID:           "defeat_enemy",
				Title:        "Defeat an Enemy",
				Instructions: "Win a fight! Defeat any hostile NPC.",
				Hint:         "Find and defeat an enemy using attacks and skills.",
				Trigger:      TriggerKill,
			},
			{
				ID:           "check_xp",
				Title:        "Track Progress",
				Instructions: "Check your experience with 'stats'. Level up to unlock new abilities!",
				Hint:         "Type: stats",
				Trigger:      TriggerCommand,
				TriggerData:  map[string]interface{}{"command": "stats"},
			},
		},
		Rewards: TutorialRewards{
			XP:    100,
			Title: "Warrior",
		},
	}
	
	// Crafting Tutorial
	m.tutorials["crafting_101"] = &Tutorial{
		ID:            "crafting_101",
		Name:          "Crafting 101",
		Description:   "Learn to create items from components.",
		Prerequisites: []string{"new_player"},
		Skippable:     true,
		Steps: []TutorialStep{
			{
				ID:           "recipes",
				Title:        "View Recipes",
				Instructions: "Type 'recipes' to see what you can craft.",
				Hint:         "Type: recipes",
				Trigger:      TriggerCommand,
				TriggerData:  map[string]interface{}{"command": "recipes"},
			},
			{
				ID:           "gather",
				Title:        "Gather Materials",
				Instructions: "Collect crafting components from defeated enemies or exploration.",
				Hint:         "Defeat enemies or search rooms for components.",
				Trigger:      TriggerPickup,
			},
			{
				ID:           "craft",
				Title:        "Craft an Item",
				Instructions: "When you have materials, type 'craft <item>' to create something new.",
				Hint:         "Type: craft <recipe name>",
				Trigger:      TriggerCommand,
				TriggerData:  map[string]interface{}{"command": "craft"},
			},
		},
		Rewards: TutorialRewards{
			XP:    75,
			Items: []string{"code_fragment"},
			Title: "Crafter",
		},
	}
	
	// Faction Tutorial
	m.tutorials["faction_guide"] = &Tutorial{
		ID:            "faction_guide",
		Name:          "Faction Guide",
		Description:   "Learn about the three factions vying for control.",
		Prerequisites: []string{"new_player"},
		Skippable:     true,
		Steps: []TutorialStep{
			{
				ID:           "factions",
				Title:        "View Factions",
				Instructions: "Type 'factions' to see the available factions and your standing.",
				Hint:         "Type: factions",
				Trigger:      TriggerCommand,
				TriggerData:  map[string]interface{}{"command": "factions"},
			},
			{
				ID:           "reputation",
				Title:        "Understand Reputation",
				Instructions: "Actions affect your reputation. Help a faction to increase standing.",
				Hint:         "Complete quests or defeat faction enemies.",
				Trigger:      TriggerAny,
			},
		},
		Rewards: TutorialRewards{
			XP:    50,
			Title: "Diplomat",
		},
	}
	
	// Party Tutorial
	m.tutorials["party_play"] = &Tutorial{
		ID:            "party_play",
		Name:          "Party Play",
		Description:   "Learn to team up with other players.",
		Prerequisites: []string{"new_player"},
		Skippable:     true,
		Steps: []TutorialStep{
			{
				ID:           "party_create",
				Title:        "Create a Party",
				Instructions: "Type 'party create' to form a group, or 'party join <player>' to join one.",
				Hint:         "Type: party create",
				Trigger:      TriggerCommand,
				TriggerData:  map[string]interface{}{"command": "party"},
			},
			{
				ID:           "party_chat",
				Title:        "Party Communication",
				Instructions: "Use 'party say <message>' to talk to your party members.",
				Hint:         "Type: party say Hello team!",
				Trigger:      TriggerCommand,
				TriggerData:  map[string]interface{}{"command": "party"},
			},
		},
		Rewards: TutorialRewards{
			XP:    50,
			Title: "Team Player",
		},
	}
}

// Global manager instance
var GlobalManager = NewManager()
