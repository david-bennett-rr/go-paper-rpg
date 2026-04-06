package overworld

import (
	"math"

	"github.com/solarlune/tetra3d"
)

const (
	EnemySpeed      = 0.03
	EnemyChaseRange = 8.0
	CollisionDist   = 0.6
)

// Enemy is an overworld enemy that chases the player.
type Enemy struct {
	X, Y, Z  float64
	Yaw      float64
	Root     tetra3d.INode
	Shadow   tetra3d.INode
	Prefab   string
	MapID    string
	Group    []string // battle group prefab IDs
	Defeated bool
}

func NewEnemy(root tetra3d.INode, prefab, mapID string, group []string) *Enemy {
	return &Enemy{
		Root:   root,
		Prefab: prefab,
		MapID:  mapID,
		Group:  group,
	}
}

func (e *Enemy) SetPlacement(x, y, z, yaw float64) {
	e.X = x
	e.Y = y
	e.Z = z
	e.Yaw = yaw
	e.syncModel()
}

// Update moves the enemy toward the player if in range.
func (e *Enemy) Update(playerX, playerZ float64) {
	if e.Defeated {
		return
	}

	dx := playerX - e.X
	dz := playerZ - e.Z
	dist := math.Sqrt(dx*dx + dz*dz)

	if dist > EnemyChaseRange || dist < 0.01 {
		return
	}

	nx := dx / dist
	nz := dz / dist
	e.X += nx * EnemySpeed
	e.Z += nz * EnemySpeed
	e.Yaw = math.Atan2(nx, nz)
	e.syncModel()
}

// CollidesWithPlayer returns true if enemy is close enough to trigger a battle.
func (e *Enemy) CollidesWithPlayer(playerX, playerZ float64) bool {
	if e.Defeated {
		return false
	}
	dx := playerX - e.X
	dz := playerZ - e.Z
	return dx*dx+dz*dz < CollisionDist*CollisionDist
}

func (e *Enemy) syncModel() {
	if e.Root != nil {
		e.Root.SetLocalPositionVec(tetra3d.NewVector3(float32(e.X), float32(e.Y), float32(e.Z)))
		e.Root.SetLocalRotation(tetra3d.NewMatrix4Rotate(0, 1, 0, float32(e.Yaw)))
	}
	if e.Shadow != nil {
		e.Shadow.SetLocalPositionVec(tetra3d.NewVector3(float32(e.X), 0.02, float32(e.Z)))
	}
}

func (e *Enemy) Hide() {
	if e.Root != nil {
		e.Root.SetLocalPosition(0, -100, 0)
	}
	if e.Shadow != nil {
		e.Shadow.SetLocalPosition(0, -100, 0)
	}
}
