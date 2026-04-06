package overworld

import (
	"math"

	"github.com/davidbennett/go-paper-rpg/internal/data"
)

// WallCollider provides cheap 2D collision against wall tiles authored in the map editor.
type WallCollider struct {
	occupied map[[2]int]struct{}
}

func NewWallCollider(walls []data.WallDef) *WallCollider {
	collider := &WallCollider{
		occupied: map[[2]int]struct{}{},
	}

	for _, wall := range walls {
		countX := int(math.Max(1, math.Round(wall.Size[0])))
		countZ := int(math.Max(1, math.Round(wall.Size[2])))
		startX := wall.Position[0] - float64(countX-1)/2
		startZ := wall.Position[2] - float64(countZ-1)/2

		for x := 0; x < countX; x++ {
			for z := 0; z < countZ; z++ {
				cellX := int(math.Round(startX + float64(x)))
				cellZ := int(math.Round(startZ + float64(z)))
				collider.occupied[[2]int{cellX, cellZ}] = struct{}{}
			}
		}
	}

	if len(collider.occupied) == 0 {
		return nil
	}
	return collider
}

func (c *WallCollider) BlocksCircle(x, z, radius float64) bool {
	if c == nil || len(c.occupied) == 0 {
		return false
	}

	minCellX := int(math.Floor(x - radius - 0.5))
	maxCellX := int(math.Ceil(x + radius + 0.5))
	minCellZ := int(math.Floor(z - radius - 0.5))
	maxCellZ := int(math.Ceil(z + radius + 0.5))

	for cellX := minCellX; cellX <= maxCellX; cellX++ {
		for cellZ := minCellZ; cellZ <= maxCellZ; cellZ++ {
			if _, ok := c.occupied[[2]int{cellX, cellZ}]; !ok {
				continue
			}
			if circleIntersectsTile(x, z, radius, float64(cellX), float64(cellZ)) {
				return true
			}
		}
	}

	return false
}

func moveWithCollision(x, z, moveX, moveZ, radius float64, blocked func(x, z, radius float64) bool) (float64, float64) {
	if blocked == nil {
		return x + moveX, z + moveZ
	}

	nextX := x
	if !blocked(x+moveX, z, radius) {
		nextX = x + moveX
	}

	nextZ := z
	if !blocked(nextX, z+moveZ, radius) {
		nextZ = z + moveZ
	}

	return nextX, nextZ
}

func circleIntersectsTile(x, z, radius, tileX, tileZ float64) bool {
	nearestX := clamp(x, tileX-0.5, tileX+0.5)
	nearestZ := clamp(z, tileZ-0.5, tileZ+0.5)
	dx := x - nearestX
	dz := z - nearestZ
	return dx*dx+dz*dz < radius*radius
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
