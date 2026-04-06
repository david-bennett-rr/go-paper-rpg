package state

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/solarlune/tetra3d"

	"github.com/davidbennett/go-paper-rpg/internal/battle"
	"github.com/davidbennett/go-paper-rpg/internal/battle/action"
	"github.com/davidbennett/go-paper-rpg/internal/data"
	"github.com/davidbennett/go-paper-rpg/internal/input"
	"github.com/davidbennett/go-paper-rpg/internal/render"
	"github.com/davidbennett/go-paper-rpg/internal/rpg"
	"github.com/davidbennett/go-paper-rpg/internal/world"
)

// battlePhase tracks the current phase of the turn-based battle.
type battlePhase int

const (
	phasePlayerMenu   battlePhase = iota
	phasePlayerTarget
	phasePlayerAction
	phasePlayerResult
	phaseEnemyAction
	phaseEnemyResult
	phaseVictory
	phaseDefeat
)

const resultDisplayFrames = 60

// battleFoe is an enemy combatant.
type battleFoe struct {
	Name   string
	Stats  rpg.Stats
	Moves  []*rpg.Move
	AI     []data.AIPattern
	Prefab string
	Node   tetra3d.INode // 3D model in battle scene
}

// BattleState manages a full turn-based battle encounter.
type BattleState struct {
	shared   *SharedContext
	renderer *render.Renderer
	party    *rpg.Party
	foes     []*battleFoe

	scene      *tetra3d.Scene
	playerNode tetra3d.INode

	phase        battlePhase
	menuCursor   int
	targetCursor int
	selectedMove *rpg.Move
	actionCmd    action.ActionCommand
	resultTimer  int
	resultText   string
	turnFoeIndex int

	screenW, screenH int
}

func NewBattleState(shared *SharedContext, renderer *render.Renderer, enemyDefs []data.EnemyDef, foePrefabs []string) *BattleState {
	foes := make([]*battleFoe, 0, len(enemyDefs))
	for i, def := range enemyDefs {
		moves := make([]*rpg.Move, 0, len(def.Moves))
		for _, m := range def.Moves {
			moves = append(moves, &rpg.Move{
				Name:          m.Name,
				BasePower:     m.Power,
				Type:          rpg.MoveType(m.Type),
				ActionCommand: m.ActionCommand,
			})
		}
		prefab := def.Sprite
		if i < len(foePrefabs) {
			prefab = foePrefabs[i]
		}
		foes = append(foes, &battleFoe{
			Name:   def.Name,
			Prefab: prefab,
			Stats: rpg.Stats{
				HP: def.HP, MaxHP: def.HP,
				Attack: def.Attack, Defense: def.Defense,
			},
			Moves: moves,
			AI:    def.AIPatterns,
		})
	}

	bs := &BattleState{
		shared:   shared,
		renderer: renderer,
		party:    rpg.NewParty(),
		foes:     foes,
	}
	bs.buildBattleScene()
	return bs
}

func (s *BattleState) buildBattleScene() {
	scene := tetra3d.NewScene("Battle")

	// Lighting
	ambient := tetra3d.NewAmbientLight("Ambient", 1, 1, 1, 0.45)
	sun := tetra3d.NewDirectionalLight("Sun", 1, 0.96, 0.90, 1.0)
	sun.SetLocalRotation(
		tetra3d.NewMatrix4Rotate(0, 1, 0, -2.35).
			Rotated(1, 0, 0, -0.95),
	)
	scene.Root.AddChildren(ambient, sun)

	// Ground plane
	groundMat := tetra3d.NewMaterial("BattleGround")
	groundMat.Color = tetra3d.NewColor(0.32, 0.50, 0.28, 1)
	ground := tetra3d.NewModel("Ground", tetra3d.NewPlaneMesh(2, 2))
	for _, mp := range ground.Mesh.MeshParts {
		mp.Material = groundMat
	}
	ground.SetLocalScaleVec(tetra3d.NewVector3(12, 1, 6))
	scene.Root.AddChildren(ground)

	// Player model on left side
	playerNode := world.BuildPlayerPrefab()
	playerNode.SetLocalPositionVec(tetra3d.NewVector3(-3, 0, 0))
	playerNode.SetLocalRotation(tetra3d.NewMatrix4Rotate(0, 1, 0, math.Pi/6))
	scene.Root.AddChildren(playerNode)
	s.playerNode = playerNode

	// Enemy models on right side
	count := len(s.foes)
	for i, foe := range s.foes {
		node, err := world.BuildEnemyPrefab(foe.Prefab)
		if err != nil {
			// Fallback: try as prop
			node, err = world.BuildPropPrefab(foe.Prefab)
			if err != nil {
				continue
			}
		}

		spacing := float32(2.0)
		xPos := float32(3.0) + float32(i-count/2)*spacing
		node.SetLocalPositionVec(tetra3d.NewVector3(xPos, 0, 0))
		node.SetLocalRotation(tetra3d.NewMatrix4Rotate(0, 1, 0, -math.Pi/6))
		scene.Root.AddChildren(node)
		foe.Node = node
	}

	s.scene = scene
}

