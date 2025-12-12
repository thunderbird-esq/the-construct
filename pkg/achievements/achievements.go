// Package achievements implements the achievement and title system for Matrix MUD.
// Players earn achievements for milestones and can unlock titles.
package achievements

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// AchievementID uniquely identifies an achievement
type AchievementID string

// Achievement categories
const (
	CatCombat      = "combat"
	CatExploration = "exploration"
	CatSocial      = "social"
	CatProgression = "progression"
	CatSecret      = "secret"
)

// Achievement definitions
const (
	AchFirstBlood     AchievementID = "first_blood"
	AchAgentSlayer    AchievementID = "agent_slayer"
	AchPacifist       AchievementID = "pacifist"
	AchAwakened       AchievementID = "awakened"
	AchReachZion      AchievementID = "reach_zion"
	AchMeetOracle     AchievementID = "meet_oracle"
	AchPhoneMaster    AchievementID = "phone_master"
	AchPartyAnimal    AchievementID = "party_animal"
	AchCrafter        AchievementID = "crafter"
	AchMillionaire    AchievementID = "millionaire"
	AchLevel10        AchievementID = "level_10"
	AchLevel25        AchievementID = "level_25"
	AchQuestComplete5 AchievementID = "quest_complete_5"
	AchQuestComplete10 AchievementID = "quest_complete_10"
	AchSurvivor       AchievementID = "survivor"
	AchTheOne         AchievementID = "the_one"
)

