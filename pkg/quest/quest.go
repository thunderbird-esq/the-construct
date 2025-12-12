// Package quest implements a multi-stage quest system for The Construct.
// Quests have stages, prerequisites, objectives, and rewards.
package quest

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
	"time"
)

// ObjectiveType defines what kind of objective this is
type ObjectiveType string

const (
	ObjKill     ObjectiveType = "kill"     // Kill N of target
	ObjCollect  ObjectiveType = "collect"  // Collect N of item
	ObjDeliver  ObjectiveType = "deliver"  // Deliver item to NPC
	ObjVisit    ObjectiveType = "visit"    // Visit a room
	ObjTalk     ObjectiveType = "talk"     // Talk to NPC
	ObjUse      ObjectiveType = "use"      // Use an item
	ObjChoice   ObjectiveType = "choice"   // Make a choice (red/blue pill)
)

// Objective represents a single quest objective
type Objective struct {
	ID          string        `json:"id"`
	Type        ObjectiveType `json:"type"`
	Target      string        `json:"target"`       // NPC ID, item ID, or room ID
	Count       int           `json:"count"`        // How many (for kill/collect)
	Description string        `json:"description"`  // Human-readable description
}

// Reward represents quest completion rewards
type Reward struct {
	XP     int      `json:"xp"`
	Money  int      `json:"money"`
	Items  []string `json:"items"`   // Item IDs to give
	Title  string   `json:"title"`   // Title to unlock
	Unlock string   `json:"unlock"`  // Quest ID to unlock
}

// Stage represents a stage of a multi-stage quest
type Stage struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Objectives  []Objective `json:"objectives"`
	Dialogue    string      `json:"dialogue"`     // Text shown when starting stage
	Completion  string      `json:"completion"`   // Text shown when completing stage
}

// Quest represents a complete quest definition
type Quest struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Giver         string   `json:"giver"`          // NPC ID who gives quest
	Prerequisites []string `json:"prerequisites"`  // Quest IDs that must be completed first
	Stages        []Stage  `json:"stages"`
	Reward        Reward   `json:"reward"`
	Repeatable    bool     `json:"repeatable"`
	MinLevel      int      `json:"min_level"`
}

// Progress tracks a player's quest progress
type Progress struct {
	QuestID       string             `json:"quest_id"`
	CurrentStage  int                `json:"current_stage"`
	ObjectiveProgress map[string]int `json:"objective_progress"` // objective ID -> count
	StartedAt     time.Time          `json:"started_at"`
	CompletedAt   time.Time          `json:"completed_at,omitempty"`
}

// PlayerQuests holds all quest data for a player
type PlayerQuests struct {
	Active    map[string]*Progress `json:"active"`
	Completed []string             `json:"completed"`
}

// Manager handles quest operations
type Manager struct {
	mu      sync.RWMutex
	Quests  map[string]*Quest          // All available quests
	Players map[string]*PlayerQuests   // Player name -> quest progress
}

// NewManager creates a new quest manager and loads quests
func NewManager() *Manager {
	m := &Manager{
		Quests:  make(map[string]*Quest),
		Players: make(map[string]*PlayerQuests),
	}
	m.loadQuests()
	return m
}

// loadQuests loads quest definitions from data/quests.json
func (m *Manager) loadQuests() {
	data, err := os.ReadFile("data/quests.json")
	if err != nil {
		// Use default quests if file not found
		m.loadDefaultQuests()
		return
	}

	var questData struct {
		Quests map[string]*Quest `json:"quests"`
	}
	if err := json.Unmarshal(data, &questData); err != nil {
		m.loadDefaultQuests()
		return
	}

	m.Quests = questData.Quests
}

