package editor

import (
	"errors"
	"fmt"
	"image/color"
	"io/fs"
	"math"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/davidbennett/go-paper-rpg/internal/data"
	"github.com/davidbennett/go-paper-rpg/internal/input"
	"github.com/davidbennett/go-paper-rpg/internal/world"
)

const (
	DefaultWindowW = 1400
	DefaultWindowH = 900
	sidebarWidth   = 320
	buttonHeight   = 28
	gridSpacing    = 1.0
	defaultZoom    = 36.0
	minZoom        = 12.0
	maxZoom        = 96.0
	snapIncrement  = 1.0
	wallHeight     = 1.5
	rotationStep   = math.Pi / 12
	defaultRadius  = 1.0
	iconTileSize   = 18.0
)

const flashDuration = 120 // frames (~2 seconds at 60fps)

type Tool string

const (
	ToolSpawn    Tool = "spawn"
	ToolLocation Tool = "location"
	ToolProp     Tool = "prop"
	ToolWall     Tool = "wall"
	ToolEnemy    Tool = "enemy"
	ToolTerrain  Tool = "terrain"
	ToolErase    Tool = "erase"
)

type Layer string

const (
	LayerGrid      Layer = "grid"
	LayerSpawn     Layer = "spawn"
	LayerLocations Layer = "locations"
	LayerProps     Layer = "props"
	LayerWalls     Layer = "walls"
	LayerEnemies   Layer = "enemies"
)

type terrainType struct {
	id        string
	name      string
	editorClr color.RGBA
}

var terrainTypes = []terrainType{
	{id: "grass", name: "Grass", editorClr: color.RGBA{R: 90, G: 130, B: 82, A: 255}},
	{id: "dirt", name: "Dirt", editorClr: color.RGBA{R: 140, G: 110, B: 70, A: 255}},
	{id: "stone", name: "Stone", editorClr: color.RGBA{R: 130, G: 130, B: 135, A: 255}},
	{id: "sand", name: "Sand", editorClr: color.RGBA{R: 194, G: 178, B: 128, A: 255}},
	{id: "water", name: "Water", editorClr: color.RGBA{R: 70, G: 120, B: 170, A: 255}},
}

func terrainColor(id string) color.RGBA {
	for _, t := range terrainTypes {
		if t.id == id {
			return t.editorClr
		}
	}
	return terrainTypes[0].editorClr
}

type App struct {
	assetsFS       fs.FS
	assetsDir      string
	rooms          map[string]data.RoomDef
	roomIDs        []string
	currentRoom    string
	currentMap     *data.MapDef
	currentMapPath string

	selectedTool    Tool
	selectedProp    string
	selectedEnemy   string
	selectedTerrain string
	brushYaw        float64

	layerVisible map[Layer]bool

	zoom float64
	camX float64
	camZ float64

	layoutW int
	layoutH int

	panning        bool
	lastMouse      point
	brushCellValid bool
	lastBrushCell  point

	onClose  func(roomID string) error
	input    *input.Manager
	ownInput bool

	unsaved    bool
	status     string
	flashMsg   string
	flashTimer int
}

type point struct {
	x float64
	z float64
}

type rect struct {
	x float64
	y float64
	w float64
	h float64
}

type uiButton struct {
	kind   string
	value  string
	label  string
	active bool
	bounds rect
}

type layout struct {
	sidebar  rect
	viewport rect
	buttons  []uiButton
}

func NewApp() (*App, error) {
	fsys, assetsDir, err := data.OpenAssetsFS()
	if err != nil {
		return nil, err
	}

	gameData, err := data.LoadGameData(fsys)
	if err != nil {
		return nil, err
	}

	return newAppWithResources(fsys, assetsDir, gameData, "", DefaultWindowW, DefaultWindowH, nil, nil)
}

func NewEmbeddedApp(assetsFS fs.FS, assetsDir string, gameData *data.GameData, roomID string, onClose func(roomID string) error, inputMgr *input.Manager) (*App, error) {
	return newAppWithResources(assetsFS, assetsDir, gameData, roomID, DefaultWindowW, DefaultWindowH, onClose, inputMgr)
}

