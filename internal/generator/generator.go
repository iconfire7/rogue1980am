// Package generator implements procedural dungeon level generation.
// The dungeon is divided into a 3×3 grid of sections; each section gets one room.
// Rooms are connected by corridors to form a fully-connected graph.
package generator

import (
	"math/rand"
	"rogue1980am/internal/domain"
)

const (
	MapWidth  = 80
	MapHeight = 40
	Sections  = 3 // grid dimension; produces 3×3 = 9 rooms
)

// itemIDCounter generates monotonically increasing item IDs across a session.
var itemIDCounter int

func nextItemID() int {
	itemIDCounter++
	return itemIDCounter
}

var enemyIDCounter int

func nextEnemyID() int {
	enemyIDCounter++
	return enemyIDCounter
}

// GenerateLevel creates a complete dungeon level scaled to its depth.
func GenerateLevel(depth int, rng *rand.Rand) *domain.Level {
	level := domain.NewLevel(depth, MapWidth, MapHeight)

	// 1. partition map into 3×3 sections and place a room in each
	rooms := placeRooms(level, rng)
	level.Rooms = rooms

	// 2. carve rooms into the tile grid
	for _, r := range rooms {
		carveRoom(level, r)
	}

	// 3. connect rooms with corridors (spanning tree + optional extra edge)
	corridors := connectRooms(level, rooms, rng)
	level.Corridors = corridors

	// 4. pick start and exit rooms (guaranteed to be different)
	startIdx := rng.Intn(len(rooms))
	exitIdx := startIdx
	for exitIdx == startIdx {
		exitIdx = rng.Intn(len(rooms))
	}
	rooms[startIdx].IsStart = true
	rooms[exitIdx].IsExit = true

	// 5. place staircase in exit room
	stairsPos := randomFloorInRoom(rooms[exitIdx], rng)
	level.ExitPos = stairsPos
	t := level.TileAt(stairsPos.X, stairsPos.Y)
	if t != nil {
		t.Type = domain.TileStairs
	}

	// 6. place player start position
	level.StartPos = randomFloorInRoom(rooms[startIdx], rng)

	// 7. populate rooms with enemies and items (skip start room)
	for i, r := range rooms {
		if i == startIdx {
			continue
		}
		populateRoom(level, r, depth, rng)
	}

	return level
}

// placeRooms divides the map into a 3×3 section grid and places one room per section.
func placeRooms(level *domain.Level, rng *rand.Rand) []*domain.Room {
	secW := level.Width / Sections
	secH := level.Height / Sections
	rooms := make([]*domain.Room, 0, Sections*Sections)
	id := 0
	for sy := 0; sy < Sections; sy++ {
		for sx := 0; sx < Sections; sx++ {
			// section bounds with a 1-tile border on each side for walls
			sxMin := sx*secW + 1
			syMin := sy*secH + 1
			sxMax := (sx+1)*secW - 2
			syMax := (sy+1)*secH - 2

			maxW := sxMax - sxMin - 1
			maxH := syMax - syMin - 1
			if maxW < 3 {
				maxW = 3
			}
			if maxH < 3 {
				maxH = 3
			}

			w := 3 + rng.Intn(maxW-2)
			h := 3 + rng.Intn(maxH-2)

			// ensure room fits inside section
			roomXMax := sxMax - w
			roomYMax := syMax - h
			if roomXMax < sxMin {
				roomXMax = sxMin
			}
			if roomYMax < syMin {
				roomYMax = syMin
			}

			rx := sxMin + rng.Intn(roomXMax-sxMin+1)
			ry := syMin + rng.Intn(roomYMax-syMin+1)

			rooms = append(rooms, &domain.Room{
				ID: id,
				X:  rx,
				Y:  ry,
				W:  w,
				H:  h,
			})
			id++
		}
	}
	return rooms
}

// carveRoom writes wall and floor tiles into the level grid.
func carveRoom(level *domain.Level, r *domain.Room) {
	// draw outer walls
	for x := r.X - 1; x <= r.X+r.W; x++ {
		setTile(level, x, r.Y-1, domain.TileWall)
		setTile(level, x, r.Y+r.H, domain.TileWall)
	}
	for y := r.Y - 1; y <= r.Y+r.H; y++ {
		setTile(level, r.X-1, y, domain.TileWall)
		setTile(level, r.X+r.W, y, domain.TileWall)
	}
	// draw floor
	for y := r.Y; y < r.Y+r.H; y++ {
		for x := r.X; x < r.X+r.W; x++ {
			setTile(level, x, y, domain.TileFloor)
		}
	}
}

