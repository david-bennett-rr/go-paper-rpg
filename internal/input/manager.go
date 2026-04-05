package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	einput "github.com/quasilyte/ebitengine-input"
)

type Manager struct {
	system  einput.System
	handler *einput.Handler
}

func NewManager() *Manager {
	m := &Manager{}
	m.system.Init(einput.SystemConfig{
		DevicesEnabled: einput.AnyDevice,
	})

	keymap := einput.Keymap{
		ActionMoveUp:    {einput.KeyGamepadLStickUp, einput.KeyGamepadUp, einput.KeyUp, einput.KeyW},
		ActionMoveDown:  {einput.KeyGamepadLStickDown, einput.KeyGamepadDown, einput.KeyDown, einput.KeyS},
		ActionMoveLeft:  {einput.KeyGamepadLStickLeft, einput.KeyGamepadLeft, einput.KeyLeft, einput.KeyA},
		ActionMoveRight: {einput.KeyGamepadLStickRight, einput.KeyGamepadRight, einput.KeyRight, einput.KeyD},
		ActionConfirm:   {einput.KeyGamepadA, einput.KeyZ, einput.KeyEnter},
		ActionCancel:    {einput.KeyGamepadB, einput.KeyX, einput.KeyBackspace},
		ActionMenu:      {einput.KeyGamepadStart, einput.KeyEscape},
		ActionPartner:   {einput.KeyGamepadY, einput.KeyC},
	}

	m.handler = m.system.NewHandler(0, keymap)
	return m
}

func (m *Manager) Update() {
	m.system.Update()
}

func (m *Manager) Handler() *einput.Handler {
	return m.handler
}

// MoveDir returns the normalized movement direction as (dx, dz) for 3D space.
// Returns (0, 0) if no movement input is active.
func (m *Manager) MoveDir() (float64, float64) {
	var dx, dz float64
	if m.handler.ActionIsPressed(ActionMoveUp) {
		dz -= 1
	}
	if m.handler.ActionIsPressed(ActionMoveDown) {
		dz += 1
	}
	if m.handler.ActionIsPressed(ActionMoveLeft) {
		dx -= 1
	}
	if m.handler.ActionIsPressed(ActionMoveRight) {
		dx += 1
	}
	// Normalize diagonal movement
	if dx != 0 && dz != 0 {
		dx *= 0.7071 // 1/sqrt(2)
		dz *= 0.7071
	}
	return dx, dz
}

// GamepadConnected returns true if at least one gamepad is connected.
func (m *Manager) GamepadConnected() bool {
	return len(ebiten.AppendGamepadIDs(nil)) > 0
}
