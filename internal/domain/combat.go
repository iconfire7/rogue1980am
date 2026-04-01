package domain

import "math/rand"

// CombatResult carries the outcome of a single attack.
type CombatResult struct {
	Hit      bool
	Damage   int
	Missed   bool   // first-hit vampire miss
	Message  string
}

// hitChance returns the probability (0–100) for attacker to hit defender.
// Formula: base 50%, +3 per attacker dex point, −3 per defender dex point, clamped 5–95.
func hitChance(attackerDex, defenderDex int) int {
	p := 50 + attackerDex*3 - defenderDex*3
	if p < 5 {
		p = 5
	}
	if p > 95 {
		p = 95
	}
	return p
}

// rollDamage returns random damage based on attacker strength and optional weapon.
func rollDamage(str int, weapon *Item) int {
	base := 1 + rand.Intn(str)
	if weapon != nil {
		base += rand.Intn(weapon.Strength + 1)
	}
	return base
}

// CharacterAttacksEnemy resolves a player → enemy hit.
func CharacterAttacksEnemy(c *Character, e *Enemy) CombatResult {
	// Vampire: first hit always misses
	if e.Type == EnemyVampire && !e.FirstHitTaken {
		e.FirstHitTaken = true
		return CombatResult{Hit: false, Missed: true, Message: "Your attack phases through the Vampire!"}
	}
	chance := hitChance(c.Dexterity, e.Dexterity)
	if rand.Intn(100) >= chance {
		return CombatResult{Hit: false, Message: "You miss!"}
	}
	dmg := rollDamage(c.Strength, c.Weapon)
	e.Health -= dmg
	if !e.IsAlive() {
		e.Alive = false
	}
	c.HitsDealt++
	return CombatResult{Hit: true, Damage: dmg}
}

// EnemyAttacksCharacter resolves an enemy → player hit.
// Returns the CombatResult; special effects are applied in-place.
func EnemyAttacksCharacter(e *Enemy, c *Character) CombatResult {
	// Ogre: guaranteed counter-attack after rest (handled in AI)
	chance := hitChance(e.Dexterity, c.Dexterity)
	if rand.Intn(100) >= chance {
		return CombatResult{Hit: false, Message: e.Type.Name() + " misses you."}
	}
	dmg := rollDamage(e.Strength, nil)
	c.Health -= dmg
	c.HitsReceived++
	msg := ""
	// special per-type side effects
	switch e.Type {
	case EnemyVampire:
		// Vampire drains max health
		c.MaxHealth--
		if c.MaxHealth < 1 {
			c.MaxHealth = 1
		}
		if c.Health > c.MaxHealth {
			c.Health = c.MaxHealth
		}
		msg = "The Vampire drains your life force! Max HP reduced."
	case EnemySnakeMage:
		// Chance to put player to sleep
		if rand.Intn(100) < 35 {
			if c.SleepTurns == 0 {
				c.SleepTurns = 1
			}
			msg = "You feel drowsy... (sleep 1 turn)"
		}
	}
	return CombatResult{Hit: true, Damage: dmg, Message: msg}
}
