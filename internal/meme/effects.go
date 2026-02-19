package meme

import (
	"image/color"
	"math"
	"math/rand/v2"
)

const (
	TotalFrames = 16
	FrameDelay  = 8 // centiseconds (80ms per frame ≈ 12.5fps, ~1.9s loop)
)

// Sparkle represents a single sparkle overlay.
type Sparkle struct {
	X, Y  int
	Size  int     // radius in pixels
	Alpha float64 // 0.0 to 1.0
}

// FrameParams holds all computed animation values for a single frame.
type FrameParams struct {
	// Text color cycling — rainbow color for text fill
	TextColor color.Color
	// Text pulse — font size multiplier (oscillates around 1.0)
	FontScale float64
	// Potato bounce — Y offset for potato position (negative = up)
	PotatoBounceY int
	// Potato rotation angle in radians (gentle wobble)
	PotatoRotation float64
	// Screen shake — small X,Y offset for the whole scene
	ShakeDX, ShakeDY int
	// Sparkle overlays
	Sparkles []Sparkle
}

// ComputeFrameParams calculates animation parameters for a given frame.
func ComputeFrameParams(frame, totalFrames, canvasW, canvasH int) FrameParams {
	t := float64(frame) / float64(totalFrames) // 0.0 to ~1.0

	// Rainbow text color — cycle hue through 360°
	hue := t * 360.0
	textColor := hslToRGB(hue, 1.0, 0.55) // fully saturated, slightly bright

	// Text pulse — oscillate font size ±15%
	fontScale := 1.0 + 0.15*math.Sin(2*math.Pi*t)

	// Potato bounce — absolute sine wave bounce
	bounceHeight := 40.0
	potatoBounceY := -int(math.Abs(math.Sin(2*math.Pi*t)) * bounceHeight)

	// Potato wobble — gentle rotation ±10°
	potatoRotation := 0.17 * math.Sin(2*math.Pi*t) // ~10° in radians

	// Screen shake — random ±3px jitter (deterministic per frame)
	rng := rand.New(rand.NewPCG(uint64(frame*7919), uint64(frame*6271)))
	shakeDX := rng.IntN(7) - 3
	shakeDY := rng.IntN(7) - 3

	// Sparkles — 6-8 random sparkles per frame
	numSparkles := 6 + rng.IntN(3)
	sparkles := make([]Sparkle, numSparkles)
	for i := range sparkles {
		sparkles[i] = Sparkle{
			X:     rng.IntN(canvasW),
			Y:     rng.IntN(canvasH),
			Size:  4 + rng.IntN(8),         // 4-11px radius
			Alpha: 0.5 + rng.Float64()*0.5, // 0.5-1.0
		}
	}

	return FrameParams{
		TextColor:      textColor,
		FontScale:      fontScale,
		PotatoBounceY:  potatoBounceY,
		PotatoRotation: potatoRotation,
		ShakeDX:        shakeDX,
		ShakeDY:        shakeDY,
		Sparkles:       sparkles,
	}
}

// hslToRGB converts HSL (hue 0-360, saturation 0-1, lightness 0-1) to an RGB color.
func hslToRGB(h, s, l float64) color.RGBA {
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}

	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2

	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return color.RGBA{
		R: uint8((r + m) * 255),
		G: uint8((g + m) * 255),
		B: uint8((b + m) * 255),
		A: 255,
	}
}
