package render

import (
	_ "embed"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed postprocess.kage
var postProcessSrc []byte

// PostProcess applies bloom and vignette to the final frame.
type PostProcess struct {
	shader     *ebiten.Shader
	offscreen  *ebiten.Image
	lastW      int
	lastH      int
	compileErr bool

	// Tunable parameters
	VignetteStrength float32
	BloomThreshold   float32
	BloomIntensity   float32
}

func NewPostProcess() *PostProcess {
	return &PostProcess{
		VignetteStrength: 0.45,
		BloomThreshold:   0.8,
		BloomIntensity:   0.12,
	}
}

func (p *PostProcess) ensureShader() *ebiten.Shader {
	if p.shader != nil || p.compileErr {
		return p.shader
	}
	s, err := ebiten.NewShader(postProcessSrc)
	if err != nil {
		p.compileErr = true
		return nil
	}
	p.shader = s
	return s
}

func (p *PostProcess) ensureOffscreen(w, h int) *ebiten.Image {
	if p.offscreen != nil && p.lastW == w && p.lastH == h {
		return p.offscreen
	}
	p.offscreen = ebiten.NewImage(w, h)
	p.lastW = w
	p.lastH = h
	return p.offscreen
}

// Apply renders the scene from src onto dst with post-processing.
// If the shader fails to compile, it falls back to a plain copy.
func (p *PostProcess) Apply(dst, src *ebiten.Image) {
	shader := p.ensureShader()
	if shader == nil {
		// Fallback: plain copy
		dst.DrawImage(src, nil)
		return
	}

	w := dst.Bounds().Dx()
	h := dst.Bounds().Dy()

	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = map[string]any{
		"ScreenSize":       []float32{float32(w), float32(h)},
		"VignetteStrength": p.VignetteStrength,
		"BloomThreshold":   p.BloomThreshold,
		"BloomIntensity":   p.BloomIntensity,
	}
	op.Images[0] = src
	dst.DrawRectShader(w, h, shader, op)
}