func newAppWithResources(assetsFS fs.FS, assetsDir string, gameData *data.GameData, roomID string, layoutW, layoutH int, onClose func(roomID string) error, inputMgr *input.Manager) (*App, error) {
	if gameData == nil {
		return nil, errors.New("missing game data")
	}
	ownInput := false
	if inputMgr == nil {
		inputMgr = input.NewManager()
		ownInput = true
	}

	roomIDs := make([]string, 0, len(gameData.Rooms))
	for id, room := range gameData.Rooms {
		if room.MapFile != "" {
			roomIDs = append(roomIDs, id)
		}
	}
	sort.Strings(roomIDs)
	if len(roomIDs) == 0 {
		return nil, errors.New("no rooms with map_file entries found in assets/data/rooms.json")
	}

	app := &App{
		assetsFS:      assetsFS,
		assetsDir:     assetsDir,
		rooms:         gameData.Rooms,
		roomIDs:       roomIDs,
		selectedTool:    ToolProp,
		selectedProp:    firstPrefabID(world.PropPrefabInfos()),
		selectedEnemy:   firstPrefabID(world.EnemyPrefabInfos()),
		selectedTerrain: "dirt",
		layerVisible: map[Layer]bool{
			LayerGrid:      true,
			LayerSpawn:     true,
			LayerLocations: true,
			LayerProps:     true,
			LayerWalls:     true,
			LayerEnemies:   true,
		},
		zoom:     defaultZoom,
		layoutW:  layoutW,
		layoutH:  layoutH,
		onClose:  onClose,
		input:    inputMgr,
		ownInput: ownInput,
		status:   "Ready",
	}

	if roomID == "" {
		roomID = preferredRoomID(roomIDs)
	}

	if err := app.loadRoom(roomID); err != nil {
		return nil, err
	}

	return app, nil
}

func (a *App) flash(msg string) {
	a.flashMsg = msg
	a.flashTimer = flashDuration
}

func (a *App) Update() error {
	if a.ownInput && a.input != nil {
		a.input.Update()
	}

	if a.flashTimer > 0 {
		a.flashTimer--
	}

	l := a.buildLayout(a.layoutW, a.layoutH)
	mx, my := ebiten.CursorPosition()
	cursor := point{x: float64(mx), z: float64(my)}

	a.handleZoom(l, cursor)
	a.handlePan(l, cursor)
	if handled := a.handleShortcuts(); handled {
		return nil
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if clicked, ok := a.buttonAt(l, cursor); ok {
			if err := a.handleButton(clicked); err != nil {
				a.status = err.Error()
			}
			a.brushCellValid = false
			return nil
		}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && l.viewport.contains(cursor.x, cursor.z) {
		worldPos, ok := a.screenToWorld(l.viewport, cursor)
		if ok {
			a.handleViewportAction(worldPos, inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft))
		}
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		a.brushCellValid = false
	}

	a.lastMouse = cursor
	return nil
}

func (a *App) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 22, G: 24, B: 28, A: 255})

	l := a.buildLayout(a.layoutW, a.layoutH)
	a.drawViewport(screen, l.viewport)
	a.drawSidebar(screen, l)

	if a.flashTimer > 0 {
		a.drawFlash(screen, l.viewport)
	}
}

func (a *App) drawFlash(screen *ebiten.Image, viewport rect) {
	alpha := uint8(255)
	if a.flashTimer < 30 {
		alpha = uint8(a.flashTimer * 255 / 30)
	}
	msgW := float64(len(a.flashMsg)*7 + 24)
	msgH := 28.0
	x := viewport.x + viewport.w/2 - msgW/2
	y := viewport.y + 16
	ebitenutil.DrawRect(screen, x, y, msgW, msgH, color.RGBA{R: 40, G: 100, B: 40, A: alpha})
	vector.StrokeRect(screen, float32(x), float32(y), float32(msgW), float32(msgH), 1, color.RGBA{R: 80, G: 180, B: 80, A: alpha}, false)
	ebitenutil.DebugPrintAt(screen, a.flashMsg, int(x)+12, int(y)+8)
}

func (a *App) Layout(_, _ int) (int, int) {
	return a.layoutW, a.layoutH
}

