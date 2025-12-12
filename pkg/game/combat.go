// Package game implements core game mechanics for Matrix MUD.
package game

import (
	"fmt"
	"math/rand"
)

// Combat constants
const (
	CombatTickRate  = 500 // milliseconds between combat rounds
	BaseHitChance   = 65  // percent chance to hit with equal AC
	CriticalChance  = 5   // percent chance for critical hit
	CriticalMulti   = 2   // damage multiplier for critical hits
	FleeChance      = 50  // percent chance to successfully flee
	RespawnTime     = 60  // seconds until NPC respawns
)

// CombatResult represents the outcome of a combat action
type CombatResult struct {
	Hit      bool
	Critical bool
	Damage   int
	Message  string
}

// CalculateHit determines if an attack hits based on attacker strength and defender AC
func CalculateHit(attackerStrength, defenderAC int) bool {
	// Base hit chance modified by AC difference
	hitChance := BaseHitChance + (attackerStrength - defenderAC)
	if hitChance < 5 {
		hitChance = 5 // Minimum 5% hit chance
	}
	if hitChance > 95 {
		hitChance = 95 // Maximum 95% hit chance
	}
	return rand.Intn(100) < hitChance
}

// CalculateDamage computes damage dealt by an attack
func CalculateDamage(baseDamage, strength int) (damage int, critical bool) {
	// Base damage plus strength modifier
	damage = baseDamage + (strength / 4)
	if damage < 1 {
		damage = 1
	}

	// Check for critical hit
	if rand.Intn(100) < CriticalChance {
		critical = true
		damage *= CriticalMulti
	}

	// Add some variance (+/- 20%)
	variance := damage / 5
	if variance > 0 {
		damage += rand.Intn(variance*2+1) - variance
	}

	return damage, critical
}

// AttackNPC performs a player attack against an NPC
func AttackNPC(player *Player, npc *NPC, weaponDamage int) CombatResult {
	result := CombatResult{}

	// Calculate player's effective damage
	playerDamage := weaponDamage
	if playerDamage == 0 {
		playerDamage = 1 + (player.Strength / 5) // Unarmed damage
	}

	// Check if attack hits
	if !CalculateHit(player.Strength, npc.AC) {
		result.Hit = false
		result.Message = fmt.Sprintf("You swing at %s but miss!", npc.Name)
		return result
	}

	// Calculate damage
	damage, critical := CalculateDamage(playerDamage, player.Strength)
	result.Hit = true
	result.Critical = critical
	result.Damage = damage

	// Apply damage to NPC
	npc.HP -= damage

	if critical {
		result.Message = fmt.Sprintf("CRITICAL HIT! You strike %s for %d damage!", npc.Name, damage)
	} else {
		result.Message = fmt.Sprintf("You hit %s for %d damage.", npc.Name, damage)
	}

	if npc.HP <= 0 {
		result.Message += fmt.Sprintf("\r\n%s has been defeated!", npc.Name)
	}

	return result
}

// NPCAttackPlayer performs an NPC attack against a player
func NPCAttackPlayer(npc *NPC, player *Player) CombatResult {
	result := CombatResult{}

	// Calculate player's effective AC
	playerAC := player.BaseAC
	for _, item := range player.Equipment {
		if item != nil {
			playerAC += item.AC
		}
	}

	// Check if attack hits
	if !CalculateHit(npc.Damage, playerAC) {
		result.Hit = false
		result.Message = fmt.Sprintf("%s swings at you but misses!", npc.Name)
		return result
	}

	// Calculate damage (NPCs use Damage stat as base)
	damage, critical := CalculateDamage(npc.Damage, npc.Damage)
	result.Hit = true
	result.Critical = critical
	result.Damage = damage

	// Apply damage to player
	player.HP -= damage

	if critical {
		result.Message = fmt.Sprintf("CRITICAL! %s strikes you for %d damage!", npc.Name, damage)
	} else {
		result.Message = fmt.Sprintf("%s hits you for %d damage.", npc.Name, damage)
	}

	if player.HP <= 0 {
		result.Message += "\r\nYou have been defeated!"
	}

	return result
}

// TryFlee attempts to flee from combat
func TryFlee() bool {
	return rand.Intn(100) < FleeChance
}
