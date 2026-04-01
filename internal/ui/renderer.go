// Package ui provides a tcell-based terminal rendering layer for the game.
package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"rogue1980am/internal/domain"
)

const (
	StatusPanelWidth = 24
	MsgPanelHeight   = 5
)

// Renderer wraps a tcell.Screen and provides game-drawing methods.
type Renderer struct {
	Screen tcell.Screen
}

// NewRenderer creates and initialises a new terminal renderer.
func NewRenderer() (*Renderer, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := s.Init(); err != nil {
		return nil, err
	}
	s.SetStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite))
	s.Clear()
	return &Renderer{Screen: s}, nil
}

// Close releases the terminal.
func (r *Renderer) Close() {
	r.Screen.Fini()
}

// PollEvent delegates to tcell.
func (r *Renderer) PollEvent() tcell.Event {
	return r.Screen.PollEvent()
}

// ---- Scene rendering ----------------------------------------------------------

// DrawGame renders the complete game view: map, status panel, and message log.
func (r *Renderer) DrawGame(gs *domain.GameSession, visSet map[domain.Position]bool) {
	r.Screen.Clear()
	r.drawMap(gs, visSet)
	r.drawStatusPanel(gs)
	r.drawMessages(gs)
	r.Screen.Show()
}

// drawMap renders the dungeon tiles, items, enemies and player onto the screen.
func (r *Renderer) drawMap(gs *domain.GameSession, visSet map[domain.Position]bool) {
	lvl := gs.Level
	_, sh := r.Screen.Size()
	mapH := sh - MsgPanelHeight

	for y := 0; y < mapH && y < lvl.Height; y++ {
		for x := 0; x < lvl.Width && x < r.mapWidth(); x++ {
			pos := domain.Position{X: x, Y: y}
			tile := lvl.TileAt(x, y)
			if tile == nil {
				continue
			}

			inVis := visSet[pos]
			inExplored := tile.Explored

			if !inVis && !inExplored {
				continue // fog of war: never seen
			}

			if !inVis && inExplored {
				// previously visited: draw walls/structure only
				ch, st := tileGlyph(tile, false)
				if tile.Type == domain.TileWall || tile.Type == domain.TileOpeningH || tile.Type == domain.TileOpeningV {
					r.Screen.SetContent(x, y, ch, nil, st.Dim(true))
				}
				continue
			}

			// fully visible
			ch, st := tileGlyph(tile, true)
			r.Screen.SetContent(x, y, ch, nil, st)
		}
	}

	// Draw floor items (only in visible area)
	for _, it := range lvl.Items {
		if !it.OnFloor {
			continue
		}
		if !visSet[it.Pos] {
			continue
		}
		ch, st := itemGlyph(it)
		r.Screen.SetContent(it.Pos.X, it.Pos.Y, ch, nil, st)
	}

	// Draw enemies (only visible ones; invisible ghosts are hidden unless adjacent)
	for _, e := range lvl.Enemies {
		if !e.IsAlive() {
			continue
		}
		if !visSet[e.Pos] {
			continue
		}
		if e.Invisible && e.Pos.ManhattanDistance(gs.Player.Pos) > 1 {
			continue
		}
		ch, st := enemyGlyph(e)
		r.Screen.SetContent(e.Pos.X, e.Pos.Y, ch, nil, st)
	}

	// Draw player
	pSt := tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true)
	r.Screen.SetContent(gs.Player.Pos.X, gs.Player.Pos.Y, '@', nil, pSt)
}

func (r *Renderer) mapWidth() int {
	sw, _ := r.Screen.Size()
	return sw - StatusPanelWidth
}

