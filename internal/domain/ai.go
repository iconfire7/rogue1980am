package domain

import (
	"math/rand"
)

// MoveEnemy decides and executes one move for the given enemy.
// Returns a combat result if it attacked the player (nil otherwise).
func MoveEnemy(e *Enemy, gs *GameSession, rng *rand.Rand) *CombatResult {
	if !e.IsAlive() {
		return nil
	}
	// Ogre rests after attacking
	if e.RestTurns > 0 {
		e.RestTurns--
		return nil
	}

	player := gs.Player
	lvl := gs.Level
	dist := e.Pos.ManhattanDistance(player.Pos)

	// Ghost: periodically teleport within its room; become invisible out of combat
	if e.Type == EnemyGhost {
		e.TeleportTimer--
		if e.TeleportTimer <= 0 {
			teleportGhost(e, lvl, rng)
			e.TeleportTimer = 2 + rng.Intn(4)
		}
		e.InvisTimer--
		if e.InvisTimer <= 0 {
			if dist > 2 {
				e.Invisible = !e.Invisible
			} else {
				e.Invisible = false
			}
			e.InvisTimer = 3 + rng.Intn(5)
		}
	}

	// Update alerted state
	if dist <= e.Hostility {
		e.Alerted = true
	}
	// Check line of sight for alerted enemies: if no path, continue random movement
	if e.Alerted && dist > e.Hostility*2 {
		e.Alerted = false
	}

	// Adjacent to player: attack
	if dist == 1 {
		res := EnemyAttacksCharacter(e, player)
		if e.Type == EnemyOgre {
			e.RestTurns = 1 // rest after attacking
		}
		return &res
	}

	// Move toward player if alerted, else move according to type pattern
	if e.Alerted {
		next := bfsStep(e.Pos, player.Pos, lvl, rng)
		if next != nil && lvl.EnemyAt(*next) == nil {
			e.Pos = *next
		}
	} else {
		moveByPattern(e, lvl, rng)
	}
	return nil
}

// moveByPattern moves the enemy according to its type-specific pattern.
func moveByPattern(e *Enemy, lvl *Level, rng *rand.Rand) {
	switch e.Type {
	case EnemyOgre:
		// moves two tiles per turn
		for i := 0; i < 2; i++ {
			randomStep(e, lvl, rng)
		}
	case EnemySnakeMage:
		// moves diagonally, bouncing off walls
		next := e.Pos.Add(e.DiagDir.X, e.DiagDir.Y)
		if !lvl.IsWalkable(next.X, next.Y) || lvl.EnemyAt(next) != nil {
			// bounce: try reversals
			e.DiagDir.X = -e.DiagDir.X
			next = e.Pos.Add(e.DiagDir.X, e.DiagDir.Y)
			if !lvl.IsWalkable(next.X, next.Y) || lvl.EnemyAt(next) != nil {
				e.DiagDir.Y = -e.DiagDir.Y
				next = e.Pos.Add(e.DiagDir.X, e.DiagDir.Y)
				if !lvl.IsWalkable(next.X, next.Y) || lvl.EnemyAt(next) != nil {
					e.DiagDir.X = -e.DiagDir.X
					next = e.Pos.Add(e.DiagDir.X, e.DiagDir.Y)
				}
			}
		}
		if lvl.IsWalkable(next.X, next.Y) && lvl.EnemyAt(next) == nil {
			e.Pos = next
		}
	default:
		randomStep(e, lvl, rng)
	}
}

func randomStep(e *Enemy, lvl *Level, rng *rand.Rand) {
	dirs := []Position{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	rng.Shuffle(len(dirs), func(i, j int) { dirs[i], dirs[j] = dirs[j], dirs[i] })
	for _, d := range dirs {
		next := e.Pos.Add(d.X, d.Y)
		if lvl.IsWalkable(next.X, next.Y) && lvl.EnemyAt(next) == nil {
			e.Pos = next
			return
		}
	}
}

func teleportGhost(e *Enemy, lvl *Level, rng *rand.Rand) {
	// find the room the ghost is in and teleport to a random floor in it
	room := lvl.RoomAt(e.Pos)
	if room == nil {
		return
	}
	for attempts := 0; attempts < 20; attempts++ {
		x := room.X + rng.Intn(room.W)
		y := room.Y + rng.Intn(room.H)
		pos := Position{X: x, Y: y}
		if lvl.IsWalkable(x, y) && lvl.EnemyAt(pos) == nil {
			e.Pos = pos
			return
		}
	}
}

// bfsStep returns the next position on the shortest path from src to dst,
// using BFS over walkable tiles.  Returns nil if no path exists.
func bfsStep(src, dst Position, lvl *Level, rng *rand.Rand) *Position {
	if src.Equal(dst) {
		return nil
	}
	type node struct {
		pos    Position
		parent *Position
	}
	visited := make(map[Position]bool)
	visited[src] = true
	queue := []node{{pos: src}}
	parentOf := make(map[Position]Position)

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, nb := range cur.pos.Neighbors() {
			if nb.Equal(dst) {
				// reconstruct first step
				step := nb
				prev := cur.pos
				for !prev.Equal(src) {
					step = prev
					prev = parentOf[prev]
				}
				return &step
			}
			if visited[nb] {
				continue
			}
			if !lvl.IsWalkable(nb.X, nb.Y) {
				continue
			}
			visited[nb] = true
			parentOf[nb] = cur.pos
			queue = append(queue, node{pos: nb})
		}
	}
	return nil
}
