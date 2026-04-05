package world

import (
	"fmt"
	"image/color"
	"math"

	"github.com/solarlune/tetra3d"
)

type PrefabCategory string

const (
	PrefabCategoryProp  PrefabCategory = "prop"
	PrefabCategoryEnemy PrefabCategory = "enemy"
)

type PrefabInfo struct {
	ID       string
	Name     string
	Category PrefabCategory
	Icon     string
	Color    color.RGBA
	Radius   float64
}

var prefabInfos = []PrefabInfo{
	{ID: "tree_pine", Name: "Pine Tree", Category: PrefabCategoryProp, Icon: "P", Color: color.RGBA{R: 46, G: 112, B: 45, A: 255}, Radius: 0.7},
	{ID: "tree_round", Name: "Round Tree", Category: PrefabCategoryProp, Icon: "T", Color: color.RGBA{R: 72, G: 144, B: 62, A: 255}, Radius: 0.75},
	{ID: "rock", Name: "Rock", Category: PrefabCategoryProp, Icon: "R", Color: color.RGBA{R: 115, G: 117, B: 126, A: 255}, Radius: 0.45},
	{ID: "wolf", Name: "Wolf", Category: PrefabCategoryEnemy, Icon: "W", Color: color.RGBA{R: 144, G: 149, B: 160, A: 255}, Radius: 0.55},
}

type scenePalette struct {
	grass      *tetra3d.Material
	dirt       *tetra3d.Material
	bark       *tetra3d.Material
	leaves     *tetra3d.Material
	leavesAlt  *tetra3d.Material
	stone      *tetra3d.Material
	wall       *tetra3d.Material
	playerCoat *tetra3d.Material
	playerSkin *tetra3d.Material
	playerHat  *tetra3d.Material
	playerBoot *tetra3d.Material
	wolfFur    *tetra3d.Material
	wolfDark   *tetra3d.Material
	eyes       *tetra3d.Material
}

func PropPrefabInfos() []PrefabInfo {
	return prefabInfosFor(PrefabCategoryProp)
}

func EnemyPrefabInfos() []PrefabInfo {
	return prefabInfosFor(PrefabCategoryEnemy)
}

func PrefabInfoByID(id string) (PrefabInfo, bool) {
	for _, info := range prefabInfos {
		if info.ID == id {
			return info, true
		}
	}
	return PrefabInfo{}, false
}

func BuildPropPrefab(id string) (tetra3d.INode, error) {
	return buildPrefab(id, newScenePalette())
}

func BuildEnemyPrefab(id string) (tetra3d.INode, error) {
	return buildPrefab(id, newScenePalette())
}

func BuildPlayerPrefab() tetra3d.INode {
	return anchorNodeToTile(buildPlayerPrefab(newScenePalette()))
}

func BuildWall(size tetra3d.Vector3) tetra3d.INode {
	palette := newScenePalette()
	return newWallModel("Wall", palette.wall, size)
}

func prefabInfosFor(category PrefabCategory) []PrefabInfo {
	out := make([]PrefabInfo, 0, len(prefabInfos))
	for _, info := range prefabInfos {
		if info.Category == category {
			out = append(out, info)
		}
	}
	return out
}

func buildPrefab(id string, palette scenePalette) (tetra3d.INode, error) {
	var node tetra3d.INode
	switch id {
	case "tree_pine":
		node = buildPineTreePrefab(palette)
	case "tree_round":
		node = buildRoundTreePrefab(palette)
	case "rock":
		node = buildRockPrefab(palette)
	case "wolf":
		node = buildWolfPrefab(palette)
	default:
		return nil, fmt.Errorf("unknown prefab %q", id)
	}
	return anchorNodeToTile(node), nil
}

