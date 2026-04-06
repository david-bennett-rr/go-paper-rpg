package render

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/tetra3d"

	"github.com/davidbennett/go-paper-rpg/internal/util"
)

const (
	// Three-quarter RPG camera with a slight perspective and diagonal yaw.
	camDist        = 25.0
	camAngleY      = math.Pi / 4.0
	camAngleX      = math.Pi / 5.0
	camFieldOfView = 32.0
)

// ScreenToWorldMove converts screen-relative movement input into world-space ground movement
// for the current fixed camera yaw. screenY should be negative when moving up the screen.
func ScreenToWorldMove(screenX, screenY float64) (worldX, worldZ float64) {
	sinY := math.Sin(camAngleY)
	cosY := math.Cos(camAngleY)
	worldX = cosY*screenX + sinY*screenY
	worldZ = -sinY*screenX + cosY*screenY
	return worldX, worldZ
}

// Renderer manages Tetra3D scene rendering and sprite billboarding.
type Renderer struct {
	Camera      *tetra3d.Camera
	Scene       *tetra3d.Scene
	PostProcess *PostProcess
	logged      bool
}

func NewRenderer() *Renderer {
	// Start with a reasonable default; Resize() will update to match the window.
	camera := tetra3d.NewCamera(1920, 1080)
	camera.SetPerspective(true)
	camera.SetFieldOfView(camFieldOfView)
	camera.SetFar(200)
	camera.SetNear(0.1)

	util.DebugLog("Camera created: perspective, fov=%.1f", camFieldOfView)

	return &Renderer{
		Camera:      camera,
		PostProcess: NewPostProcess(),
	}
}

// Resize updates the camera render resolution. Call from Layout().
func (r *Renderer) Resize(w, h int) {
	r.Camera.Resize(w, h)
}

// SetScene sets the active 3D scene and adds the camera to it.
func (r *Renderer) SetScene(scene *tetra3d.Scene) {
	if r.Scene != nil {
		r.Camera.Unparent()
	}
	r.Scene = scene
	if scene != nil {
		scene.Root.AddChildren(r.Camera)
		r.orientCamera(0, 0, 0)
	}
}

// orientCamera positions and rotates the camera to look at a target from an isometric angle.
func (r *Renderer) orientCamera(tx, ty, tz float64) {
	// Spherical coordinates to cartesian offset.
	cx := tx + camDist*math.Sin(camAngleY)*math.Cos(camAngleX)
	cy := ty + camDist*math.Sin(camAngleX)
	cz := tz + camDist*math.Cos(camAngleY)*math.Cos(camAngleX)

	r.Camera.SetWorldPositionVec(tetra3d.NewVector3(float32(cx), float32(cy), float32(cz)))

	// Order matters here: pitch is applied relative to the yawed camera.
	rot := tetra3d.NewMatrix4Rotate(0, 1, 0, float32(camAngleY)).
		Rotated(1, 0, 0, float32(-camAngleX))
	r.Camera.SetLocalRotation(rot)

	if !r.logged {
		util.DebugLog("Camera oriented: pos=(%.1f, %.1f, %.1f) target=(%.1f, %.1f, %.1f)",
			cx, cy, cz, tx, ty, tz)
	}
}

// DrawSceneWithSprites renders the scene, billboard sprites, then composites to screen.
func (r *Renderer) DrawSceneWithSprites(screen *ebiten.Image, sprites []SpriteInstance) {
	r.Camera.Clear()

	if r.Scene != nil {
		r.Camera.RenderScene(r.Scene)
	}

	colorTex := r.Camera.ColorTexture()

	for _, s := range sprites {
		r.Camera.RenderSprite3D(
			colorTex,
			tetra3d.DrawSprite3dSettings{
				Image:         s.Image,
				WorldPosition: tetra3d.NewVector3(float32(s.X), float32(s.Y), float32(s.Z)),
			},
		)
	}

	// Apply post-processing (bloom + vignette)
	if r.PostProcess != nil {
		r.PostProcess.Apply(screen, colorTex)
	} else {
		screen.DrawImage(colorTex, nil)
	}

	if !r.logged {
		r.logged = true
		util.DebugLog("First frame rendered. Sprites: %d", len(sprites))
	}
}

// DrawScene renders the scene without any billboard sprites.
func (r *Renderer) DrawScene(screen *ebiten.Image) {
	r.DrawSceneWithSprites(screen, nil)
}

// SpriteInstance represents a sprite to be drawn at a 3D position.
type SpriteInstance struct {
	Image   *ebiten.Image
	X, Y, Z float64
}

// SetCameraFollow moves the camera to center on a target position.
func (r *Renderer) SetCameraFollow(targetX, targetY, targetZ float64) {
	r.orientCamera(targetX, targetY, targetZ)
}

// RenderBattleScene renders a standalone scene with a front-facing camera.
// The camera sits at eye level, looking straight at the origin.
func (r *Renderer) RenderBattleScene(screen *ebiten.Image, scene *tetra3d.Scene) {
	if scene == nil {
		return
	}

	// Temporarily attach camera to the battle scene
	prevScene := r.Scene
	r.Scene = scene
	scene.Root.AddChildren(r.Camera)

	// Position camera to view the battle stage from the front
	r.Camera.SetWorldPositionVec(tetra3d.NewVector3(0, 1.8, 8))
	rot := tetra3d.NewMatrix4Rotate(0, 1, 0, 0).
		Rotated(1, 0, 0, -0.12)
	r.Camera.SetLocalRotation(rot)

	r.Camera.Clear()
	r.Camera.RenderScene(scene)

	colorTex := r.Camera.ColorTexture()
	if r.PostProcess != nil {
		r.PostProcess.Apply(screen, colorTex)
	} else {
		screen.DrawImage(colorTex, nil)
	}

	// Restore
	r.Camera.Unparent()
	r.Scene = prevScene
	if prevScene != nil {
		prevScene.Root.AddChildren(r.Camera)
	}
}