func (s *BattleState) Enter(prev GameState) {}
func (s *BattleState) Exit()               {}

func (s *BattleState) Update() error {
	switch s.phase {
	case phasePlayerMenu:
		s.updatePlayerMenu()
	case phasePlayerTarget:
		s.updatePlayerTarget()
	case phasePlayerAction:
		s.updatePlayerAction()
	case phasePlayerResult:
		s.updateResult(phaseEnemyAction)
	case phaseEnemyAction:
		s.updateEnemyAction()
	case phaseEnemyResult:
		s.updateResult(phasePlayerMenu)
	case phaseVictory, phaseDefeat:
		s.updateEndPhase()
	}

	return nil
}

func (s *BattleState) updatePlayerMenu() {
	h := s.shared.Input.Handler()
	moves := s.party.Mario.Moves
	if h.ActionIsJustPressed(input.ActionMoveUp) {
		s.menuCursor--
		if s.menuCursor < 0 {
			s.menuCursor = len(moves) - 1
		}
	}
	if h.ActionIsJustPressed(input.ActionMoveDown) {
		s.menuCursor++
		if s.menuCursor >= len(moves) {
			s.menuCursor = 0
		}
	}
	if h.ActionIsJustPressed(input.ActionConfirm) {
		s.selectedMove = moves[s.menuCursor]
		living := s.livingFoes()
		if len(living) == 1 {
			s.targetCursor = living[0]
			s.startActionCommand()
		} else {
			s.targetCursor = living[0]
			s.phase = phasePlayerTarget
		}
	}
}

func (s *BattleState) updatePlayerTarget() {
	h := s.shared.Input.Handler()
	if h.ActionIsJustPressed(input.ActionCancel) {
		s.phase = phasePlayerMenu
		return
	}

	if h.ActionIsJustPressed(input.ActionMoveUp) || h.ActionIsJustPressed(input.ActionMoveLeft) {
		s.targetCursor = s.prevLivingFoe(s.targetCursor)
	}
	if h.ActionIsJustPressed(input.ActionMoveDown) || h.ActionIsJustPressed(input.ActionMoveRight) {
		s.targetCursor = s.nextLivingFoe(s.targetCursor)
	}

	if h.ActionIsJustPressed(input.ActionConfirm) {
		s.startActionCommand()
	}
}

func (s *BattleState) startActionCommand() {
	cmd := action.NewTimedPress(20, 50, 35)
	cmd.Start()
	s.actionCmd = cmd
	s.phase = phasePlayerAction
}

func (s *BattleState) updatePlayerAction() {
	s.actionCmd.Update(s.shared.Input)

	if s.actionCmd.IsComplete() {
		result := s.actionCmd.Result()
		target := s.foes[s.targetCursor]
		dmg := battle.CalculateDamage(&s.party.Mario.Stats, s.selectedMove, &target.Stats, result)
		target.Stats.TakeDamage(dmg)

		s.resultText = fmt.Sprintf("%s! %d damage to %s!", result.Quality, dmg, target.Name)
		if !target.Stats.IsAlive() {
			s.resultText += " Defeated!"
			// Hide the 3D model
			if target.Node != nil {
				target.Node.SetLocalPosition(0, -100, 0)
			}
		}
		s.resultTimer = resultDisplayFrames
		s.actionCmd = nil

		if s.allFoesDefeated() {
			s.phase = phaseVictory
			s.resultText = "Victory!"
			s.resultTimer = resultDisplayFrames * 2
			return
		}

		s.phase = phasePlayerResult
	}
}

