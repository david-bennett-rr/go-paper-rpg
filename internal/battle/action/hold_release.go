package action

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/davidbennett/go-paper-rpg/internal/input"
)

// HoldRelease is an action command where the player holds Confirm
// and releases at the right moment when a visual cue appears.
type HoldRelease struct {
	releaseWindowStart int
	releaseWindowEnd   int
	sweetSpot          int
	maxTick            int // Auto-fail if held too long
	currentTick        int
	holding            bool
	released           bool
	complete           bool
	result             CommandResult
}

func NewHoldRelease(windowStart, windowEnd, sweetSpot, maxTick int) *HoldRelease {
	return &HoldRelease{
		releaseWindowStart: windowStart,
		releaseWindowEnd:   windowEnd,
		sweetSpot:          sweetSpot,
		maxTick:            maxTick,
	}
}

func (hr *HoldRelease) Start() {
	hr.currentTick = 0
	hr.holding = false
	hr.released = false
	hr.complete = false
	hr.result = CommandResult{Quality: QualityMiss, BonusMult: 1.0}
}

func (hr *HoldRelease) Update(inp *input.Manager) {
	if hr.complete {
		return
	}

	hr.currentTick++

	// Track holding state
	if inp.Handler().ActionIsPressed(input.ActionConfirm) {
		hr.holding = true
	}

	// Detect release
	if hr.holding && !inp.Handler().ActionIsPressed(input.ActionConfirm) {
		hr.released = true
		hr.complete = true

		if hr.currentTick >= hr.releaseWindowStart && hr.currentTick <= hr.releaseWindowEnd {
			distance := hr.currentTick - hr.sweetSpot
			if distance < 0 {
				distance = -distance
			}
			hr.result = ResultFromQuality(qualityFromDistance(distance, hr.releaseWindowEnd-hr.releaseWindowStart))
		} else {
			hr.result = ResultFromQuality(QualityMiss)
		}
		return
	}

	// Auto-fail if held too long
	if hr.currentTick >= hr.maxTick {
		hr.complete = true
		hr.result = ResultFromQuality(QualityMiss)
	}
}

func (hr *HoldRelease) Draw(screen *ebiten.Image) {
	if hr.complete {
		return
	}

	barX := float32(160)
	barY := float32(200)
	barW := float32(160)
	barH := float32(14)

	// Background
	vector.DrawFilledRect(screen, barX, barY, barW, barH, color.RGBA{R: 40, G: 40, B: 40, A: 200}, true)

	// Charging meter (fills as you hold)
	chargePct := float32(hr.currentTick) / float32(hr.maxTick)
	if chargePct > 1 {
		chargePct = 1
	}

	chargeColor := color.RGBA{R: 80, G: 80, B: 220, A: 220}
	if hr.currentTick >= hr.releaseWindowStart && hr.currentTick <= hr.releaseWindowEnd {
		chargeColor = color.RGBA{R: 255, G: 200, B: 0, A: 255} // Glow during window
	}
	vector.DrawFilledRect(screen, barX, barY, barW*chargePct, barH, chargeColor, true)

	// Release window zone
	totalDuration := float32(hr.maxTick)
	winStartPct := float32(hr.releaseWindowStart) / totalDuration
	winEndPct := float32(hr.releaseWindowEnd) / totalDuration
	winX := barX + winStartPct*barW
	winW := (winEndPct - winStartPct) * barW
	vector.StrokeRect(screen, winX, barY-2, winW, barH+4, 2, color.RGBA{R: 255, G: 255, B: 0, A: 200}, true)

	// Border
	vector.StrokeRect(screen, barX, barY, barW, barH, 2, color.RGBA{R: 200, G: 200, B: 200, A: 255}, true)
}

func (hr *HoldRelease) IsComplete() bool {
	return hr.complete
}

func (hr *HoldRelease) Result() CommandResult {
	return hr.result
}
