package state

import (
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/tetra3d"

	"github.com/davidbennett/go-paper-rpg/internal/data"
	"github.com/davidbennett/go-paper-rpg/internal/input"
	"github.com/davidbennett/go-paper-rpg/internal/overworld"
	"github.com/davidbennett/go-paper-rpg/internal/render"
)

// OverworldState handles free-roaming exploration in the 3D isometric world.
type OverworldState struct {
	shared     *SharedContext
	renderer   *render.Renderer
	roomID     string
	player     *overworld.Player
	enemies    []*overworld.Enemy
	scene      *tetra3d.Scene
	locations  []data.LocationDef
	activeLoc  int
	collider   *overworld.WallCollider
	gameData   *data.GameData
	openEditor func(roomID string) error
}

func NewOverworldState(shared *SharedContext, renderer *render.Renderer, roomID string, mapDef *data.MapDef, player *overworld.Player, enemies []*overworld.Enemy, scene *tetra3d.Scene, gameData *data.GameData, openEditor func(roomID string) error) *OverworldState {
	s := &OverworldState{
		shared:     shared,
		renderer:   renderer,
		roomID:     roomID,
		player:     player,
		enemies:    enemies,
		scene:      scene,
		activeLoc:  -1,
		gameData:   gameData,
		openEditor: openEditor,
	}
	if mapDef != nil {
		s.locations = append([]data.LocationDef(nil), mapDef.Locations...)
		s.collider = overworld.NewWallCollider(mapDef.Walls)
	}
	return s
}

func (s *OverworldState) Enter(prev GameState) {
	s.renderer.SetScene(s.scene)
	s.updateActiveLocation()
}

func (s *OverworldState) Exit() {}

func (s *OverworldState) Update() error {
	if s.shared.Input.Handler().ActionIsJustPressed(input.ActionMenu) {
		s.shared.States.Push(NewMenuState(s.shared, s.roomID, s.openEditor))
		return nil
	}

	s.player.Update(s.shared.Input, s.blocksMovement)
	s.updateActiveLocation()

	// Update enemies and check collisions
	for _, enemy := range s.enemies {
		enemy.Update(s.player.X, s.player.Z, s.blocksMovement)

		if enemy.CollidesWithPlayer(s.player.X, s.player.Z) {
			// Roll 1-3 copies of the enemy's battle group
			count := 1 + rand.IntN(3)
			group := make([]data.EnemyDef, 0, count)
			prefabs := make([]string, 0, count)
			for i := 0; i < count; i++ {
				for _, prefabID := range enemy.Group {
					if def, ok := s.gameData.Enemies[prefabID]; ok {
						group = append(group, def)
						prefabs = append(prefabs, prefabID)
					}
				}
			}
			if len(group) == 0 {
				continue
			}

			enemy.Defeated = true
			enemy.Hide()

			s.shared.States.Push(NewBattleState(s.shared, s.renderer, group, prefabs))
			return nil
		}
	}

	// Camera follows player
	x, y, z := s.player.CameraTarget()
	s.renderer.SetCameraFollow(x, y, z)

	return nil
}

func (s *OverworldState) Draw(screen *ebiten.Image) {
	s.renderer.DrawScene(screen)

	if active := s.activeLocationDef(); active != nil {
		label := active.ID
		if active.Note != "" {
			label = active.Note
		}
		render.DrawDebugText(screen, "Pos: %.1f, %.1f\nLocation: %s", s.player.X, s.player.Z, label)
		return
	}

	render.DrawDebugText(screen, "Pos: %.1f, %.1f", s.player.X, s.player.Z)
}

func (s *OverworldState) blocksMovement(x, z, radius float64) bool {
	return s.collider != nil && s.collider.BlocksCircle(x, z, radius)
}

func (s *OverworldState) updateActiveLocation() {
	best := -1
	bestRadius := 0.0

	for i := range s.locations {
		loc := s.locations[i]
		dx := s.player.X - loc.Position[0]
		dz := s.player.Z - loc.Position[2]
		if dx*dx+dz*dz > loc.Radius*loc.Radius {
			continue
		}
		if best == -1 || loc.Radius < bestRadius {
			best = i
			bestRadius = loc.Radius
		}
	}

	s.activeLoc = best
}

func (s *OverworldState) activeLocationDef() *data.LocationDef {
	if s.activeLoc < 0 || s.activeLoc >= len(s.locations) {
		return nil
	}
	return &s.locations[s.activeLoc]
}