func newScenePalette() scenePalette {
	return scenePalette{
		grass:      newMaterial("Grass", 0.29, 0.52, 0.23),
		dirt:       newMaterial("Dirt", 0.45, 0.34, 0.20),
		bark:       newMaterial("Bark", 0.38, 0.24, 0.13),
		leaves:     newMaterial("Leaves", 0.20, 0.43, 0.18),
		leavesAlt:  newMaterial("LeavesAlt", 0.28, 0.52, 0.20),
		stone:      newMaterial("Stone", 0.48, 0.50, 0.56),
		wall:       newMaterial("Wall", 0.62, 0.55, 0.42),
		playerCoat: newMaterial("PlayerCoat", 0.16, 0.27, 0.66),
		playerSkin: newMaterial("PlayerSkin", 0.88, 0.74, 0.60),
		playerHat:  newMaterial("PlayerHat", 0.70, 0.18, 0.14),
		playerBoot: newMaterial("PlayerBoot", 0.20, 0.12, 0.08),
		wolfFur:    newMaterial("WolfFur", 0.48, 0.50, 0.54),
		wolfDark:   newMaterial("WolfDark", 0.25, 0.27, 0.31),
		eyes:       newMaterial("Eyes", 0.08, 0.08, 0.10),
	}
}

func buildPineTreePrefab(palette scenePalette) tetra3d.INode {
	root := tetra3d.NewNode("TreePine")
	trunkMesh := tetra3d.NewCylinderMesh(6, 0.18, 1.5, true)
	foliageMesh := tetra3d.NewPrismMesh()

	root.AddChildren(
		newPrimitiveModel("Trunk", trunkMesh, palette.bark, v3(0, 0.75, 0), v3(1, 1, 1)),
		newPrimitiveModel("CanopyLow", foliageMesh, palette.leaves, v3(0, 1.65, 0), v3(0.95, 0.78, 0.95)),
		newPrimitiveModel("CanopyMid", foliageMesh, palette.leaves, v3(0, 2.25, 0), v3(0.72, 0.62, 0.72)),
		newPrimitiveModel("CanopyTop", foliageMesh, palette.leavesAlt, v3(0, 2.75, 0), v3(0.46, 0.42, 0.46)),
	)
	return root
}

func buildRoundTreePrefab(palette scenePalette) tetra3d.INode {
	root := tetra3d.NewNode("TreeRound")
	trunkMesh := tetra3d.NewCylinderMesh(6, 0.16, 1.2, true)
	foliageMesh := tetra3d.NewIcosphereMesh(0)

	root.AddChildren(
		newPrimitiveModel("Trunk", trunkMesh, palette.bark, v3(0, 0.6, 0), v3(1, 1, 1)),
		newPrimitiveModel("CanopyBase", foliageMesh, palette.leavesAlt, v3(0, 1.55, 0), v3(0.55, 0.48, 0.55)),
		newPrimitiveModel("CanopyTop", foliageMesh, palette.leaves, v3(0, 2.05, 0), v3(0.42, 0.38, 0.42)),
	)
	return root
}

func buildRockPrefab(palette scenePalette) tetra3d.INode {
	root := tetra3d.NewNode("Rock")
	body := newPrimitiveModel("Body", tetra3d.NewIcosphereMesh(0), palette.stone, v3(0, 0.28, 0), v3(0.5, 0.32, 0.42))
	body.SetLocalRotation(
		tetra3d.NewMatrix4Rotate(0, 1, 0, 0.35).
			Rotated(1, 0, 0, -0.18),
	)
	root.AddChildren(body)
	return root
}

func buildPlayerPrefab(palette scenePalette) tetra3d.INode {
	root := tetra3d.NewNode("Player")

	cubeMesh := tetra3d.NewCubeMesh()
	headMesh := tetra3d.NewIcosphereMesh(0)
	hatMesh := tetra3d.NewPrismMesh()

	root.AddChildren(
		newPrimitiveModel("LeftBoot", cubeMesh, palette.playerBoot, v3(-0.13, 0.08, 0.02), v3(0.13, 0.10, 0.18)),
		newPrimitiveModel("RightBoot", cubeMesh, palette.playerBoot, v3(0.13, 0.08, 0.02), v3(0.13, 0.10, 0.18)),
		newPrimitiveModel("LeftLeg", cubeMesh, palette.playerCoat, v3(-0.13, 0.34, 0), v3(0.10, 0.28, 0.10)),
		newPrimitiveModel("RightLeg", cubeMesh, palette.playerCoat, v3(0.13, 0.34, 0), v3(0.10, 0.28, 0.10)),
		newPrimitiveModel("Torso", cubeMesh, palette.playerCoat, v3(0, 0.92, 0), v3(0.34, 0.48, 0.24)),
		newPrimitiveModel("LeftArm", cubeMesh, palette.playerSkin, v3(-0.29, 0.94, 0), v3(0.08, 0.30, 0.08)),
		newPrimitiveModel("RightArm", cubeMesh, palette.playerSkin, v3(0.29, 0.94, 0), v3(0.08, 0.30, 0.08)),
		newPrimitiveModel("Head", headMesh, palette.playerSkin, v3(0, 1.56, 0.02), v3(0.19, 0.22, 0.19)),
		newPrimitiveModel("Nose", cubeMesh, palette.playerSkin, v3(0, 1.50, 0.19), v3(0.04, 0.04, 0.05)),
		newPrimitiveModel("Hat", hatMesh, palette.playerHat, v3(0, 1.90, 0.02), v3(0.23, 0.16, 0.23)),
	)

	return root
}

