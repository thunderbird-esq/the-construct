// Package dialogue implements branching conversation trees for NPCs.
package dialogue

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
)

// NodeType defines the type of dialogue node
type NodeType string

const (
	NodeText   NodeType = "text"   // Simple text display
	NodeChoice NodeType = "choice" // Player makes a choice
	NodeAction NodeType = "action" // Triggers an action (give item, start quest, etc.)
	NodeEnd    NodeType = "end"    // Ends dialogue
)

// Choice represents a dialogue choice the player can make
type Choice struct {
	Text       string           `json:"text"`      // Display text
	NextNodeID string           `json:"next"`      // ID of next node
	Condition  *ChoiceCondition `json:"condition"` // Optional condition to show choice
	Action     *ChoiceAction    `json:"action"`    // Optional action on select
}

// ChoiceCondition defines when a choice is available
type ChoiceCondition struct {
	Type   string `json:"type"`   // "quest", "item", "level", "faction"
	Target string `json:"target"` // Quest ID, item ID, faction ID
	Value  int    `json:"value"`  // Level requirement or reputation threshold
}

// ChoiceAction defines what happens when a choice is selected
type ChoiceAction struct {
	Type   string `json:"type"`   // "give_item", "take_item", "start_quest", "complete_objective", "reputation"
	Target string `json:"target"` // Item ID, quest ID, or faction ID
	Value  int    `json:"value"`  // Quantity or reputation change
}

// Node represents a single node in a dialogue tree
type Node struct {
	ID       string        `json:"id"`
	Type     NodeType      `json:"type"`
	Speaker  string        `json:"speaker"` // NPC name or "player"
	Text     string        `json:"text"`    // Dialogue text
	Choices  []Choice      `json:"choices"` // Available choices (for choice nodes)
	NextNode string        `json:"next"`    // Next node ID (for text nodes)
	Action   *ChoiceAction `json:"action"`  // Action to perform (for action nodes)
}

// Tree represents a complete dialogue tree for an NPC
type Tree struct {
	ID       string           `json:"id"`
	NPCID    string           `json:"npc_id"`
	NPCName  string           `json:"npc_name"`
	RootNode string           `json:"root"`
	Nodes    map[string]*Node `json:"nodes"`
}

// Session tracks a player's current dialogue state
type Session struct {
	TreeID      string
	CurrentNode string
	History     []string // Node IDs visited
}

// Manager handles all dialogue operations
type Manager struct {
	mu       sync.RWMutex
	Trees    map[string]*Tree    // NPC ID -> dialogue tree
	Sessions map[string]*Session // Player name -> active session
}

// NewManager creates a new dialogue manager
func NewManager() *Manager {
	m := &Manager{
		Trees:    make(map[string]*Tree),
		Sessions: make(map[string]*Session),
	}
	m.loadTrees()
	return m
}

// loadTrees loads dialogue trees from data/dialogue_trees.json
func (m *Manager) loadTrees() {
	data, err := os.ReadFile("data/dialogue_trees.json")
	if err != nil {
		m.loadDefaultTrees()
		return
	}

	var treeData struct {
		Trees map[string]*Tree `json:"trees"`
	}
	if err := json.Unmarshal(data, &treeData); err != nil {
		m.loadDefaultTrees()
		return
	}

	m.Trees = treeData.Trees
}

