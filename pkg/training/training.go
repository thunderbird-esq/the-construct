// Package training implements training programs and PvP arenas for Matrix MUD.
// Training programs are instanced combat zones where players can practice
// skills and engage in consensual PvP without death penalties.
package training

import (
	"fmt"
	"sync"
	"time"
)

// ProgramType represents a type of training program
type ProgramType string

const (
	ProgramCombat   ProgramType = "combat"   // Basic combat training
	ProgramSurvival ProgramType = "survival" // Survive waves of enemies
	ProgramPvP      ProgramType = "pvp"      // Player vs Player arena
	ProgramTrial    ProgramType = "trial"    // Timed challenge
)

// Program represents a training program instance
type Program struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Type        ProgramType `json:"type"`
	Difficulty  int         `json:"difficulty"` // 1-5
	TimeLimit   int         `json:"time_limit"` // seconds, 0 = no limit
	Rewards     Rewards     `json:"rewards"`
}

// Rewards for completing a program
type Rewards struct {
	XP    int `json:"xp"`
	Money int `json:"money"`
}

// Instance represents an active training program session
type Instance struct {
	ID          string
	Program     *Program
	Players     map[string]bool // player name -> in combat
	StartTime   time.Time
	EndTime     time.Time
	IsActive    bool
	IsPvP       bool
	Scores      map[string]int // player -> score
	mu          sync.RWMutex
}

// Challenge represents a combat challenge with leaderboard
type Challenge struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	BestTime    int      `json:"best_time"`    // seconds
	BestPlayer  string   `json:"best_player"`
	Attempts    int      `json:"attempts"`
}

// Manager handles training programs
type Manager struct {
	mu         sync.RWMutex
	Programs   map[string]*Program
	Instances  map[string]*Instance
	Challenges map[string]*Challenge
	InProgram  map[string]string // player -> instance ID
}

// NewManager creates a new training manager
func NewManager() *Manager {
	m := &Manager{
		Programs:   make(map[string]*Program),
		Instances:  make(map[string]*Instance),
		Challenges: make(map[string]*Challenge),
		InProgram:  make(map[string]string),
	}
	m.loadPrograms()
	return m
}

// loadPrograms initializes available training programs
func (m *Manager) loadPrograms() {
	m.Programs = map[string]*Program{
		"combat_basic": {
			ID:          "combat_basic",
			Name:        "Basic Combat Training",
			Description: "Learn the fundamentals of combat in the Matrix.",
			Type:        ProgramCombat,
			Difficulty:  1,
			TimeLimit:   0,
			Rewards:     Rewards{XP: 50, Money: 25},
		},
		"combat_advanced": {
			ID:          "combat_advanced",
			Name:        "Advanced Combat Training",
			Description: "Master advanced combat techniques.",
			Type:        ProgramCombat,
			Difficulty:  3,
			TimeLimit:   300,
			Rewards:     Rewards{XP: 150, Money: 100},
		},
		"survival_wave": {
			ID:          "survival_wave",
			Name:        "Survival: Wave Defense",
			Description: "Survive increasingly difficult waves of enemies.",
			Type:        ProgramSurvival,
			Difficulty:  2,
			TimeLimit:   600,
			Rewards:     Rewards{XP: 200, Money: 150},
		},
		"pvp_arena": {
			ID:          "pvp_arena",
			Name:        "Combat Arena",
			Description: "Face other awakened minds in combat.",
			Type:        ProgramPvP,
			Difficulty:  0,
			TimeLimit:   180,
			Rewards:     Rewards{XP: 100, Money: 50},
		},
		"trial_speed": {
			ID:          "trial_speed",
			Name:        "Speed Trial",
			Description: "Defeat all opponents as fast as possible.",
			Type:        ProgramTrial,
			Difficulty:  3,
			TimeLimit:   120,
			Rewards:     Rewards{XP: 250, Money: 200},
		},
		"kung_fu": {
			ID:          "kung_fu",
			Name:        "Kung Fu Program",
			Description: "\"I know kung fu.\" - Neo",
			Type:        ProgramCombat,
			Difficulty:  2,
			TimeLimit:   0,
			Rewards:     Rewards{XP: 100, Money: 50},
		},
	}

	// Initialize challenges
	m.Challenges = map[string]*Challenge{
		"speed_trial": {
			ID:          "speed_trial",
			Name:        "Speed Trial",
			Description: "Complete the speed trial in the fastest time.",
		},
		"survival_record": {
			ID:          "survival_record",
			Name:        "Survival Record",
			Description: "Survive the most waves.",
		},
		"pvp_champion": {
			ID:          "pvp_champion",
			Name:        "Arena Champion",
			Description: "Most PvP victories.",
		},
	}
}

