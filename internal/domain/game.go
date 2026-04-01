package domain

// Statistics tracks one complete game run.
type Statistics struct {
	RunID          int    `json:"run_id"`
	DeepestLevel   int    `json:"deepest_level"`
	TreasureTotal  int    `json:"treasure_total"`
	EnemiesKilled  int    `json:"enemies_killed"`
	FoodEaten      int    `json:"food_eaten"`
	ElixirsDrunk   int    `json:"elixirs_drunk"`
	ScrollsRead    int    `json:"scrolls_read"`
	HitsDealt      int    `json:"hits_dealt"`
	HitsReceived   int    `json:"hits_received"`
	TilesWalked    int    `json:"tiles_walked"`
	Completed      bool   `json:"completed"` // true = player finished all 21 levels
}

const TotalLevels = 21

// GameSession holds the complete state of an ongoing run.
type GameSession struct {
	ID          int
	CurrentLevel int
	Level       *Level
	Player      *Character
	Stats       *Statistics
	Messages    []string // recent log messages shown in the UI
	GameOver    bool
	Won         bool
	Paused      bool     // in menu / item selection
}

// NewGameSession creates a fresh game session.
func NewGameSession(id int) *GameSession {
	return &GameSession{
		ID:           id,
		CurrentLevel: 1,
		Player:       NewCharacter(),
		Stats:        &Statistics{RunID: id},
		Messages:     make([]string, 0),
	}
}

// AddMessage appends a message to the log, keeping only the last 20.
func (gs *GameSession) AddMessage(msg string) {
	if msg == "" {
		return
	}
	gs.Messages = append(gs.Messages, msg)
	if len(gs.Messages) > 20 {
		gs.Messages = gs.Messages[len(gs.Messages)-20:]
	}
}

// SyncStatsFromPlayer copies live character counters into the session stats.
func (gs *GameSession) SyncStatsFromPlayer() {
	p := gs.Player
	s := gs.Stats
	s.EnemiesKilled = p.EnemiesKilled
	s.FoodEaten = p.FoodEaten
	s.ElixirsDrunk = p.ElixirsDrunk
	s.ScrollsRead = p.ScrollsRead
	s.HitsDealt = p.HitsDealt
	s.HitsReceived = p.HitsReceived
	s.TilesWalked = p.TilesWalked
	s.TreasureTotal = p.Pack.Treasure
	if gs.CurrentLevel > s.DeepestLevel {
		s.DeepestLevel = gs.CurrentLevel
	}
}
