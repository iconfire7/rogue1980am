package ui

import (
	"math"

	"rogue1980am/internal/domain"
)

// ComputeVisibility returns the set of positions visible from the player's position.
// Full rooms are revealed when the player is inside them; corridors use ray-casting.
func ComputeVisibility(gs *domain.GameSession) map[domain.Position]bool {
	lvl := gs.Level
	player := gs.Player
	vis := make(map[domain.Position]bool)

	// Mark the player's tile as visible
	vis[player.Pos] = true

	// Determine current room
	currentRoom := lvl.RoomAt(player.Pos)
	if currentRoom != nil {
		// Inside a room: reveal the entire room
		for y := currentRoom.Y - 1; y <= currentRoom.Y+currentRoom.H; y++ {
			for x := currentRoom.X - 1; x <= currentRoom.X+currentRoom.W; x++ {
				pos := domain.Position{X: x, Y: y}
				vis[pos] = true
				t := lvl.TileAt(x, y)
				if t != nil {
					t.Explored = true
				}
			}
		}
		// also reveal just outside doorways so adjacent corridors are hinted
		revealAdjacentCorridor(lvl, currentRoom, vis)
		lvl.ExploredRooms[currentRoom.ID] = true
	} else {
		// In a corridor: use ray-casting to determine visible tiles
		castRays(player.Pos, 8, lvl, vis)
	}

	// Mark all newly visible tiles as explored
	for pos := range vis {
		t := lvl.TileAt(pos.X, pos.Y)
		if t != nil {
			t.Explored = true
		}
	}
	return vis
}

// revealAdjacentCorridor reveals one tile outside each doorway of the room.
func revealAdjacentCorridor(lvl *domain.Level, r *domain.Room, vis map[domain.Position]bool) {
	// check the four walls for openings and peek one tile beyond
	for x := r.X; x < r.X+r.W; x++ {
		check := []domain.Position{{x, r.Y - 1}, {x, r.Y + r.H}}
		for _, p := range check {
			t := lvl.TileAt(p.X, p.Y)
			if t != nil && (t.Type == domain.TileOpeningH || t.Type == domain.TileOpeningV || t.Type == domain.TileCorridor) {
				vis[p] = true
				// one tile further
				dy := -1
				if p.Y > r.Y {
					dy = 1
				}
				outer := domain.Position{X: p.X, Y: p.Y + dy}
				vis[outer] = true
			}
		}
	}
	for y := r.Y; y < r.Y+r.H; y++ {
		check := []domain.Position{{r.X - 1, y}, {r.X + r.W, y}}
		for _, p := range check {
			t := lvl.TileAt(p.X, p.Y)
			if t != nil && (t.Type == domain.TileOpeningH || t.Type == domain.TileOpeningV || t.Type == domain.TileCorridor) {
				vis[p] = true
				dx := -1
				if p.X > r.X {
					dx = 1
				}
				outer := domain.Position{X: p.X + dx, Y: p.Y}
				vis[outer] = true
			}
		}
	}
}

// castRays shoots rays in all directions from pos using Bresenham's line algorithm.
func castRays(origin domain.Position, radius int, lvl *domain.Level, vis map[domain.Position]bool) {
	// cast 360 rays (1-degree resolution)
	numRays := 360
	for i := 0; i < numRays; i++ {
		angle := float64(i) * 2 * math.Pi / float64(numRays)
		dx := math.Cos(angle)
		dy := math.Sin(angle)
		castRay(origin, dx, dy, radius, lvl, vis)
	}
}

// castRay traces a single ray from origin in direction (dx,dy) up to maxDist tiles.
func castRay(origin domain.Position, dx, dy float64, maxDist int, lvl *domain.Level, vis map[domain.Position]bool) {
	x := float64(origin.X) + 0.5
	y := float64(origin.Y) + 0.5
	for step := 0; step < maxDist; step++ {
		ix := int(x)
		iy := int(y)
		pos := domain.Position{X: ix, Y: iy}
		vis[pos] = true
		t := lvl.TileAt(ix, iy)
		if t == nil || t.Type == domain.TileWall {
			break // ray blocked
		}
		x += dx
		y += dy
	}
}
