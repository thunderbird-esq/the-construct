package main

// content_expansion.go - Commands for dialogue, instances, and enhanced content
// Part of Option A: Content Expansion implementation

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/yourusername/matrix-mud/pkg/dialogue"
	"github.com/yourusername/matrix-mud/pkg/events"
	"github.com/yourusername/matrix-mud/pkg/instance"
	"github.com/yourusername/matrix-mud/pkg/quest"
)

// HandleTalkCommand handles the "talk" command for NPC dialogue
func HandleTalkCommand(world *World, player *Player, arg string) string {
	if arg == "" {
		return "Talk to whom?\r\n"
	}

	// Check if player is already in dialogue
	if dialogue.GlobalDialogue.IsInDialogue(player.Name) {
		return "You are already in conversation. Type a number to respond or 'bye' to end.\r\n"
	}

	// Find NPC in current room
	room := world.Rooms[player.RoomID]
	if room == nil {
		return "Error: Invalid room.\r\n"
	}

	argLower := strings.ToLower(arg)
	var targetNPC *NPC
	for _, npc := range room.NPCMap {
		if npc == nil || npc.IsDead {
			continue
		}
		if strings.Contains(strings.ToLower(npc.Name), argLower) ||
			strings.Contains(strings.ToLower(npc.ID), argLower) {
			targetNPC = npc
			break
		}
	}

	if targetNPC == nil {
		return fmt.Sprintf("There is no '%s' here to talk to.\r\n", arg)
	}

	// Start dialogue
	node, err := dialogue.GlobalDialogue.StartDialogue(player.Name, strings.ToLower(targetNPC.ID))
	if err != nil {
		return fmt.Sprintf("Error starting dialogue: %v\r\n", err)
	}
	if node == nil {
		// No dialogue tree, use simple dialogue
		return handleSimpleDialogue(world, player, targetNPC, "")
	}

	return dialogue.FormatNode(node)
}

// HandleDialogueChoice handles numbered choices in dialogue
func HandleDialogueChoice(world *World, player *Player, choiceStr string) string {
	if !dialogue.GlobalDialogue.IsInDialogue(player.Name) {
		return "" // Not in dialogue, let normal command processing happen
	}

	// Parse choice number
	choice, err := strconv.Atoi(choiceStr)
	if err != nil {
		return "Please enter a number to choose, or 'bye' to end conversation.\r\n"
	}

	// Convert to 0-indexed
	choice--

	node, action, _ := dialogue.GlobalDialogue.SelectChoice(player.Name, choice)

	// Handle any actions
	if action != nil {
		handleDialogueAction(world, player, action)
	}

	if node == nil {
		// Dialogue ended
		return "The conversation ends.\r\n"
	}

	// For text nodes, auto-advance to next
	if node.Type == dialogue.NodeText {
		output := dialogue.FormatNode(node)
		// Auto-advance
		nextNode, nextAction, _ := dialogue.GlobalDialogue.SelectChoice(player.Name, 0)
		if nextAction != nil {
			handleDialogueAction(world, player, nextAction)
		}
		if nextNode != nil {
			output += dialogue.FormatNode(nextNode)
		} else {
			output += "\r\nThe conversation ends.\r\n"
		}
		return output
	}

	return dialogue.FormatNode(node)
}

// HandleByeCommand handles ending dialogue
func HandleByeCommand(player *Player) string {
	if dialogue.GlobalDialogue.IsInDialogue(player.Name) {
		dialogue.GlobalDialogue.EndDialogue(player.Name)
		return "You end the conversation.\r\n"
	}
	return "" // Not in dialogue
}

// handleDialogueAction processes actions triggered by dialogue choices
func handleDialogueAction(world *World, player *Player, action *dialogue.ChoiceAction) {
	if action == nil {
		return
	}

	switch action.Type {
	case "give_item":
		// Give item to player
		if item := world.GetItemTemplate(action.Target); item != nil {
			newItem := *item
			newItem.ID = fmt.Sprintf("%s_%d", item.ID, len(player.Inventory))
			player.Inventory = append(player.Inventory, &newItem)
		}
	case "take_item":
		// Remove item from player
		for i, item := range player.Inventory {
			if item.ID == action.Target || strings.Contains(item.ID, action.Target) {
				player.Inventory = append(player.Inventory[:i], player.Inventory[i+1:]...)
				break
			}
		}
	case "start_quest":
		// Start a quest
		quest.GlobalQuests.StartQuest(player.Name, action.Target)
	case "complete_objective":
		// Complete quest objective
		quest.GlobalQuests.UpdateProgress(player.Name, "talk", action.Target, 1)
	case "reputation":
		// Modify faction reputation
		// TODO: Integrate with faction system
	}
}

// handleSimpleDialogue handles NPCs without full dialogue trees
func handleSimpleDialogue(world *World, player *Player, npc *NPC, topic string) string {
	// Use the existing dialogue.json data
	dialogueMap := world.GetNPCDialogue(npc.ID)
	if dialogueMap == nil {
		return fmt.Sprintf("%s has nothing to say.\r\n", npc.Name)
	}

	// Get hello/greeting by default
	if greeting, ok := dialogueMap["hello"]; ok {
		return fmt.Sprintf("\r\n%s: \"%s\"\r\n", npc.Name, greeting)
	}

	return fmt.Sprintf("%s nods at you.\r\n", npc.Name)
}

