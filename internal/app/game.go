package app

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/davidbennett/go-paper-rpg/internal/data"
	"github.com/davidbennett/go-paper-rpg/internal/input"
	"github.com/davidbennett/go-paper-rpg/internal/render"
	"github.com/davidbennett/go-paper-rpg/internal/state"
	"github.com/davidbennett/go-paper-rpg/internal/world"
)

const startRoomID = "test_room"

type Game struct {
	inputMgr *input.Manager
	stateMgr *state.Manager
	renderer *render.Renderer
}

func NewGame() *Game {
	g := &Game{}
	g.inputMgr = input.NewManager()
	g.stateMgr = state.NewManager(g.inputMgr)
	g.renderer = render.NewRenderer()

	fsys, _, err := data.OpenAssetsFS()
	if err != nil {
		panic(err)
	}
	gameData, err := data.LoadGameData(fsys)
	if err != nil {
		panic(err)
	}

	roomDef, ok := gameData.Rooms[startRoomID]
	if !ok {
		panic("start room not found: " + startRoomID)
	}
	if roomDef.MapFile == "" {
		panic("start room missing map_file: " + startRoomID)
	}

	mapDef, err := data.LoadMap(fsys, roomDef.MapFile)
	if err != nil {
		panic(err)
	}

	scene, player, err := world.BuildScene(mapDef)
	if err != nil {
		panic(err)
	}

	ow := state.NewOverworldState(g.stateMgr.Shared(), g.renderer, player, scene)
	g.stateMgr.Push(ow)

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
	return render.InternalW, render.InternalH
}
