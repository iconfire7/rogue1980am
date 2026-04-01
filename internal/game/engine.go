// Package game implements the high-level game loop that connects domain, ui, and storage layers.
package game

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gdamore/tcell/v2"
	"rogue1980am/internal/domain"
	"rogue1980am/internal/generator"
	"rogue1980am/internal/storage"
	"rogue1980am/internal/ui"
)

// Engine is the central game controller.
type Engine struct {
	renderer  *ui.Renderer
	session   *domain.GameSession
	rng       *rand.Rand
	seed      int64
	sessionID int
}

// NewEngine creates a new game engine with a renderer.
func NewEngine(r *ui.Renderer) *Engine {
	return &Engine{renderer: r}
}

// RunMainMenu shows the main menu and handles navigation.
func (e *Engine) RunMainMenu() {
	for {
		hasSave := storage.HasSave()
		e.renderer.DrawMainMenu(hasSave)
		ev := e.renderer.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Rune() {
			case 'n', 'N':
				e.startNewGame()
				e.runGameLoop()
			case 'c', 'C':
				if hasSave {
					if err := e.loadGame(); err == nil {
						e.runGameLoop()
					}
				}
			case 's', 'S':
				e.showStatistics()
			case 'q', 'Q':
				return
			}
		case *tcell.EventResize:
			// handled on next render
		}
	}
}

// startNewGame initialises a brand-new session.
func (e *Engine) startNewGame() {
	e.sessionID++
	e.seed = time.Now().UnixNano()
	e.rng = rand.New(rand.NewSource(e.seed))
	e.session = domain.NewGameSession(e.sessionID)
	e.loadLevel(1)
}

// loadGame restores a saved session.
func (e *Engine) loadGame() error {
	sd, err := storage.LoadGame()
	if err != nil {
		return err
	}
	e.sessionID = sd.SessionID
	e.seed = sd.Level.Seed
	e.rng = rand.New(rand.NewSource(e.seed))
	e.session = domain.NewGameSession(e.sessionID)
	e.session.CurrentLevel = sd.LevelNumber
	e.session.Stats = sd.Stats
	e.session.Player = storage.RestorePlayer(sd.Player)
	// re-generate the level with the saved seed to get identical geometry
	lvl := generator.GenerateLevel(sd.LevelNumber, e.rng)
	// restore enemy positions and state from saved data
	if sd.Level != nil {
		applyLevelSave(lvl, sd.Level)
	}
	e.session.Level = lvl
	e.session.Player.Pos = lvl.StartPos
	if sd.Player != nil {
		e.session.Player.Pos = domain.Position{X: sd.Player.PosX, Y: sd.Player.PosY}
	}
	return nil
}

// applyLevelSave overwrites generated enemy/item data with saved state.
func applyLevelSave(lvl *domain.Level, ls *storage.LevelSave) {
	// Build a map by ID for fast lookup
	savedByID := make(map[int]*storage.EnemySave)
	for _, es := range ls.Enemies {
		savedByID[es.ID] = es
	}
	for _, en := range lvl.Enemies {
		if saved, ok := savedByID[en.ID]; ok {
			en.Pos = domain.Position{X: saved.PosX, Y: saved.PosY}
			en.Health = saved.Health
			en.MaxHealth = saved.MaxHealth
			en.Alive = saved.Alive
		} else {
			en.Alive = false // enemy was killed
		}
	}
}

// loadLevel generates a fresh level and places the player in the start room.
func (e *Engine) loadLevel(depth int) {
	lvl := generator.GenerateLevel(depth, e.rng)
	e.session.Level = lvl
	e.session.CurrentLevel = depth
	e.session.Player.Pos = lvl.StartPos
}

// runGameLoop is the main event loop for an active game session.
func (e *Engine) runGameLoop() {
	for {
		vis := ui.ComputeVisibility(e.session)
		e.renderer.DrawGame(e.session, vis)

		ev := e.renderer.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			done := e.handleKey(ev)
			if done {
				return
			}
		case *tcell.EventResize:
			e.renderer.Screen.Sync()
		}
	}
}