func (a *App) buildLayout(width, height int) layout {
	sidebar := rect{x: 0, y: 0, w: sidebarWidth, h: float64(height)}
	viewport := rect{x: sidebar.w, y: 0, w: float64(width) - sidebar.w, h: float64(height)}

	buttons := make([]uiButton, 0, 48)
	y := 12.0

	buttons = append(buttons, uiButton{
		kind:   "save",
		value:  "save",
		label:  "Save",
		active: false,
		bounds: rect{x: 12, y: y, w: sidebar.w - 24, h: buttonHeight},
	})

	if a.onClose != nil {
		y += buttonHeight + 6
		buttons = append(buttons, uiButton{
			kind:   "close",
			value:  "close",
			label:  "Back To Game",
			active: false,
			bounds: rect{x: 12, y: y, w: sidebar.w - 24, h: buttonHeight},
		})
	}

	y += buttonHeight + 18
	for _, roomID := range a.roomIDs {
		buttons = append(buttons, uiButton{
			kind:   "room",
			value:  roomID,
			label:  roomID,
			active: roomID == a.currentRoom,
			bounds: rect{x: 12, y: y, w: sidebar.w - 24, h: buttonHeight},
		})
		y += buttonHeight + 6
	}

	y += 12
	toolButtons := []struct {
		tool  Tool
		label string
	}{
		{ToolSpawn, "Spawn"},
		{ToolLocation, "Location"},
		{ToolProp, "Prop"},
		{ToolWall, "Wall Brush"},
		{ToolEnemy, "Enemy"},
		{ToolTerrain, "Terrain"},
		{ToolErase, "Erase Brush"},
	}
	for i, tool := range toolButtons {
		col := float64(i % 2)
		row := float64(i / 2)
		buttons = append(buttons, uiButton{
			kind:   "tool",
			value:  string(tool.tool),
			label:  tool.label,
			active: a.selectedTool == tool.tool,
			bounds: rect{
				x: 12 + col*((sidebar.w-36)/2+12),
				y: y + row*(buttonHeight+6),
				w: (sidebar.w - 36) / 2,
				h: buttonHeight,
			},
		})
	}
	y += 4*(buttonHeight+6) + 14

	layerButtons := []struct {
		layer Layer
		label string
	}{
		{LayerGrid, "Grid"},
		{LayerSpawn, "Spawn"},
		{LayerLocations, "Locations"},
		{LayerProps, "Props"},
		{LayerWalls, "Walls"},
		{LayerEnemies, "Enemies"},
	}
	for i, layer := range layerButtons {
		col := float64(i % 2)
		row := float64(i / 2)
		buttons = append(buttons, uiButton{
			kind:   "layer",
			value:  string(layer.layer),
			label:  layer.label,
			active: a.layerVisible[layer.layer],
			bounds: rect{
				x: 12 + col*((sidebar.w-36)/2+12),
				y: y + row*(buttonHeight+6),
				w: (sidebar.w - 36) / 2,
				h: buttonHeight,
			},
		})
	}
	y += 3*(buttonHeight+6) + 14

	prefabs := a.currentPrefabButtons(sidebar.w, y)
	buttons = append(buttons, prefabs...)

	return layout{
		sidebar:  sidebar,
		viewport: viewport,
		buttons:  buttons,
	}
}

func (a *App) currentPrefabButtons(sidebarW, startY float64) []uiButton {
	buttons := []uiButton{}
	currentY := startY

	switch a.selectedTool {
	case ToolProp:
		for _, info := range world.PropPrefabInfos() {
			buttons = append(buttons, uiButton{
				kind:   "prop",
				value:  info.ID,
				label:  info.Icon + "  " + info.Name,
				active: a.selectedProp == info.ID,
				bounds: rect{x: 12, y: currentY, w: sidebarW - 24, h: buttonHeight},
			})
			currentY += buttonHeight + 6
		}
	case ToolEnemy:
		for _, info := range world.EnemyPrefabInfos() {
			buttons = append(buttons, uiButton{
				kind:   "enemy",
				value:  info.ID,
				label:  info.Icon + "  " + info.Name,
				active: a.selectedEnemy == info.ID,
				bounds: rect{x: 12, y: currentY, w: sidebarW - 24, h: buttonHeight},
			})
			currentY += buttonHeight + 6
		}
	case ToolTerrain:
		for _, t := range terrainTypes {
			buttons = append(buttons, uiButton{
				kind:   "terrain",
				value:  t.id,
				label:  t.name,
				active: a.selectedTerrain == t.id,
				bounds: rect{x: 12, y: currentY, w: sidebarW - 24, h: buttonHeight},
			})
			currentY += buttonHeight + 6
		}
	}

	return buttons
}

func (a *App) handleZoom(view layout, cursor point) {
	_, wheelY := ebiten.Wheel()
	if wheelY == 0 || !view.viewport.contains(cursor.x, cursor.z) {
		return
	}

	oldZoom := a.zoom
	a.zoom *= math.Pow(1.1, wheelY)
	a.zoom = clamp(a.zoom, minZoom, maxZoom)
	if oldZoom == a.zoom {
		return
	}

	worldPos, ok := a.screenToWorld(view.viewport, cursor)
	if !ok {
		return
	}
	a.camX = worldPos.x - (cursor.x-(view.viewport.x+view.viewport.w/2))/a.zoom
	a.camZ = worldPos.z - (cursor.z-(view.viewport.y+view.viewport.h/2))/a.zoom
}