func setTile(level *domain.Level, x, y int, t domain.TileType) {
	tile := level.TileAt(x, y)
	if tile == nil {
		return
	}
	// don't overwrite floor with wall
	if tile.Type == domain.TileFloor && t == domain.TileWall {
		return
	}
	tile.Type = t
}

// connectRooms builds a minimum spanning tree over the 9 rooms and carves corridors.
func connectRooms(level *domain.Level, rooms []*domain.Room, rng *rand.Rand) []*domain.Corridor {
	n := len(rooms)
	connected := make([]bool, n)
	connected[0] = true
	var corridors []*domain.Corridor

	for connCount := 1; connCount < n; connCount++ {
		// pick a random already-connected room
		var fromIdx int
		for {
			fromIdx = rng.Intn(n)
			if connected[fromIdx] {
				break
			}
		}
		// pick a random not-yet-connected room
		toIdx := rng.Intn(n - connCount)
		unconnected := []int{}
		for i := 0; i < n; i++ {
			if !connected[i] {
				unconnected = append(unconnected, i)
			}
		}
		toIdx = unconnected[rng.Intn(len(unconnected))]
		connected[toIdx] = true

		cor := carveCorridor(level, rooms[fromIdx], rooms[toIdx])
		corridors = append(corridors, cor)
	}
	return corridors
}

// carveCorridor builds an L-shaped corridor between two rooms.
func carveCorridor(level *domain.Level, a, b *domain.Room) *domain.Corridor {
	ac := a.Center()
	bc := b.Center()

	cor := &domain.Corridor{RoomA: a.ID, RoomB: b.ID}

	// horizontal then vertical (or vice versa chosen randomly)
	var path []domain.Position
	if level.Rooms != nil { // always true; just to avoid import cycle lint
		// carve horizontal segment first, then vertical
		path = hThenV(ac, bc)
	}

	for _, pos := range path {
		t := level.TileAt(pos.X, pos.Y)
		if t == nil {
			continue
		}
		if t.Type == domain.TileEmpty || t.Type == domain.TileWall {
			// check if this is a wall adjacent to a room floor; if so make it an opening
			if t.Type == domain.TileWall {
				// determine orientation from neighbours
				above := level.TileAt(pos.X, pos.Y-1)
				below := level.TileAt(pos.X, pos.Y+1)
				if (above != nil && above.Type == domain.TileFloor) ||
					(below != nil && below.Type == domain.TileFloor) {
					t.Type = domain.TileOpeningH
				} else {
					t.Type = domain.TileOpeningV
				}
			} else {
				t.Type = domain.TileCorridor
			}
		}
		cor.Tiles = append(cor.Tiles, domain.CorridorTile{Pos: pos})
	}
	return cor
}

func hThenV(a, b domain.Position) []domain.Position {
	var path []domain.Position
	x, y := a.X, a.Y
	dx := 1
	if b.X < x {
		dx = -1
	}
	for x != b.X {
		path = append(path, domain.Position{X: x, Y: y})
		x += dx
	}
	dy := 1
	if b.Y < y {
		dy = -1
	}
	for y != b.Y {
		path = append(path, domain.Position{X: x, Y: y})
		y += dy
	}
	path = append(path, domain.Position{X: x, Y: y})
	return path
}

// randomFloorInRoom returns a random floor position inside the room.
func randomFloorInRoom(r *domain.Room, rng *rand.Rand) domain.Position {
	x := r.X + rng.Intn(r.W)
	y := r.Y + rng.Intn(r.H)
	return domain.Position{X: x, Y: y}
}

// populateRoom fills a room with enemies and items scaled to depth.
func populateRoom(level *domain.Level, r *domain.Room, depth int, rng *rand.Rand) {
	// enemy count scales up with depth, item count scales down
	maxEnemies := 1 + depth/4
	if maxEnemies > 5 {
		maxEnemies = 5
	}
	numEnemies := rng.Intn(maxEnemies + 1)

	maxItems := 3 - depth/8
	if maxItems < 0 {
		maxItems = 0
	}
	numItems := rng.Intn(maxItems + 2)

	used := make(map[domain.Position]bool)
	used[r.Center()] = true // reserve center

	placeAt := func() domain.Position {
		for attempts := 0; attempts < 50; attempts++ {
			pos := randomFloorInRoom(r, rng)
			if !used[pos] {
				used[pos] = true
				return pos
			}
		}
		return r.Center()
	}

	for i := 0; i < numEnemies; i++ {
		et := randomEnemyType(depth, rng)
		pos := placeAt()
		e := domain.NewEnemy(nextEnemyID(), et, depth, pos)
		level.Enemies = append(level.Enemies, e)
		r.Enemies = append(r.Enemies, e)
	}

	for i := 0; i < numItems; i++ {
		it := randomItem(depth, rng)
		pos := placeAt()
		it.Pos = pos
		it.OnFloor = true
		level.Items = append(level.Items, it)
		r.Items = append(r.Items, it)
	}
}

