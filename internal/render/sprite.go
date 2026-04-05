package render

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// Animation defines a sequence of frames within a sprite sheet.
type Animation struct {
	Frames   []image.Rectangle // Sub-image rects within the sheet
	Duration int               // Ticks per frame
	Loop     bool
}

// AnimatedSprite manages sprite sheet animation playback.
type AnimatedSprite struct {
	sheet      *ebiten.Image
	animations map[string]*Animation
	current    string
	frameIndex int
	tickCount  int
}

func NewAnimatedSprite(sheet *ebiten.Image) *AnimatedSprite {
	return &AnimatedSprite{
		sheet:      sheet,
		animations: make(map[string]*Animation),
	}
}

func (s *AnimatedSprite) AddAnimation(name string, anim *Animation) {
	s.animations[name] = anim
}

func (s *AnimatedSprite) Play(name string) {
	if s.current == name {
		return
	}
	s.current = name
	s.frameIndex = 0
	s.tickCount = 0
}

func (s *AnimatedSprite) Tick() {
	anim, ok := s.animations[s.current]
	if !ok {
		return
	}

	s.tickCount++
	if s.tickCount >= anim.Duration {
		s.tickCount = 0
		s.frameIndex++
		if s.frameIndex >= len(anim.Frames) {
			if anim.Loop {
				s.frameIndex = 0
			} else {
				s.frameIndex = len(anim.Frames) - 1
			}
		}
	}
}

func (s *AnimatedSprite) CurrentFrame() *ebiten.Image {
	anim, ok := s.animations[s.current]
	if !ok || len(anim.Frames) == 0 {
		return s.sheet
	}
	rect := anim.Frames[s.frameIndex]
	return s.sheet.SubImage(rect).(*ebiten.Image)
}

func (s *AnimatedSprite) Sheet() *ebiten.Image {
	return s.sheet
}

// GenerateGridFrames creates frame rectangles for a uniform grid sprite sheet.
func GenerateGridFrames(frameW, frameH, cols, startRow, startCol, count int) []image.Rectangle {
	frames := make([]image.Rectangle, 0, count)
	col := startCol
	row := startRow
	for i := 0; i < count; i++ {
		x := col * frameW
		y := row * frameH
		frames = append(frames, image.Rect(x, y, x+frameW, y+frameH))
		col++
		if col >= cols {
			col = 0
			row++
		}
	}
	return frames
}