func (a *App) handlePan(view layout, cursor point) {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) && view.viewport.contains(cursor.x, cursor.z) {
		a.panning = true
		a.lastMouse = cursor
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonRight) {
		a.panning = false
	}
	if !a.panning || !ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		return
	}

	dx := cursor.x - a.lastMouse.x
	dz := cursor.z - a.lastMouse.z
	a.camX -= dx / a.zoom
	a.camZ -= dz / a.zoom
}

func (a *App) handleShortcuts() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		a.brushYaw -= rotationStep
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		a.brushYaw += rotationStep
	}

	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	if ctrl && inpututil.IsKeyJustPressed(ebiten.KeyS) {
		if err := a.saveCurrentMap(); err != nil {
			a.status = err.Error()
		}
		return true
	}

	menuPressed := inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyF1)
	cancelPressed := false
	if a.input != nil && a.input.Handler() != nil {
		menuPressed = menuPressed || a.input.Handler().ActionIsJustPressed(input.ActionMenu)
		cancelPressed = a.input.Handler().ActionIsJustPressed(input.ActionCancel)
	}

	if a.onClose != nil && (menuPressed || cancelPressed) {
		if err := a.exitToGame(); err != nil {
			a.status = err.Error()
		}
		return true
	}

	return false
}

func (a *App) handleButton(button uiButton) error {
	switch button.kind {
	case "save":
		return a.saveCurrentMap()
	case "close":
		return a.exitToGame()
	case "room":
		if a.unsaved {
			if err := a.saveCurrentMap(); err != nil {
				return err
			}
		}
		return a.loadRoom(button.value)
	case "tool":
		a.selectedTool = Tool(button.value)
		return nil
	case "prop":
		a.selectedProp = button.value
		return nil
	case "enemy":
		a.selectedEnemy = button.value
		return nil
	case "terrain":
		a.selectedTerrain = button.value
		return nil
	case "layer":
		layer := Layer(button.value)
		a.layerVisible[layer] = !a.layerVisible[layer]
		return nil
	default:
		return nil
	}
}

func (a *App) handleViewportAction(pos point, justPressed bool) {
	cell := point{x: snap(pos.x), z: snap(pos.z)}

	switch a.selectedTool {
	case ToolWall:
		a.paintWallCell(cell)
	case ToolErase:
		a.eraseAtCell(cell)
	case ToolSpawn:
		if justPressed {
			a.currentMap.PlayerSpawn.Position = [3]float64{cell.x, 0, cell.z}
			a.currentMap.PlayerSpawn.Yaw = a.brushYaw
			a.unsaved = true
			a.status = "Updated player spawn"
		}
	case ToolLocation:
		if justPressed {
			a.currentMap.Locations = append(a.currentMap.Locations, data.LocationDef{
				ID:       nextID("location", locationIDs(a.currentMap.Locations)),
				Position: [3]float64{cell.x, 0, cell.z},
				Radius:   defaultRadius,
			})
			a.unsaved = true
			a.status = "Added location"
		}
	case ToolProp:
		if justPressed {
			a.currentMap.Props = append(a.currentMap.Props, data.PropDef{
				ID:       nextID("prop", propIDs(a.currentMap.Props)),
				Prefab:   a.selectedProp,
				Position: [3]float64{cell.x, 0, cell.z},
				Yaw:      a.brushYaw,
				Scale:    [3]float64{1, 1, 1},
			})
			a.unsaved = true
			a.status = "Added prop"
		}
	case ToolEnemy:
		if justPressed {
			a.currentMap.Enemies = append(a.currentMap.Enemies, data.MapEnemyDef{
				ID:          nextID("enemy", enemyIDs(a.currentMap.Enemies)),
				Prefab:      a.selectedEnemy,
				Position:    [3]float64{cell.x, 0, cell.z},
				Yaw:         a.brushYaw,
				BattleGroup: []string{a.selectedEnemy},
			})
			a.unsaved = true
			a.status = "Added enemy"
		}
	case ToolTerrain:
		a.paintTerrain(cell)
	}
}

func (a *App) paintTerrain(cell point) {
	if a.sameBrushCell(cell) {
		return
	}
	a.lastBrushCell = cell
	a.brushCellValid = true

	pos := [2]float64{cell.x, cell.z}
	// If painting grass, remove the tile (grass is the default)
	if a.selectedTerrain == "grass" {
		for i, t := range a.currentMap.Ground.Terrain {
			if t.Position == pos {
				a.currentMap.Ground.Terrain = append(a.currentMap.Ground.Terrain[:i], a.currentMap.Ground.Terrain[i+1:]...)
				a.unsaved = true
				a.status = "Cleared terrain"
				return
			}
		}
		return
	}
	// Update existing or append new
	for i, t := range a.currentMap.Ground.Terrain {
		if t.Position == pos {
			if t.Type == a.selectedTerrain {
				return
			}
			a.currentMap.Ground.Terrain[i].Type = a.selectedTerrain
			a.unsaved = true
			a.status = "Painted " + a.selectedTerrain
			return
		}
	}
	a.currentMap.Ground.Terrain = append(a.currentMap.Ground.Terrain, data.TerrainDef{
		Position: pos,
		Type:     a.selectedTerrain,
	})
	a.unsaved = true
	a.status = "Painted " + a.selectedTerrain
}