// loadDefaultQuests creates the core story quests
func (m *Manager) loadDefaultQuests() {
	m.Quests = map[string]*Quest{
		"free_your_mind": {
			ID:          "free_your_mind",
			Name:        "Free Your Mind",
			Description: "Begin your journey to awakening. Find Morpheus and learn the truth.",
			Giver:       "morpheus",
			Stages: []Stage{
				{
					ID:          "find_morpheus",
					Name:        "Find Morpheus",
					Description: "Locate the rebel leader Morpheus.",
					Objectives:  []Objective{{ID: "visit_dojo", Type: ObjVisit, Target: "dojo", Count: 1, Description: "Find the Dojo"}},
					Dialogue:    "You've heard whispers about a man named Morpheus who knows the truth...",
				},
				{
					ID:          "make_choice",
					Name:        "The Choice",
					Description: "Take the red pill or the blue pill.",
					Objectives:  []Objective{{ID: "take_pill", Type: ObjChoice, Target: "red_pill", Count: 1, Description: "Make your choice"}},
					Dialogue:    "This is your last chance. After this, there is no turning back.",
					Completion:  "You have chosen to see how deep the rabbit hole goes.",
				},
				{
					ID:          "first_combat",
					Name:        "Combat Training",
					Description: "Prove yourself in the training program.",
					Objectives:  []Objective{{ID: "kill_training", Type: ObjKill, Target: "riot_cop", Count: 1, Description: "Defeat a training opponent"}},
					Dialogue:    "Now, show me what you've learned.",
					Completion:  "You're beginning to believe.",
				},
			},
			Reward: Reward{XP: 100, Money: 50, Title: "Awakened"},
		},
		"the_oracle": {
			ID:            "the_oracle",
			Name:          "The Oracle",
			Description:   "Seek the Oracle and learn your destiny.",
			Giver:         "morpheus",
			Prerequisites: []string{"free_your_mind"},
			MinLevel:      3,
			Stages: []Stage{
				{
					ID:          "find_oracle",
					Name:        "Find the Oracle",
					Description: "Navigate to the Oracle's apartment.",
					Objectives:  []Objective{{ID: "visit_oracle", Type: ObjVisit, Target: "oracle_apartment", Count: 1, Description: "Find the Oracle"}},
					Dialogue:    "The Oracle can tell you what you need to know. But first, you must find her.",
				},
				{
					ID:          "defeat_seraph",
					Name:        "Prove Yourself",
					Description: "Seraph guards the Oracle. You must prove your worth.",
					Objectives:  []Objective{{ID: "beat_seraph", Type: ObjKill, Target: "seraph", Count: 1, Description: "Defeat Seraph"}},
					Dialogue:    "You do not truly know someone until you fight them.",
				},
				{
					ID:          "hear_prophecy",
					Name:        "The Prophecy",
					Description: "Receive the Oracle's wisdom.",
					Objectives:  []Objective{{ID: "talk_oracle", Type: ObjTalk, Target: "oracle", Count: 1, Description: "Speak with the Oracle"}},
					Completion:  "Being The One is just like being in love. No one can tell you you're in love, you just know it.",
				},
			},
			Reward: Reward{XP: 250, Money: 100, Unlock: "rescue_morpheus"},
		},
		"rescue_morpheus": {
			ID:            "rescue_morpheus",
			Name:          "Rescue Morpheus",
			Description:   "Morpheus has been captured by Agents. Save him before it's too late.",
			Giver:         "trinity",
			Prerequisites: []string{"the_oracle"},
			MinLevel:      5,
			Stages: []Stage{
				{
					ID:          "gather_team",
					Name:        "Assemble",
					Description: "Meet Trinity at the extraction point.",
					Objectives:  []Objective{{ID: "meet_trinity", Type: ObjVisit, Target: "helipad", Count: 1, Description: "Meet Trinity at the helipad"}},
					Dialogue:    "Morpheus believed something I'm not sure I believe. But I believe he was willing to give his life for what he believed.",
				},
				{
					ID:          "infiltrate",
					Name:        "Infiltrate",
					Description: "Fight through the government building.",
					Objectives: []Objective{
						{ID: "kill_guards", Type: ObjKill, Target: "riot_cop", Count: 5, Description: "Clear the guards"},
						{ID: "reach_floor", Type: ObjVisit, Target: "agent_floor", Count: 1, Description: "Reach the interrogation floor"},
					},
					Dialogue:    "There are only two ways out of this building. One is that scaffold, the other is in their custody.",
				},
				{
					ID:          "face_agents",
					Name:        "Face the Agents",
					Description: "Defeat the Agents and free Morpheus.",
					Objectives:  []Objective{{ID: "beat_agent", Type: ObjKill, Target: "agent", Count: 1, Description: "Defeat an Agent"}},
					Completion:  "You moved like they do. I've never seen anyone move that fast.",
				},
			},
			Reward: Reward{XP: 500, Money: 250, Items: []string{"code_blade"}, Title: "The One?"},
		},
	}
}

// GetPlayerQuests gets or creates quest tracking for a player
func (m *Manager) GetPlayerQuests(playerName string) *PlayerQuests {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := strings.ToLower(playerName)
	if pq, ok := m.Players[name]; ok {
		return pq
	}

	pq := &PlayerQuests{
		Active:    make(map[string]*Progress),
		Completed: []string{},
	}
	m.Players[name] = pq
	return pq
}

// CanStart checks if a player can start a quest
func (m *Manager) CanStart(playerName string, questID string, playerLevel int) (bool, string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	quest, ok := m.Quests[questID]
	if !ok {
		return false, "Quest not found."
	}

	pq := m.Players[strings.ToLower(playerName)]
	if pq == nil {
		pq = &PlayerQuests{Active: make(map[string]*Progress), Completed: []string{}}
	}

	// Check if already active
	if _, active := pq.Active[questID]; active {
		return false, "You are already on this quest."
	}

	// Check if completed and not repeatable
	for _, completed := range pq.Completed {
		if completed == questID && !quest.Repeatable {
			return false, "You have already completed this quest."
		}
	}

	// Check level
	if playerLevel < quest.MinLevel {
		return false, "You are not experienced enough for this quest."
	}

	// Check prerequisites
	for _, prereq := range quest.Prerequisites {
		found := false
		for _, completed := range pq.Completed {
			if completed == prereq {
				found = true
				break
			}
		}
		if !found {
			prereqQuest := m.Quests[prereq]
			prereqName := prereq
			if prereqQuest != nil {
				prereqName = prereqQuest.Name
			}
			return false, "You must complete \"" + prereqName + "\" first."
		}
	}

	return true, ""
}

