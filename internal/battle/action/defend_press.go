package action

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/davidbennett/go-paper-rpg/internal/input"
)

// DefendPress is an action command where the player must press Cancel (B)
// within a timing window to reduce incoming damage. Same mechanic as TimedPress
// but uses the B button and yields a defense reduction multiplier.
type DefendPress struct {
	windowStart int
	windowEnd   int
	sweetSpot   int
	currentTick int
	pressed     bool
	complete    bool
	result      CommandResult
}

func NewDefendPress(windowStart, windowEnd, sweetSpot int) *DefendPress {
	return &DefendPress{
		windowStart: windowStart,
		windowEnd:   windowEnd,
		sweetSpot:   sweetSpot,
	}
}

func (dp *DefendPress) Start() {
	dp.currentTick = 0
	dp.pressed = false
	dp.complete = false
	dp.result = CommandResult{Quality: QualityMiss, BonusMult: 1.0}
}

func (dp *DefendPress) Update(inp *input.Manager) {
	if dp.complete {
		return
	}

	dp.currentTick++

	// Listen for B (Cancel) button
	if inp.Handler().ActionIsJustPressed(input.ActionCancel) && !dp.pressed {
		dp.pressed = true

		if dp.currentTick >= dp.windowStart && dp.currentTick <= dp.windowEnd {
			distance := dp.currentTick - dp.sweetSpot
			if distance < 0 {
				distance = -distance
			}
			dp.result = defendResultFromQuality(qualityFromDistance(distance, dp.windowEnd-dp.windowStart))
		} else {
			dp.result = defendResultFromQuality(QualityMiss)
		}

		dp.complete = true
	}

	// Auto-miss if window passes without press
	if dp.currentTick > dp.windowEnd && !dp.pressed {
		dp.pressed = true
		dp.result = defendResultFromQuality(QualityMiss)
		dp.complete = true
	}
}

func (dp *DefendPress) Draw(screen *ebiten.Image) {
	if dp.complete {
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

	if dp.windowEnd > 0 {
		totalDuration := float32(dp.windowEnd + 12)

		// Window zone (blue for defense)
		winStartPct := float32(dp.windowStart) / totalDuration
		winEndPct := float32(dp.windowEnd) / totalDuration
		winX := barX + winStartPct*barW
		winW := (winEndPct - winStartPct) * barW
		vector.DrawFilledRect(screen, winX, barY, winW, barH, color.RGBA{R: 50, G: 100, B: 180, A: 220}, false)

		// Sweet spot marker
		sweetPct := float32(dp.sweetSpot) / totalDuration
		sweetX := barX + sweetPct*barW
		vector.DrawFilledRect(screen, sweetX-2, barY-4, 4, barH+8, color.RGBA{R: 100, G: 200, B: 255, A: 255}, false)

		// Current position cursor (white)
		curPct := float32(dp.currentTick) / totalDuration
		curX := barX + curPct*barW
		vector.DrawFilledRect(screen, curX-3, barY-6, 6, barH+12, color.RGBA{R: 255, G: 255, B: 255, A: 255}, false)
	}
}

func (dp *DefendPress) IsComplete() bool {
	return dp.complete
}

func (dp *DefendPress) Result() CommandResult {
	return dp.result
}

// DefenseReduction returns the damage multiplier (lower = better defense).
// Miss = 1.0 (full damage), Excellent = 0.25 (quarter damage).
func (dp *DefendPress) DefenseReduction() float64 {
	return dp.result.BonusMult
}

// defendResultFromQuality maps quality to a defense reduction multiplier.
// Lower is better for the defender.
func defendResultFromQuality(q Quality) CommandResult {
	mults := map[Quality]float64{
		QualityMiss:      1.0,
		QualityNice:      0.75,
		QualityGood:      0.5,
		QualityGreat:     0.35,
		QualityExcellent: 0.25,
	}
	return CommandResult{
		Success:   q > QualityMiss,
		Quality:   q,
		BonusMult: mults[q],
	}
}