func (a *App) paintWallCell(cell point) {
	if a.sameBrushCell(cell) {
		return
	}
	a.lastBrushCell = cell
	a.brushCellValid = true

	if a.wallContainsCell(cell) {
		return
	}

	a.currentMap.Walls = append(a.currentMap.Walls, data.WallDef{
		ID:       nextID("wall", wallIDs(a.currentMap.Walls)),
		Position: [3]float64{cell.x, wallHeight / 2, cell.z},
		Size:     [3]float64{1, wallHeight, 1},
		Yaw:      0,
	})
	a.unsaved = true
	a.status = "Painted wall"
}

func (a *App) eraseAtCell(cell point) {
	if a.sameBrushCell(cell) {
		return
	}
	a.lastBrushCell = cell
	a.brushCellValid = true

	if a.eraseWallAtCell(cell) || a.erasePropAtCell(cell) || a.eraseEnemyAtCell(cell) || a.eraseLocationAtCell(cell) || a.eraseTerrainAtCell(cell) {
		a.unsaved = true
		a.status = "Erased"
	}
}

func (a *App) sameBrushCell(cell point) bool {
	return a.brushCellValid && a.lastBrushCell == cell
}

func (a *App) wallContainsCell(cell point) bool {
	for _, wall := range a.currentMap.Walls {
		if pointInWall(cell, wall) {
			return true
		}
	}
	return false
}

func (a *App) eraseWallAtCell(cell point) bool {
	for i, wall := range a.currentMap.Walls {
		if pointInWall(cell, wall) {
			a.currentMap.Walls = append(a.currentMap.Walls[:i], a.currentMap.Walls[i+1:]...)
			return true
		}
	}
	return false
}

func (a *App) erasePropAtCell(cell point) bool {
	for i, prop := range a.currentMap.Props {
		if sameCell(cell, point{x: prop.Position[0], z: prop.Position[2]}) {
			a.currentMap.Props = append(a.currentMap.Props[:i], a.currentMap.Props[i+1:]...)
			return true
		}
	}
	return false
}

func (a *App) eraseEnemyAtCell(cell point) bool {
	for i, enemy := range a.currentMap.Enemies {
		if sameCell(cell, point{x: enemy.Position[0], z: enemy.Position[2]}) {
			a.currentMap.Enemies = append(a.currentMap.Enemies[:i], a.currentMap.Enemies[i+1:]...)
			return true
		}
	}
	return false
}

func (a *App) eraseTerrainAtCell(cell point) bool {
	pos := [2]float64{cell.x, cell.z}
	for i, t := range a.currentMap.Ground.Terrain {
		if t.Position == pos {
			a.currentMap.Ground.Terrain = append(a.currentMap.Ground.Terrain[:i], a.currentMap.Ground.Terrain[i+1:]...)
			return true
		}
	}
	return false
}

func (a *App) eraseLocationAtCell(cell point) bool {
	for i, location := range a.currentMap.Locations {
		if sameCell(cell, point{x: location.Position[0], z: location.Position[2]}) {
			a.currentMap.Locations = append(a.currentMap.Locations[:i], a.currentMap.Locations[i+1:]...)
			return true
		}
	}
	return false
}

func (a *App) drawViewport(screen *ebiten.Image, viewport rect) {
	ebitenutil.DrawRect(screen, viewport.x, viewport.y, viewport.w, viewport.h, color.RGBA{R: 58, G: 70, B: 77, A: 255})
	a.drawGround(screen, viewport)
	if a.layerVisible[LayerGrid] {
		a.drawGrid(screen, viewport)
	}
	if a.layerVisible[LayerWalls] {
		a.drawWalls(screen, viewport)
	}
	if a.layerVisible[LayerProps] {
		a.drawProps(screen, viewport)
	}
	if a.layerVisible[LayerEnemies] {
		a.drawEnemies(screen, viewport)
	}
	if a.layerVisible[LayerLocations] {
		a.drawLocations(screen, viewport)
	}
	if a.layerVisible[LayerSpawn] {
		a.drawSpawn(screen, viewport)
	}
	a.drawCursorCell(screen, viewport)
	vector.StrokeRect(screen, float32(viewport.x), float32(viewport.y), float32(viewport.w), float32(viewport.h), 2, color.RGBA{R: 20, G: 24, B: 28, A: 255}, false)
}

