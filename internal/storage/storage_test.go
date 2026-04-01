package storage_test

import (
	"os"
	"testing"

	"rogue1980am/internal/domain"
	"rogue1980am/internal/storage"
)

func TestSaveAndLoadStatistics(t *testing.T) {
	// Use a temp dir to avoid polluting user config
	tmp := t.TempDir()
	os.Setenv("HOME", tmp)
	os.Setenv("XDG_CONFIG_HOME", tmp)

	s := &domain.Statistics{
		RunID:         1,
		TreasureTotal: 500,
		DeepestLevel:  10,
		EnemiesKilled: 25,
	}
	if err := storage.SaveStatistics(s); err != nil {
		t.Fatalf("SaveStatistics failed: %v", err)
	}

	all, err := storage.LoadAllStatistics()
	if err != nil {
		t.Fatalf("LoadAllStatistics failed: %v", err)
	}
	if len(all) == 0 {
		t.Fatal("expected at least one statistic entry")
	}
	if all[0].TreasureTotal != 500 {
		t.Fatalf("expected treasure 500, got %d", all[0].TreasureTotal)
	}
}

func TestStatisticsLeaderboardOrder(t *testing.T) {
	tmp := t.TempDir()
	os.Setenv("HOME", tmp)
	os.Setenv("XDG_CONFIG_HOME", tmp)

	for i, gold := range []int{100, 500, 300} {
		_ = storage.SaveStatistics(&domain.Statistics{RunID: i + 1, TreasureTotal: gold})
	}

	all, err := storage.LoadAllStatistics()
	if err != nil {
		t.Fatalf("LoadAllStatistics failed: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(all))
	}
	// Should be sorted descending
	if all[0].TreasureTotal < all[1].TreasureTotal || all[1].TreasureTotal < all[2].TreasureTotal {
		t.Fatal("statistics should be sorted by treasure descending")
	}
}

func TestPlayerSaveRestore(t *testing.T) {
	c := domain.NewCharacter()
	c.MaxHealth = 30
	c.Health = 20
	c.Dexterity = 7
	c.Strength = 9
	c.Pack.Foods = append(c.Pack.Foods, &domain.Item{
		ID: 42, Type: domain.ItemFood, Name: "Ration", Health: 8,
	})

	ps := playerToSave(c)
	restored := storage.RestorePlayer(ps)

	if restored.MaxHealth != 30 || restored.Health != 20 {
		t.Fatalf("health mismatch: max=%d health=%d", restored.MaxHealth, restored.Health)
	}
	if restored.Dexterity != 7 || restored.Strength != 9 {
		t.Fatalf("stat mismatch: dex=%d str=%d", restored.Dexterity, restored.Strength)
	}
	if len(restored.Pack.Foods) != 1 {
		t.Fatalf("expected 1 food item, got %d", len(restored.Pack.Foods))
	}
}

// playerToSave is a helper that replicates the unexported conversion.
func playerToSave(c *domain.Character) *storage.PlayerSave {
	ps := &storage.PlayerSave{
		PosX:      c.Pos.X,
		PosY:      c.Pos.Y,
		MaxHealth: c.MaxHealth,
		Health:    c.Health,
		Dexterity: c.Dexterity,
		Strength:  c.Strength,
		Treasure:  c.Pack.Treasure,
	}
	for _, it := range c.Pack.Foods {
		ps.Foods = append(ps.Foods, &storage.ItemSave{
			ID: it.ID, Type: int(it.Type), Name: it.Name, Health: it.Health,
		})
	}
	return ps
}
