package domain

// ItemType categorises an item.
type ItemType int

const (
	ItemTreasure ItemType = iota
	ItemFood
	ItemElixir
	ItemScroll
	ItemWeapon
)

// ItemSubtype further specialises elixirs and scrolls.
type ItemSubtype int

const (
	SubtypeNone ItemSubtype = iota
	SubtypeDexterity
	SubtypeStrength
	SubtypeMaxHealth
)

// BackpackCapacity is the maximum number of each item type (except treasure).
const BackpackCapacity = 9

// Item represents a collectible object in the dungeon.
type Item struct {
	ID         int
	Type       ItemType
	Subtype    ItemSubtype
	Name       string
	Health     int // HP restored (food)
	MaxHealth  int // max HP increase (scrolls / elixirs)
	Dexterity  int // dex increase (scrolls / elixirs)
	Strength   int // str increase (scrolls / elixirs / weapons)
	Value      int // gold value (treasure)
	Duration   int // turns remaining for temporary effects (elixirs); 0 = permanent
	Pos        Position
	OnFloor    bool
}

// IsTemporary returns true if the item has a time-limited effect.
func (it *Item) IsTemporary() bool {
	return it.Type == ItemElixir && it.Duration > 0
}

// Backpack stores the player's carried items.
type Backpack struct {
	Treasure  int    // cumulative gold
	Foods     []*Item
	Elixirs   []*Item
	Scrolls   []*Item
	Weapons   []*Item
}

// NewBackpack creates an empty backpack.
func NewBackpack() *Backpack {
	return &Backpack{}
}

// CanAdd returns true if there is room for the given item type.
func (b *Backpack) CanAdd(t ItemType) bool {
	switch t {
	case ItemTreasure:
		return true
	case ItemFood:
		return len(b.Foods) < BackpackCapacity
	case ItemElixir:
		return len(b.Elixirs) < BackpackCapacity
	case ItemScroll:
		return len(b.Scrolls) < BackpackCapacity
	case ItemWeapon:
		return len(b.Weapons) < BackpackCapacity
	}
	return false
}

// Add places an item into the backpack.  Returns true on success.
func (b *Backpack) Add(item *Item) bool {
	if !b.CanAdd(item.Type) {
		return false
	}
	item.OnFloor = false
	switch item.Type {
	case ItemTreasure:
		b.Treasure += item.Value
		return true
	case ItemFood:
		b.Foods = append(b.Foods, item)
	case ItemElixir:
		b.Elixirs = append(b.Elixirs, item)
	case ItemScroll:
		b.Scrolls = append(b.Scrolls, item)
	case ItemWeapon:
		b.Weapons = append(b.Weapons, item)
	}
	return true
}

// RemoveFood removes the food at index i and returns it.
func (b *Backpack) RemoveFood(i int) *Item {
	if i < 0 || i >= len(b.Foods) {
		return nil
	}
	it := b.Foods[i]
	b.Foods = append(b.Foods[:i], b.Foods[i+1:]...)
	return it
}

// RemoveElixir removes the elixir at index i and returns it.
func (b *Backpack) RemoveElixir(i int) *Item {
	if i < 0 || i >= len(b.Elixirs) {
		return nil
	}
	it := b.Elixirs[i]
	b.Elixirs = append(b.Elixirs[:i], b.Elixirs[i+1:]...)
	return it
}

// RemoveScroll removes the scroll at index i and returns it.
func (b *Backpack) RemoveScroll(i int) *Item {
	if i < 0 || i >= len(b.Scrolls) {
		return nil
	}
	it := b.Scrolls[i]
	b.Scrolls = append(b.Scrolls[:i], b.Scrolls[i+1:]...)
	return it
}
