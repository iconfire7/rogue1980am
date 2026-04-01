package domain

// Character represents the player's hero.
type Character struct {
	Pos           Position
	MaxHealth     int
	Health        int
	Dexterity     int
	Strength      int
	Weapon        *Item // currently equipped weapon (nil = unarmed)
	Pack          *Backpack
	SleepTurns    int  // turns remaining while asleep
	ActiveEffects []*ActiveEffect
	// stats
	TilesWalked   int
	HitsDealt     int
	HitsReceived  int
	EnemiesKilled int
	FoodEaten     int
	ElixirsDrunk  int
	ScrollsRead   int
}

// ActiveEffect holds a temporary stat modifier applied by an elixir.
type ActiveEffect struct {
	Subtype  ItemSubtype
	Amount   int
	Turns    int
}

// NewCharacter creates a starting character with default stats.
func NewCharacter() *Character {
	return &Character{
		MaxHealth: 20,
		Health:    20,
		Dexterity: 5,
		Strength:  5,
		Pack:      NewBackpack(),
	}
}

// IsAlive returns true when the character is alive.
func (c *Character) IsAlive() bool {
	return c.Health > 0
}

// IsSleeping returns true when the character cannot act.
func (c *Character) IsSleeping() bool {
	return c.SleepTurns > 0
}

// TickEffects decrements active elixir timers and removes expired ones.
// If max-health drops and current health would fall to 0 or below it is clamped to 1.
func (c *Character) TickEffects() {
	remaining := c.ActiveEffects[:0]
	for _, e := range c.ActiveEffects {
		e.Turns--
		if e.Turns <= 0 {
			// reverse the effect
			switch e.Subtype {
			case SubtypeMaxHealth:
				c.MaxHealth -= e.Amount
				if c.Health > c.MaxHealth {
					c.Health = c.MaxHealth
				}
				if c.Health <= 0 {
					c.Health = 1
				}
			case SubtypeDexterity:
				c.Dexterity -= e.Amount
				if c.Dexterity < 1 {
					c.Dexterity = 1
				}
			case SubtypeStrength:
				c.Strength -= e.Amount
				if c.Strength < 1 {
					c.Strength = 1
				}
			}
		} else {
			remaining = append(remaining, e)
		}
	}
	c.ActiveEffects = remaining
	// tick sleep
	if c.SleepTurns > 0 {
		c.SleepTurns--
	}
}

// ApplyItem applies the effect of an item to the character.
// Returns a descriptive message.
func (c *Character) ApplyItem(item *Item) string {
	switch item.Type {
	case ItemFood:
		heal := item.Health
		c.Health += heal
		if c.Health > c.MaxHealth {
			c.Health = c.MaxHealth
		}
		c.FoodEaten++
		return "You eat the food and restore some health."
	case ItemElixir:
		c.ElixirsDrunk++
		switch item.Subtype {
		case SubtypeMaxHealth:
			c.MaxHealth += item.MaxHealth
			c.Health += item.MaxHealth
			if item.Duration > 0 {
				c.ActiveEffects = append(c.ActiveEffects, &ActiveEffect{
					Subtype: SubtypeMaxHealth, Amount: item.MaxHealth, Turns: item.Duration,
				})
				return "Your maximum health temporarily increases!"
			}
			return "Your maximum health increases!"
		case SubtypeDexterity:
			c.Dexterity += item.Dexterity
			if item.Duration > 0 {
				c.ActiveEffects = append(c.ActiveEffects, &ActiveEffect{
					Subtype: SubtypeDexterity, Amount: item.Dexterity, Turns: item.Duration,
				})
				return "Your dexterity temporarily increases!"
			}
			return "Your dexterity increases!"
		case SubtypeStrength:
			c.Strength += item.Strength
			if item.Duration > 0 {
				c.ActiveEffects = append(c.ActiveEffects, &ActiveEffect{
					Subtype: SubtypeStrength, Amount: item.Strength, Turns: item.Duration,
				})
				return "Your strength temporarily increases!"
			}
			return "Your strength increases!"
		}
	case ItemScroll:
		c.ScrollsRead++
		switch item.Subtype {
		case SubtypeMaxHealth:
			c.MaxHealth += item.MaxHealth
			c.Health += item.MaxHealth
			return "You read the scroll. Your max health permanently increases!"
		case SubtypeDexterity:
			c.Dexterity += item.Dexterity
			return "You read the scroll. Your dexterity permanently increases!"
		case SubtypeStrength:
			c.Strength += item.Strength
			return "You read the scroll. Your strength permanently increases!"
		}
	}
	return ""
}
