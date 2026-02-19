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

// PotatoClone represents a smaller potato copy bouncing independently.
type PotatoClone struct {
	X, Y     int     // position
	Scale    float64 // size multiplier (0.15-0.3)
	Rotation float64 // rotation in radians
	BounceY  int     // vertical bounce offset
}

// ComicBurst represents a starburst word that flashes on/off.
type ComicBurst struct {
	X, Y     int
	Text     string
	Rotation float64
	Scale    float64
	Visible  bool // only show on some frames for flashing effect
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

	// Potato clones — smaller copies bouncing independently
	Clones []PotatoClone

	// Divine glow behind main potato
	GlowAlpha  float64 // 0.2-0.6, pulsing
	GlowRadius int     // radius of the glow circle

	// Scrolling news ticker
	TickerX float64 // horizontal position of ticker text (scrolls left)

	// Zoom pulse — entire scene breathes in/out
	ZoomScale float64 // 1.0-1.08, oscillating

	// Rotating hypno wheel
	SpiralAngle float64 // rotation angle in radians, increases each frame

	// Comic book bursts
	Bursts []ComicBurst
}

var burstWords = []string{"SPUD!", "WOW!", "TATER!", "POW!", "NICE!", "EPIC!", "YEET!", "BRUH!", "OMG!", "SPICY!"}

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

	// Potato clones — 3 smaller copies at different positions and phases
	clones := make([]PotatoClone, 3)
	cloneRNG := rand.New(rand.NewPCG(uint64(frame*3571+1), uint64(frame*2903+1)))
	clonePositions := [][2]int{
		{60, 80},                       // upper-left area
		{canvasW - 180, 60},            // upper-right area
		{canvasW/2 - 50, canvasH - 90}, // bottom-center area
	}
	for i := range clones {
		phase := float64(i) * 2.0 * math.Pi / 3.0 // 120° phase offset between clones
		scale := 0.15 + float64(i)*0.05           // 0.15, 0.20, 0.25
		bounceFreq := 1.5 + float64(i)*0.7        // different bounce frequencies
		cloneBounce := -int(math.Abs(math.Sin(2*math.Pi*t*bounceFreq+phase)) * 25.0)
		cloneRotation := 0.25 * math.Sin(2*math.Pi*t*1.3+phase)
		_ = cloneRNG // seeded but positions are deterministic from clonePositions
		clones[i] = PotatoClone{
			X:        clonePositions[i][0],
			Y:        clonePositions[i][1],
			Scale:    scale,
			Rotation: cloneRotation,
			BounceY:  cloneBounce,
		}
	}

	// Divine glow — pulsing alpha behind main potato
	glowAlpha := 0.4 + 0.2*math.Sin(2*math.Pi*t*1.5)
	glowRadius := 120 + int(20*math.Sin(2*math.Pi*t*1.5))

	// Ticker — scrolls from right to left across frames
	// TickerX starts at canvasWidth and decreases. The speed is ~40px per frame.
	tickerSpeed := 40.0
	tickerX := float64(canvasW) - float64(frame)*tickerSpeed

	// Zoom pulse — oscillates between 1.0 and 1.08
	zoomScale := 1.0 + 0.04*(1.0+math.Sin(2*math.Pi*t*0.8))

	// Hypno wheel — rotation angle increases each frame
	spiralAngle := 2.0 * math.Pi * t * 0.5 // half rotation per loop

	// Comic bursts — 2 bursts per frame, flashing on alternating frames
	burstRNG := rand.New(rand.NewPCG(uint64(frame*4219+7), uint64(frame*3137+13)))
	bursts := make([]ComicBurst, 2)
	for i := range bursts {
		bursts[i] = ComicBurst{
			X:        80 + burstRNG.IntN(canvasW-160),
			Y:        80 + burstRNG.IntN(canvasH-200),
			Text:     burstWords[burstRNG.IntN(len(burstWords))],
			Rotation: (burstRNG.Float64() - 0.5) * 0.5, // ±0.25 radians
			Scale:    0.7 + burstRNG.Float64()*0.6,     // 0.7-1.3
			Visible:  (frame+i)%3 != 0,                 // visible 2 out of 3 frames
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
		Clones:         clones,
		GlowAlpha:      glowAlpha,
		GlowRadius:     glowRadius,
		TickerX:        tickerX,
		ZoomScale:      zoomScale,
		SpiralAngle:    spiralAngle,
		Bursts:         bursts,
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