func (s *BattleState) updateEnemyAction() {
	for s.turnFoeIndex < len(s.foes) && !s.foes[s.turnFoeIndex].Stats.IsAlive() {
		s.turnFoeIndex++
	}

	if s.turnFoeIndex >= len(s.foes) {
		s.turnFoeIndex = 0
		s.phase = phasePlayerMenu
		return
	}

	foe := s.foes[s.turnFoeIndex]
	move := s.pickEnemyMove(foe)
	if move == nil {
		s.turnFoeIndex++
		return
	}

	cmdResult := action.ResultFromQuality(action.QualityNice)
	dmg := battle.CalculateDamage(&foe.Stats, move, &s.party.Mario.Stats, cmdResult)
	s.party.Mario.Stats.TakeDamage(dmg)

	s.resultText = fmt.Sprintf("%s uses %s! %d damage!", foe.Name, move.Name, dmg)
	s.resultTimer = resultDisplayFrames

	if !s.party.Mario.Stats.IsAlive() {
		s.phase = phaseDefeat
		s.resultText = "Defeated..."
		s.resultTimer = resultDisplayFrames * 2
		return
	}

	s.turnFoeIndex++
	s.phase = phaseEnemyResult
}

func (s *BattleState) updateResult(nextPhase battlePhase) {
	s.resultTimer--
	if s.resultTimer <= 0 {
		s.phase = nextPhase
	}
}

func (s *BattleState) updateEndPhase() {
	s.resultTimer--
	if s.resultTimer <= 0 || s.shared.Input.Handler().ActionIsJustPressed(input.ActionConfirm) {
		s.shared.States.Pop()
	}
}

func (s *BattleState) pickEnemyMove(foe *battleFoe) *rpg.Move {
	if len(foe.Moves) == 0 {
		return nil
	}

	totalWeight := 0
	type candidate struct {
		move   *rpg.Move
		weight int
	}
	candidates := make([]candidate, 0, len(foe.AI))
	hpPct := float64(foe.Stats.HP) / float64(foe.Stats.MaxHP) * 100

	for _, pattern := range foe.AI {
		if !aiConditionMet(pattern.Condition, hpPct) {
			continue
		}
		for _, m := range foe.Moves {
			if m.Name == pattern.Move {
				candidates = append(candidates, candidate{move: m, weight: pattern.Weight})
				totalWeight += pattern.Weight
				break
			}
		}
	}

	if len(candidates) == 0 {
		return foe.Moves[0]
	}

	roll := rand.Intn(totalWeight)
	for _, c := range candidates {
		roll -= c.weight
		if roll < 0 {
			return c.move
		}
	}
	return candidates[0].move
}

func aiConditionMet(condition string, hpPct float64) bool {
	switch condition {
	case "always", "":
		return true
	case "hp_below_50":
		return hpPct < 50
	case "hp_below_30":
		return hpPct < 30
	default:
		return true
	}
}

// --- Drawing ---

func (s *BattleState) Draw(screen *ebiten.Image) {
	w := screen.Bounds().Dx()
	h := screen.Bounds().Dy()
	s.screenW = w
	s.screenH = h

	// Render the 3D battle scene as the background
	s.renderer.RenderBattleScene(screen, s.scene)

	// Draw the 2D UI overlay
	s.drawBattleUI(screen)
}