func (a *App) drawSidebar(screen *ebiten.Image, l layout) {
	ebitenutil.DrawRect(screen, l.sidebar.x, l.sidebar.y, l.sidebar.w, l.sidebar.h, color.RGBA{R: 32, G: 34, B: 38, A: 255})
	vector.StrokeRect(screen, float32(l.sidebar.x), float32(l.sidebar.y), float32(l.sidebar.w), float32(l.sidebar.h), 2, color.RGBA{R: 55, G: 58, B: 64, A: 255}, false)

	info := []string{
		fmt.Sprintf("Room: %s", a.currentRoom),
		fmt.Sprintf("Tool: %s", toolLabel(a.selectedTool)),
		fmt.Sprintf("Yaw: %.0f deg", a.brushYaw*180/math.Pi),
	}
	if a.unsaved {
		info = append(info, "* Unsaved changes")
	}
	info = append(info,
		"",
		"LMB: paint  RMB: pan  Wheel: zoom",
		"Q/E: rotate  Ctrl+S: save",
	)
	if a.onClose != nil {
		info = append(info, "Esc: back to game")
	}
	info = append(info, "", a.status)
	ebitenutil.DebugPrintAt(screen, strings.Join(info, "\n"), 12, int(l.sidebar.h)-200)

	for _, button := range l.buttons {
		a.drawButton(screen, button)
	}
}

func (a *App) drawButton(screen *ebiten.Image, button uiButton) {
	bg := color.RGBA{R: 48, G: 51, B: 56, A: 255}
	border := color.RGBA{R: 68, G: 72, B: 80, A: 255}
	if button.active {
		bg = color.RGBA{R: 56, G: 90, B: 56, A: 255}
		border = color.RGBA{R: 80, G: 130, B: 80, A: 255}
	}
	ebitenutil.DrawRect(screen, button.bounds.x, button.bounds.y, button.bounds.w, button.bounds.h, bg)
	vector.StrokeRect(screen, float32(button.bounds.x), float32(button.bounds.y), float32(button.bounds.w), float32(button.bounds.h), 1, border, false)
	ebitenutil.DebugPrintAt(screen, button.label, int(button.bounds.x)+8, int(button.bounds.y)+8)
}

func (a *App) drawGround(screen *ebiten.Image, viewport rect) {
	ebitenutil.DrawRect(screen, viewport.x, viewport.y, viewport.w, viewport.h, color.RGBA{R: 90, G: 130, B: 82, A: 255})

	for _, t := range a.currentMap.Ground.Terrain {
		clr := terrainColor(t.Type)
		cell := point{x: t.Position[0], z: t.Position[1]}
		cellRect := a.cellScreenRect(viewport, cell)
		ebitenutil.DrawRect(screen, cellRect.x, cellRect.y, cellRect.w, cellRect.h, clr)
	}
}

func (a *App) drawGrid(screen *ebiten.Image, viewport rect) {
	gridColor := color.RGBA{R: 123, G: 155, B: 118, A: 255}
	minX, maxX, minZ, maxZ := a.visibleCellRange(viewport)
	for x := minX; x <= maxX; x++ {
		for z := minZ; z <= maxZ; z++ {
			cell := point{x: float64(x), z: float64(z)}
			cellRect := a.cellScreenRect(viewport, cell)
			vector.StrokeRect(screen, float32(cellRect.x), float32(cellRect.y), float32(cellRect.w), float32(cellRect.h), 1, gridColor, false)
		}
	}
}

func (a *App) drawWalls(screen *ebiten.Image, viewport rect) {
	for _, wall := range a.currentMap.Walls {
		forEachWallCell(wall, func(cell point) {
			center := a.worldToScreen(viewport, cell)
			size := a.zoom
			ebitenutil.DrawRect(screen, center.x-size/2, center.z-size/2, size, size, color.RGBA{R: 167, G: 148, B: 117, A: 255})
			vector.StrokeRect(screen, float32(center.x-size/2), float32(center.z-size/2), float32(size), float32(size), 1, color.RGBA{R: 95, G: 78, B: 56, A: 255}, false)
		})
	}
}

func (a *App) drawProps(screen *ebiten.Image, viewport rect) {
	for _, prop := range a.currentMap.Props {
		info, ok := world.PrefabInfoByID(prop.Prefab)
		if !ok {
			continue
		}
		center := a.worldToScreen(viewport, point{x: prop.Position[0], z: prop.Position[2]})
		a.drawIconMarker(screen, center, info.Icon, info.Color, color.RGBA{R: 28, G: 53, B: 23, A: 255})
	}
}