// handleKey processes a keypress and returns true if the game loop should exit.
func (e *Engine) handleKey(ev *tcell.EventKey) bool {
	gs := e.session
	p := gs.Player

	switch ev.Key() {
	case tcell.KeyEscape:
		return true
	}

	switch ev.Rune() {
	case 'Q':
		gs.SyncStatsFromPlayer()
		_ = storage.SaveStatistics(gs.Stats)
		storage.DeleteSave()
		return true
	case '?':
		e.showStatistics()
		return false
	case 'h':
		e.useItemMenu("Weapons", &p.Pack.Weapons, true)
		return false
	case 'j':
		e.useItemMenuList("Food", p.Pack.Foods, func(i int) {
			it := p.Pack.RemoveFood(i)
			if it != nil {
				msg := p.ApplyItem(it)
				gs.AddMessage(msg)
			}
		})
		return false
	case 'k':
		e.useItemMenuList("Elixirs", p.Pack.Elixirs, func(i int) {
			it := p.Pack.RemoveElixir(i)
			if it != nil {
				msg := p.ApplyItem(it)
				gs.AddMessage(msg)
			}
		})
		return false
	case 'e':
		e.useItemMenuList("Scrolls", p.Pack.Scrolls, func(i int) {
			it := p.Pack.RemoveScroll(i)
			if it != nil {
				msg := p.ApplyItem(it)
				gs.AddMessage(msg)
			}
		})
		return false
	case 'w', 'W':
		e.tryMove(0, -1)
	case 's', 'S':
		e.tryMove(0, 1)
	case 'a', 'A':
		e.tryMove(-1, 0)
	case 'd', 'D':
		e.tryMove(1, 0)
	}

	// Check win/loss after each action
	if gs.GameOver {
		gs.SyncStatsFromPlayer()
		_ = storage.SaveStatistics(gs.Stats)
		storage.DeleteSave()
		e.renderer.DrawGameOver(gs)
		e.renderer.PollEvent()
		return true
	}
	if gs.Won {
		gs.SyncStatsFromPlayer()
		_ = storage.SaveStatistics(gs.Stats)
		storage.DeleteSave()
		e.renderer.DrawWin(gs)
		e.renderer.PollEvent()
		return true
	}
	return false
}

// tryMove attempts to move the player by (dx,dy).
// If target is an enemy, attack it.  If target is stairs, advance the level.
func (e *Engine) tryMove(dx, dy int) {
	gs := e.session
	p := gs.Player
	if p.IsSleeping() {
		p.TickEffects()
		gs.AddMessage("You are asleep and cannot move!")
		return
	}

	lvl := gs.Level
	target := p.Pos.Add(dx, dy)

	// Check for enemy
	enemy := lvl.EnemyAt(target)
	if enemy != nil {
		res := domain.CharacterAttacksEnemy(p, enemy)
		if res.Missed {
			gs.AddMessage(res.Message)
		} else if !res.Hit {
			gs.AddMessage("You miss the " + enemy.Type.Name() + "!")
		} else {
			gs.AddMessage(fmt.Sprintf("You hit the %s for %d damage!", enemy.Type.Name(), res.Damage))
			if !enemy.IsAlive() {
				gs.AddMessage(fmt.Sprintf("You killed the %s!", enemy.Type.Name()))
				p.EnemiesKilled++
				// drop treasure
				gold := enemy.TreasureDrop()
				treasure := &domain.Item{
					ID:      gold,
					Type:    domain.ItemTreasure,
					Name:    "Gold",
					Value:   gold,
					Pos:     enemy.Pos,
					OnFloor: true,
				}
				lvl.Items = append(lvl.Items, treasure)
			}
		}
		e.processTurn()
		return
	}

	// Move
	if !lvl.IsWalkable(target.X, target.Y) {
		return
	}
	p.Pos = target
	p.TilesWalked++

	// Check for stairs
	tile := lvl.TileAt(target.X, target.Y)
	if tile != nil && tile.Type == domain.TileStairs {
		e.advanceLevel()
		return
	}

	// Auto-pick up items
	items := lvl.ItemsAt(target)
	for _, it := range items {
		if p.Pack.CanAdd(it.Type) {
			if p.Pack.Add(it) {
				lvl.RemoveItem(it)
				if it.Type == domain.ItemTreasure {
					gs.AddMessage(fmt.Sprintf("You pick up %d gold!", it.Value))
				} else {
					gs.AddMessage(fmt.Sprintf("You pick up %s.", it.Name))
				}
			} else {
				gs.AddMessage("Your backpack is full!")
			}
		} else {
			gs.AddMessage("Your backpack is full!")
		}
	}

	e.processTurn()
}