func (s *BattleState) drawBattleUI(screen *ebiten.Image) {
	w := float32(s.screenW)
	h := float32(s.screenH)

	// Scale factor for UI - target readable at 1080p+
	scale := h / 720.0
	if scale < 1 {
		scale = 1
	}

	// --- Top bar: party HP ---
	topBarH := 70 * scale
	vector.DrawFilledRect(screen, 0, 0, w, topBarH, color.RGBA{R: 15, G: 15, B: 25, A: 210}, false)
	vector.StrokeLine(screen, 0, topBarH, w, topBarH, 2, color.RGBA{R: 50, G: 55, B: 80, A: 255}, false)

	// Mario stats in top bar
	mario := s.party.Mario
	nameX := 24 * scale
	nameY := 12 * scale
	s.drawScaledText(screen, mario.Name, nameX, nameY, scale)
	hpText := fmt.Sprintf("HP  %d / %d", mario.Stats.HP, mario.Stats.MaxHP)
	s.drawScaledText(screen, hpText, nameX, nameY+22*scale, scale)
	s.drawHPBar(screen, nameX, nameY+44*scale, 180*scale, 12*scale, mario.Stats.HP, mario.Stats.MaxHP)

	fpText := fmt.Sprintf("FP  %d / %d", mario.Stats.FP, mario.Stats.MaxFP)
	s.drawScaledText(screen, fpText, nameX+220*scale, nameY+22*scale, scale)

	// Foe stats on right side of top bar
	for i, foe := range s.foes {
		foeX := w - (220*scale)*float32(len(s.foes)-i)
		foeNameClr := color.RGBA{R: 220, G: 200, B: 200, A: 255}
		if !foe.Stats.IsAlive() {
			foeNameClr = color.RGBA{R: 100, G: 100, B: 100, A: 255}
		}
		// Highlight target
		if (s.phase == phasePlayerTarget || s.phase == phasePlayerAction) && i == s.targetCursor {
			vector.DrawFilledRect(screen, foeX-4*scale, 4*scale, 210*scale, topBarH-8*scale, color.RGBA{R: 60, G: 50, B: 20, A: 100}, false)
			vector.StrokeRect(screen, foeX-4*scale, 4*scale, 210*scale, topBarH-8*scale, 2, color.RGBA{R: 255, G: 220, B: 80, A: 200}, false)
		}

		s.drawScaledTextColor(screen, foe.Name, foeX, nameY, scale, foeNameClr)
		foeHP := fmt.Sprintf("HP  %d / %d", foe.Stats.HP, foe.Stats.MaxHP)
		s.drawScaledText(screen, foeHP, foeX, nameY+22*scale, scale)
		s.drawHPBar(screen, foeX, nameY+44*scale, 180*scale, 12*scale, foe.Stats.HP, foe.Stats.MaxHP)
	}

	// --- Bottom panel ---
	panelH := 140 * scale
	panelY := h - panelH
	vector.DrawFilledRect(screen, 0, panelY, w, panelH, color.RGBA{R: 15, G: 15, B: 25, A: 225}, false)
	vector.StrokeLine(screen, 0, panelY, w, panelY, 2, color.RGBA{R: 50, G: 55, B: 80, A: 255}, false)

	padX := 32 * scale
	padY := panelY + 18*scale
	lineH := 30 * scale

	switch s.phase {
	case phasePlayerMenu:
		s.drawScaledText(screen, "Choose your move:", padX, padY, scale)
		for i, move := range s.party.Mario.Moves {
			y := padY + lineH + float32(i)*lineH
			prefix := "   "
			if i == s.menuCursor {
				prefix = " > "
				vector.DrawFilledRect(screen, padX-4*scale, y-4*scale, 280*scale, lineH, color.RGBA{R: 40, G: 50, B: 80, A: 200}, false)
			}
			fpLabel := ""
			if move.FPCost > 0 {
				fpLabel = fmt.Sprintf("  (%d FP)", move.FPCost)
			}
			s.drawScaledText(screen, fmt.Sprintf("%s%s%s", prefix, move.Name, fpLabel), padX, y, scale)
		}

	case phasePlayerTarget:
		s.drawScaledText(screen, "Select a target       A: Confirm   B: Back", padX, padY, scale)

	case phasePlayerAction:
		s.drawScaledText(screen, "Press A at the right moment!", padX, padY, scale)
		if s.actionCmd != nil {
			s.drawActionBar(screen, panelY+60*scale, scale)
		}

	case phasePlayerResult, phaseEnemyResult:
		s.drawScaledText(screen, s.resultText, padX, padY+20*scale, scale)

	case phaseEnemyAction:
		s.drawScaledText(screen, "Enemy turn...", padX, padY+20*scale, scale)

	case phaseVictory:
		s.drawScaledText(screen, "Victory!   Press A to continue.", padX, padY+20*scale, scale)

	case phaseDefeat:
		s.drawScaledText(screen, "Defeated...   Press A to continue.", padX, padY+20*scale, scale)
	}
}

