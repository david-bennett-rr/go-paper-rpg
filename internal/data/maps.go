package data

import (
	"encoding/json"
	"io/fs"
	"math"
	"os"
	"path/filepath"
)

// MapDef is the prefab-based source of truth for a room's traversable layout and entities.
type MapDef struct {
	Name        string        `json:"name"`
	Ground      GroundDef     `json:"ground"`
	PlayerSpawn SpawnPointDef `json:"player_spawn"`
	Locations   []LocationDef `json:"locations"`
	Props       []PropDef     `json:"props"`
	Walls       []WallDef     `json:"walls"`
	Enemies     []MapEnemyDef `json:"enemies"`
}

type GroundDef struct {
	Size    [2]float64   `json:"size"`
	Terrain []TerrainDef `json:"terrain,omitempty"`
}

type TerrainDef struct {
	Position [2]float64 `json:"position"`
	Type     string     `json:"type"`
}

type SpawnPointDef struct {
	Position [3]float64 `json:"position"`
	Yaw      float64    `json:"yaw"`
}

type LocationDef struct {
	ID       string     `json:"id"`
	Position [3]float64 `json:"position"`
	Radius   float64    `json:"radius"`
	Note     string     `json:"note,omitempty"`
}

type PropDef struct {
	ID       string     `json:"id"`
	Prefab   string     `json:"prefab"`
	Position [3]float64 `json:"position"`
	Yaw      float64    `json:"yaw"`
	Scale    [3]float64 `json:"scale,omitempty"`
}

type WallDef struct {
	ID       string     `json:"id"`
	Position [3]float64 `json:"position"`
	Size     [3]float64 `json:"size"`
	Yaw      float64    `json:"yaw"`
}

type MapEnemyDef struct {
	ID          string       `json:"id"`
	Prefab      string       `json:"prefab"`
	Position    [3]float64   `json:"position"`
	Yaw         float64      `json:"yaw"`
	BattleGroup []string     `json:"battle_group,omitempty"`
	Patrol      [][3]float64 `json:"patrol,omitempty"`
}

func DefaultMap(name string) *MapDef {
	return &MapDef{
		Name: name,
		Ground: GroundDef{
			Size: [2]float64{28, 28},
		},
		PlayerSpawn: SpawnPointDef{
			Position: [3]float64{0, 0, 0},
			Yaw:      0,
		},
		Locations: []LocationDef{},
		Props:     []PropDef{},
		Walls:     []WallDef{},
		Enemies:   []MapEnemyDef{},
	}
}

func LoadMap(fsys fs.FS, path string) (*MapDef, error) {
	var out MapDef
	if err := loadJSON(fsys, path, &out); err != nil {
		return nil, err
	}
	NormalizeMap(&out)
	return &out, nil
}

func SaveMap(path string, mapDef *MapDef) error {
	NormalizeMap(mapDef)

	data, err := json.MarshalIndent(mapDef, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func NormalizeMap(mapDef *MapDef) {
	if mapDef == nil {
		return
	}
	if mapDef.Ground.Size[0] <= 0 {
		mapDef.Ground.Size[0] = 28
	}
	if mapDef.Ground.Size[1] <= 0 {
		mapDef.Ground.Size[1] = 28
	}
	mapDef.Ground.Size[0] = snapWhole(mapDef.Ground.Size[0])
	mapDef.Ground.Size[1] = snapWhole(mapDef.Ground.Size[1])

	mapDef.PlayerSpawn.Position[0] = snapTileCenter(mapDef.PlayerSpawn.Position[0])
	mapDef.PlayerSpawn.Position[2] = snapTileCenter(mapDef.PlayerSpawn.Position[2])

	for i := range mapDef.Locations {
		mapDef.Locations[i].Position[0] = snapTileCenter(mapDef.Locations[i].Position[0])
		mapDef.Locations[i].Position[2] = snapTileCenter(mapDef.Locations[i].Position[2])
		mapDef.Locations[i].Radius = snapWhole(mapDef.Locations[i].Radius)
		if mapDef.Locations[i].Radius <= 0 {
			mapDef.Locations[i].Radius = 1
		}
	}

	for i := range mapDef.Props {
		mapDef.Props[i].Position[0] = snapTileCenter(mapDef.Props[i].Position[0])
		mapDef.Props[i].Position[2] = snapTileCenter(mapDef.Props[i].Position[2])
		if mapDef.Props[i].Scale == [3]float64{} {
			mapDef.Props[i].Scale = [3]float64{1, 1, 1}
		}
		mapDef.Props[i].Scale[0] = snapWhole(mapDef.Props[i].Scale[0])
		mapDef.Props[i].Scale[1] = snapWhole(mapDef.Props[i].Scale[1])
		mapDef.Props[i].Scale[2] = snapWhole(mapDef.Props[i].Scale[2])
	}

	for i := range mapDef.Walls {
		mapDef.Walls[i].Size[0] = snapWhole(mapDef.Walls[i].Size[0])
		mapDef.Walls[i].Size[2] = snapWhole(mapDef.Walls[i].Size[2])
		if mapDef.Walls[i].Size[0] <= 0 {
			mapDef.Walls[i].Size[0] = 1
		}
		if mapDef.Walls[i].Size[1] <= 0 {
			mapDef.Walls[i].Size[1] = 1.5
		}
		if mapDef.Walls[i].Size[2] <= 0 {
			mapDef.Walls[i].Size[2] = 1
		}
		mapDef.Walls[i].Position[0] = snapWallCenter(mapDef.Walls[i].Position[0], mapDef.Walls[i].Size[0])
		mapDef.Walls[i].Position[1] = mapDef.Walls[i].Size[1] / 2
		mapDef.Walls[i].Position[2] = snapWallCenter(mapDef.Walls[i].Position[2], mapDef.Walls[i].Size[2])
	}

	for i := range mapDef.Enemies {
		mapDef.Enemies[i].Position[0] = snapTileCenter(mapDef.Enemies[i].Position[0])
		mapDef.Enemies[i].Position[2] = snapTileCenter(mapDef.Enemies[i].Position[2])
	}
}

func snapWhole(v float64) float64 {
	return math.Round(v)
}

func snapTileCenter(v float64) float64 {
	return math.Round(v)
}

func snapWallCenter(position, span float64) float64 {
	count := int(math.Max(1, math.Round(span)))
	if count%2 == 0 {
		return math.Round(position-0.5) + 0.5
	}
	return snapTileCenter(position)
}
