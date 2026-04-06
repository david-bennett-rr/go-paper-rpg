package rpg

// PartyMember represents a playable character (Mario or a partner).
type PartyMember struct {
	Name   string
	Stats  Stats
	Moves  []*Move
	IsMain bool // true for Mario
}

// Party manages the player's active party.
type Party struct {
	Mario         *PartyMember
	Partners      []*PartyMember
	ActivePartner int // Index into Partners
	Coins         int
	StarPoints    int
	StarPower     float64
	MaxStarPower  float64
}

func NewParty() *Party {
	mario := &PartyMember{
		Name:   "Neu",
		IsMain: true,
		Stats: Stats{
			HP: 10, MaxHP: 10,
			FP: 5, MaxFP: 5,
			Attack: 1, Defense: 0,
			Level: 1,
		},
		Moves: []*Move{
			{
				Name:          "Slash",
				BasePower:     1,
				FPCost:        0,
				Type:          MoveTypeSword,
				ActionCommand: "double_slash",
				Description:   "Two quick sword cuts. Press A on each swing.",
			},
		},
	}

	return &Party{
		Mario:        mario,
		Partners:     make([]*PartyMember, 0),
		MaxStarPower: 1.0,
	}
}

func (p *Party) ActivePartnerMember() *PartyMember {
	if len(p.Partners) == 0 || p.ActivePartner >= len(p.Partners) {
		return nil
	}
	return p.Partners[p.ActivePartner]
}

func (p *Party) AddPartner(partner *PartyMember) {
	p.Partners = append(p.Partners, partner)
}
