package overworld

import (
	"testing"

	"github.com/davidbennett/go-paper-rpg/internal/data"
)

func TestWallColliderBlocksOccupiedTile(t *testing.T) {
	collider := NewWallCollider([]data.WallDef{
		{
			ID:       "wall_1",
			Position: [3]float64{0, 0.75, 0},
			Size:     [3]float64{1, 1.5, 1},
		},
	})

	if collider == nil {
		t.Fatal("expected collider to be created")
	}
	if !collider.BlocksCircle(0, 0, PlayerRadius) {
		t.Fatal("expected occupied wall tile to block movement")
	}
	if collider.BlocksCircle(2, 2, PlayerRadius) {
		t.Fatal("expected open tile to allow movement")
	}
}

func TestMoveWithCollisionSlidesAlongWall(t *testing.T) {
	collider := NewWallCollider([]data.WallDef{
		{
			ID:       "wall_1",
			Position: [3]float64{1, 0.75, 0},
			Size:     [3]float64{1, 1.5, 1},
		},
	})

	x, z := moveWithCollision(0, 0, 0.9, 0.4, PlayerRadius, collider.BlocksCircle)
	if x != 0 {
		t.Fatalf("expected X movement to be blocked by wall, got %v", x)
	}
	if z == 0 {
		t.Fatalf("expected Z movement to slide through open space, got %v", z)
	}
}
