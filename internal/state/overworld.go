package state

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/tetra3d"

	"github.com/davidbennett/go-paper-rpg/internal/overworld"
	"github.com/davidbennett/go-paper-rpg/internal/render"
)

// OverworldState handles free-roaming exploration in the 3D isometric world.
type OverworldState struct {
	shared   *SharedContext
	renderer *render.Renderer
	player   *overworld.Player
	scene    *tetra3d.Scene
}

func NewOverworldState(shared *SharedContext, renderer *render.Renderer, player *overworld.Player, scene *tetra3d.Scene) *OverworldState {
	return &OverworldState{
		shared:   shared,
		renderer: renderer,
		player:   player,
		scene:    scene,
	}
}

func (s *OverworldState) Enter(prev GameState) {
	s.renderer.SetScene(s.scene)
}

func (s *OverworldState) Exit() {}

func (s *OverworldState) Update() error {
	s.player.Update(s.shared.Input)

	// Camera follows player
	x, y, z := s.player.CameraTarget()
	s.renderer.SetCameraFollow(x, y, z)

	return nil
}

func (s *OverworldState) Draw(screen *ebiten.Image) {
	s.renderer.DrawScene(screen)

	// Debug HUD
	render.DrawDebugText(screen, "Paper RPG - Arrow keys/WASD move relative to camera\nPos: %.1f, %.1f", s.player.X, s.player.Z)
}
