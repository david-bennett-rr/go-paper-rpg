package state

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/davidbennett/go-paper-rpg/internal/input"
)

// GameState represents a discrete game state (overworld, battle, dialogue, etc.).
type GameState interface {
	Enter(prev GameState)
	Exit()
	Update() error
	Draw(screen *ebiten.Image)
}

// SharedContext holds cross-cutting dependencies available to all states.
type SharedContext struct {
	Input    *input.Manager
	States   *Manager
}

// Manager is a stack-based state manager.
// The top of the stack is the active state that receives Update/Draw calls.
type Manager struct {
	stack  []GameState
	shared *SharedContext
}

func NewManager(inputMgr *input.Manager) *Manager {
	m := &Manager{}
	m.shared = &SharedContext{
		Input:  inputMgr,
		States: m,
	}
	return m
}

func (m *Manager) Shared() *SharedContext {
	return m.shared
}

// Push adds a state on top of the stack (e.g., opening a dialogue over the overworld).
func (m *Manager) Push(s GameState) {
	var prev GameState
	if len(m.stack) > 0 {
		prev = m.stack[len(m.stack)-1]
	}
	m.stack = append(m.stack, s)
	s.Enter(prev)
}

// Pop removes the top state and returns to the one below.
func (m *Manager) Pop() {
	if len(m.stack) == 0 {
		return
	}
	top := m.stack[len(m.stack)-1]
	top.Exit()
	m.stack = m.stack[:len(m.stack)-1]
}

// Switch replaces the top state (e.g., overworld -> battle).
func (m *Manager) Switch(s GameState) {
	var prev GameState
	if len(m.stack) > 0 {
		prev = m.stack[len(m.stack)-1]
		prev.Exit()
		m.stack[len(m.stack)-1] = s
	} else {
		m.stack = append(m.stack, s)
	}
	s.Enter(prev)
}

// Current returns the active (top) state, or nil if empty.
func (m *Manager) Current() GameState {
	if len(m.stack) == 0 {
		return nil
	}
	return m.stack[len(m.stack)-1]
}

func (m *Manager) Update() error {
	if s := m.Current(); s != nil {
		return s.Update()
	}
	return nil
}

func (m *Manager) Draw(screen *ebiten.Image) {
	// Draw all states from bottom to top so overlays render on top.
	for _, s := range m.stack {
		s.Draw(screen)
	}
}
