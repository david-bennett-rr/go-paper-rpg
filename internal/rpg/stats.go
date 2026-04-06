package rpg

// Stats represents the combat statistics of a character or enemy.
type Stats struct {
	HP      int
	MaxHP   int
	FP      int
	MaxFP   int
	Attack  int
	Defense int
	Level   int
}

// IsAlive returns true if the character has HP remaining.
func (s *Stats) IsAlive() bool {
	return s.HP > 0
}

// TakeDamage reduces HP by the given amount, clamping to 0.
func (s *Stats) TakeDamage(amount int) {
	s.HP -= amount
	if s.HP < 0 {
		s.HP = 0
	}
}

// Heal restores HP up to MaxHP.
func (s *Stats) Heal(amount int) {
	s.HP += amount
	if s.HP > s.MaxHP {
		s.HP = s.MaxHP
	}
}

// RestoreFP restores FP up to MaxFP.
func (s *Stats) RestoreFP(amount int) {
	s.FP += amount
	if s.FP > s.MaxFP {
		s.FP = s.MaxFP
	}
}

// SpendFP deducts FP. Returns false if not enough FP.
func (s *Stats) SpendFP(amount int) bool {
	if s.FP < amount {
		return false
	}
	s.FP -= amount
	return true
}

// Move represents an attack or ability.
type Move struct {
	Name          string
	BasePower     int
	FPCost        int
	Type          MoveType // "jump", "hammer", "special", etc.
	ActionCommand string   // Which action command type to use
	Description   string
}

type MoveType string

const (
	MoveTypeJump    MoveType = "jump"
	MoveTypeHammer  MoveType = "hammer"
	MoveTypeSword   MoveType = "sword"
	MoveTypeSpecial MoveType = "special"
	MoveTypeItem    MoveType = "item"
)
