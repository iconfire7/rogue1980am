package domain_test

import (
	"testing"

	"rogue1980am/internal/domain"
)

func TestCharacterAlive(t *testing.T) {
	c := domain.NewCharacter()
	if !c.IsAlive() {
		t.Fatal("new character should be alive")
	}
	c.Health = 0
	if c.IsAlive() {
		t.Fatal("character with 0 HP should not be alive")
	}
}

func TestBackpackAdd(t *testing.T) {
	bp := domain.NewBackpack()
	food := &domain.Item{ID: 1, Type: domain.ItemFood, Name: "Food", Health: 5}
	if !bp.Add(food) {
		t.Fatal("should add food to empty backpack")
	}
	if len(bp.Foods) != 1 {
		t.Fatalf("expected 1 food, got %d", len(bp.Foods))
	}
}

func TestBackpackCapacity(t *testing.T) {
	bp := domain.NewBackpack()
	for i := 0; i < domain.BackpackCapacity; i++ {
		bp.Add(&domain.Item{ID: i, Type: domain.ItemFood, Name: "Food"})
	}
	if bp.CanAdd(domain.ItemFood) {
		t.Fatal("backpack should be full")
	}
}

func TestBackpackTreasureCumulative(t *testing.T) {
	bp := domain.NewBackpack()
	for i := 0; i < 20; i++ {
		bp.Add(&domain.Item{ID: i, Type: domain.ItemTreasure, Value: 10})
	}
	if bp.Treasure != 200 {
		t.Fatalf("expected 200 gold, got %d", bp.Treasure)
	}
}

func TestCharacterApplyFood(t *testing.T) {
	c := domain.NewCharacter()
	c.Health = 5
	food := &domain.Item{Type: domain.ItemFood, Health: 10}
	c.ApplyItem(food)
	if c.Health != 15 {
		t.Fatalf("expected 15 HP, got %d", c.Health)
	}
}

func TestCharacterApplyFoodCap(t *testing.T) {
	c := domain.NewCharacter()
	c.Health = 18
	food := &domain.Item{Type: domain.ItemFood, Health: 10}
	c.ApplyItem(food)
	if c.Health != c.MaxHealth {
		t.Fatalf("expected HP capped at %d, got %d", c.MaxHealth, c.Health)
	}
}

func TestCharacterApplyScroll(t *testing.T) {
	c := domain.NewCharacter()
	oldStr := c.Strength
	scroll := &domain.Item{Type: domain.ItemScroll, Subtype: domain.SubtypeStrength, Strength: 3}
	c.ApplyItem(scroll)
	if c.Strength != oldStr+3 {
		t.Fatalf("expected strength %d, got %d", oldStr+3, c.Strength)
	}
	if c.ScrollsRead != 1 {
		t.Fatal("scrolls read counter should be 1")
	}
}

func TestElixirTemporaryEffect(t *testing.T) {
	c := domain.NewCharacter()
	elixir := &domain.Item{
		Type:    domain.ItemElixir,
		Subtype: domain.SubtypeDexterity,
		Dexterity: 4,
		Duration: 2,
	}
	origDex := c.Dexterity
	c.ApplyItem(elixir)
	if c.Dexterity != origDex+4 {
		t.Fatalf("elixir should add 4 dex; got %d", c.Dexterity)
	}
	// tick twice to expire
	c.TickEffects()
	c.TickEffects()
	if c.Dexterity != origDex {
		t.Fatalf("elixir effect should have expired; dex is %d", c.Dexterity)
	}
}

func TestElixirExpiryClampsHealth(t *testing.T) {
	c := domain.NewCharacter()
	c.MaxHealth = 5
	c.Health = 5
	elixir := &domain.Item{
		Type:      domain.ItemElixir,
		Subtype:   domain.SubtypeMaxHealth,
		MaxHealth: 10,
		Duration:  1,
	}
	c.ApplyItem(elixir)
	// now max=15, health=15
	if c.MaxHealth != 15 || c.Health != 15 {
		t.Fatalf("expected max=15, health=15; got max=%d, health=%d", c.MaxHealth, c.Health)
	}
	c.Health = 3 // reduce before expiry
	c.TickEffects() // expires the elixir; max goes back to 5
	if c.Health < 1 {
		t.Fatalf("health should not drop below 1 on elixir expiry; got %d", c.Health)
	}
}

func TestPositionAdd(t *testing.T) {
	p := domain.Position{X: 3, Y: 4}
	q := p.Add(1, -1)
	if q.X != 4 || q.Y != 3 {
		t.Fatalf("unexpected position %v", q)
	}
}

func TestManhattanDistance(t *testing.T) {
	a := domain.Position{X: 0, Y: 0}
	b := domain.Position{X: 3, Y: 4}
	if a.ManhattanDistance(b) != 7 {
		t.Fatalf("expected 7, got %d", a.ManhattanDistance(b))
	}
}

func TestNewEnemyAlive(t *testing.T) {
	e := domain.NewEnemy(1, domain.EnemyZombie, 1, domain.Position{})
	if !e.IsAlive() {
		t.Fatal("new enemy should be alive")
	}
}

func TestGameSessionMessages(t *testing.T) {
	gs := domain.NewGameSession(1)
	for i := 0; i < 25; i++ {
		gs.AddMessage("msg")
	}
	if len(gs.Messages) > 20 {
		t.Fatalf("messages should be capped at 20, got %d", len(gs.Messages))
	}
}

func TestLevelTileAt(t *testing.T) {
	lvl := domain.NewLevel(1, 10, 10)
	tile := lvl.TileAt(5, 5)
	if tile == nil {
		t.Fatal("tile should not be nil")
	}
	if lvl.TileAt(-1, 0) != nil {
		t.Fatal("out-of-bounds should return nil")
	}
}

func TestLevelWalkable(t *testing.T) {
	lvl := domain.NewLevel(1, 10, 10)
	lvl.Tiles[3][3].Type = domain.TileFloor
	if !lvl.IsWalkable(3, 3) {
		t.Fatal("floor should be walkable")
	}
	if lvl.IsWalkable(0, 0) {
		t.Fatal("empty tile should not be walkable")
	}
}
