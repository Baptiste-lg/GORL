package game

import "math/rand"

// CombatResult holds the outcome of a single attack.
type CombatResult struct {
	Damage  int
	IsCrit  bool
	IsDodge bool
	IsDeath bool
}

// ResolveAttack calculates damage from attacker to defender.
func ResolveAttack(attacker, defender *Entity, rng *rand.Rand) CombatResult {
	result := CombatResult{}

	// Check dodge
	if rng.Float64() < defender.Stats.DodgeChance() {
		result.IsDodge = true
		return result
	}

	// Base damage
	damage := attacker.Stats.Damage() - defender.Stats.Defense()
	if damage < 1 {
		damage = 1
	}

	// Check critical hit
	if rng.Float64() < attacker.Stats.CritChance() {
		damage *= 2
		result.IsCrit = true
	}

	result.Damage = defender.TakeDamage(damage)
	result.IsDeath = !defender.IsAlive
	return result
}

// XPForKill returns XP gained from killing an enemy on a given floor.
func XPForKill(baseXP, floor int) int {
	return int(float64(baseXP) * (1.0 + float64(floor)*0.2))
}
