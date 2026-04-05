package data

import (
	"encoding/json"
	"errors"
	"io/fs"

	"github.com/davidbennett/go-paper-rpg/internal/rpg"
)

// GameData holds all loaded game content definitions.
type GameData struct {
	Enemies map[string]EnemyDef `json:"enemies"`
	Items   map[string]ItemDef  `json:"items"`
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
	StarPoints int         `json:"star_points"`
	CoinsDrop  [2]int      `json:"coins_drop"` // [min, max]
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

// ItemDef defines an item loaded from JSON.
type ItemDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Type        string         `json:"type"`
	Effect      rpg.ItemEffect `json:"effect"`
	BuyPrice    int            `json:"buy_price"`
	SellPrice   int            `json:"sell_price"`
}

// RoomDef defines a room registry entry loaded from JSON.
type RoomDef struct {
	Name      string `json:"name"`
	MapFile   string `json:"map_file"`
	SceneFile string `json:"scene_file,omitempty"` // Legacy field for older data.
	SceneName string `json:"scene_name,omitempty"` // Legacy field for older data.
	BGM       string `json:"bgm"`
}

type NPCDef struct {
	ID       string     `json:"id"`
	Position [3]float64 `json:"position"`
	Dialogue string     `json:"dialogue"`
}

type FieldEnemyDef struct {
	Type        string   `json:"type"`
	PatrolPath  string   `json:"patrol_path"`
	BattleGroup []string `json:"battle_group"`
}

type TransitionDef struct {
	TriggerNode string `json:"trigger_node"`
	TargetRoom  string `json:"target_room"`
	Spawn       string `json:"spawn"`
}

// LoadGameData loads all game data from the given filesystem.
func LoadGameData(fsys fs.FS) (*GameData, error) {
	data := &GameData{
		Enemies: make(map[string]EnemyDef),
		Items:   make(map[string]ItemDef),
		Rooms:   make(map[string]RoomDef),
	}

	// Load enemies
	if err := loadJSON(fsys, "data/enemies.json", &data.Enemies); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	}

	// Load items
	if err := loadJSON(fsys, "data/items.json", &data.Items); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	}

	// Load rooms
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
