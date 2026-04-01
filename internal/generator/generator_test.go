package generator_test

import (
	"math/rand"
	"testing"

	"rogue1980am/internal/generator"
)

func TestGenerateLevelStructure(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	lvl := generator.GenerateLevel(1, rng)

	if lvl == nil {
		t.Fatal("level should not be nil")
	}
	if len(lvl.Rooms) != 9 {
		t.Fatalf("expected 9 rooms, got %d", len(lvl.Rooms))
	}
	if len(lvl.Corridors) < 8 {
		t.Fatalf("expected at least 8 corridors for a spanning tree, got %d", len(lvl.Corridors))
	}
}

func TestGenerateLevelHasStartAndExit(t *testing.T) {
	rng := rand.New(rand.NewSource(123))
	lvl := generator.GenerateLevel(3, rng)

	startCount := 0
	exitCount := 0
	for _, r := range lvl.Rooms {
		if r.IsStart {
			startCount++
		}
		if r.IsExit {
			exitCount++
		}
	}
	if startCount != 1 {
		t.Fatalf("expected 1 start room, got %d", startCount)
	}
	if exitCount != 1 {
		t.Fatalf("expected 1 exit room, got %d", exitCount)
	}
}

func TestGenerateLevelStartAndExitDiffer(t *testing.T) {
	rng := rand.New(rand.NewSource(999))
	lvl := generator.GenerateLevel(1, rng)

	var startRoom, exitRoom interface{}
	for _, r := range lvl.Rooms {
		if r.IsStart {
			startRoom = r
		}
		if r.IsExit {
			exitRoom = r
		}
	}
	if startRoom == exitRoom {
		t.Fatal("start and exit rooms should be different")
	}
}

func TestGenerateLevelStartPlayerPos(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	lvl := generator.GenerateLevel(1, rng)

	if !lvl.IsWalkable(lvl.StartPos.X, lvl.StartPos.Y) {
		t.Fatal("player start position should be walkable")
	}
}

func TestGenerateLevelStairsPresent(t *testing.T) {
	rng := rand.New(rand.NewSource(55))
	lvl := generator.GenerateLevel(1, rng)

	stairsFound := false
	for y := 0; y < lvl.Height; y++ {
		for x := 0; x < lvl.Width; x++ {
			t := lvl.TileAt(x, y)
			if t != nil && t.Type == 6 { // TileStairs = 6
				stairsFound = true
			}
		}
	}
	if !stairsFound {
		t.Fatal("level should contain stairs")
	}
}

func TestGenerateLevelDepthScaling(t *testing.T) {
	rng1 := rand.New(rand.NewSource(1))
	lvl1 := generator.GenerateLevel(1, rng1)
	rng21 := rand.New(rand.NewSource(1))
	lvl21 := generator.GenerateLevel(21, rng21)

	totalHP1 := 0
	for _, e := range lvl1.Enemies {
		totalHP1 += e.MaxHealth
	}
	totalHP21 := 0
	for _, e := range lvl21.Enemies {
		totalHP21 += e.MaxHealth
	}

	// Deeper levels should generally have higher total enemy HP
	// (This is probabilistic but reliable with fixed seed and many enemies)
	if len(lvl21.Enemies) > 0 && len(lvl1.Enemies) > 0 {
		avgHP1 := totalHP1 / len(lvl1.Enemies)
		avgHP21 := totalHP21 / len(lvl21.Enemies)
		if avgHP21 <= avgHP1 {
			t.Errorf("expected deeper level enemies to have more HP on average; lvl1 avg=%d, lvl21 avg=%d", avgHP1, avgHP21)
		}
	}
}