func (a *App) drawEnemies(screen *ebiten.Image, viewport rect) {
	for _, enemy := range a.currentMap.Enemies {
		info, ok := world.PrefabInfoByID(enemy.Prefab)
		if !ok {
			continue
		}
		center := a.worldToScreen(viewport, point{x: enemy.Position[0], z: enemy.Position[2]})
		a.drawIconMarker(screen, center, info.Icon, info.Color, color.RGBA{R: 75, G: 34, B: 34, A: 255})
	}
}

func (a *App) drawLocations(screen *ebiten.Image, viewport rect) {
	for _, location := range a.currentMap.Locations {
		center := a.worldToScreen(viewport, point{x: location.Position[0], z: location.Position[2]})
		a.drawIconMarker(screen, center, "L", color.RGBA{R: 247, G: 214, B: 118, A: 255}, color.RGBA{R: 117, G: 82, B: 26, A: 255})
	}
}

func (a *App) drawSpawn(screen *ebiten.Image, viewport rect) {
	center := a.worldToScreen(viewport, point{
		x: a.currentMap.PlayerSpawn.Position[0],
		z: a.currentMap.PlayerSpawn.Position[2],
	})
	a.drawIconMarker(screen, center, "S", color.RGBA{R: 85, G: 157, B: 255, A: 255}, color.RGBA{R: 17, G: 51, B: 96, A: 255})
	a.drawYawLine(screen, center, a.currentMap.PlayerSpawn.Yaw, 16, color.RGBA{R: 17, G: 51, B: 96, A: 255})
}

func (a *App) drawCursorCell(screen *ebiten.Image, viewport rect) {
	mx, my := ebiten.CursorPosition()
	cursor := point{x: float64(mx), z: float64(my)}
	if !viewport.contains(cursor.x, cursor.z) {
		return
	}
	cell, ok := a.screenToWorld(viewport, cursor)
	if !ok {
		return
	}
	cell.x = snap(cell.x)
	cell.z = snap(cell.z)
	cellRect := a.cellScreenRect(viewport, cell)
	inset := math.Max(2, a.zoom*0.08)
	vector.StrokeRect(
		screen,
		float32(cellRect.x+inset),
		float32(cellRect.y+inset),
		float32(cellRect.w-(inset*2)),
		float32(cellRect.h-(inset*2)),
		2,
		color.RGBA{R: 255, G: 245, B: 179, A: 255},
		false,
	)
}

func (a *App) drawIconMarker(screen *ebiten.Image, center point, label string, fill, border color.Color) {
	size := math.Min(iconTileSize, math.Max(10, a.zoom*0.55))
	ebitenutil.DrawRect(screen, center.x-size/2, center.z-size/2, size, size, fill)
	vector.StrokeRect(screen, float32(center.x-size/2), float32(center.z-size/2), float32(size), float32(size), 1, border, false)
	ebitenutil.DebugPrintAt(screen, label, int(center.x)-3, int(center.z)-4)
}

func (a *App) drawYawLine(screen *ebiten.Image, center point, yaw, length float64, clr color.Color) {
	end := point{
		x: center.x + math.Sin(yaw)*length,
		z: center.z + math.Cos(yaw)*length,
	}
	vector.StrokeLine(screen, float32(center.x), float32(center.z), float32(end.x), float32(end.z), 2, clr, false)
}

func (a *App) worldToScreen(viewport rect, p point) point {
	return point{
		x: viewport.x + viewport.w/2 + (p.x-a.camX)*a.zoom,
		z: viewport.y + viewport.h/2 + (p.z-a.camZ)*a.zoom,
	}
}

func (a *App) screenToWorld(viewport rect, p point) (point, bool) {
	if !viewport.contains(p.x, p.z) {
		return point{}, false
	}
	return point{
		x: a.camX + (p.x-(viewport.x+viewport.w/2))/a.zoom,
		z: a.camZ + (p.z-(viewport.y+viewport.h/2))/a.zoom,
	}, true
}

func (a *App) cellScreenRect(viewport rect, cell point) rect {
	center := a.worldToScreen(viewport, cell)
	return rect{
		x: center.x - a.zoom/2,
		y: center.z - a.zoom/2,
		w: a.zoom,
		h: a.zoom,
	}
}

