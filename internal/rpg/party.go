package rpg

// PartyMember represents a playable character.
type PartyMember struct {
	Name  string
	Stats Stats
	Moves []*Move
}

// Party manages the player's active party.
type Party struct {
	Mario *PartyMember
}

func NewParty() *Party {
	mario := &PartyMember{
		Name: "Neu",
		Stats: Stats{
			HP: 10, MaxHP: 10,
			FP: 5, MaxFP: 5,
			Attack: 1, Defense: 0,
		},
		Moves: []*Move{
			{
				Name:          "Slash",
				BasePower:     1,
				FPCost:        0,
				Type:          MoveTypeSword,
				ActionCommand: "double_slash",
			},
		},
	}

	return &Party{
		Mario: mario,
	}
}