// drawStatusPanel renders the HUD on the right side of the screen.
func (r *Renderer) drawStatusPanel(gs *domain.GameSession) {
	sw, _ := r.Screen.Size()
	x := sw - StatusPanelWidth
	p := gs.Player
	borderSt := tcell.StyleDefault.Foreground(tcell.ColorGray)
	for y := 0; y < 20; y++ {
		r.Screen.SetContent(x, y, '│', nil, borderSt)
	}
	y := 0
	r.drawText(x+1, y, fmt.Sprintf("Level: %2d / %2d", gs.CurrentLevel, domain.TotalLevels), tcell.StyleDefault.Bold(true))
	y++
	hpColor := tcell.ColorGreen
	if p.Health*3 < p.MaxHealth {
		hpColor = tcell.ColorRed
	} else if p.Health*2 < p.MaxHealth {
		hpColor = tcell.ColorYellow
	}
	r.drawText(x+1, y, fmt.Sprintf("HP:  %3d / %3d", p.Health, p.MaxHealth), tcell.StyleDefault.Foreground(hpColor))
	y++
	r.drawText(x+1, y, fmt.Sprintf("Str: %3d  Dex: %3d", p.Strength, p.Dexterity), tcell.StyleDefault)
	y++
	wpnName := "Unarmed"
	if p.Weapon != nil {
		wpnName = p.Weapon.Name[:min(len(p.Weapon.Name), StatusPanelWidth-4)]
	}
	r.drawText(x+1, y, fmt.Sprintf("Wpn: %s", wpnName), tcell.StyleDefault)
	y++
	r.drawText(x+1, y, fmt.Sprintf("Gold:%4d", p.Pack.Treasure), tcell.StyleDefault.Foreground(tcell.ColorYellow))
	y++
	y++ // blank
	r.drawText(x+1, y, "--- Backpack ---", tcell.StyleDefault.Dim(true))
	y++
	r.drawText(x+1, y, fmt.Sprintf("Food:   %d", len(p.Pack.Foods)), tcell.StyleDefault)
	y++
	r.drawText(x+1, y, fmt.Sprintf("Elixir: %d", len(p.Pack.Elixirs)), tcell.StyleDefault)
	y++
	r.drawText(x+1, y, fmt.Sprintf("Scroll: %d", len(p.Pack.Scrolls)), tcell.StyleDefault)
	y++
	r.drawText(x+1, y, fmt.Sprintf("Weapon: %d", len(p.Pack.Weapons)), tcell.StyleDefault)
	y++
	y++
	r.drawText(x+1, y, "--- Controls ---", tcell.StyleDefault.Dim(true))
	y++
	r.drawText(x+1, y, "WASD: move", tcell.StyleDefault.Dim(true))
	y++
	r.drawText(x+1, y, "h: weapon  j: food", tcell.StyleDefault.Dim(true))
	y++
	r.drawText(x+1, y, "k: elixir  e: scroll", tcell.StyleDefault.Dim(true))
	y++
	r.drawText(x+1, y, "?: stats   Q: quit", tcell.StyleDefault.Dim(true))
	if p.IsSleeping() {
		y++
		r.drawText(x+1, y, "*** SLEEPING ***", tcell.StyleDefault.Foreground(tcell.ColorBlue).Bold(true))
	}
}

// drawMessages renders the bottom message log.
func (r *Renderer) drawMessages(gs *domain.GameSession) {
	_, sh := r.Screen.Size()
	msgY := sh - MsgPanelHeight
	sw, _ := r.Screen.Size()
	mw := sw - StatusPanelWidth

	// separator
	for x := 0; x < mw; x++ {
		r.Screen.SetContent(x, msgY, '─', nil, tcell.StyleDefault.Foreground(tcell.ColorGray))
	}
	msgY++
	msgs := gs.Messages
	start := len(msgs) - (MsgPanelHeight - 1)
	if start < 0 {
		start = 0
	}
	for i, msg := range msgs[start:] {
		if i >= MsgPanelHeight-1 {
			break
		}
		r.drawText(0, msgY+i, truncate(msg, mw), tcell.StyleDefault)
	}
}

// ---- Overlay screens ----------------------------------------------------------

// DrawMainMenu draws the main menu.
func (r *Renderer) DrawMainMenu(hasSave bool) {
	r.Screen.Clear()
	sw, sh := r.Screen.Size()
	cx := sw / 2
	cy := sh / 2

	r.drawCentered(cx, cy-4, "╔══════════════════════════╗", tcell.StyleDefault.Foreground(tcell.ColorYellow))
	r.drawCentered(cx, cy-3, "║     ROGUE  1980          ║", tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true))
	r.drawCentered(cx, cy-2, "╚══════════════════════════╝", tcell.StyleDefault.Foreground(tcell.ColorYellow))
	r.drawCentered(cx, cy, "  [N] New Game", tcell.StyleDefault)
	if hasSave {
		r.drawCentered(cx, cy+1, "  [C] Continue", tcell.StyleDefault)
	}
	r.drawCentered(cx, cy+2, "  [S] Statistics", tcell.StyleDefault)
	r.drawCentered(cx, cy+3, "  [Q] Quit", tcell.StyleDefault)
	r.Screen.Show()
}

// DrawGameOver renders the game-over screen.
func (r *Renderer) DrawGameOver(gs *domain.GameSession) {
	r.Screen.Clear()
	sw, sh := r.Screen.Size()
	cx, cy := sw/2, sh/2
	r.drawCentered(cx, cy-2, "  *** GAME OVER ***  ", tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true))
	r.drawCentered(cx, cy, fmt.Sprintf("Level reached: %d", gs.Stats.DeepestLevel), tcell.StyleDefault)
	r.drawCentered(cx, cy+1, fmt.Sprintf("Treasure: %d", gs.Stats.TreasureTotal), tcell.StyleDefault)
	r.drawCentered(cx, cy+3, "Press any key to continue", tcell.StyleDefault.Dim(true))
	r.Screen.Show()
}

// DrawWin renders the victory screen.
func (r *Renderer) DrawWin(gs *domain.GameSession) {
	r.Screen.Clear()
	sw, sh := r.Screen.Size()
	cx, cy := sw/2, sh/2
	r.drawCentered(cx, cy-2, "  *** YOU WIN! ***  ", tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true))
	r.drawCentered(cx, cy, fmt.Sprintf("Treasure collected: %d", gs.Stats.TreasureTotal), tcell.StyleDefault)
	r.drawCentered(cx, cy+1, fmt.Sprintf("Enemies slain: %d", gs.Stats.EnemiesKilled), tcell.StyleDefault)
	r.drawCentered(cx, cy+3, "Press any key to continue", tcell.StyleDefault.Dim(true))
	r.Screen.Show()
}