func buildWolfPrefab(palette scenePalette) tetra3d.INode {
	root := tetra3d.NewNode("Wolf")

	cubeMesh := tetra3d.NewCubeMesh()
	tailMesh := tetra3d.NewPrismMesh()

	root.AddChildren(
		newPrimitiveModel("Body", cubeMesh, palette.wolfFur, v3(0, 0.52, 0), v3(0.56, 0.24, 0.28)),
		newPrimitiveModel("Shoulders", cubeMesh, palette.wolfFur, v3(0, 0.58, 0.28), v3(0.36, 0.20, 0.22)),
		newPrimitiveModel("Head", cubeMesh, palette.wolfFur, v3(0, 0.60, 0.72), v3(0.28, 0.20, 0.22)),
		newPrimitiveModel("Snout", cubeMesh, palette.wolfDark, v3(0, 0.54, 0.95), v3(0.14, 0.10, 0.16)),
		newPrimitiveModel("LeftFrontLeg", cubeMesh, palette.wolfDark, v3(-0.18, 0.17, 0.24), v3(0.08, 0.22, 0.08)),
		newPrimitiveModel("RightFrontLeg", cubeMesh, palette.wolfDark, v3(0.18, 0.17, 0.24), v3(0.08, 0.22, 0.08)),
		newPrimitiveModel("LeftBackLeg", cubeMesh, palette.wolfDark, v3(-0.18, 0.17, -0.20), v3(0.08, 0.22, 0.08)),
		newPrimitiveModel("RightBackLeg", cubeMesh, palette.wolfDark, v3(0.18, 0.17, -0.20), v3(0.08, 0.22, 0.08)),
		newPrimitiveModel("LeftEye", cubeMesh, palette.eyes, v3(-0.06, 0.64, 0.86), v3(0.03, 0.03, 0.03)),
		newPrimitiveModel("RightEye", cubeMesh, palette.eyes, v3(0.06, 0.64, 0.86), v3(0.03, 0.03, 0.03)),
	)

	tail := newPrimitiveModel("Tail", tailMesh, palette.wolfFur, v3(0, 0.72, -0.70), v3(0.09, 0.11, 0.23))
	tail.SetLocalRotation(tetra3d.NewMatrix4Rotate(1, 0, 0, -0.9))
	root.AddChildren(tail)

	leftEar := newPrimitiveModel("LeftEar", tailMesh, palette.wolfDark, v3(-0.10, 0.82, 0.72), v3(0.06, 0.10, 0.06))
	rightEar := newPrimitiveModel("RightEar", tailMesh, palette.wolfDark, v3(0.10, 0.82, 0.72), v3(0.06, 0.10, 0.06))
	root.AddChildren(leftEar, rightEar)

	return root
}

func newPrimitiveModel(name string, mesh *tetra3d.Mesh, material *tetra3d.Material, position, scale tetra3d.Vector3) *tetra3d.Model {
	model := tetra3d.NewModel(name, mesh)
	applyMaterial(model, material)
	model.SetLocalPositionVec(position)
	model.SetLocalScaleVec(scale)
	return model
}

func newWallModel(name string, material *tetra3d.Material, size tetra3d.Vector3) *tetra3d.Model {
	return newPrimitiveModel(
		name,
		tetra3d.NewCubeMesh(),
		material,
		tetra3d.Vector3{},
		v3(size.X/2, size.Y/2, size.Z/2),
	)
}

func applyMaterial(model *tetra3d.Model, material *tetra3d.Material) {
	for _, meshPart := range model.Mesh.MeshParts {
		meshPart.Material = material
	}
}

