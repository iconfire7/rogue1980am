package domain

// EnemyType represents the kind of enemy.
type EnemyType int

const (
	EnemyZombie    EnemyType = iota
	EnemyVampire
	EnemyGhost
	EnemyOgre
	EnemySnakeMage
)

// Enemy represents a dungeon monster.
type Enemy struct {
	ID              int
	Type            EnemyType
	Pos             Position
	Health          int
	MaxHealth       int
	Dexterity       int
	Strength        int
	Hostility       int  // aggro radius (tiles)
	Alive           bool
	// behavioural state
	Alerted         bool // is actively chasing player
	RestTurns       int  // turns remaining before the enemy can act (Ogre rest)
	Invisible       bool // Ghost invisibility
	InvisTimer      int  // countdown until next visibility change
	DiagDir         Position // Snake-Mage diagonal movement direction
	FirstHitTaken   bool // Vampire: first hit misses
	TeleportTimer   int  // Ghost teleport countdown
}

// EnemyConfig holds base stats per enemy type scaled by depth.
type EnemyConfig struct {
	BaseHealth    int
	BaseDexterity int
	BaseStrength  int
	BaseHostility int
}

var enemyConfigs = map[EnemyType]EnemyConfig{
	EnemyZombie:    {BaseHealth: 20, BaseDexterity: 2, BaseStrength: 5, BaseHostility: 5},
	EnemyVampire:   {BaseHealth: 18, BaseDexterity: 8, BaseStrength: 5, BaseHostility: 8},
	EnemyGhost:     {BaseHealth: 8,  BaseDexterity: 8, BaseStrength: 2, BaseHostility: 3},
	EnemyOgre:      {BaseHealth: 30, BaseDexterity: 2, BaseStrength: 12, BaseHostility: 5},
	EnemySnakeMage: {BaseHealth: 14, BaseDexterity: 10, BaseStrength: 4, BaseHostility: 7},
}

var enemyNames = map[EnemyType]string{
	EnemyZombie:    "Zombie",
	EnemyVampire:   "Vampire",
	EnemyGhost:     "Ghost",
	EnemyOgre:      "Ogre",
	EnemySnakeMage: "Snake-Mage",
}

// Name returns the display name of the enemy type.
func (et EnemyType) Name() string {
	if n, ok := enemyNames[et]; ok {
		return n
	}
	return "Unknown"
}

// NewEnemy creates an enemy of the given type scaled to dungeon depth.
func NewEnemy(id int, et EnemyType, depth int, pos Position) *Enemy {
	cfg := enemyConfigs[et]
	scale := 1 + (depth-1)/5 // boost every 5 levels
	hp := cfg.BaseHealth + depth*2 + scale*3
	dex := cfg.BaseDexterity + depth/4
	str := cfg.BaseStrength + depth/3
	host := cfg.BaseHostility + depth/5
	diag := Position{X: 1, Y: 1}
	return &Enemy{
		ID:         id,
		Type:       et,
		Pos:        pos,
		Health:     hp,
		MaxHealth:  hp,
		Dexterity:  dex,
		Strength:   str,
		Hostility:  host,
		Alive:      true,
		InvisTimer: 5,
		DiagDir:    diag,
		TeleportTimer: 3,
	}
}

// IsAlive returns true when the enemy is still alive.
func (e *Enemy) IsAlive() bool {
	return e.Alive && e.Health > 0
}

// Glyph returns the display rune for the enemy.
func (e *Enemy) Glyph() rune {
	switch e.Type {
	case EnemyZombie:
		return 'z'
	case EnemyVampire:
		return 'v'
	case EnemyGhost:
		return 'g'
	case EnemyOgre:
		return 'O'
	case EnemySnakeMage:
		return 's'
	}
	return '?'
}

// TreasureDrop returns the gold value dropped when this enemy dies.
func (e *Enemy) TreasureDrop() int {
	base := e.MaxHealth/4 + e.Strength/2 + e.Dexterity/2 + e.Hostility
	return base + 1
}
