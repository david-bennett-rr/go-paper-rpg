package input

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	einput "github.com/quasilyte/ebitengine-input"
)

const moveStickDeadzone = 0.2

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
	if dx, dz, ok := m.analogMoveDir(); ok {
		return dx, dz
	}

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

func (m *Manager) analogMoveDir() (float64, float64, bool) {
	gamepadIDs := ebiten.AppendGamepadIDs(nil)
	for _, id := range gamepadIDs {
		dx, dz, ok := standardGamepadMoveDir(id)
		if !ok {
			dx, dz, ok = rawGamepadMoveDir(id)
		}
		if !ok {
			continue
		}
		return applyDeadzone(dx, dz, moveStickDeadzone)
	}
	return 0, 0, false
}

func standardGamepadMoveDir(id ebiten.GamepadID) (float64, float64, bool) {
	if !ebiten.IsStandardGamepadLayoutAvailable(id) {
		return 0, 0, false
	}
	if !ebiten.IsStandardGamepadAxisAvailable(id, ebiten.StandardGamepadAxisLeftStickHorizontal) ||
		!ebiten.IsStandardGamepadAxisAvailable(id, ebiten.StandardGamepadAxisLeftStickVertical) {
		return 0, 0, false
	}

	return ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickHorizontal),
		ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickVertical),
		true
}

func rawGamepadMoveDir(id ebiten.GamepadID) (float64, float64, bool) {
	if ebiten.GamepadAxisCount(id) < 2 {
		return 0, 0, false
	}
	return ebiten.GamepadAxisValue(id, 0), ebiten.GamepadAxisValue(id, 1), true
}

func applyDeadzone(dx, dz, deadzone float64) (float64, float64, bool) {
	mag := math.Hypot(dx, dz)
	if mag <= deadzone {
		return 0, 0, false
	}

	if mag > 1 {
		dx /= mag
		dz /= mag
		mag = 1
	}

	adjusted := (mag - deadzone) / (1 - deadzone)
	if adjusted < 0 {
		adjusted = 0
	}
	scale := adjusted / mag
	return dx * scale, dz * scale, true
}

// GamepadConnected returns true if at least one gamepad is connected.
func (m *Manager) GamepadConnected() bool {
	return len(ebiten.AppendGamepadIDs(nil)) > 0
}
