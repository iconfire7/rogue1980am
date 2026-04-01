package domain

// Position represents a 2D coordinate on the map.
type Position struct {
	X, Y int
}

// Add returns a new position offset by dx, dy.
func (p Position) Add(dx, dy int) Position {
	return Position{X: p.X + dx, Y: p.Y + dy}
}

// Equal returns true if two positions are the same.
func (p Position) Equal(other Position) bool {
	return p.X == other.X && p.Y == other.Y
}

// Neighbors returns the four cardinal adjacent positions.
func (p Position) Neighbors() []Position {
	return []Position{
		{p.X, p.Y - 1},
		{p.X, p.Y + 1},
		{p.X - 1, p.Y},
		{p.X + 1, p.Y},
	}
}

// AllNeighbors returns all 8 surrounding positions (cardinal + diagonal).
func (p Position) AllNeighbors() []Position {
	dirs := []Position{}
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			dirs = append(dirs, Position{p.X + dx, p.Y + dy})
		}
	}
	return dirs
}

// ManhattanDistance returns the Manhattan distance between two positions.
func (p Position) ManhattanDistance(other Position) int {
	dx := p.X - other.X
	if dx < 0 {
		dx = -dx
	}
	dy := p.Y - other.Y
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}
