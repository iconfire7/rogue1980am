// Package storage handles persistence of game state and statistics to JSON.
package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"rogue1980am/internal/domain"
)

const (
	statsFile = "rogue_stats.json"
	saveFile  = "rogue_save.json"
)

// dataDir returns a writable directory for game data files.
func dataDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	d := filepath.Join(dir, "rogue1980am")
	_ = os.MkdirAll(d, 0o700)
	return d
}

// ---- Statistics / Leaderboard -------------------------------------------------

// SaveStatistics appends or updates stats for a completed/ended run.
func SaveStatistics(s *domain.Statistics) error {
	all, _ := LoadAllStatistics()
	// replace existing entry for the same run ID
	found := false
	for i, existing := range all {
		if existing.RunID == s.RunID {
			all[i] = s
			found = true
			break
		}
	}
	if !found {
		all = append(all, s)
	}
	return writeJSON(filepath.Join(dataDir(), statsFile), all)
}

// LoadAllStatistics reads all recorded run statistics sorted by treasure (desc).
func LoadAllStatistics() ([]*domain.Statistics, error) {
	var all []*domain.Statistics
	err := readJSON(filepath.Join(dataDir(), statsFile), &all)
	if err != nil {
		return nil, err
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].TreasureTotal > all[j].TreasureTotal
	})
	return all, nil
}

// ---- Save Game ----------------------------------------------------------------

// SaveData is the full serialisable game state.
type SaveData struct {
	SessionID    int                `json:"session_id"`
	LevelNumber  int                `json:"level_number"`
	Player       *PlayerSave        `json:"player"`
	Level        *LevelSave         `json:"level"`
	Stats        *domain.Statistics `json:"stats"`
}

// PlayerSave captures the serialisable portion of the player character.
type PlayerSave struct {
	PosX, PosY    int               `json:"pos_x,omitempty"`
	MaxHealth     int               `json:"max_health"`
	Health        int               `json:"health"`
	Dexterity     int               `json:"dexterity"`
	Strength      int               `json:"strength"`
	SleepTurns    int               `json:"sleep_turns"`
	Treasure      int               `json:"treasure"`
	WeaponID      int               `json:"weapon_id"` // 0 = unarmed
	Foods         []*ItemSave       `json:"foods"`
	Elixirs       []*ItemSave       `json:"elixirs"`
	Scrolls       []*ItemSave       `json:"scrolls"`
	Weapons       []*ItemSave       `json:"weapons"`
	ActiveEffects []*EffectSave     `json:"active_effects"`
	// counters
	TilesWalked   int `json:"tiles_walked"`
	HitsDealt     int `json:"hits_dealt"`
	HitsReceived  int `json:"hits_received"`
	EnemiesKilled int `json:"enemies_killed"`
	FoodEaten     int `json:"food_eaten"`
	ElixirsDrunk  int `json:"elixirs_drunk"`
	ScrollsRead   int `json:"scrolls_read"`
}

// EffectSave persists an active elixir effect.
type EffectSave struct {
	Subtype int `json:"subtype"`
	Amount  int `json:"amount"`
	Turns   int `json:"turns"`
}

// ItemSave is a compact item representation.
type ItemSave struct {
	ID        int `json:"id"`
	Type      int `json:"type"`
	Subtype   int `json:"subtype"`
	Name      string `json:"name"`
	Health    int `json:"health,omitempty"`
	MaxHealth int `json:"max_health,omitempty"`
	Dexterity int `json:"dexterity,omitempty"`
	Strength  int `json:"strength,omitempty"`
	Value     int `json:"value,omitempty"`
	Duration  int `json:"duration,omitempty"`
}

// LevelSave stores only the RNG seed and enemy/item positions — we re-generate
// from seed and then apply the saved positions so the layout is identical.
type LevelSave struct {
	Seed    int64        `json:"seed"`
	Enemies []*EnemySave `json:"enemies"`
	Items   []*ItemSave  `json:"items"`
}

// EnemySave stores a snapshot of an enemy.
type EnemySave struct {
	ID         int  `json:"id"`
	Type       int  `json:"type"`
	PosX, PosY int  `json:"pos_x,omitempty"`
	Health     int  `json:"health"`
	MaxHealth  int  `json:"max_health"`
	Dexterity  int  `json:"dexterity"`
	Strength   int  `json:"strength"`
	Hostility  int  `json:"hostility"`
	Alive      bool `json:"alive"`
}

// SaveGame persists the current game state to disk.
func SaveGame(gs *domain.GameSession, seed int64) error {
	sd := &SaveData{
		SessionID:   gs.ID,
		LevelNumber: gs.CurrentLevel,
		Player:      playerToSave(gs.Player),
		Level:       levelToSave(gs.Level, seed),
		Stats:       gs.Stats,
	}
	return writeJSON(filepath.Join(dataDir(), saveFile), sd)
}

// LoadGame reads the saved game state from disk.
func LoadGame() (*SaveData, error) {
	var sd SaveData
	if err := readJSON(filepath.Join(dataDir(), saveFile), &sd); err != nil {
		return nil, err
	}
	return &sd, nil
}

// HasSave returns true if a save file exists.
func HasSave() bool {
	_, err := os.Stat(filepath.Join(dataDir(), saveFile))
	return err == nil
}

// DeleteSave removes the save file (call on game-over or completion).
func DeleteSave() {
	_ = os.Remove(filepath.Join(dataDir(), saveFile))
}

