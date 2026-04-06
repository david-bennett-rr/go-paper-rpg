package overworld

import (
	"math"

	"github.com/solarlune/tetra3d"

	"github.com/davidbennett/go-paper-rpg/internal/input"
	"github.com/davidbennett/go-paper-rpg/internal/render"
)

const (
	PlayerSpeed    = 0.08
	PlayerEyeLevel = 0.9
	PlayerRadius   = 0.26
)

// Player represents the player character in the overworld.
type Player struct {
	X, Y, Z float64
	Yaw     float64
	Root    tetra3d.INode
	Shadow  tetra3d.INode
}

func NewPlayer(root tetra3d.INode) *Player {
	p := &Player{
		Root: root,
	}
	p.syncModel()
	return p
}

func (p *Player) Update(inp *input.Manager, blocked func(x, z, radius float64) bool) {
	screenX, screenY := inp.MoveDir()
	dx, dz := render.ScreenToWorldMove(screenX, screenY)

	if dx != 0 || dz != 0 {
		p.X, p.Z = moveWithCollision(p.X, p.Z, dx*PlayerSpeed, dz*PlayerSpeed, PlayerRadius, blocked)
		p.faceMovement(dx, dz)
	}

	p.syncModel()
}

func (p *Player) CameraTarget() (x, y, z float64) {
	return p.X, p.Y + PlayerEyeLevel, p.Z
}

func (p *Player) SetPlacement(x, y, z, yaw float64) {
	p.X = x
	p.Y = y
	p.Z = z
	p.Yaw = yaw
	p.syncModel()
}

func (p *Player) syncModel() {
	if p.Root != nil {
		p.Root.SetLocalPositionVec(tetra3d.NewVector3(float32(p.X), float32(p.Y), float32(p.Z)))
		p.Root.SetLocalRotation(tetra3d.NewMatrix4Rotate(0, 1, 0, float32(p.Yaw)))
	}
	if p.Shadow != nil {
		p.Shadow.SetLocalPositionVec(tetra3d.NewVector3(float32(p.X), 0.02, float32(p.Z)))
	}
}

func (p *Player) faceMovement(dx, dz float64) {
	if p.Root == nil {
		return
	}
	p.Yaw = math.Atan2(dx, dz)
}