// HandleInstanceCommand handles the "instance" command
func HandleInstanceCommand(world *World, player *Player, args string) string {
	parts := strings.Fields(args)
	if len(parts) == 0 {
		return instanceHelp()
	}

	subCmd := strings.ToLower(parts[0])
	subArg := ""
	if len(parts) > 1 {
		subArg = strings.Join(parts[1:], " ")
	}

	switch subCmd {
	case "list":
		return instance.GlobalInstance.ListTemplates(player.Level)

	case "create", "start":
		if subArg == "" {
			return "Usage: instance create <template_id>\r\n" + instance.GlobalInstance.ListTemplates(player.Level)
		}
		inst, errMsg := instance.GlobalInstance.CreateInstance(subArg, player.Name, player.Level)
		if errMsg != "" {
			return errMsg + "\r\n"
		}
		return fmt.Sprintf("=== ENTERING %s ===\r\n\r\n%s\r\n", inst.Name, instance.GlobalInstance.LookInInstance(player.Name))

	case "leave", "exit":
		result := instance.GlobalInstance.LeaveInstance(player.Name)
		return result + "\r\n"

	case "look":
		look := instance.GlobalInstance.LookInInstance(player.Name)
		if look == "" {
			return "You are not in an instance.\r\n"
		}
		return look

	case "rewards", "claim":
		rewards, errMsg := instance.GlobalInstance.GetRewards(player.Name)
		if errMsg != "" {
			return errMsg + "\r\n"
		}
		// Apply rewards
		var sb strings.Builder
		sb.WriteString("=== INSTANCE REWARDS ===\r\n\r\n")
		if rewards.XP > 0 {
			player.XP += rewards.XP
			sb.WriteString(fmt.Sprintf("  +%d XP\r\n", rewards.XP))
			// Check level up
			checkLevelUp(player)
		}
		if rewards.Money > 0 {
			player.Money += rewards.Money
			sb.WriteString(fmt.Sprintf("  +%d bits\r\n", rewards.Money))
		}
		for _, itemID := range rewards.Items {
			if item := world.GetItemTemplate(itemID); item != nil {
				newItem := *item
				newItem.ID = fmt.Sprintf("%s_%d", item.ID, len(player.Inventory))
				player.Inventory = append(player.Inventory, &newItem)
				sb.WriteString(fmt.Sprintf("  +%s\r\n", item.Name))
			}
		}
		if rewards.Title != "" {
			// Add title to player (would need player.Titles field)
			sb.WriteString(fmt.Sprintf("  +Title: %s\r\n", rewards.Title))
		}
		sb.WriteString("\r\nCongratulations!\r\n")
		return sb.String()

	default:
		return instanceHelp()
	}
}

// instanceHelp returns help text for instance commands
func instanceHelp() string {
	return `=== INSTANCE COMMANDS ===

instance list          - Show available instances
instance create <id>   - Enter an instance
instance look          - Look around in current instance
instance leave         - Leave current instance
instance rewards       - Claim rewards after completion

Instances are special challenge areas with their own maps and enemies.
Complete all rooms and defeat the boss to earn rewards!
`
}

// HandleInstanceMove handles movement within an instance
func HandleInstanceMove(player *Player, direction string) (string, bool) {
	if !instance.GlobalInstance.IsInInstance(player.Name) {
		return "", false // Not in instance, use normal movement
	}

	result, ok := instance.GlobalInstance.MoveInInstance(player.Name, direction)
	if ok {
		return result + "\r\n" + instance.GlobalInstance.LookInInstance(player.Name), true
	}
	return result + "\r\n", true
}

// HandleInstanceAttack handles attacks within an instance
func HandleInstanceAttack(world *World, player *Player, target string) (string, bool) {
	if !instance.GlobalInstance.IsInInstance(player.Name) {
		return "", false // Not in instance, use normal combat
	}

	// Calculate player damage
	damage := player.Strength
	if player.Equipment != nil && player.Equipment["hand"] != nil {
		damage += player.Equipment["hand"].Damage
	}

	result, ok := instance.GlobalInstance.AttackInInstance(player.Name, target, damage)
	return result + "\r\n", ok
}

// HandleInstanceLook handles looking within an instance
func HandleInstanceLook(player *Player) (string, bool) {
	if !instance.GlobalInstance.IsInInstance(player.Name) {
		return "", false // Not in instance
	}

	return instance.GlobalInstance.LookInInstance(player.Name), true
}

// IsInDialogue checks if player is in dialogue mode
func IsInDialogue(playerName string) bool {
	return dialogue.GlobalDialogue.IsInDialogue(playerName)
}

// IsInInstance checks if player is in an instance
func IsInInstance(playerName string) bool {
	return instance.GlobalInstance.IsInInstance(playerName)
}

// checkLevelUp checks if player should level up
func checkLevelUp(player *Player) {
	xpForLevel := player.Level * 100
	for player.XP >= xpForLevel {
		player.Level++
		player.MaxHP += 10
		player.HP = player.MaxHP

		// Emit level up event
		events.GlobalEventBus.Publish(&events.Event{
			Type:       events.EventPlayerLevelUp,
			Timestamp:  time.Now(),
			PlayerName: player.Name,
		})
		xpForLevel = player.Level * 100
	}
}
