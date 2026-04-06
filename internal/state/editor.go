package state

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/davidbennett/go-paper-rpg/internal/editor"
)

// EditorState wraps the map editor so it can run inside the normal game state flow.
type EditorState struct {
	app     *editor.App
	onEnter func()
}

func NewEditorState(app *editor.App, onEnter func()) *EditorState {
	return &EditorState{
		app:     app,
		onEnter: onEnter,
	}
}

func (s *EditorState) Enter(prev GameState) {
	if s.onEnter != nil {
		s.onEnter()
	}
}

func (s *EditorState) Exit() {}

func (s *EditorState) Update() error {
	return s.app.Update()
}

func (s *EditorState) Draw(screen *ebiten.Image) {
	s.app.Draw(screen)
}