// processTurn advances enemy AI and player effect timers.
func (e *Engine) processTurn() {
	gs := e.session
	p := gs.Player

	p.TickEffects()

	for _, enemy := range gs.Level.Enemies {
		if !enemy.IsAlive() {
			continue
		}
		res := domain.MoveEnemy(enemy, gs, e.rng)
		if res != nil {
			if res.Hit {
				gs.AddMessage(fmt.Sprintf("The %s hits you for %d damage! %s", enemy.Type.Name(), res.Damage, res.Message))
			} else {
				gs.AddMessage(fmt.Sprintf("The %s misses. %s", enemy.Type.Name(), res.Message))
			}
			if !p.IsAlive() {
				gs.GameOver = true
				return
			}
		}
	}
}

// advanceLevel moves the player to the next dungeon level.
func (e *Engine) advanceLevel() {
	gs := e.session
	gs.SyncStatsFromPlayer()
	next := gs.CurrentLevel + 1
	if next > domain.TotalLevels {
		gs.Won = true
		return
	}
	_ = storage.SaveGame(gs, e.seed)
	e.loadLevel(next)
	gs.AddMessage(fmt.Sprintf("You descend to level %d.", next))
}

// showStatistics displays the leaderboard overlay.
func (e *Engine) showStatistics() {
	stats, _ := storage.LoadAllStatistics()
	e.renderer.DrawStatistics(stats)
	e.renderer.PollEvent()
}

// useItemMenu shows a weapon-selection menu (with optional unequip option).
func (e *Engine) useItemMenu(title string, items *[]*domain.Item, canUnequip bool) {
	p := e.session.Player
	names := make([]string, len(*items))
	for i, it := range *items {
		if p.Weapon == it {
			names[i] = it.Name + " [equipped]"
		} else {
			names[i] = it.Name
		}
	}
	e.renderer.DrawItemMenu(title, names, canUnequip)
	ev := e.renderer.PollEvent()
	kev, ok := ev.(*tcell.EventKey)
	if !ok {
		return
	}
	ch := kev.Rune()
	if ch == '0' && canUnequip {
		// drop current weapon to adjacent tile
		if p.Weapon != nil {
			dropWeapon(e.session, p.Weapon)
			p.Weapon = nil
			e.session.AddMessage("You unequip your weapon.")
		}
		return
	}
	if ch >= '1' && ch <= '9' {
		idx := int(ch - '1')
		if idx < len(*items) {
			p.Weapon = (*items)[idx]
			e.session.AddMessage(fmt.Sprintf("You equip the %s.", p.Weapon.Name))
		}
	}
}

// useItemMenuList shows a menu for consumable items and invokes cb on selection.
func (e *Engine) useItemMenuList(title string, items []*domain.Item, cb func(int)) {
	if len(items) == 0 {
		e.session.AddMessage("You have no " + title + ".")
		return
	}
	names := make([]string, len(items))
	for i, it := range items {
		names[i] = it.Name
	}
	e.renderer.DrawItemMenu(title, names, false)
	ev := e.renderer.PollEvent()
	kev, ok := ev.(*tcell.EventKey)
	if !ok {
		return
	}
	ch := kev.Rune()
	if ch >= '1' && ch <= '9' {
		idx := int(ch - '1')
		cb(idx)
	}
}

// dropWeapon places the weapon on an adjacent free floor tile.
func dropWeapon(gs *domain.GameSession, w *domain.Item) {
	lvl := gs.Level
	p := gs.Player
	for _, nb := range p.Pos.AllNeighbors() {
		t := lvl.TileAt(nb.X, nb.Y)
		if t != nil && t.Walkable() && lvl.EnemyAt(nb) == nil {
			w.Pos = nb
			w.OnFloor = true
			lvl.Items = append(lvl.Items, w)
			// remove from backpack
			for i, it := range p.Pack.Weapons {
				if it == w {
					p.Pack.Weapons = append(p.Pack.Weapons[:i], p.Pack.Weapons[i+1:]...)
					break
				}
			}
			return
		}
	}
}
