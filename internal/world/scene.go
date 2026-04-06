package world

import (
	"fmt"
	"math"

	"github.com/solarlune/tetra3d"

	"github.com/davidbennett/go-paper-rpg/internal/data"
	"github.com/davidbennett/go-paper-rpg/internal/overworld"
)

const (
	sunYaw          = -3 * math.Pi / 4
	sunPitch        = -0.95
	groundTileDepth = 0.08
	wallHeightScale = 1.22
	wallWidthScale  = 1.12
)

func BuildScene(mapDef *data.MapDef) (*tetra3d.Scene, *overworld.Player, []*overworld.Enemy, error) {
	data.NormalizeMap(mapDef)

	scene := tetra3d.NewScene(mapDef.Name)
	palette := newScenePalette()

	scene.Root.AddChildren(buildGround(mapDef.Ground, palette))
	scene.Root.AddChildren(buildLights())

	for _, prop := range mapDef.Props {
		instance, err := buildPrefab(prop.Prefab, palette)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("build prop %s: %w", prop.ID, err)
		}
		placeInstance(instance, prop.Position, prop.Yaw, clampScale(prop.Scale))
		instance.SetName(prop.ID)
		scene.Root.AddChildren(instance)

		// Blob shadow for props
		info, ok := PrefabInfoByID(prop.Prefab)
		if ok {
			r := float32(info.Radius) * 2.0
			shadow := NewBlobShadow(prop.ID+"_shadow", r, r, palette)
			shadow.SetLocalPositionVec(v3(float32(prop.Position[0]), 0.02, float32(prop.Position[2])))
			scene.Root.AddChildren(shadow)
		}
	}

	for _, wallModel := range buildWalls(mapDef.Walls, palette) {
		scene.Root.AddChildren(wallModel)
	}

	enemies := make([]*overworld.Enemy, 0, len(mapDef.Enemies))
	for _, enemyDef := range mapDef.Enemies {
		instance, err := buildPrefab(enemyDef.Prefab, palette)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("build enemy %s: %w", enemyDef.ID, err)
		}

		group := enemyDef.BattleGroup
		if len(group) == 0 {
			group = []string{enemyDef.Prefab}
		}

		enemy := overworld.NewEnemy(instance, enemyDef.Prefab, enemyDef.ID, group)
		enemy.SetPlacement(
			enemyDef.Position[0],
			enemyDef.Position[1],
			enemyDef.Position[2],
			enemyDef.Yaw,
		)
		instance.SetName(enemyDef.ID)
		scene.Root.AddChildren(instance)

		// Blob shadow for enemy
		info, ok := PrefabInfoByID(enemyDef.Prefab)
		if ok {
			r := float32(info.Radius) * 2.0
			shadow := NewBlobShadow(enemyDef.ID+"_shadow", r, r, palette)
			shadow.SetLocalPositionVec(v3(float32(enemyDef.Position[0]), 0.02, float32(enemyDef.Position[2])))
			scene.Root.AddChildren(shadow)
			enemy.Shadow = shadow
		}

		enemies = append(enemies, enemy)
	}

	player := overworld.NewPlayer(BuildPlayerPrefab())
	spawn := mapDef.PlayerSpawn
	player.SetPlacement(
		spawn.Position[0],
		spawn.Position[1],
		spawn.Position[2],
		spawn.Yaw,
	)

	// Warm point light follows the player
	playerLight := tetra3d.NewPointLight("PlayerLight", 1.0, 0.92, 0.75, 0.4)
	playerLight.Range = 5
	playerLight.SetLocalPosition(0, 1.5, 0)
	player.Root.AddChildren(playerLight)

	// Player blob shadow
	playerShadow := NewBlobShadow("player_shadow", 1.2, 1.2, palette)
	scene.Root.AddChildren(playerShadow)
	player.Shadow = playerShadow

	scene.Root.AddChildren(player.Root)

	return scene, player, enemies, nil
}