// ---- conversion helpers -------------------------------------------------------

func playerToSave(p *domain.Character) *PlayerSave {
	ps := &PlayerSave{
		PosX:          p.Pos.X,
		PosY:          p.Pos.Y,
		MaxHealth:     p.MaxHealth,
		Health:        p.Health,
		Dexterity:     p.Dexterity,
		Strength:      p.Strength,
		SleepTurns:    p.SleepTurns,
		Treasure:      p.Pack.Treasure,
		TilesWalked:   p.TilesWalked,
		HitsDealt:     p.HitsDealt,
		HitsReceived:  p.HitsReceived,
		EnemiesKilled: p.EnemiesKilled,
		FoodEaten:     p.FoodEaten,
		ElixirsDrunk:  p.ElixirsDrunk,
		ScrollsRead:   p.ScrollsRead,
	}
	if p.Weapon != nil {
		ps.WeaponID = p.Weapon.ID
	}
	for _, it := range p.Pack.Foods {
		ps.Foods = append(ps.Foods, itemToSave(it))
	}
	for _, it := range p.Pack.Elixirs {
		ps.Elixirs = append(ps.Elixirs, itemToSave(it))
	}
	for _, it := range p.Pack.Scrolls {
		ps.Scrolls = append(ps.Scrolls, itemToSave(it))
	}
	for _, it := range p.Pack.Weapons {
		ps.Weapons = append(ps.Weapons, itemToSave(it))
	}
	for _, e := range p.ActiveEffects {
		ps.ActiveEffects = append(ps.ActiveEffects, &EffectSave{
			Subtype: int(e.Subtype),
			Amount:  e.Amount,
			Turns:   e.Turns,
		})
	}
	return ps
}

// RestorePlayer applies a PlayerSave to a Character.
func RestorePlayer(ps *PlayerSave) *domain.Character {
	c := &domain.Character{
		Pos:          domain.Position{X: ps.PosX, Y: ps.PosY},
		MaxHealth:    ps.MaxHealth,
		Health:       ps.Health,
		Dexterity:    ps.Dexterity,
		Strength:     ps.Strength,
		SleepTurns:   ps.SleepTurns,
		TilesWalked:  ps.TilesWalked,
		HitsDealt:    ps.HitsDealt,
		HitsReceived: ps.HitsReceived,
		EnemiesKilled: ps.EnemiesKilled,
		FoodEaten:    ps.FoodEaten,
		ElixirsDrunk: ps.ElixirsDrunk,
		ScrollsRead:  ps.ScrollsRead,
		Pack:         domain.NewBackpack(),
	}
	c.Pack.Treasure = ps.Treasure
	for _, is := range ps.Foods {
		c.Pack.Foods = append(c.Pack.Foods, saveToItem(is))
	}
	for _, is := range ps.Elixirs {
		c.Pack.Elixirs = append(c.Pack.Elixirs, saveToItem(is))
	}
	for _, is := range ps.Scrolls {
		c.Pack.Scrolls = append(c.Pack.Scrolls, saveToItem(is))
	}
	for _, is := range ps.Weapons {
		it := saveToItem(is)
		c.Pack.Weapons = append(c.Pack.Weapons, it)
		if is.ID == ps.WeaponID {
			c.Weapon = it
		}
	}
	for _, es := range ps.ActiveEffects {
		c.ActiveEffects = append(c.ActiveEffects, &domain.ActiveEffect{
			Subtype: domain.ItemSubtype(es.Subtype),
			Amount:  es.Amount,
			Turns:   es.Turns,
		})
	}
	return c
}

func itemToSave(it *domain.Item) *ItemSave {
	return &ItemSave{
		ID:        it.ID,
		Type:      int(it.Type),
		Subtype:   int(it.Subtype),
		Name:      it.Name,
		Health:    it.Health,
		MaxHealth: it.MaxHealth,
		Dexterity: it.Dexterity,
		Strength:  it.Strength,
		Value:     it.Value,
		Duration:  it.Duration,
	}
}

func saveToItem(is *ItemSave) *domain.Item {
	return &domain.Item{
		ID:        is.ID,
		Type:      domain.ItemType(is.Type),
		Subtype:   domain.ItemSubtype(is.Subtype),
		Name:      is.Name,
		Health:    is.Health,
		MaxHealth: is.MaxHealth,
		Dexterity: is.Dexterity,
		Strength:  is.Strength,
		Value:     is.Value,
		Duration:  is.Duration,
	}
}

func levelToSave(l *domain.Level, seed int64) *LevelSave {
	ls := &LevelSave{Seed: seed}
	for _, e := range l.Enemies {
		ls.Enemies = append(ls.Enemies, &EnemySave{
			ID:        e.ID,
			Type:      int(e.Type),
			PosX:      e.Pos.X,
			PosY:      e.Pos.Y,
			Health:    e.Health,
			MaxHealth: e.MaxHealth,
			Dexterity: e.Dexterity,
			Strength:  e.Strength,
			Hostility: e.Hostility,
			Alive:     e.Alive,
		})
	}
	for _, it := range l.Items {
		if it.OnFloor {
			is := itemToSave(it)
			ls.Items = append(ls.Items, is)
		}
	}
	return ls
}

// ---- JSON helpers -------------------------------------------------------------

func writeJSON(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func readJSON(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
