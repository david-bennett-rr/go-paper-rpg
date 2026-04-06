package app

import (
	"fmt"
	"image/color"
	"io/fs"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/davidbennett/go-paper-rpg/internal/data"
	"github.com/davidbennett/go-paper-rpg/internal/editor"
	"github.com/davidbennett/go-paper-rpg/internal/input"
	"github.com/davidbennett/go-paper-rpg/internal/render"
	"github.com/davidbennett/go-paper-rpg/internal/state"
	"github.com/davidbennett/go-paper-rpg/internal/world"
)

const startRoomID = "test_room"

type Game struct {
	inputMgr  *input.Manager
	stateMgr  *state.Manager
	renderer  *render.Renderer
	assetsFS  fs.FS
	assetsDir string
	gameData  *data.GameData
	layoutW   int
	layoutH   int
}

func NewGame() *Game {
	g := &Game{}
	g.inputMgr = input.NewManager()
	g.stateMgr = state.NewManager(g.inputMgr)
	g.renderer = render.NewRenderer()

	fsys, assetsDir, err := data.OpenAssetsFS()
	if err != nil {
		panic(err)
	}
	gameData, err := data.LoadGameData(fsys)
	if err != nil {
		panic(err)
	}
	g.assetsFS = fsys
	g.assetsDir = assetsDir
	g.gameData = gameData
	g.setPlayWindow()

	if err := g.switchToOverworld(startRoomID); err != nil {
		panic(err)
	}

	return g
}

func (g *Game) Update() error {
	g.inputMgr.Update()
	return g.stateMgr.Update()
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 100, G: 149, B: 237, A: 255})
	g.stateMgr.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	if g.layoutW > 0 && g.layoutH > 0 {
		// Editor mode uses a fixed layout.
		return g.layoutW, g.layoutH
	}
	g.renderer.Resize(outsideWidth, outsideHeight)
	return outsideWidth, outsideHeight
}

func (g *Game) openEditor(roomID string) error {
	editorApp, err := editor.NewEmbeddedApp(g.assetsFS, g.assetsDir, g.gameData, roomID, g.switchToOverworld, g.inputMgr)
	if err != nil {
		return err
	}

	g.stateMgr.Switch(state.NewEditorState(editorApp, g.setEditorWindow))
	return nil
}

func (g *Game) switchToOverworld(roomID string) error {
	roomDef, ok := g.gameData.Rooms[roomID]
	if !ok {
		return fmt.Errorf("room not found: %s", roomID)
	}
	if roomDef.MapFile == "" {
		return fmt.Errorf("room missing map_file: %s", roomID)
	}

	mapDef, err := data.LoadMap(g.assetsFS, roomDef.MapFile)
	if err != nil {
		return err
	}

	scene, player, enemies, err := world.BuildScene(mapDef)
	if err != nil {
		return err
	}

	ow := state.NewOverworldState(g.stateMgr.Shared(), g.renderer, roomID, player, enemies, scene, g.gameData, g.openEditor)
	g.setPlayWindow()
	if g.stateMgr.Current() == nil {
		g.stateMgr.Push(ow)
	} else {
		g.stateMgr.Switch(ow)
	}
	return nil
}

func (g *Game) setPlayWindow() {
	g.layoutW = 0
	g.layoutH = 0
	ebiten.SetWindowTitle("Paper RPG")
}

func (g *Game) setEditorWindow() {
	g.layoutW = editor.DefaultWindowW
	g.layoutH = editor.DefaultWindowH
	ebiten.SetWindowTitle("Paper RPG - Map Editor")
}
