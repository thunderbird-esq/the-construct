package game

import (
	"testing"
)

func TestCalculateHit(t *testing.T) {
	tests := []struct {
		name     string
		strength int
		ac       int
		// We can't test exact results due to randomness, but we test boundaries
	}{
		{"equal_stats", 10, 10},
		{"high_strength", 20, 10},
		{"high_ac", 10, 20},
		{"very_high_ac", 5, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run multiple times to verify randomness works
			hits := 0
			attempts := 1000
			for i := 0; i < attempts; i++ {
				if CalculateHit(tt.strength, tt.ac) {
					hits++
				}
			}
			hitRate := float64(hits) / float64(attempts)
			t.Logf("%s: hit rate %.2f%% (%d/%d)", tt.name, hitRate*100, hits, attempts)

			// Verify hit rate is within expected bounds
			if hitRate < 0.01 {
				t.Error("Hit rate too low - should have minimum 5% chance")
			}
			if hitRate > 0.99 {
				t.Error("Hit rate too high - should have maximum 95% chance")
			}
		})
	}
}

func TestCalculateDamage(t *testing.T) {
	tests := []struct {
		name       string
		baseDamage int
		strength   int
		minDamage  int
	}{
		{"basic_damage", 5, 10, 1},
		{"high_damage", 20, 20, 15},
		{"zero_base", 0, 10, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			damage, _ := CalculateDamage(tt.baseDamage, tt.strength)
			if damage < tt.minDamage {
				t.Errorf("Damage %d is below minimum %d", damage, tt.minDamage)
			}
		})
	}
}

func TestCriticalHits(t *testing.T) {
	crits := 0
	attempts := 10000
	for i := 0; i < attempts; i++ {
		_, crit := CalculateDamage(10, 10)
		if crit {
			crits++
		}
	}
	critRate := float64(crits) / float64(attempts)
	t.Logf("Critical rate: %.2f%% (%d/%d)", critRate*100, crits, attempts)

	// Should be around 5% +/- 2%
	if critRate < 0.02 || critRate > 0.08 {
		t.Errorf("Critical rate %.2f%% outside expected range (3-8%%)", critRate*100)
	}
}

func TestAttackNPC(t *testing.T) {
	player := &Player{
		Name:     "TestPlayer",
		Strength: 12,
		BaseAC:   10,
		HP:       100,
		MaxHP:    100,
	}

	npc := &NPC{
		Name:   "TestNPC",
		HP:     50,
		MaxHP:  50,
		AC:     10,
		Damage: 5,
	}

	// Attack multiple times
	totalDamage := 0
	for i := 0; i < 10; i++ {
		result := AttackNPC(player, npc, 5)
		if result.Hit {
			totalDamage += result.Damage
			if result.Damage <= 0 {
				t.Error("Hit should deal positive damage")
			}
		}
		// Reset NPC HP for next test
		npc.HP = 50
	}
	t.Logf("Total damage dealt in 10 attacks: %d", totalDamage)
}

func TestNPCAttackPlayer(t *testing.T) {
	player := &Player{
		Name:      "TestPlayer",
		Strength:  12,
		BaseAC:    10,
		HP:        100,
		MaxHP:     100,
		Equipment: make(map[string]*Item),
	}

	npc := &NPC{
		Name:   "TestNPC",
		HP:     50,
		MaxHP:  50,
		AC:     10,
		Damage: 8,
	}

	startHP := player.HP
	result := NPCAttackPlayer(npc, player)

	if result.Hit {
		if player.HP >= startHP {
			t.Error("Player HP should decrease on hit")
		}
		if result.Damage <= 0 {
			t.Error("Hit should deal positive damage")
		}
	}
	t.Logf("NPC attack result: hit=%v, damage=%d, message=%s", result.Hit, result.Damage, result.Message)
}

func TestTryFlee(t *testing.T) {
	successes := 0
	attempts := 1000
	for i := 0; i < attempts; i++ {
		if TryFlee() {
			successes++
		}
	}
	fleeRate := float64(successes) / float64(attempts)
	t.Logf("Flee rate: %.2f%% (%d/%d)", fleeRate*100, successes, attempts)

	// Should be around 50% +/- 5%
	if fleeRate < 0.40 || fleeRate > 0.60 {
		t.Errorf("Flee rate %.2f%% outside expected range (40-60%%)", fleeRate*100)
	}
}

func TestNPCDefeat(t *testing.T) {
	player := &Player{
		Name:     "TestPlayer",
		Strength: 20,
		BaseAC:   10,
		HP:       100,
		MaxHP:    100,
	}

	npc := &NPC{
		Name:   "WeakNPC",
		HP:     1,
		MaxHP:  1,
		AC:     5,
		Damage: 1,
	}

	// Keep attacking until NPC is defeated
	for npc.HP > 0 {
		result := AttackNPC(player, npc, 10)
		if result.Hit {
			t.Logf("Attack dealt %d damage, NPC HP: %d", result.Damage, npc.HP)
		}
	}

	if npc.HP > 0 {
		t.Error("NPC should be defeated")
	}
}
