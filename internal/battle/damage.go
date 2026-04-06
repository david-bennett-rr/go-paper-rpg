package battle

import (
	"github.com/davidbennett/go-paper-rpg/internal/battle/action"
	"github.com/davidbennett/go-paper-rpg/internal/rpg"
)

// CalculateDamage computes damage using the Paper Mario formula:
// damage = max(attack + basePower - defense, 0), minimum 1 if attack > 0.
// The action command bonus multiplies the base power before calculation.
func CalculateDamage(attacker *rpg.Stats, move *rpg.Move, defender *rpg.Stats, cmdResult action.CommandResult) int {
	// Apply the action command multiplier to the full offensive total (attack + basePower)
	// so timing bonuses are meaningful even with low base power.
	rawOffense := float64(move.BasePower+attacker.Attack) * cmdResult.BonusMult
	raw := int(rawOffense) - defender.Defense

	if raw <= 0 {
		if move.BasePower+attacker.Attack > 0 {
			return 1
		}
		return 0
	}

	return raw
}
