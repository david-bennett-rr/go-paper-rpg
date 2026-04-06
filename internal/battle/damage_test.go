package battle

import (
	"testing"

	"github.com/davidbennett/go-paper-rpg/internal/battle/action"
	"github.com/davidbennett/go-paper-rpg/internal/rpg"
)

func TestCalculateDamage(t *testing.T) {
	miss := action.ResultFromQuality(action.QualityMiss)
	excellent := action.ResultFromQuality(action.QualityExcellent)

	tests := []struct {
		name     string
		attacker rpg.Stats
		move     rpg.Move
		defender rpg.Stats
		cmd      action.CommandResult
		want     int
	}{
		{
			name:     "basic attack miss",
			attacker: rpg.Stats{Attack: 1},
			move:     rpg.Move{BasePower: 1},
			defender: rpg.Stats{Defense: 0},
			cmd:      miss,
			want:     2, // (1+1)*1.0 - 0 = 2
		},
		{
			name:     "excellent attack",
			attacker: rpg.Stats{Attack: 1},
			move:     rpg.Move{BasePower: 1},
			defender: rpg.Stats{Defense: 0},
			cmd:      excellent,
			want:     4, // (1+1)*2.0 - 0 = 4
		},
		{
			name:     "high defense reduces to 1",
			attacker: rpg.Stats{Attack: 1},
			move:     rpg.Move{BasePower: 1},
			defender: rpg.Stats{Defense: 10},
			cmd:      miss,
			want:     1, // Minimum 1 since attacker has power
		},
		{
			name:     "zero power zero damage",
			attacker: rpg.Stats{Attack: 0},
			move:     rpg.Move{BasePower: 0},
			defender: rpg.Stats{Defense: 5},
			cmd:      miss,
			want:     0, // No power at all = 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateDamage(&tt.attacker, &tt.move, &tt.defender, tt.cmd)
			if got != tt.want {
				t.Errorf("CalculateDamage() = %d, want %d", got, tt.want)
			}
		})
	}
}
