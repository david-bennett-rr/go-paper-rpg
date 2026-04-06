package action

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
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

	screenW := float32(screen.Bounds().Dx())
	screenH := float32(screen.Bounds().Dy())

	barW := screenW * 0.4
	if barW < 300 {
		barW = 300
	}
	barH := float32(20)
	barX := (screenW - barW) / 2
	barY := screenH/2 - barH/2

	// Background bar
	vector.DrawFilledRect(screen, barX-4, barY-4, barW+8, barH+8, color.RGBA{R: 20, G: 22, B: 30, A: 220}, false)
	vector.DrawFilledRect(screen, barX, barY, barW, barH, color.RGBA{R: 40, G: 40, B: 50, A: 230}, false)
	vector.StrokeRect(screen, barX, barY, barW, barH, 2, color.RGBA{R: 70, G: 72, B: 90, A: 255}, false)

	if tp.windowEnd > 0 {
		totalDuration := float32(tp.windowEnd + 12)

		// Window zone (green area)
		winStartPct := float32(tp.windowStart) / totalDuration
		winEndPct := float32(tp.windowEnd) / totalDuration
		winX := barX + winStartPct*barW
		winW := (winEndPct - winStartPct) * barW
		vector.DrawFilledRect(screen, winX, barY, winW, barH, color.RGBA{R: 50, G: 160, B: 60, A: 220}, false)

		// Sweet spot marker (bright yellow line)
		sweetPct := float32(tp.sweetSpot) / totalDuration
		sweetX := barX + sweetPct*barW
		vector.DrawFilledRect(screen, sweetX-2, barY-4, 4, barH+8, color.RGBA{R: 255, G: 220, B: 60, A: 255}, false)

		// Current position cursor (white)
		curPct := float32(tp.currentTick) / totalDuration
		curX := barX + curPct*barW
		vector.DrawFilledRect(screen, curX-3, barY-6, 6, barH+12, color.RGBA{R: 255, G: 255, B: 255, A: 255}, false)
	}

	// "Press A!" label centered above bar
	label := "Press A!"
	labelX := int(barX+barW/2) - len(label)*3
	ebitenutil.DebugPrintAt(screen, label, labelX, int(barY)-22)
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