func buildGround(ground data.GroundDef, palette scenePalette) tetra3d.INode {
	root := tetra3d.NewNode("Ground")

	sizeX := ground.Size[0]
	sizeZ := ground.Size[1]
	if sizeX <= 0 {
		sizeX = 28
	}
	if sizeZ <= 0 {
		sizeZ = 28
	}

	cellTypes, minX, maxX, minZ, maxZ := groundCellTypes(ground)
	cubeMesh := tetra3d.NewCubeMesh()
	for _, terrainType := range []string{"grass", "dirt", "stone", "sand", "water"} {
		mat := terrainMaterial(terrainType, palette)
		if mat == nil {
			continue
		}

		merged := tetra3d.NewModel(
			fmt.Sprintf("Ground_%s", terrainType),
			tetra3d.NewMesh(fmt.Sprintf("Ground_%s_Mesh", terrainType)),
		)
		partModels := make([]*tetra3d.Model, 0, (maxX-minX+1)*(maxZ-minZ+1))

		for x := minX; x <= maxX; x++ {
			for z := minZ; z <= maxZ; z++ {
				if cellTypes[[2]int{x, z}] != terrainType {
					continue
				}
				tile := newPrimitiveModel(
					"GroundTile",
					cubeMesh,
					mat,
					v3(float32(x), float32(-groundTileDepth/2), float32(z)),
					v3(0.5, float32(groundTileDepth/2), 0.5),
				)
				partModels = append(partModels, tile)
			}
		}

		if len(partModels) == 0 {
			continue
		}

		merged.StaticMerge(partModels...)
		root.AddChildren(merged)
	}

	return root
}

func terrainMaterial(terrainType string, palette scenePalette) *tetra3d.Material {
	switch terrainType {
	case "grass":
		return palette.grass
	case "dirt":
		return palette.dirt
	case "stone":
		return palette.stone
	case "sand":
		return newMaterial("Sand", 0.76, 0.70, 0.50)
	case "water":
		return newMaterial("Water", 0.27, 0.47, 0.67)
	default:
		return nil
	}
}

func buildLights() tetra3d.INode {
	root := tetra3d.NewNode("Lights")

	ambient := tetra3d.NewAmbientLight("Ambient", 1, 1, 1, 0.34)
	sun := tetra3d.NewDirectionalLight("Sun", 1, 0.96, 0.90, 1.1)
	sun.SetLocalRotation(
		tetra3d.NewMatrix4Rotate(0, 1, 0, float32(sunYaw)).
			Rotated(1, 0, 0, float32(sunPitch)),
	)

	root.AddChildren(ambient, sun)
	return root
}

func placeInstance(instance tetra3d.INode, position [3]float64, yaw float64, scale tetra3d.Vector3) {
	instance.SetLocalPositionVec(v3(
		float32(position[0]),
		float32(position[1]),
		float32(position[2]),
	))
	instance.SetLocalRotation(yawRotation(yaw))
	instance.SetLocalScaleVec(scale)
}

type wallProfileKey struct {
	y      int
	height int
	yaw    int
}

type wallTileKey struct {
	x       int
	z       int
	profile wallProfileKey
}

type wallTile struct {
	key      wallTileKey
	position [3]float64
	height   float64
	yaw      float64
}

func buildWalls(walls []data.WallDef, palette scenePalette) []tetra3d.INode {
	tiles := collectWallTiles(walls)
	if len(tiles) == 0 {
		return nil
	}

	components := splitWallComponents(tiles)
	cubeMesh := tetra3d.NewCubeMesh()
	models := make([]tetra3d.INode, 0, len(components))

	for i, component := range components {
		merged := tetra3d.NewModel(
			fmt.Sprintf("WallGroup_%d", i+1),
			tetra3d.NewMesh(fmt.Sprintf("WallGroup_%d_Mesh", i+1)),
		)
		partModels := make([]*tetra3d.Model, 0, len(component))

		for _, tile := range component {
			scaledHeight := tile.height * wallHeightScale
			heightLift := (scaledHeight - tile.height) / 2
			part := newPrimitiveModel(
				"WallTile",
				cubeMesh,
				palette.wall,
				tetra3d.Vector3{},
				v3(0.5*wallWidthScale, float32(scaledHeight)/2, 0.5*wallWidthScale),
			)
			part.SetLocalPositionVec(v3(
				float32(tile.position[0]),
				float32(tile.position[1]+heightLift),
				float32(tile.position[2]),
			))
			part.SetLocalRotation(yawRotation(tile.yaw))
			partModels = append(partModels, part)
		}

		merged.StaticMerge(partModels...)
		models = append(models, merged)
	}

	return models
}

