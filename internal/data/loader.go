package data

import (
	"encoding/json"
	"errors"
	"io/fs"
)

// GameData holds all loaded game content definitions.
type GameData struct {
	Enemies map[string]EnemyDef `json:"enemies"`
	Rooms   map[string]RoomDef  `json:"rooms"`
}

// EnemyDef defines an enemy type loaded from JSON.
type EnemyDef struct {
	Name       string      `json:"name"`
	HP         int         `json:"hp"`
	Attack     int         `json:"attack"`
	Defense    int         `json:"defense"`
	Sprite     string      `json:"sprite"`
	Moves      []MoveDef   `json:"moves"`
	AIPatterns []AIPattern `json:"ai_patterns"`
}

type MoveDef struct {
	Name          string `json:"name"`
	Power         int    `json:"power"`
	Type          string `json:"type"`
	ActionCommand string `json:"action_command"`
}

type AIPattern struct {
	Move      string `json:"move"`
	Weight    int    `json:"weight"`
	Condition string `json:"condition"`
}

// RoomDef defines a room registry entry loaded from JSON.
type RoomDef struct {
	Name    string `json:"name"`
	MapFile string `json:"map_file"`
	BGM     string `json:"bgm"`
}

// LoadGameData loads all game data from the given filesystem.
func LoadGameData(fsys fs.FS) (*GameData, error) {
	data := &GameData{
		Enemies: make(map[string]EnemyDef),
		Rooms:   make(map[string]RoomDef),
	}

	if err := loadJSON(fsys, "data/enemies.json", &data.Enemies); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	}

	if err := loadJSON(fsys, "data/rooms.json", &data.Rooms); err != nil {
		return nil, err
	}

	return data, nil
}

func loadJSON(fsys fs.FS, path string, dest any) error {
	f, err := fsys.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewDecoder(f).Decode(dest)
}