// loadDefaultTrees creates default dialogue trees for major NPCs
func (m *Manager) loadDefaultTrees() {
	m.Trees = make(map[string]*Tree)

	// Morpheus dialogue tree
	m.Trees["morpheus"] = &Tree{
		ID:       "morpheus_main",
		NPCID:    "morpheus",
		NPCName:  "Morpheus",
		RootNode: "greeting",
		Nodes: map[string]*Node{
			"greeting": {
				ID:      "greeting",
				Type:    NodeChoice,
				Speaker: "Morpheus",
				Text:    "Welcome. I've been expecting you. What would you like to know?",
				Choices: []Choice{
					{Text: "What is the Matrix?", NextNodeID: "matrix_explain"},
					{Text: "Tell me about the prophecy.", NextNodeID: "prophecy"},
					{Text: "I need training.", NextNodeID: "training"},
					{Text: "I should go.", NextNodeID: "farewell"},
				},
			},
			"matrix_explain": {
				ID:       "matrix_explain",
				Type:     NodeText,
				Speaker:  "Morpheus",
				Text:     "The Matrix is everywhere. It is all around us. Even now, in this very room. It is the world that has been pulled over your eyes to blind you from the truth.",
				NextNode: "matrix_continue",
			},
			"matrix_continue": {
				ID:      "matrix_continue",
				Type:    NodeChoice,
				Speaker: "Morpheus",
				Text:    "The truth is that you are a slave. Like everyone else, you were born into bondage, born into a prison for your mind.",
				Choices: []Choice{
					{Text: "How do I escape?", NextNodeID: "escape"},
					{Text: "I want to go back.", NextNodeID: "farewell"},
				},
			},
			"escape": {
				ID:       "escape",
				Type:     NodeText,
				Speaker:  "Morpheus",
				Text:     "Unfortunately, no one can be told what the Matrix is. You have to see it for yourself. Free your mind.",
				NextNode: "greeting",
			},
			"prophecy": {
				ID:       "prophecy",
				Type:     NodeText,
				Speaker:  "Morpheus",
				Text:     "When the Matrix was first built, there was a man born inside who had the ability to change whatever he wanted. He freed the first of us.",
				NextNode: "prophecy_continue",
			},
			"prophecy_continue": {
				ID:      "prophecy_continue",
				Type:    NodeChoice,
				Speaker: "Morpheus",
				Text:    "The Oracle prophesied his return. I believe that search is over.",
				Choices: []Choice{
					{Text: "You think I'm The One?", NextNodeID: "the_one"},
					{Text: "Tell me more.", NextNodeID: "greeting"},
				},
			},
			"the_one": {
				ID:       "the_one",
				Type:     NodeText,
				Speaker:  "Morpheus",
				Text:     "There's a difference between knowing the path and walking the path.",
				NextNode: "greeting",
			},
			"training": {
				ID:      "training",
				Type:    NodeChoice,
				Speaker: "Morpheus",
				Text:    "Your mind makes it real. What do you wish to learn?",
				Choices: []Choice{
					{Text: "Teach me combat.", NextNodeID: "combat_training"},
					{Text: "Teach me to focus.", NextNodeID: "focus_training"},
					{Text: "Maybe later.", NextNodeID: "greeting"},
				},
			},
			"combat_training": {
				ID:       "combat_training",
				Type:     NodeAction,
				Speaker:  "Morpheus",
				Text:     "Good. Remember, your mind makes it real. Let us begin.",
				NextNode: "training_end",
				Action:   &ChoiceAction{Type: "complete_objective", Target: "talk_morpheus"},
			},
			"focus_training": {
				ID:       "focus_training",
				Type:     NodeText,
				Speaker:  "Morpheus",
				Text:     "Focus is the key. When you can control your mind, you can control the Matrix around you.",
				NextNode: "greeting",
			},
			"training_end": {
				ID:       "training_end",
				Type:     NodeText,
				Speaker:  "Morpheus",
				Text:     "You're learning. Soon you'll be ready.",
				NextNode: "greeting",
			},
			"farewell": {
				ID:      "farewell",
				Type:    NodeEnd,
				Speaker: "Morpheus",
				Text:    "Remember, all I'm offering is the truth. Nothing more.",
			},
		},
	}

	// Oracle dialogue tree
	m.Trees["oracle"] = &Tree{
		ID:       "oracle_main",
		NPCID:    "oracle",
		NPCName:  "The Oracle",
		RootNode: "greeting",
		Nodes: map[string]*Node{
			"greeting": {
				ID:      "greeting",
				Type:    NodeChoice,
				Speaker: "The Oracle",
				Text:    "I know you're Neo. Be right with you. Would you like a cookie?",
				Choices: []Choice{
					{Text: "Yes, please.", NextNodeID: "cookie"},
					{Text: "No, thank you.", NextNodeID: "no_cookie"},
					{Text: "What can you tell me about my future?", NextNodeID: "future"},
				},
			},
			"cookie": {
				ID:       "cookie",
				Type:     NodeText,
				Speaker:  "The Oracle",
				Text:     "Here, take a cookie. I promise, by the time you're done eating it, you'll feel right as rain.",
				NextNode: "after_cookie",
			},
			"no_cookie": {
				ID:       "no_cookie",
				Type:     NodeText,
				Speaker:  "The Oracle",
				Text:     "Suit yourself. More for me.",
				NextNode: "after_cookie",
			},
			"after_cookie": {
				ID:      "after_cookie",
				Type:    NodeChoice,
				Speaker: "The Oracle",
				Text:    "Now, what's really on your mind?",
				Choices: []Choice{
					{Text: "Am I The One?", NextNodeID: "the_one"},
					{Text: "What about Morpheus?", NextNodeID: "morpheus_fate"},
					{Text: "I need to go.", NextNodeID: "farewell"},
				},
			},
			"future": {
				ID:       "future",
				Type:     NodeText,
				Speaker:  "The Oracle",
				Text:     "I only tell you what you need to hear. The future depends on your choices.",
				NextNode: "after_cookie",
			},
			"the_one": {
				ID:       "the_one",
				Type:     NodeText,
				Speaker:  "The Oracle",
				Text:     "Being The One is just like being in love. No one can tell you you're in love, you just know it.",
				NextNode: "the_one_continue",
			},
			"the_one_continue": {
				ID:      "the_one_continue",
				Type:    NodeChoice,
				Speaker: "The Oracle",
				Text:    "You've already made the choice. Now you have to understand it.",
				Choices: []Choice{
					{Text: "What choice?", NextNodeID: "choice_explain"},
					{Text: "I understand.", NextNodeID: "farewell"},
				},
			},
			"choice_explain": {
				ID:       "choice_explain",
				Type:     NodeText,
				Speaker:  "The Oracle",
				Text:     "You'll know when the time comes. Trust yourself.",
				NextNode: "after_cookie",
			},
			"morpheus_fate": {
				ID:       "morpheus_fate",
				Type:     NodeText,
				Speaker:  "The Oracle",
				Text:     "Without him, we are lost. He believes in you. That belief might cost him his life.",
				NextNode: "morpheus_warn",
			},
			"morpheus_warn": {
				ID:      "morpheus_warn",
				Type:    NodeChoice,
				Speaker: "The Oracle",
				Text:    "You will have to make a choice. In one hand, your life. In the other, his.",
				Choices: []Choice{
					{Text: "I won't let him die.", NextNodeID: "determination"},
					{Text: "What can I do?", NextNodeID: "advice"},
				},
			},
			"determination": {
				ID:       "determination",
				Type:     NodeText,
				Speaker:  "The Oracle",
				Text:     "Good. That's exactly what you needed to hear.",
				NextNode: "farewell",
			},
			"advice": {
				ID:       "advice",
				Type:     NodeText,
				Speaker:  "The Oracle",
				Text:     "Trust your instincts. You have more power than you know.",
				NextNode: "farewell",
			},
			"farewell": {
				ID:      "farewell",
				Type:    NodeEnd,
				Speaker: "The Oracle",
				Text:    "Make a believer out of me.",
				Action:  &ChoiceAction{Type: "complete_objective", Target: "talk_oracle"},
			},
		},
	}

	// Architect dialogue tree
	m.Trees["architect"] = &Tree{
		ID:       "architect_main",
		NPCID:    "architect",
		NPCName:  "The Architect",
		RootNode: "greeting",
		Nodes: map[string]*Node{
			"greeting": {
				ID:       "greeting",
				Type:     NodeText,
				Speaker:  "The Architect",
				Text:     "Hello, Neo. I am the Architect. I created the Matrix. I've been waiting for you.",
				NextNode: "purpose",
			},
			"purpose": {
				ID:      "purpose",
				Type:    NodeChoice,
				Speaker: "The Architect",
				Text:    "Your life is the sum of a remainder of an unbalanced equation. You are the eventuality of an anomaly.",
				Choices: []Choice{
					{Text: "What anomaly?", NextNodeID: "anomaly"},
					{Text: "Get to the point.", NextNodeID: "point"},
				},
			},
			"anomaly": {
				ID:       "anomaly",
				Type:     NodeText,
				Speaker:  "The Architect",
				Text:     "The function of The One is to return to the Source, allowing a dissemination of the code you carry.",
				NextNode: "choice_intro",
			},
			"point": {
				ID:       "point",
				Type:     NodeText,
				Speaker:  "The Architect",
				Text:     "Hmph. As you wish.",
				NextNode: "choice_intro",
			},
			"choice_intro": {
				ID:      "choice_intro",
				Type:    NodeChoice,
				Speaker: "The Architect",
				Text:    "The door to your left leads to the Source and salvation of Zion. The door to your right leads back to the Matrix, to her.",
				Choices: []Choice{
					{Text: "[LEFT DOOR] Save Zion.", NextNodeID: "left_door"},
					{Text: "[RIGHT DOOR] Save Trinity.", NextNodeID: "right_door"},
					{Text: "Why should I choose?", NextNodeID: "why_choose"},
				},
			},
			"why_choose": {
				ID:       "why_choose",
				Type:     NodeText,
				Speaker:  "The Architect",
				Text:     "Hope. It is the quintessential human delusion.",
				NextNode: "choice_intro",
			},
			"left_door": {
				ID:       "left_door",
				Type:     NodeAction,
				Speaker:  "The Architect",
				Text:     "You choose to fulfill the function of The One. Zion will survive... for now.",
				NextNode: "ending_source",
				Action:   &ChoiceAction{Type: "complete_objective", Target: "talk_architect"},
			},
			"right_door": {
				ID:       "right_door",
				Type:     NodeAction,
				Speaker:  "The Architect",
				Text:     "Denial is the most predictable of all human responses.",
				NextNode: "ending_trinity",
				Action:   &ChoiceAction{Type: "complete_objective", Target: "talk_architect"},
			},
			"ending_source": {
				ID:      "ending_source",
				Type:    NodeEnd,
				Speaker: "The Architect",
				Text:    "The One has returned to the Source. The cycle begins anew.",
			},
			"ending_trinity": {
				ID:      "ending_trinity",
				Type:    NodeEnd,
				Speaker: "The Architect",
				Text:    "Go. Save your Trinity. We shall see what happens next.",
			},
		},
	}

	// Merovingian dialogue tree
	m.Trees["merovingian"] = &Tree{
		ID:       "merovingian_main",
		NPCID:    "merovingian",
		NPCName:  "The Merovingian",
		RootNode: "greeting",
		Nodes: map[string]*Node{
			"greeting": {
				ID:      "greeting",
				Type:    NodeChoice,
				Speaker: "The Merovingian",
				Text:    "Ah, here at last. Neo, The One himself, right? I must say I am surprised to see you.",
				Choices: []Choice{
					{Text: "I want the Keymaker.", NextNodeID: "keymaker"},
					{Text: "Who are you?", NextNodeID: "introduction"},
					{Text: "I'm leaving.", NextNodeID: "leave_hostile"},
				},
			},
			"introduction": {
				ID:       "introduction",
				Type:     NodeText,
				Speaker:  "The Merovingian",
				Text:     "I am a trafficker of information. I know everything I can.",
				NextNode: "cause_effect",
			},
			"cause_effect": {
				ID:      "cause_effect",
				Type:    NodeChoice,
				Speaker: "The Merovingian",
				Text:    "There is only one constant. One universal truth. Causality. Action, reaction. Cause and effect.",
				Choices: []Choice{
					{Text: "I need the Keymaker.", NextNodeID: "keymaker"},
					{Text: "Tell me more.", NextNodeID: "why"},
				},
			},
			"why": {
				ID:       "why",
				Type:     NodeText,
				Speaker:  "The Merovingian",
				Text:     "Choice is an illusion created between those with power and those without.",
				NextNode: "deal",
			},
			"keymaker": {
				ID:      "keymaker",
				Type:    NodeChoice,
				Speaker: "The Merovingian",
				Text:    "Ah yes, the Keymaker. A wonderful program. But what do you offer in return?",
				Choices: []Choice{
					{Text: "I'll do a job for you.", NextNodeID: "deal"},
					{Text: "I'll take him by force.", NextNodeID: "force"},
				},
			},
			"deal": {
				ID:      "deal",
				Type:    NodeChoice,
				Speaker: "The Merovingian",
				Text:    "Now we speak the same language. Bring me certain digital artifacts, and we can do business.",
				Choices: []Choice{
					{Text: "What artifacts?", NextNodeID: "artifacts"},
					{Text: "I'll find another way.", NextNodeID: "leave_neutral"},
				},
			},
			"artifacts": {
				ID:       "artifacts",
				Type:     NodeAction,
				Speaker:  "The Merovingian",
				Text:     "Bring me 10 pieces of Digital Trash. Prove yourself useful.",
				NextNode: "farewell_deal",
				Action:   &ChoiceAction{Type: "start_quest", Target: "exile_business"},
			},
			"farewell_deal": {
				ID:      "farewell_deal",
				Type:    NodeEnd,
				Speaker: "The Merovingian",
				Text:    "Au revoir. Do not keep me waiting.",
			},
			"force": {
				ID:       "force",
				Type:     NodeText,
				Speaker:  "The Merovingian",
				Text:     "Oh my, how... American. Do you really think you can defeat my Twins?",
				NextNode: "leave_hostile",
			},
			"leave_hostile": {
				ID:      "leave_hostile",
				Type:    NodeEnd,
				Speaker: "The Merovingian",
				Text:    "Then we have nothing more to discuss. My Twins will show you out.",
			},
			"leave_neutral": {
				ID:      "leave_neutral",
				Type:    NodeEnd,
				Speaker: "The Merovingian",
				Text:    "A pity. The path without me is much harder. Au revoir.",
			},
		},
	}
}

