package action

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/davidbennett/go-paper-rpg/internal/input"
)

// Mash is an action command where the player mashes Confirm
// as many times as possible within a time window.
type Mash struct {
	duration    int    // Total ticks allowed
	currentTick int
	pressCount  int
	thresholds  [4]int // Press counts for Nice/Good/Great/Excellent
	complete    bool
	result      CommandResult
}

func NewMash(duration int, nice, good, great, excellent int) *Mash {
	return &Mash{
		duration:   duration,
		thresholds: [4]int{nice, good, great, excellent},
	}
}

func (m *Mash) Start() {
	m.currentTick = 0
	m.pressCount = 0
	m.complete = false
	m.result = CommandResult{Quality: QualityMiss, BonusMult: 1.0}
}

func (m *Mash) Update(inp *input.Manager) {
	if m.complete {
		return
	}

	m.currentTick++

	if inp.Handler().ActionIsJustPressed(input.ActionConfirm) {
		m.pressCount++
	}

	if m.currentTick >= m.duration {
		m.complete = true

		var q Quality
		switch {
		case m.pressCount >= m.thresholds[3]:
			q = QualityExcellent
		case m.pressCount >= m.thresholds[2]:
			q = QualityGreat
		case m.pressCount >= m.thresholds[1]:
			q = QualityGood
		case m.pressCount >= m.thresholds[0]:
			q = QualityNice
		default:
			q = QualityMiss
		}
		m.result = ResultFromQuality(q)
	}
}

func (m *Mash) Draw(screen *ebiten.Image) {
	if m.complete {
		return
	}

	// Draw a gauge that fills based on press count
	barX := float32(160)
	barY := float32(200)
	barW := float32(160)
	barH := float32(16)

	// Background
	vector.DrawFilledRect(screen, barX, barY, barW, barH, color.RGBA{R: 40, G: 40, B: 40, A: 200}, true)

	// Fill based on progress toward Excellent threshold
	fillPct := float32(m.pressCount) / float32(m.thresholds[3])
	if fillPct > 1 {
		fillPct = 1
	}
	fillColor := color.RGBA{R: 50, G: 180, B: 50, A: 220}
	if fillPct > 0.75 {
		fillColor = color.RGBA{R: 255, G: 200, B: 0, A: 220}
	}
	vector.DrawFilledRect(screen, barX, barY, barW*fillPct, barH, fillColor, true)

	// Border
	vector.StrokeRect(screen, barX, barY, barW, barH, 2, color.RGBA{R: 200, G: 200, B: 200, A: 255}, true)

	// Time remaining indicator
	timePct := float32(m.currentTick) / float32(m.duration)
	timeX := barX + timePct*barW
	vector.DrawFilledRect(screen, timeX-1, barY-4, 2, barH+8, color.RGBA{R: 255, G: 80, B: 80, A: 255}, true)
}

func (m *Mash) IsComplete() bool {
	return m.complete
}

func (m *Mash) Result() CommandResult {
	return m.result
}