func expandWallTiles(wall data.WallDef) [][3]float64 {
	countX := int(math.Max(1, math.Round(wall.Size[0])))
	countZ := int(math.Max(1, math.Round(wall.Size[2])))
	startX := wall.Position[0] - float64(countX-1)/2
	startZ := wall.Position[2] - float64(countZ-1)/2

	tiles := make([][3]float64, 0, countX*countZ)
	for x := 0; x < countX; x++ {
		for z := 0; z < countZ; z++ {
			tiles = append(tiles, [3]float64{
				startX + float64(x),
				wall.Position[1],
				startZ + float64(z),
			})
		}
	}
	return tiles
}

func collectWallTiles(walls []data.WallDef) []wallTile {
	tiles := make([]wallTile, 0, len(walls))
	seen := map[wallTileKey]struct{}{}

	for _, wall := range walls {
		profile := wallProfileKey{
			y:      quantizeWallValue(wall.Position[1]),
			height: quantizeWallValue(wall.Size[1]),
			yaw:    quantizeWallValue(normalizeYaw(wall.Yaw)),
		}

		for _, position := range expandWallTiles(wall) {
			key := wallTileKey{
				x:       quantizeWallCoord(position[0]),
				z:       quantizeWallCoord(position[2]),
				profile: profile,
			}
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}

			tiles = append(tiles, wallTile{
				key: key,
				position: [3]float64{
					dequantizeWallCoord(key.x),
					wall.Position[1],
					dequantizeWallCoord(key.z),
				},
				height: wall.Size[1],
				yaw:    wall.Yaw,
			})
		}
	}

	return tiles
}

func splitWallComponents(tiles []wallTile) [][]wallTile {
	indexByKey := make(map[wallTileKey]int, len(tiles))
	for i, tile := range tiles {
		indexByKey[tile.key] = i
	}

	visited := make([]bool, len(tiles))
	components := make([][]wallTile, 0, len(tiles))

	for i := range tiles {
		if visited[i] {
			continue
		}

		queue := []int{i}
		visited[i] = true
		component := make([]wallTile, 0, 8)

		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			component = append(component, tiles[current])

			for _, neighbor := range wallNeighborKeys(tiles[current].key) {
				next, exists := indexByKey[neighbor]
				if !exists || visited[next] {
					continue
				}
				visited[next] = true
				queue = append(queue, next)
			}
		}

		components = append(components, component)
	}

	return components
}

func wallNeighborKeys(key wallTileKey) []wallTileKey {
	return []wallTileKey{
		{x: key.x - 2, z: key.z, profile: key.profile},
		{x: key.x + 2, z: key.z, profile: key.profile},
		{x: key.x, z: key.z - 2, profile: key.profile},
		{x: key.x, z: key.z + 2, profile: key.profile},
	}
}

func quantizeWallCoord(v float64) int {
	return int(math.Round(v * 2))
}

func dequantizeWallCoord(v int) float64 {
	return float64(v) / 2
}

func quantizeWallValue(v float64) int {
	return int(math.Round(v * 1000))
}

func groundCellTypes(ground data.GroundDef) (map[[2]int]string, int, int, int, int) {
	terrainByCell := map[[2]int]string{}
	width := int(math.Max(1, math.Round(ground.Size[0])))
	depth := int(math.Max(1, math.Round(ground.Size[1])))
	minX := -int(math.Floor(float64(width-1) / 2))
	maxX := minX + width - 1
	minZ := -int(math.Floor(float64(depth-1) / 2))
	maxZ := minZ + depth - 1

	for _, tile := range ground.Terrain {
		cell := [2]int{
			int(math.Round(tile.Position[0])),
			int(math.Round(tile.Position[1])),
		}
		terrainByCell[cell] = tile.Type
		if cell[0] < minX {
			minX = cell[0]
		}
		if cell[0] > maxX {
			maxX = cell[0]
		}
		if cell[1] < minZ {
			minZ = cell[1]
		}
		if cell[1] > maxZ {
			maxZ = cell[1]
		}
	}

	cellTypes := make(map[[2]int]string, (maxX-minX+1)*(maxZ-minZ+1))
	for x := minX; x <= maxX; x++ {
		for z := minZ; z <= maxZ; z++ {
			cell := [2]int{x, z}
			terrainType, ok := terrainByCell[cell]
			if !ok || terrainType == "" {
				terrainType = "grass"
			}
			cellTypes[cell] = terrainType
		}
	}

	return cellTypes, minX, maxX, minZ, maxZ
}
