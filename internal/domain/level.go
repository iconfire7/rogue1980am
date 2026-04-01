package domain

// TileType describes what occupies a map cell.
type TileType int

const (
	TileEmpty      TileType = iota
	TileFloor               // walkable room floor
	TileWall                // room boundary wall
	TileCorridor            // corridor passage
	TileOpeningH            // horizontal wall opening (doorway)
	TileOpeningV            // vertical wall opening (doorway)
	TileStairs              // exit to next level
)

// Tile is a single map cell.
type Tile struct {
	Type     TileType
	Explored bool // player has seen this tile
}

// Walkable returns true if characters may occupy this tile.
func (t *Tile) Walkable() bool {
	return t.Type == TileFloor || t.Type == TileCorridor ||
		t.Type == TileOpeningH || t.Type == TileOpeningV ||
		t.Type == TileStairs
}

// Room represents a rectangular chamber in the dungeon.
type Room struct {
	ID      int
	X, Y    int // top-left corner (inside walls)
	W, H    int // inner dimensions (floor only)
	Enemies []*Enemy
	Items   []*Item
	IsStart bool
	IsExit  bool
}

// InnerBounds returns the inclusive bounds of the floor tiles.
func (r *Room) InnerBounds() (minX, minY, maxX, maxY int) {
	return r.X, r.Y, r.X + r.W - 1, r.Y + r.H - 1
}

// OuterBounds returns the inclusive bounds including the walls.
func (r *Room) OuterBounds() (minX, minY, maxX, maxY int) {
	return r.X - 1, r.Y - 1, r.X + r.W, r.Y + r.H
}

// Contains returns true if pos is inside the floor area.
func (r *Room) Contains(pos Position) bool {
	return pos.X >= r.X && pos.X < r.X+r.W && pos.Y >= r.Y && pos.Y < r.Y+r.H
}

// Center returns the center floor tile of the room.
func (r *Room) Center() Position {
	return Position{X: r.X + r.W/2, Y: r.Y + r.H/2}
}

// CorridorTile holds one cell of a corridor.
type CorridorTile struct {
	Pos Position
}

// Corridor connects two rooms via a list of floor tiles.
type Corridor struct {
	RoomA, RoomB int // room IDs
	Tiles        []CorridorTile
}

// Level contains all geometry and entities for one dungeon floor.
type Level struct {
	Number   int
	Width    int
	Height   int
	Tiles    [][]Tile // [y][x]
	Rooms    []*Room
	Corridors []*Corridor
	Enemies  []*Enemy
	Items    []*Item // items lying on the floor
	StartPos Position
	ExitPos  Position
	// fog-of-war: which rooms have been entered
	ExploredRooms map[int]bool
}

// NewLevel creates an empty level grid.
func NewLevel(number, width, height int) *Level {
	tiles := make([][]Tile, height)
	for y := range tiles {
		tiles[y] = make([]Tile, width)
	}
	return &Level{
		Number:        number,
		Width:         width,
		Height:        height,
		Tiles:         tiles,
		ExploredRooms: make(map[int]bool),
	}
}

// TileAt returns a pointer to the tile at (x,y) or nil if out of bounds.
func (l *Level) TileAt(x, y int) *Tile {
	if x < 0 || y < 0 || x >= l.Width || y >= l.Height {
		return nil
	}
	return &l.Tiles[y][x]
}

// RoomAt returns the room that contains pos, or nil.
func (l *Level) RoomAt(pos Position) *Room {
	for _, r := range l.Rooms {
		if r.Contains(pos) {
			return r
		}
	}
	return nil
}

// EnemyAt returns the living enemy at pos, or nil.
func (l *Level) EnemyAt(pos Position) *Enemy {
	for _, e := range l.Enemies {
		if e.IsAlive() && e.Pos.Equal(pos) {
			return e
		}
	}
	return nil
}

// ItemsAt returns all items lying on the floor at pos.
func (l *Level) ItemsAt(pos Position) []*Item {
	var out []*Item
	for _, it := range l.Items {
		if it.OnFloor && it.Pos.Equal(pos) {
			out = append(out, it)
		}
	}
	return out
}

// RemoveItem removes an item from the level's item list.
func (l *Level) RemoveItem(item *Item) {
	for i, it := range l.Items {
		if it == item {
			l.Items = append(l.Items[:i], l.Items[i+1:]...)
			return
		}
	}
}

// IsWalkable returns true if (x,y) is within bounds and walkable.
func (l *Level) IsWalkable(x, y int) bool {
	t := l.TileAt(x, y)
	return t != nil && t.Walkable()
}
