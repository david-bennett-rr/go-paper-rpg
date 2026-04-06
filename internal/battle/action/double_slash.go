package action

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/davidbennett/go-paper-rpg/internal/input"
)

type SlashBeat struct {
	WindowStart int
	WindowEnd   int
	SweetSpot   int
}

// DoubleSlash asks the player to press Confirm twice, once for each swing.
type DoubleSlash struct {
	beats        []SlashBeat
	currentTick  int
	nextBeat     int
	complete     bool
	result       CommandResult
	slashResults []CommandResult
}

func NewDoubleSlash(beats ...SlashBeat) *DoubleSlash {
	if len(beats) == 0 {
		beats = []SlashBeat{
			{WindowStart: 25, WindowEnd: 45, SweetSpot: 35},
			{WindowStart: 60, WindowEnd: 80, SweetSpot: 70},
		}
	}

	return &DoubleSlash{
		beats:        append([]SlashBeat(nil), beats...),
		slashResults: make([]CommandResult, len(beats)),
	}
}

func (ds *DoubleSlash) Start() {
	ds.currentTick = 0
	ds.nextBeat = 0
	ds.complete = false
	ds.result = CommandResult{Quality: QualityMiss, BonusMult: 0}
	for i := range ds.slashResults {
		ds.slashResults[i] = CommandResult{Quality: QualityMiss, BonusMult: 0}
	}
}

func (ds *DoubleSlash) Update(inp *input.Manager) {
	if ds.complete {
		return
	}

	ds.currentTick++

	for ds.nextBeat < len(ds.beats) && ds.currentTick > ds.beats[ds.nextBeat].WindowEnd {
		ds.slashResults[ds.nextBeat] = CommandResult{Quality: QualityMiss, BonusMult: 0}
		ds.nextBeat++
	}

	if inp.Handler().ActionIsJustPressed(input.ActionConfirm) && ds.nextBeat < len(ds.beats) {
		beat := ds.beats[ds.nextBeat]
		if ds.currentTick >= beat.WindowStart && ds.currentTick <= beat.WindowEnd {
			distance := ds.currentTick - beat.SweetSpot
			if distance < 0 {
				distance = -distance
			}
			ds.slashResults[ds.nextBeat] = ResultFromQuality(qualityFromDistance(distance, beat.WindowEnd-beat.WindowStart))
		} else {
			ds.slashResults[ds.nextBeat] = CommandResult{Quality: QualityMiss, BonusMult: 0}
		}
		ds.nextBeat++
	}

	if ds.nextBeat >= len(ds.beats) {
		ds.finish()
	}
}

func (ds *DoubleSlash) Draw(screen *ebiten.Image) {
	if ds.complete || len(ds.beats) == 0 {
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
	totalDuration := float32(ds.beats[len(ds.beats)-1].WindowEnd + 12)

	vector.DrawFilledRect(screen, barX-4, barY-4, barW+8, barH+8, color.RGBA{R: 20, G: 22, B: 30, A: 220}, false)
	vector.DrawFilledRect(screen, barX, barY, barW, barH, color.RGBA{R: 40, G: 40, B: 50, A: 230}, false)
	vector.StrokeRect(screen, barX, barY, barW, barH, 2, color.RGBA{R: 70, G: 72, B: 90, A: 255}, false)

	for i, beat := range ds.beats {
		startPct := float32(beat.WindowStart) / totalDuration
		endPct := float32(beat.WindowEnd) / totalDuration
		sweetPct := float32(beat.SweetSpot) / totalDuration
		windowX := barX + startPct*barW
		windowW := (endPct - startPct) * barW
		windowColor := color.RGBA{R: 74, G: 138, B: 82, A: 220}
		if i < ds.nextBeat {
			if ds.slashResults[i].Quality == QualityMiss {
				windowColor = color.RGBA{R: 140, G: 62, B: 62, A: 220}
			} else {
				windowColor = color.RGBA{R: 98, G: 172, B: 108, A: 220}
			}
		}
		vector.DrawFilledRect(screen, windowX, barY, windowW, barH, windowColor, false)
		sweetX := barX + sweetPct*barW
		vector.DrawFilledRect(screen, sweetX-2, barY-4, 4, barH+8, color.RGBA{R: 255, G: 223, B: 92, A: 255}, false)

		// Dim completed windows
		if i < ds.nextBeat {
			vector.DrawFilledRect(screen, windowX, barY, windowW, barH, color.RGBA{R: 0, G: 0, B: 0, A: 60}, false)
		}
	}

	cursorPct := float32(ds.currentTick) / totalDuration
	if cursorPct > 1 {
		cursorPct = 1
	}
	cursorX := barX + cursorPct*barW
	vector.DrawFilledRect(screen, cursorX-3, barY-6, 6, barH+12, color.RGBA{R: 245, G: 245, B: 255, A: 255}, false)
}

func (ds *DoubleSlash) IsComplete() bool {
	return ds.complete
}

func (ds *DoubleSlash) Result() CommandResult {
	return ds.result
}

func (ds *DoubleSlash) CurrentTick() int {
	return ds.currentTick
}

func (ds *DoubleSlash) Beats() []SlashBeat {
	return append([]SlashBeat(nil), ds.beats...)
}

func (ds *DoubleSlash) SlashResults() []CommandResult {
	return append([]CommandResult(nil), ds.slashResults...)
}

func (ds *DoubleSlash) finish() {
	score := 0
	for _, slash := range ds.slashResults {
		score += qualityScore(slash.Quality)
	}

	var quality Quality
	switch {
	case score >= 7:
		quality = QualityExcellent
	case score >= 5:
		quality = QualityGreat
	case score >= 3:
		quality = QualityGood
	case score >= 1:
		quality = QualityNice
	default:
		quality = QualityMiss
	}

	ds.result = ResultFromQuality(quality)
	if quality == QualityMiss {
		ds.result.BonusMult = 0
		ds.result.Success = false
	}
	ds.complete = true
}

func qualityScore(q Quality) int {
	switch q {
	case QualityNice:
		return 1
	case QualityGood:
		return 2
	case QualityGreat:
		return 3
	case QualityExcellent:
		return 4
	default:
		return 0
	}
}