// randomEnemyType returns an enemy type biased by depth.
func randomEnemyType(depth int, rng *rand.Rand) domain.EnemyType {
	types := []domain.EnemyType{domain.EnemyZombie}
	if depth >= 3 {
		types = append(types, domain.EnemyVampire)
	}
	if depth >= 5 {
		types = append(types, domain.EnemyGhost)
	}
	if depth >= 8 {
		types = append(types, domain.EnemyOgre)
	}
	if depth >= 12 {
		types = append(types, domain.EnemySnakeMage)
	}
	return types[rng.Intn(len(types))]
}

// randomItem creates a random item appropriate for the given depth.
func randomItem(depth int, rng *rand.Rand) *domain.Item {
	id := nextItemID()
	// item type distribution shifts toward fewer useful items at greater depth
	roll := rng.Intn(100)
	if roll < 35 {
		return makeFood(id, rng)
	} else if roll < 55 {
		return makeElixir(id, rng)
	} else if roll < 70 {
		return makeScroll(id, rng)
	} else if roll < 85 {
		return makeWeapon(id, depth, rng)
	}
	// leftover: extra food
	return makeFood(id, rng)
}

func makeFood(id int, rng *rand.Rand) *domain.Item {
	return &domain.Item{
		ID:     id,
		Type:   domain.ItemFood,
		Name:   "Food Ration",
		Health: 8 + rng.Intn(6),
	}
}

func makeElixir(id int, rng *rand.Rand) *domain.Item {
	subtypes := []domain.ItemSubtype{domain.SubtypeDexterity, domain.SubtypeStrength, domain.SubtypeMaxHealth}
	sub := subtypes[rng.Intn(len(subtypes))]
	names := map[domain.ItemSubtype]string{
		domain.SubtypeDexterity: "Elixir of Agility",
		domain.SubtypeStrength:  "Elixir of Power",
		domain.SubtypeMaxHealth: "Elixir of Vitality",
	}
	it := &domain.Item{
		ID:       id,
		Type:     domain.ItemElixir,
		Subtype:  sub,
		Name:     names[sub],
		Duration: 15 + rng.Intn(10),
	}
	switch sub {
	case domain.SubtypeDexterity:
		it.Dexterity = 2 + rng.Intn(3)
	case domain.SubtypeStrength:
		it.Strength = 2 + rng.Intn(3)
	case domain.SubtypeMaxHealth:
		it.MaxHealth = 4 + rng.Intn(4)
	}
	return it
}

func makeScroll(id int, rng *rand.Rand) *domain.Item {
	subtypes := []domain.ItemSubtype{domain.SubtypeDexterity, domain.SubtypeStrength, domain.SubtypeMaxHealth}
	sub := subtypes[rng.Intn(len(subtypes))]
	names := map[domain.ItemSubtype]string{
		domain.SubtypeDexterity: "Scroll of Agility",
		domain.SubtypeStrength:  "Scroll of Power",
		domain.SubtypeMaxHealth: "Scroll of Fortitude",
	}
	it := &domain.Item{
		ID:      id,
		Type:    domain.ItemScroll,
		Subtype: sub,
		Name:    names[sub],
	}
	switch sub {
	case domain.SubtypeDexterity:
		it.Dexterity = 1 + rng.Intn(2)
	case domain.SubtypeStrength:
		it.Strength = 1 + rng.Intn(2)
	case domain.SubtypeMaxHealth:
		it.MaxHealth = 2 + rng.Intn(3)
	}
	return it
}

func makeWeapon(id int, depth int, rng *rand.Rand) *domain.Item {
	names := []string{"Dagger", "Shortsword", "Mace", "Battleaxe", "Longsword"}
	n := names[rng.Intn(len(names))]
	str := 2 + depth/5 + rng.Intn(4)
	return &domain.Item{
		ID:       id,
		Type:     domain.ItemWeapon,
		Name:     n,
		Strength: str,
	}
}