// ListPrograms returns all available programs
func (m *Manager) ListPrograms() []*Program {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Program, 0, len(m.Programs))
	for _, p := range m.Programs {
		result = append(result, p)
	}
	return result
}

// GetProgram returns a program by ID
func (m *Manager) GetProgram(id string) *Program {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Programs[id]
}

// StartProgram creates a new training instance for a player
func (m *Manager) StartProgram(playerName, programID string) (*Instance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already in a program
	if _, ok := m.InProgram[playerName]; ok {
		return nil, fmt.Errorf("you are already in a training program")
	}

	program := m.Programs[programID]
	if program == nil {
		return nil, fmt.Errorf("unknown program: %s", programID)
	}

	instance := &Instance{
		ID:        fmt.Sprintf("%s_%d", programID, time.Now().UnixNano()),
		Program:   program,
		Players:   map[string]bool{playerName: false},
		StartTime: time.Now(),
		IsActive:  true,
		IsPvP:     program.Type == ProgramPvP,
		Scores:    map[string]int{playerName: 0},
	}

	m.Instances[instance.ID] = instance
	m.InProgram[playerName] = instance.ID

	return instance, nil
}

// JoinProgram adds a player to an existing PvP instance
func (m *Manager) JoinProgram(playerName, instanceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.InProgram[playerName]; ok {
		return fmt.Errorf("you are already in a training program")
	}

	instance := m.Instances[instanceID]
	if instance == nil {
		return fmt.Errorf("instance not found")
	}

	if !instance.IsPvP {
		return fmt.Errorf("this is not a PvP program")
	}

	if len(instance.Players) >= 2 {
		return fmt.Errorf("arena is full")
	}

	instance.Players[playerName] = false
	instance.Scores[playerName] = 0
	m.InProgram[playerName] = instanceID

	return nil
}

// LeaveProgram removes a player from their program
func (m *Manager) LeaveProgram(playerName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	instanceID, ok := m.InProgram[playerName]
	if !ok {
		return fmt.Errorf("you are not in a training program")
	}

	delete(m.InProgram, playerName)

	instance := m.Instances[instanceID]
	if instance != nil {
		delete(instance.Players, playerName)
		delete(instance.Scores, playerName)

		// Clean up empty instances
		if len(instance.Players) == 0 {
			delete(m.Instances, instanceID)
		}
	}

	return nil
}

// GetPlayerInstance returns the instance a player is in
func (m *Manager) GetPlayerInstance(playerName string) *Instance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instanceID, ok := m.InProgram[playerName]
	if !ok {
		return nil
	}
	return m.Instances[instanceID]
}

// IsInProgram checks if a player is in a training program
func (m *Manager) IsInProgram(playerName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.InProgram[playerName]
	return ok
}

// RecordScore adds to a player's score in their current program
func (m *Manager) RecordScore(playerName string, points int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	instanceID, ok := m.InProgram[playerName]
	if !ok {
		return
	}

	instance := m.Instances[instanceID]
	if instance != nil {
		instance.mu.Lock()
		instance.Scores[playerName] += points
		instance.mu.Unlock()
	}
}

// CompleteProgram finishes a program and returns rewards
func (m *Manager) CompleteProgram(playerName string) (*Rewards, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	instanceID, ok := m.InProgram[playerName]
	if !ok {
		return nil, 0, fmt.Errorf("you are not in a training program")
	}

	instance := m.Instances[instanceID]
	if instance == nil {
		delete(m.InProgram, playerName)
		return nil, 0, fmt.Errorf("instance not found")
	}

	score := instance.Scores[playerName]
	rewards := &instance.Program.Rewards

	// Check for challenge records
	elapsed := int(time.Since(instance.StartTime).Seconds())
	if instance.Program.Type == ProgramTrial {
		if challenge := m.Challenges["speed_trial"]; challenge != nil {
			if challenge.BestTime == 0 || elapsed < challenge.BestTime {
				challenge.BestTime = elapsed
				challenge.BestPlayer = playerName
			}
			challenge.Attempts++
		}
	}

	// Clean up
	delete(m.InProgram, playerName)
	delete(instance.Players, playerName)
	delete(instance.Scores, playerName)
	if len(instance.Players) == 0 {
		delete(m.Instances, instanceID)
	}

	return rewards, score, nil
}

// GetChallenge returns a challenge by ID
func (m *Manager) GetChallenge(id string) *Challenge {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Challenges[id]
}

// ListChallenges returns all challenges
func (m *Manager) ListChallenges() []*Challenge {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Challenge, 0, len(m.Challenges))
	for _, c := range m.Challenges {
		result = append(result, c)
	}
	return result
}

// GlobalTraining is a global training manager instance
var GlobalTraining = NewManager()
