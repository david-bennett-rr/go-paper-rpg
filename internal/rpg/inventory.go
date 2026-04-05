package rpg

const MaxItems = 10

// Item represents a usable item.
type Item struct {
	ID          string
	Name        string
	Description string
	Effect      ItemEffect
	BuyPrice    int
	SellPrice   int
}

type ItemEffect struct {
	Type   string // "heal_hp", "heal_fp", "damage", "status"
	Amount int
}

// Inventory manages the player's items with a Paper Mario-style cap.
type Inventory struct {
	Items []*Item
}

func NewInventory() *Inventory {
	return &Inventory{
		Items: make([]*Item, 0, MaxItems),
	}
}

func (inv *Inventory) IsFull() bool {
	return len(inv.Items) >= MaxItems
}

func (inv *Inventory) Add(item *Item) bool {
	if inv.IsFull() {
		return false
	}
	inv.Items = append(inv.Items, item)
	return true
}

func (inv *Inventory) Remove(index int) *Item {
	if index < 0 || index >= len(inv.Items) {
		return nil
	}
	item := inv.Items[index]
	inv.Items = append(inv.Items[:index], inv.Items[index+1:]...)
	return item
}

func (inv *Inventory) Count() int {
	return len(inv.Items)
}
