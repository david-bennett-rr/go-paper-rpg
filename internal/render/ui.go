package render

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// DrawBox draws a bordered rectangle (Paper Mario-style dialogue/menu box).
func DrawBox(screen *ebiten.Image, x, y, w, h float32, bgColor, borderColor color.Color) {
	vector.DrawFilledRect(screen, x, y, w, h, bgColor, true)
	vector.StrokeRect(screen, x, y, w, h, 2, borderColor, true)
}

// DrawDebugText draws simple debug text at the top-left of the screen.
func DrawDebugText(screen *ebiten.Image, format string, args ...any) {
	ebitenutil.DebugPrint(screen, fmt.Sprintf(format, args...))
}
