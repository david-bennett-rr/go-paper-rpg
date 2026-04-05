package rpg

// Badge represents an equippable modifier (Paper Mario badge system).
type Badge struct {
	ID          string
	Name        string
	Description string
	BPCost      int
	Effect      BadgeEffect
}

type BadgeEffect struct {
	Type   string // "stat_modifier", "grant_move", "passive"
	Stat   string // For stat_modifier: "attack", "defense"
	Amount int    // For stat_modifier
	Move   *Move  // For grant_move
}

// BadgeManager handles equipping/unequipping badges within BP limits.
type BadgeManager struct {
	Owned    []*Badge
	Equipped []*Badge
	MaxBP    int
	UsedBP   int
}

func NewBadgeManager(maxBP int) *BadgeManager {
	return &BadgeManager{
		Owned:    make([]*Badge, 0),
		Equipped: make([]*Badge, 0),
		MaxBP:    maxBP,
	}
}

func (bm *BadgeManager) AddBadge(b *Badge) {
	bm.Owned = append(bm.Owned, b)
}

func (bm *BadgeManager) Equip(b *Badge) bool {
	if bm.UsedBP+b.BPCost > bm.MaxBP {
		return false
	}
	bm.Equipped = append(bm.Equipped, b)
	bm.UsedBP += b.BPCost
	return true
}

func (bm *BadgeManager) Unequip(index int) {
	if index < 0 || index >= len(bm.Equipped) {
		return
	}
	b := bm.Equipped[index]
	bm.UsedBP -= b.BPCost
	bm.Equipped = append(bm.Equipped[:index], bm.Equipped[index+1:]...)
}

// AttackBonus returns the total attack bonus from equipped badges.
func (bm *BadgeManager) AttackBonus() int {
	bonus := 0
	for _, b := range bm.Equipped {
		if b.Effect.Type == "stat_modifier" && b.Effect.Stat == "attack" {
			bonus += b.Effect.Amount
		}
	}
	return bonus
}

// DefenseBonus returns the total defense bonus from equipped badges.
func (bm *BadgeManager) DefenseBonus() int {
	bonus := 0
	for _, b := range bm.Equipped {
		if b.Effect.Type == "stat_modifier" && b.Effect.Stat == "defense" {
			bonus += b.Effect.Amount
		}
	}
	return bonus
}

// GrantedMoves returns all moves granted by equipped badges.
func (bm *BadgeManager) GrantedMoves() []*Move {
	moves := make([]*Move, 0)
	for _, b := range bm.Equipped {
		if b.Effect.Type == "grant_move" && b.Effect.Move != nil {
			moves = append(moves, b.Effect.Move)
		}
	}
	return moves
}
