package state

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/davidbennett/go-paper-rpg/internal/input"
	"github.com/davidbennett/go-paper-rpg/internal/render"
)

type menuItem struct {
	label string
	kind  string
}

// MenuState is a controller-first pause menu layered over the overworld.
type MenuState struct {
	shared     *SharedContext
	roomID     string
	openEditor func(roomID string) error
	items      []menuItem
	selected   int
}

func NewMenuState(shared *SharedContext, roomID string, openEditor func(roomID string) error) *MenuState {
	items := []menuItem{
		{label: "Resume", kind: "resume"},
		{label: "Map Editor", kind: "editor"},
		{label: "Controls", kind: "controls"},
	}
	return &MenuState{
		shared:     shared,
		roomID:     roomID,
		openEditor: openEditor,
		items:      items,
	}
}

func (s *MenuState) Enter(prev GameState) {}

func (s *MenuState) Exit() {}

func (s *MenuState) Update() error {
	handler := s.shared.Input.Handler()

	if handler.ActionIsJustPressed(input.ActionMenu) || handler.ActionIsJustPressed(input.ActionCancel) {
		s.shared.States.Pop()
		return nil
	}

	if handler.ActionIsJustPressed(input.ActionMoveUp) {
		s.selected--
		if s.selected < 0 {
			s.selected = len(s.items) - 1
		}
	}
	if handler.ActionIsJustPressed(input.ActionMoveDown) {
		s.selected++
		if s.selected >= len(s.items) {
			s.selected = 0
		}
	}

	if handler.ActionIsJustPressed(input.ActionConfirm) {
		switch s.items[s.selected].kind {
		case "resume":
			s.shared.States.Pop()
		case "editor":
			s.shared.States.Pop()
			if s.openEditor != nil {
				return s.openEditor(s.roomID)
			}
		case "controls":
			// Read-only pane for now.
		}
	}

	return nil
}

func (s *MenuState) Draw(screen *ebiten.Image) {
	bounds := screen.Bounds()
	ebitenutil.DrawRect(screen, 0, 0, float64(bounds.Dx()), float64(bounds.Dy()), color.RGBA{A: 96})

	menuBoxX := float32(18)
	menuBoxY := float32(18)
	menuBoxW := float32(128)
	menuBoxH := float32(154)
	paneBoxX := float32(156)
	paneBoxY := float32(18)
	paneBoxW := float32(306)
	paneBoxH := float32(154)

	render.DrawBox(screen, menuBoxX, menuBoxY, menuBoxW, menuBoxH, color.RGBA{R: 248, G: 244, B: 232, A: 236}, color.RGBA{R: 93, G: 79, B: 60, A: 255})
	render.DrawBox(screen, paneBoxX, paneBoxY, paneBoxW, paneBoxH, color.RGBA{R: 248, G: 244, B: 232, A: 236}, color.RGBA{R: 93, G: 79, B: 60, A: 255})

	ebitenutil.DebugPrintAt(screen, "Menu", int(menuBoxX)+10, int(menuBoxY)+8)
	for i, item := range s.items {
		y := int(menuBoxY) + 32 + i*26
		if i == s.selected {
			render.DrawBox(screen, menuBoxX+8, float32(y-4), menuBoxW-16, 22, color.RGBA{R: 182, G: 210, B: 176, A: 255}, color.RGBA{R: 84, G: 114, B: 80, A: 255})
		}
		ebitenutil.DebugPrintAt(screen, item.label, int(menuBoxX)+16, y)
	}

	ebitenutil.DebugPrintAt(screen, s.paneText(), int(paneBoxX)+10, int(paneBoxY)+8)
}

func (s *MenuState) paneText() string {
	switch s.items[s.selected].kind {
	case "resume":
		return "Resume\n\nClose the menu and return to play."
	case "editor":
		return "Map Editor\n\nOpen the current room in the in-game editor.\nSaving there reloads the room when you back out."
	case "controls":
		return "Controls\n\nMove: Left Stick / D-Pad / WASD\nConfirm: A / Enter\nBack: B / Backspace\nMenu: Start / Escape\nMap Editor: Menu -> Map Editor"
	default:
		return ""
	}
}