// Achievement represents an earnable achievement
type Achievement struct {
	ID          AchievementID `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Category    string        `json:"category"`
	Points      int           `json:"points"`
	Title       string        `json:"title,omitempty"` // Unlocked title, if any
	Hidden      bool          `json:"hidden"`          // Secret achievements
}

// PlayerAchievements tracks a player's earned achievements
type PlayerAchievements struct {
	Earned      map[AchievementID]int64 `json:"earned"` // ID -> unix timestamp
	ActiveTitle string                  `json:"active_title"`
	TotalPoints int                     `json:"total_points"`
}

// Manager handles achievement operations
type Manager struct {
	mu           sync.RWMutex
	Achievements map[AchievementID]*Achievement
	Players      map[string]*PlayerAchievements
}

// NewManager creates a new achievement manager
func NewManager() *Manager {
	m := &Manager{
		Achievements: make(map[AchievementID]*Achievement),
		Players:      make(map[string]*PlayerAchievements),
	}
	m.loadAchievements()
	return m
}

// loadAchievements initializes all achievements
func (m *Manager) loadAchievements() {
	m.Achievements = map[AchievementID]*Achievement{
		AchFirstBlood: {
			ID:          AchFirstBlood,
			Name:        "First Blood",
			Description: "Defeat your first enemy.",
			Category:    CatCombat,
			Points:      10,
		},
		AchAgentSlayer: {
			ID:          AchAgentSlayer,
			Name:        "Agent Slayer",
			Description: "Defeat an Agent.",
			Category:    CatCombat,
			Points:      50,
			Title:       "Agent Slayer",
		},
		AchPacifist: {
			ID:          AchPacifist,
			Name:        "Pacifist",
			Description: "Reach level 5 without killing anything.",
			Category:    CatSecret,
			Points:      100,
			Title:       "The Peaceful",
			Hidden:      true,
		},
		AchAwakened: {
			ID:          AchAwakened,
			Name:        "Awakened",
			Description: "Take the red pill and leave the Matrix.",
			Category:    CatProgression,
			Points:      25,
			Title:       "Awakened",
		},
		AchReachZion: {
			ID:          AchReachZion,
			Name:        "Welcome to the Real World",
			Description: "Reach Zion for the first time.",
			Category:    CatExploration,
			Points:      30,
		},
		AchMeetOracle: {
			ID:          AchMeetOracle,
			Name:        "Seek the Oracle",
			Description: "Meet the Oracle.",
			Category:    CatExploration,
			Points:      25,
		},
		AchPhoneMaster: {
			ID:          AchPhoneMaster,
			Name:        "Phone Master",
			Description: "Discover all phone booth locations.",
			Category:    CatExploration,
			Points:      40,
			Title:       "Operator",
		},
		AchPartyAnimal: {
			ID:          AchPartyAnimal,
			Name:        "Party Animal",
			Description: "Complete a quest with a full party.",
			Category:    CatSocial,
			Points:      20,
		},
		AchCrafter: {
			ID:          AchCrafter,
			Name:        "Craftsman",
			Description: "Craft 10 items.",
			Category:    CatProgression,
			Points:      25,
		},
		AchMillionaire: {
			ID:          AchMillionaire,
			Name:        "Millionaire",
			Description: "Accumulate 1,000,000 bits.",
			Category:    CatProgression,
			Points:      100,
			Title:       "The Wealthy",
		},
		AchLevel10: {
			ID:          AchLevel10,
			Name:        "Rising Power",
			Description: "Reach level 10.",
			Category:    CatProgression,
			Points:      25,
		},
		AchLevel25: {
			ID:          AchLevel25,
			Name:        "True Potential",
			Description: "Reach level 25.",
			Category:    CatProgression,
			Points:      50,
			Title:       "The Powerful",
		},
		AchQuestComplete5: {
			ID:          AchQuestComplete5,
			Name:        "Quest Seeker",
			Description: "Complete 5 quests.",
			Category:    CatProgression,
			Points:      20,
		},
		AchQuestComplete10: {
			ID:          AchQuestComplete10,
			Name:        "Quest Master",
			Description: "Complete 10 quests.",
			Category:    CatProgression,
			Points:      40,
			Title:       "Questmaster",
		},
		AchSurvivor: {
			ID:          AchSurvivor,
			Name:        "Survivor",
			Description: "Escape from combat with less than 5 HP.",
			Category:    CatCombat,
			Points:      15,
		},
		AchTheOne: {
			ID:          AchTheOne,
			Name:        "The One",
			Description: "Complete the main storyline.",
			Category:    CatSecret,
			Points:      200,
			Title:       "The One",
			Hidden:      true,
		},
	}
}

// GetPlayerAchievements returns a player's achievements, creating if needed
func (m *Manager) GetPlayerAchievements(playerName string) *PlayerAchievements {
	m.mu.Lock()
	defer m.mu.Unlock()

	if pa, ok := m.Players[playerName]; ok {
		return pa
	}

	pa := &PlayerAchievements{
		Earned: make(map[AchievementID]int64),
	}
	m.Players[playerName] = pa
	return pa
}

// Award grants an achievement to a player
// Returns the achievement if newly earned, nil if already had
func (m *Manager) Award(playerName string, achID AchievementID) *Achievement {
	m.mu.Lock()
	defer m.mu.Unlock()

	pa := m.Players[playerName]
	if pa == nil {
		pa = &PlayerAchievements{
			Earned: make(map[AchievementID]int64),
		}
		m.Players[playerName] = pa
	}

	// Check if already earned
	if _, ok := pa.Earned[achID]; ok {
		return nil
	}

	ach := m.Achievements[achID]
	if ach == nil {
		return nil
	}

	pa.Earned[achID] = time.Now().Unix()
	pa.TotalPoints += ach.Points

	return ach
}

// HasAchievement checks if a player has earned an achievement
func (m *Manager) HasAchievement(playerName string, achID AchievementID) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pa := m.Players[playerName]
	if pa == nil {
		return false
	}
	_, ok := pa.Earned[achID]
	return ok
}

// GetEarnedAchievements returns all achievements earned by a player
func (m *Manager) GetEarnedAchievements(playerName string) []*Achievement {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pa := m.Players[playerName]
	if pa == nil {
		return nil
	}

	result := make([]*Achievement, 0, len(pa.Earned))
	for achID := range pa.Earned {
		if ach := m.Achievements[achID]; ach != nil {
			result = append(result, ach)
		}
	}
	return result
}

// GetAvailableTitles returns titles a player has unlocked
func (m *Manager) GetAvailableTitles(playerName string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pa := m.Players[playerName]
	if pa == nil {
		return nil
	}

	var titles []string
	for achID := range pa.Earned {
		if ach := m.Achievements[achID]; ach != nil && ach.Title != "" {
			titles = append(titles, ach.Title)
		}
	}
	return titles
}

// SetTitle sets a player's active title
func (m *Manager) SetTitle(playerName, title string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	pa := m.Players[playerName]
	if pa == nil {
		return false
	}

	// Verify they have unlocked this title
	if title != "" {
		found := false
		for achID := range pa.Earned {
			if ach := m.Achievements[achID]; ach != nil && ach.Title == title {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	pa.ActiveTitle = title
	return true
}

// GetTitle returns a player's active title
func (m *Manager) GetTitle(playerName string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pa := m.Players[playerName]
	if pa == nil {
		return ""
	}
	return pa.ActiveTitle
}

// GetTotalPoints returns a player's total achievement points
func (m *Manager) GetTotalPoints(playerName string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pa := m.Players[playerName]
	if pa == nil {
		return 0
	}
	return pa.TotalPoints
}

// GetByCategory returns all achievements in a category
func (m *Manager) GetByCategory(category string) []*Achievement {
	var result []*Achievement
	for _, ach := range m.Achievements {
		if ach.Category == category && !ach.Hidden {
			result = append(result, ach)
		}
	}
	return result
}

// Save persists achievement data to disk
func (m *Manager) Save() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := json.MarshalIndent(m.Players, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("data/achievements.json", data, 0644)
}

// Load reads achievement data from disk
func (m *Manager) Load() error {
	data, err := os.ReadFile("data/achievements.json")
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	return json.Unmarshal(data, &m.Players)
}

// GlobalAchievements is a global achievement manager instance
var GlobalAchievements = NewManager()