func newMaterial(name string, r, g, b float32) *tetra3d.Material {
	mat := tetra3d.NewMaterial(name)
	mat.Color = tetra3d.NewColor(r, g, b, 1)
	return mat
}

func v3(x, y, z float32) tetra3d.Vector3 {
	return tetra3d.NewVector3(x, y, z)
}

func clampScale(scale [3]float64) tetra3d.Vector3 {
	x, y, z := scale[0], scale[1], scale[2]
	if x == 0 {
		x = 1
	}
	if y == 0 {
		y = 1
	}
	if z == 0 {
		z = 1
	}
	return v3(float32(x), float32(y), float32(z))
}

func yawRotation(yaw float64) tetra3d.Matrix4 {
	return tetra3d.NewMatrix4Rotate(0, 1, 0, float32(yaw))
}

func anchorNodeToTile(node tetra3d.INode) tetra3d.INode {
	if node == nil {
		return nil
	}

	minX, maxX, minY, maxY, minZ, maxZ, ok := nodeGeometryBounds(node)
	if !ok {
		return node
	}

	centerX, centerZ := nodeSupportCenter(node, minY, maxY, minX, maxX, minZ, maxZ)

	const groundInset = 0.02

	node.SetLocalPosition(
		-centerX,
		-minY-groundInset,
		-centerZ,
	)

	anchor := tetra3d.NewNode(node.Name() + "Anchor")
	anchor.AddChildren(node)
	return anchor
}

func nodeGeometryBounds(node tetra3d.INode) (minX, maxX, minY, maxY, minZ, maxZ float32, ok bool) {
	minX = float32(math.MaxFloat32)
	minY = float32(math.MaxFloat32)
	minZ = float32(math.MaxFloat32)
	maxX = -float32(math.MaxFloat32)
	maxY = -float32(math.MaxFloat32)
	maxZ = -float32(math.MaxFloat32)

	walkModels(node, func(model *tetra3d.Model) {
		transform := model.Transform()
		for _, vertex := range model.Mesh.VertexPositions {
			world := transform.MultVec(vertex)
			if world.X < minX {
				minX = world.X
			}
			if world.X > maxX {
				maxX = world.X
			}
			if world.Y < minY {
				minY = world.Y
			}
			if world.Y > maxY {
				maxY = world.Y
			}
			if world.Z < minZ {
				minZ = world.Z
			}
			if world.Z > maxZ {
				maxZ = world.Z
			}
			ok = true
		}
	})

	return
}

func nodeSupportCenter(node tetra3d.INode, minY, maxY, fallbackMinX, fallbackMaxX, fallbackMinZ, fallbackMaxZ float32) (float32, float32) {
	spanY := maxY - minY
	supportThreshold := float32(0.05)
	if dynamicThreshold := spanY * 0.08; dynamicThreshold > supportThreshold {
		supportThreshold = dynamicThreshold
	}
	supportY := minY + supportThreshold

	minX := float32(math.MaxFloat32)
	minZ := float32(math.MaxFloat32)
	maxX := -float32(math.MaxFloat32)
	maxZ := -float32(math.MaxFloat32)
	found := false

	walkModels(node, func(model *tetra3d.Model) {
		transform := model.Transform()
		for _, vertex := range model.Mesh.VertexPositions {
			world := transform.MultVec(vertex)
			if world.Y > supportY {
				continue
			}
			if world.X < minX {
				minX = world.X
			}
			if world.X > maxX {
				maxX = world.X
			}
			if world.Z < minZ {
				minZ = world.Z
			}
			if world.Z > maxZ {
				maxZ = world.Z
			}
			found = true
		}
	})

	if !found {
		minX = fallbackMinX
		maxX = fallbackMaxX
		minZ = fallbackMinZ
		maxZ = fallbackMaxZ
	}

	return (minX + maxX) / 2, (minZ + maxZ) / 2
}

func walkModels(node tetra3d.INode, fn func(model *tetra3d.Model)) {
	model, ok := node.(*tetra3d.Model)
	if ok && model.Mesh != nil {
		fn(model)
	}
	for _, child := range node.Children() {
		walkModels(child, fn)
	}
}

func normalizeYaw(yaw float64) float64 {
	for yaw > math.Pi {
		yaw -= math.Pi * 2
	}
	for yaw <= -math.Pi {
		yaw += math.Pi * 2
	}
	return yaw
}
