package rpg

// Stats represents the combat statistics of a character or enemy.
type Stats struct {
	HP      int
	MaxHP   int
	FP      int
	MaxFP   int
	Attack  int
	Defense int
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

// Move represents an attack or ability.
type Move struct {
	Name          string
	BasePower     int
	FPCost        int
	Type          MoveType
	ActionCommand string
}

type MoveType string

const (
	MoveTypeSword MoveType = "sword"
)
