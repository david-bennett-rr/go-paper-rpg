package action

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/davidbennett/go-paper-rpg/internal/input"
)

// TimedPress is an action command where the player must press Confirm
// within a timing window. Closer to the sweet spot = better quality.
type TimedPress struct {
	windowStart int // Tick when the window opens
	windowEnd   int // Tick when the window closes
	sweetSpot   int // Tick for Excellent rating
	currentTick int
	pressed     bool
	complete    bool
	result      CommandResult
}

func NewTimedPress(windowStart, windowEnd, sweetSpot int) *TimedPress {
	return &TimedPress{
		windowStart: windowStart,
		windowEnd:   windowEnd,
		sweetSpot:   sweetSpot,
	}
}

func (tp *TimedPress) Start() {
	tp.currentTick = 0
	tp.pressed = false
	tp.complete = false
	tp.result = CommandResult{Quality: QualityMiss, BonusMult: 1.0}
}

func (tp *TimedPress) Update(inp *input.Manager) {
	if tp.complete {
		return
	}

	tp.currentTick++

	if inp.Handler().ActionIsJustPressed(input.ActionConfirm) && !tp.pressed {
		tp.pressed = true

		if tp.currentTick >= tp.windowStart && tp.currentTick <= tp.windowEnd {
			distance := tp.currentTick - tp.sweetSpot
			if distance < 0 {
				distance = -distance
			}
			tp.result = ResultFromQuality(qualityFromDistance(distance, tp.windowEnd-tp.windowStart))
		} else {
			tp.result = ResultFromQuality(QualityMiss)
		}

		tp.complete = true
	}

	// Auto-miss if window passes without press
	if tp.currentTick > tp.windowEnd && !tp.pressed {
		tp.pressed = true
		tp.result = ResultFromQuality(QualityMiss)
		tp.complete = true
	}
}

func (tp *TimedPress) Draw(screen *ebiten.Image) {
	if tp.complete {
		return
	}

	// Draw a timing bar indicator
	barX := float32(160)
	barY := float32(200)
	barW := float32(160)
	barH := float32(12)

	// Background bar
	vector.DrawFilledRect(screen, barX, barY, barW, barH, color.RGBA{R: 40, G: 40, B: 40, A: 200}, true)

	// Window zone (green area)
	if tp.windowEnd > 0 {
		totalDuration := float32(tp.windowEnd + 10) // Total animation length estimate
		winStartPct := float32(tp.windowStart) / totalDuration
		winEndPct := float32(tp.windowEnd) / totalDuration
		winX := barX + winStartPct*barW
		winW := (winEndPct - winStartPct) * barW
		vector.DrawFilledRect(screen, winX, barY, winW, barH, color.RGBA{R: 50, G: 180, B: 50, A: 200}, true)

		// Sweet spot marker (bright yellow line)
		sweetPct := float32(tp.sweetSpot) / totalDuration
		sweetX := barX + sweetPct*barW
		vector.DrawFilledRect(screen, sweetX-1, barY, 2, barH, color.RGBA{R: 255, G: 255, B: 0, A: 255}, true)
	}

	// Current position cursor (white)
	if tp.windowEnd > 0 {
		totalDuration := float32(tp.windowEnd + 10)
		curPct := float32(tp.currentTick) / totalDuration
		curX := barX + curPct*barW
		vector.DrawFilledRect(screen, curX-2, barY-2, 4, barH+4, color.RGBA{R: 255, G: 255, B: 255, A: 255}, true)
	}
}

func (tp *TimedPress) IsComplete() bool {
	return tp.complete
}

func (tp *TimedPress) Result() CommandResult {
	return tp.result
}

// qualityFromDistance maps the distance from the sweet spot to a quality rating.
func qualityFromDistance(distance, windowSize int) Quality {
	if windowSize <= 0 {
		return QualityMiss
	}

	ratio := float64(distance) / float64(windowSize)

	switch {
	case ratio <= 0.05: // Within 5% of sweet spot
		return QualityExcellent
	case ratio <= 0.15:
		return QualityGreat
	case ratio <= 0.30:
		return QualityGood
	case ratio <= 0.50:
		return QualityNice
	default:
		return QualityNice // Still in window = at least Nice
	}
}