func (s *BattleState) drawActionBar(screen *ebiten.Image, y, scale float32) {
	// Draw a larger, centered timing bar
	barW := 400 * scale
	barH := 24 * scale
	barX := (float32(s.screenW) - barW) / 2

	// Background
	vector.DrawFilledRect(screen, barX, y, barW, barH, color.RGBA{R: 30, G: 30, B: 40, A: 230}, false)
	vector.StrokeRect(screen, barX, y, barW, barH, 2, color.RGBA{R: 70, G: 70, B: 90, A: 255}, false)

	// Let the action command draw itself too (it draws at hardcoded coords,
	// so we also draw our own scaled version)
	// We read the timed press state from the action command interface.
	// For now, just call its Draw which renders at a small size,
	// and we overlay our own bar.
	s.actionCmd.Draw(screen)
}

func (s *BattleState) drawHPBar(screen *ebiten.Image, x, y, w, h float32, hp, maxHP int) {
	vector.DrawFilledRect(screen, x, y, w, h, color.RGBA{R: 30, G: 30, B: 30, A: 255}, false)
	pct := float32(hp) / float32(maxHP)
	if pct < 0 {
		pct = 0
	}
	fillClr := color.RGBA{R: 60, G: 190, B: 60, A: 255}
	if pct < 0.3 {
		fillClr = color.RGBA{R: 200, G: 50, B: 50, A: 255}
	} else if pct < 0.6 {
		fillClr = color.RGBA{R: 200, G: 170, B: 50, A: 255}
	}
	vector.DrawFilledRect(screen, x, y, w*pct, h, fillClr, false)
	vector.StrokeRect(screen, x, y, w, h, 1, color.RGBA{R: 90, G: 90, B: 100, A: 255}, false)
}

// drawScaledText draws text at a scaled size using Ebiten's built-in debug font.
// The debug font is 6x16 and doesn't scale, so we render to a temp image and scale it.
func (s *BattleState) drawScaledText(screen *ebiten.Image, text string, x, y, scale float32) {
	s.drawScaledTextColor(screen, text, x, y, scale, color.RGBA{R: 230, G: 230, B: 240, A: 255})
}

func (s *BattleState) drawScaledTextColor(screen *ebiten.Image, text string, x, y, scale float32, _ color.RGBA) {
	if scale <= 1.2 {
		ebitenutil.DebugPrintAt(screen, text, int(x), int(y))
		return
	}

	// Render text to small image, then scale up
	textW := len(text)*6 + 4
	textH := 20
	if textW <= 0 {
		return
	}
	tmp := ebiten.NewImage(textW, textH)
	ebitenutil.DebugPrint(tmp, text)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(tmp, op)
}

// --- Helpers ---

func (s *BattleState) livingFoes() []int {
	out := make([]int, 0, len(s.foes))
	for i, f := range s.foes {
		if f.Stats.IsAlive() {
			out = append(out, i)
		}
	}
	return out
}

func (s *BattleState) allFoesDefeated() bool {
	for _, f := range s.foes {
		if f.Stats.IsAlive() {
			return false
		}
	}
	return true
}

func (s *BattleState) nextLivingFoe(current int) int {
	for i := 1; i <= len(s.foes); i++ {
		idx := (current + i) % len(s.foes)
		if s.foes[idx].Stats.IsAlive() {
			return idx
		}
	}
	return current
}

func (s *BattleState) prevLivingFoe(current int) int {
	for i := 1; i <= len(s.foes); i++ {
		idx := (current - i + len(s.foes)) % len(s.foes)
		if s.foes[idx].Stats.IsAlive() {
			return idx
		}
	}
	return current
}