// StartQuest begins a quest for a player
func (m *Manager) StartQuest(playerName, questID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	quest, ok := m.Quests[questID]
	if !ok {
		return "", nil
	}

	name := strings.ToLower(playerName)
	if m.Players[name] == nil {
		m.Players[name] = &PlayerQuests{Active: make(map[string]*Progress), Completed: []string{}}
	}

	progress := &Progress{
		QuestID:          questID,
		CurrentStage:     0,
		ObjectiveProgress: make(map[string]int),
		StartedAt:        time.Now(),
	}

	m.Players[name].Active[questID] = progress

	// Return stage dialogue
	if len(quest.Stages) > 0 {
		return quest.Stages[0].Dialogue, nil
	}
	return quest.Description, nil
}

// UpdateProgress updates objective progress and checks for completion
func (m *Manager) UpdateProgress(playerName string, objType ObjectiveType, target string, count int) []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := strings.ToLower(playerName)
	pq := m.Players[name]
	if pq == nil {
		return nil
	}

	var messages []string

	for questID, progress := range pq.Active {
		quest := m.Quests[questID]
		if quest == nil || progress.CurrentStage >= len(quest.Stages) {
			continue
		}

		stage := quest.Stages[progress.CurrentStage]
		stageComplete := true

		for _, obj := range stage.Objectives {
			if obj.Type == objType && (obj.Target == target || strings.Contains(target, obj.Target)) {
				progress.ObjectiveProgress[obj.ID] += count
			}

			// Check if objective is complete
			if progress.ObjectiveProgress[obj.ID] < obj.Count {
				stageComplete = false
			}
		}

		// Advance stage if complete
		if stageComplete {
			if stage.Completion != "" {
				messages = append(messages, stage.Completion)
			}

			progress.CurrentStage++

			// Check if quest complete
			if progress.CurrentStage >= len(quest.Stages) {
				progress.CompletedAt = time.Now()
				pq.Completed = append(pq.Completed, questID)
				delete(pq.Active, questID)
				messages = append(messages, "Quest Complete: "+quest.Name)
			} else {
				// Start next stage
				nextStage := quest.Stages[progress.CurrentStage]
				messages = append(messages, "New Objective: "+nextStage.Name)
				if nextStage.Dialogue != "" {
					messages = append(messages, nextStage.Dialogue)
				}
			}
		}
	}

	return messages
}

// GetActiveQuests returns a formatted list of active quests
func (m *Manager) GetActiveQuests(playerName string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pq := m.Players[strings.ToLower(playerName)]
	if pq == nil || len(pq.Active) == 0 {
		return "You have no active quests.\r\n"
	}

	var sb strings.Builder
	sb.WriteString("=== ACTIVE QUESTS ===\r\n\r\n")

	for questID, progress := range pq.Active {
		quest := m.Quests[questID]
		if quest == nil {
			continue
		}

		sb.WriteString(quest.Name + "\r\n")
		
		if progress.CurrentStage < len(quest.Stages) {
			stage := quest.Stages[progress.CurrentStage]
			sb.WriteString("  Stage: " + stage.Name + "\r\n")
			
			for _, obj := range stage.Objectives {
				current := progress.ObjectiveProgress[obj.ID]
				status := "[ ]"
				if current >= obj.Count {
					status = "[X]"
				}
				sb.WriteString("  " + status + " " + obj.Description)
				if obj.Count > 1 {
					sb.WriteString(" (" + string(rune('0'+current)) + "/" + string(rune('0'+obj.Count)) + ")")
				}
				sb.WriteString("\r\n")
			}
		}
		sb.WriteString("\r\n")
	}

	return sb.String()
}

// GetCompletedQuests returns list of completed quest names
func (m *Manager) GetCompletedQuests(playerName string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pq := m.Players[strings.ToLower(playerName)]
	if pq == nil {
		return nil
	}

	var names []string
	for _, qid := range pq.Completed {
		if q := m.Quests[qid]; q != nil {
			names = append(names, q.Name)
		}
	}
	return names
}

// GetQuestReward returns the reward for a quest
func (m *Manager) GetQuestReward(questID string) *Reward {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if q := m.Quests[questID]; q != nil {
		return &q.Reward
	}
	return nil
}

// Global quest manager
var GlobalQuests = NewManager()