func (a *App) visibleCellRange(viewport rect) (minX, maxX, minZ, maxZ int) {
	topLeft, _ := a.screenToWorld(viewport, point{x: viewport.x, z: viewport.y})
	bottomRight, _ := a.screenToWorld(viewport, point{x: viewport.x + viewport.w, z: viewport.y + viewport.h})

	minX = int(math.Floor(topLeft.x)) - 1
	maxX = int(math.Ceil(bottomRight.x)) + 1
	minZ = int(math.Floor(topLeft.z)) - 1
	maxZ = int(math.Ceil(bottomRight.z)) + 1
	return minX, maxX, minZ, maxZ
}

func (a *App) buttonAt(l layout, p point) (uiButton, bool) {
	for _, button := range l.buttons {
		if button.bounds.contains(p.x, p.z) {
			return button, true
		}
	}
	return uiButton{}, false
}

func (a *App) loadRoom(roomID string) error {
	room, ok := a.rooms[roomID]
	if !ok {
		return fmt.Errorf("unknown room %q", roomID)
	}

	mapPath := filepath.Join(a.assetsDir, filepath.FromSlash(room.MapFile))
	mapDef, err := data.LoadMap(a.assetsFS, room.MapFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			name := room.Name
			if name == "" {
				name = roomID
			}
			mapDef = data.DefaultMap(name)
		} else {
			return err
		}
	}

	a.currentRoom = roomID
	a.currentMap = mapDef
	a.currentMapPath = mapPath
	a.camX = 0
	a.camZ = 0
	a.brushCellValid = false
	a.unsaved = false
	a.status = "Loaded " + roomID
	return nil
}

func (a *App) saveCurrentMap() error {
	if a.currentMap == nil || a.currentMapPath == "" {
		return errors.New("no map loaded")
	}
	if err := data.SaveMap(a.currentMapPath, a.currentMap); err != nil {
		return err
	}
	a.unsaved = false
	a.status = "Saved " + filepath.ToSlash(a.currentMapPath)
	a.flash("Saved!")
	return nil
}

func (a *App) exitToGame() error {
	if a.onClose == nil {
		return nil
	}
	if a.unsaved {
		if err := a.saveCurrentMap(); err != nil {
			return err
		}
	}
	return a.onClose(a.currentRoom)
}

func preferredRoomID(roomIDs []string) string {
	for _, id := range roomIDs {
		if id == "test_room" {
			return id
		}
	}
	return roomIDs[0]
}

func firstPrefabID(prefabs []world.PrefabInfo) string {
	if len(prefabs) == 0 {
		return ""
	}
	return prefabs[0].ID
}

func nextID(prefix string, used []string) string {
	maxNum := 0
	for _, id := range used {
		if strings.HasPrefix(id, prefix+"_") {
			var n int
			if _, err := fmt.Sscanf(id, prefix+"_%d", &n); err == nil && n > maxNum {
				maxNum = n
			}
		}
	}
	return fmt.Sprintf("%s_%d", prefix, maxNum+1)
}

func toolLabel(tool Tool) string {
	switch tool {
	case ToolSpawn:
		return "Spawn"
	case ToolLocation:
		return "Location"
	case ToolProp:
		return "Prop"
	case ToolWall:
		return "Wall Brush"
	case ToolEnemy:
		return "Enemy"
	case ToolTerrain:
		return "Terrain"
	case ToolErase:
		return "Erase Brush"
	default:
		return string(tool)
	}
}

func locationIDs(items []data.LocationDef) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.ID)
	}
	return out
}

func propIDs(items []data.PropDef) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.ID)
	}
	return out
}

func wallIDs(items []data.WallDef) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.ID)
	}
	return out
}

func enemyIDs(items []data.MapEnemyDef) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.ID)
	}
	return out
}

func sameCell(a, b point) bool {
	return a.x == b.x && a.z == b.z
}

func pointInWall(cell point, wall data.WallDef) bool {
	found := false
	forEachWallCell(wall, func(wallCell point) {
		if sameCell(cell, wallCell) {
			found = true
		}
	})
	return found
}

func forEachWallCell(wall data.WallDef, fn func(cell point)) {
	countX := int(math.Max(1, math.Round(wall.Size[0])))
	countZ := int(math.Max(1, math.Round(wall.Size[2])))
	startX := wall.Position[0] - float64(countX-1)/2
	startZ := wall.Position[2] - float64(countZ-1)/2

	for x := 0; x < countX; x++ {
		for z := 0; z < countZ; z++ {
			fn(point{
				x: startX + float64(x),
				z: startZ + float64(z),
			})
		}
	}
}

func snap(v float64) float64 {
	return math.Round(v/snapIncrement) * snapIncrement
}

func clamp(v, minV, maxV float64) float64 {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

func (r rect) contains(x, y float64) bool {
	return x >= r.x && x <= r.x+r.w && y >= r.y && y <= r.y+r.h
}
