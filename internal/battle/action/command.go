package action

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/davidbennett/go-paper-rpg/internal/input"
)

// Quality represents how well the player performed an action command.
type Quality int

const (
	QualityMiss Quality = iota
	QualityNice
	QualityGood
	QualityGreat
	QualityExcellent
)

func (q Quality) String() string {
	switch q {
	case QualityNice:
		return "Nice!"
	case QualityGood:
		return "Good!"
	case QualityGreat:
		return "Great!"
	case QualityExcellent:
		return "Excellent!"
	default:
		return "Miss..."
	}
}

// CommandResult holds the outcome of an action command.
type CommandResult struct {
	Success   bool
	Quality   Quality
	BonusMult float64 // 1.0 for miss, up to 2.0 for excellent
}

func ResultFromQuality(q Quality) CommandResult {
	mults := map[Quality]float64{
		QualityMiss:      1.0,
		QualityNice:      1.25,
		QualityGood:      1.5,
		QualityGreat:     1.75,
		QualityExcellent: 2.0,
	}
	return CommandResult{
		Success:   q > QualityMiss,
		Quality:   q,
		BonusMult: mults[q],
	}
}

// ActionCommand is the interface for all action command mini-games.
type ActionCommand interface {
	Start()
	Update(inp *input.Manager)
	Draw(screen *ebiten.Image)
	IsComplete() bool
	Result() CommandResult
}