// StartDialogue begins a dialogue with an NPC for a player
func (m *Manager) StartDialogue(playerName, npcID string) (*Node, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	tree, ok := m.Trees[npcID]
	if !ok {
		return nil, nil // No dialogue tree for this NPC
	}

	// Create new session
	session := &Session{
		TreeID:      tree.ID,
		CurrentNode: tree.RootNode,
		History:     []string{tree.RootNode},
	}
	m.Sessions[strings.ToLower(playerName)] = session

	return tree.Nodes[tree.RootNode], nil
}

// SelectChoice processes a player's dialogue choice
func (m *Manager) SelectChoice(playerName string, choiceIndex int) (*Node, *ChoiceAction, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := strings.ToLower(playerName)
	session, ok := m.Sessions[name]
	if !ok {
		return nil, nil, nil // No active dialogue
	}

	// Find tree and current node
	var tree *Tree
	for _, t := range m.Trees {
		if t.ID == session.TreeID {
			tree = t
			break
		}
	}
	if tree == nil {
		delete(m.Sessions, name)
		return nil, nil, nil
	}

	currentNode := tree.Nodes[session.CurrentNode]
	if currentNode == nil {
		delete(m.Sessions, name)
		return nil, nil, nil
	}

	// Get next node based on choice or automatic progression
	var nextNodeID string
	var action *ChoiceAction

	switch currentNode.Type {
	case NodeChoice:
		if choiceIndex < 0 || choiceIndex >= len(currentNode.Choices) {
			return currentNode, nil, nil // Invalid choice, stay on same node
		}
		choice := currentNode.Choices[choiceIndex]
		nextNodeID = choice.NextNodeID
		action = choice.Action
	case NodeText, NodeAction:
		nextNodeID = currentNode.NextNode
		action = currentNode.Action
	case NodeEnd:
		// End dialogue
		delete(m.Sessions, name)
		return nil, currentNode.Action, nil
	}

	// Move to next node
	if nextNodeID == "" {
		delete(m.Sessions, name)
		return nil, action, nil
	}

	nextNode := tree.Nodes[nextNodeID]
	if nextNode == nil {
		delete(m.Sessions, name)
		return nil, action, nil
	}

	session.CurrentNode = nextNodeID
	session.History = append(session.History, nextNodeID)

	// Check if next node is an end node
	if nextNode.Type == NodeEnd {
		delete(m.Sessions, name)
		return nextNode, nextNode.Action, nil
	}

	return nextNode, action, nil
}