// DrawStatistics renders the leaderboard.
func (r *Renderer) DrawStatistics(stats []*domain.Statistics) {
	r.Screen.Clear()
	sw, _ := r.Screen.Size()
	cx := sw / 2
	y := 1
	r.drawCentered(cx, y, "=== STATISTICS / LEADERBOARD ===", tcell.StyleDefault.Bold(true))
	y += 2
	hdr := fmt.Sprintf("%-4s %-8s %-7s %-7s %-6s %-7s %-6s %-6s %-6s",
		"#", "Gold", "Level", "Kills", "Food", "Elixir", "Scroll", "HitsDlt", "Tiles")
	r.drawText(2, y, hdr, tcell.StyleDefault.Underline(true))
	y++
	for i, s := range stats {
		if y > 30 {
			break
		}
		line := fmt.Sprintf("%-4d %-8d %-7d %-7d %-6d %-7d %-6d %-6d %-6d",
			i+1, s.TreasureTotal, s.DeepestLevel, s.EnemiesKilled,
			s.FoodEaten, s.ElixirsDrunk, s.ScrollsRead, s.HitsDealt, s.TilesWalked)
		st := tcell.StyleDefault
		if i == 0 {
			st = st.Foreground(tcell.ColorYellow)
		}
		r.drawText(2, y, line, st)
		y++
	}
	y += 2
	r.drawText(2, y, "Press any key to return", tcell.StyleDefault.Dim(true))
	r.Screen.Show()
}

// DrawItemMenu draws a selection list for items of a given type.
// Returns the prompt string; caller must call PollEvent to get choice.
func (r *Renderer) DrawItemMenu(title string, items []string, includeUnequip bool) {
	r.Screen.Clear()
	sw, sh := r.Screen.Size()
	cx, cy := sw/2, sh/2
	top := cy - len(items)/2 - 2
	if top < 0 {
		top = 0
	}
	r.drawCentered(cx, top, title, tcell.StyleDefault.Bold(true))
	top++
	if includeUnequip {
		r.drawCentered(cx, top, "  [0] Unequip", tcell.StyleDefault)
		top++
	}
	for i, name := range items {
		r.drawCentered(cx, top+i, fmt.Sprintf("  [%d] %s", i+1, name), tcell.StyleDefault)
	}
	r.drawCentered(cx, top+len(items)+1, "  [ESC] Cancel", tcell.StyleDefault.Dim(true))
	r.Screen.Show()
}

// ---- helpers ------------------------------------------------------------------

func (r *Renderer) drawText(x, y int, s string, st tcell.Style) {
	sw, sh := r.Screen.Size()
	for i, ch := range s {
		if x+i >= sw || y >= sh {
			break
		}
		r.Screen.SetContent(x+i, y, ch, nil, st)
	}
}

func (r *Renderer) drawCentered(cx, y int, s string, st tcell.Style) {
	x := cx - len(s)/2
	if x < 0 {
		x = 0
	}
	r.drawText(x, y, s, st)
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ---- glyph helpers ------------------------------------------------------------

func tileGlyph(t *domain.Tile, visible bool) (rune, tcell.Style) {
	dim := tcell.StyleDefault.Foreground(tcell.ColorGray)
	bright := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	_ = visible
	switch t.Type {
	case domain.TileFloor:
		return '.', dim
	case domain.TileWall:
		return '#', bright
	case domain.TileCorridor:
		return '·', dim
	case domain.TileOpeningH, domain.TileOpeningV:
		return '+', tcell.StyleDefault.Foreground(tcell.ColorBrown)
	case domain.TileStairs:
		return '>', tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true)
	}
	return ' ', tcell.StyleDefault
}

func itemGlyph(it *domain.Item) (rune, tcell.Style) {
	switch it.Type {
	case domain.ItemFood:
		return '%', tcell.StyleDefault.Foreground(tcell.ColorOlive)
	case domain.ItemElixir:
		return '!', tcell.StyleDefault.Foreground(tcell.ColorBlue).Bold(true)
	case domain.ItemScroll:
		return '?', tcell.StyleDefault.Foreground(tcell.ColorAqua)
	case domain.ItemWeapon:
		return ')', tcell.StyleDefault.Foreground(tcell.ColorSilver)
	case domain.ItemTreasure:
		return '$', tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true)
	}
	return '*', tcell.StyleDefault
}

func enemyGlyph(e *domain.Enemy) (rune, tcell.Style) {
	ch := e.Glyph()
	var color tcell.Color
	switch e.Type {
	case domain.EnemyZombie:
		color = tcell.ColorGreen
	case domain.EnemyVampire:
		color = tcell.ColorRed
	case domain.EnemyGhost:
		color = tcell.ColorWhite
	case domain.EnemyOgre:
		color = tcell.ColorYellow
	case domain.EnemySnakeMage:
		color = tcell.ColorWhite
	default:
		color = tcell.ColorFuchsia
	}
	return ch, tcell.StyleDefault.Foreground(color).Bold(true)
}