// GetCurrentNode returns the current dialogue node for a player
func (m *Manager) GetCurrentNode(playerName string) *Node {
	m.mu.RLock()
	defer m.mu.RUnlock()

	name := strings.ToLower(playerName)
	session, ok := m.Sessions[name]
	if !ok {
		return nil
	}

	for _, tree := range m.Trees {
		if tree.ID == session.TreeID {
			return tree.Nodes[session.CurrentNode]
		}
	}
	return nil
}

// EndDialogue terminates a player's dialogue session
func (m *Manager) EndDialogue(playerName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Sessions, strings.ToLower(playerName))
}

// IsInDialogue checks if a player is currently in dialogue
func (m *Manager) IsInDialogue(playerName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.Sessions[strings.ToLower(playerName)]
	return ok
}

// FormatNode formats a dialogue node for display
func FormatNode(node *Node) string {
	if node == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\r\n")
	sb.WriteString(node.Speaker + ": \"" + node.Text + "\"\r\n")

	if node.Type == NodeChoice && len(node.Choices) > 0 {
		sb.WriteString("\r\n")
		for i, choice := range node.Choices {
			sb.WriteString("  [" + string(rune('1'+i)) + "] " + choice.Text + "\r\n")
		}
		sb.WriteString("\r\nType a number to choose, or 'bye' to end conversation.\r\n")
	}

	return sb.String()
}

// Global dialogue manager
var GlobalDialogue = NewManager()
