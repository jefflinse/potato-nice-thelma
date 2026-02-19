package meme

import (
	_ "embed"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	stddraw "image/draw"
	"image/gif"
	"math"
	"math/rand/v2"
	"strings"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
)

//go:embed Anton-Regular.ttf
var fontBytes []byte

const (
	canvasWidth  = 640
	canvasHeight = 480
	fontSize     = 48
	outlineShift = 2
	topMargin    = 40
	bottomMargin = 440
	potatoScale  = 0.4 // 40% of canvas width
)

// memeTexts holds predefined text pairs for random meme generation.
var memeTexts = []struct{ Top, Bottom string }{
	{"when u a potato", "but also a cat person"},
	{"i can haz", "potato?"},
	{"tater tot", "reporting for duty"},
	{"am not cat", "am potato"},
	{"potato cat", "the hero we deserve"},
	{"one does not simply", "combine potatoes and cats"},
	{"they told me i could be anything", "so i became a potato cat"},
	{"this is fine", "everything is potato"},
	{"cat.exe has stopped working", "potato.dll loaded instead"},
	{"when the catnip hits", "and you become a potato"},
	{"i showed you my potato", "please respond"},
	{"therapist: potato cat isn't real", "potato cat:"},
	{"nobody:", "potato cat at 3am:"},
	{"me: i'm a normal person", "also me:"},
	{"the last thing you see", "before you get mashed"},
	{"you've been visited by", "the sacred potato cat"},
	{"this image is cursed", "you're welcome"},
	{"mom can we have a cat", "we have a cat at home. the cat at home:"},
	{"roses are red", "potatoes are brown. this meme is cursed. please sit down"},
	{"i have achieved", "peak internet"},
	{"delete this", "nephew"},
	{"what in tarnation", "is this abomination"},
	{"thanks i hate it", "potato cat forever"},
	{"it's not a phase mom", "i'm a potato cat now"},
	{"the prophecy is true", "the potato cat has risen"},
}

// tickerMessages are scrolling news ticker texts.
var tickerMessages = []string{
	"BREAKING: LOCAL POTATO ACHIEVES SENTIENCE, DEMANDS BELLY RUBS",
	"ALERT: SCIENTISTS CONFIRM CATS ARE 47% POTATO ON A MOLECULAR LEVEL",
	"DEVELOPING: POTATO-CAT HYBRID ESCAPES LAB, LAST SEEN HEADING TOWARD COUCH",
	"URGENT: WORLD POTATO SUPPLY NOW CONTROLLED BY CATS",
	"LIVE: POTATO ELECTED MAYOR OF INTERNET, CATS DEMAND RECOUNT",
	"THIS JUST IN: YOUR SCREEN IS NOW 100% MORE POTATO THAN BEFORE",
	"EXCLUSIVE: AREA CAT REFUSES TO ACKNOWLEDGE POTATO ROOMMATE",
}

// Generator composites a potato image and a cat image with meme text.
type Generator interface {
	Generate(potatoImg, catImg image.Image, topText, bottomText string) (*gif.GIF, error)
	GenerateRandom(potatoImg, catImg image.Image) (*gif.GIF, error)
}

// MemeGenerator implements Generator using the fogleman/gg drawing library.
type MemeGenerator struct {
	font *truetype.Font
}

// NewGenerator creates a MemeGenerator with the embedded Anton font.
func NewGenerator() (*MemeGenerator, error) {
	f, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, fmt.Errorf("parsing embedded font: %w", err)
	}
	return &MemeGenerator{font: f}, nil
}

// Generate composites catImg as the background, overlays potatoImg in the
// lower-right area, and renders topText/bottomText in classic meme style
// across multiple frames to produce an animated GIF with maximum chaos effects.
func (g *MemeGenerator) Generate(potatoImg, catImg image.Image, topText, bottomText string) (*gif.GIF, error) {
	if potatoImg == nil {
		return nil, errors.New("potato image is required")
	}
	if catImg == nil {
		return nil, errors.New("cat image is required")
	}

	// Pre-scale images once before the frame loop.
	scaledCat := scaleImage(catImg, canvasWidth, canvasHeight)

	potatoW := int(float64(canvasWidth) * potatoScale)
	potatoH := scaleHeight(potatoImg, potatoW)
	scaledPotato := scaleImage(potatoImg, potatoW, potatoH)

	// Base position for the potato (lower-right).
	potatoBaseX := canvasWidth - potatoW - 20
	potatoBaseY := canvasHeight - potatoH - 60

	topTextUpper := strings.ToUpper(topText)
	bottomTextUpper := strings.ToUpper(bottomText)

	// Pick a ticker message once for the entire animation.
	tickerMsg := tickerMessages[rand.IntN(len(tickerMessages))]

	anim := &gif.GIF{
		LoopCount: 0, // infinite loop
	}

	for i := range TotalFrames {
		params := ComputeFrameParams(i, TotalFrames, canvasWidth, canvasHeight)

		dc := gg.NewContext(canvasWidth, canvasHeight)

		// 1. Draw cat background with zoom scale and screen shake.
		drawZoomedBackground(dc, scaledCat, params.ZoomScale, params.ShakeDX, params.ShakeDY)

		// 2. Hypno wheel overlay (low alpha, rotating).
		drawHypnoWheel(dc, float64(canvasWidth)/2, float64(canvasHeight)/2,
			float64(canvasWidth)*0.8, params.SpiralAngle, 0.08)

		// 3. Divine glow behind main potato.
		potatoDrawX := potatoBaseX
		potatoDrawY := potatoBaseY + params.PotatoBounceY
		potatoCenterX := float64(potatoDrawX) + float64(potatoW)/2
		potatoCenterY := float64(potatoDrawY) + float64(potatoH)/2
		drawDivineGlow(dc, potatoCenterX, potatoCenterY, params.GlowRadius, params.GlowAlpha)

		// 4. Main potato with bounce and rotation.
		dc.Push()
		dc.RotateAbout(params.PotatoRotation, potatoCenterX, potatoCenterY)
		dc.DrawImage(scaledPotato, potatoDrawX, potatoDrawY)
		dc.Pop()

		// 5. Potato clones — smaller copies bouncing independently.
		drawPotatoClones(dc, potatoImg, params.Clones)

		// 6. Comic bursts — starburst shapes with text, flashing.
		drawComicBursts(dc, g.font, params.Bursts)

		// 7. Sparkles.
		for _, sp := range params.Sparkles {
			drawSparkle(dc, sp.X, sp.Y, sp.Size, sp.Alpha)
		}

		// 8. Meme text with animated color and size.
		scaledFontSize := fontSize * params.FontScale
		face := truetype.NewFace(g.font, &truetype.Options{Size: scaledFontSize})
		drawMemeText(dc, face, topTextUpper, canvasWidth/2, topMargin, params.TextColor)
		drawMemeText(dc, face, bottomTextUpper, canvasWidth/2, bottomMargin, params.TextColor)

		// 9. News ticker banner + scrolling text.
		drawTicker(dc, g.font, tickerMsg, params.TickerX)

		// Convert frame to paletted image.
		rgbaFrame, ok := dc.Image().(*image.RGBA)
		if !ok {
			// Fallback: copy into RGBA.
			b := dc.Image().Bounds()
			rgbaFrame = image.NewRGBA(b)
			stddraw.Draw(rgbaFrame, b, dc.Image(), b.Min, stddraw.Src)
		}

		palettedImg := image.NewPaletted(image.Rect(0, 0, canvasWidth, canvasHeight), palette.Plan9)
		stddraw.FloydSteinberg.Draw(palettedImg, palettedImg.Bounds(), rgbaFrame, image.Point{})

		anim.Image = append(anim.Image, palettedImg)
		anim.Delay = append(anim.Delay, FrameDelay)
	}

	return anim, nil
}

// GenerateRandom picks a random predefined text pair and calls Generate.
func (g *MemeGenerator) GenerateRandom(potatoImg, catImg image.Image) (*gif.GIF, error) {
	pair := memeTexts[rand.IntN(len(memeTexts))]
	return g.Generate(potatoImg, catImg, pair.Top, pair.Bottom)
}

// drawZoomedBackground draws the cat background with a zoom scale applied,
// centered on the canvas, plus screen shake offset.
func drawZoomedBackground(dc *gg.Context, catImg *image.RGBA, zoomScale float64, shakeDX, shakeDY int) {
	zoomedW := int(float64(canvasWidth) * zoomScale)
	zoomedH := int(float64(canvasHeight) * zoomScale)

	if zoomScale > 1.001 {
		zoomed := image.NewRGBA(image.Rect(0, 0, zoomedW, zoomedH))
		draw.BiLinear.Scale(zoomed, zoomed.Bounds(), catImg, catImg.Bounds(), draw.Over, nil)
		// Center the zoomed image so the zoom appears to emanate from center.
		offsetX := -(zoomedW-canvasWidth)/2 + shakeDX
		offsetY := -(zoomedH-canvasHeight)/2 + shakeDY
		dc.DrawImage(zoomed, offsetX, offsetY)
	} else {
		dc.DrawImage(catImg, shakeDX, shakeDY)
	}
}

// drawHypnoWheel draws alternating semi-transparent pie slices radiating from
// center, creating a rotating hypnotic starburst effect.
func drawHypnoWheel(dc *gg.Context, cx, cy, radius, angle, alpha float64) {
	numSlices := 8
	sliceAngle := 2 * math.Pi / float64(numSlices)
	for i := 0; i < numSlices; i++ {
		if i%2 == 0 {
			continue // skip every other slice for alternating pattern
		}
		startAngle := angle + float64(i)*sliceAngle
		endAngle := startAngle + sliceAngle
		dc.MoveTo(cx, cy)
		dc.DrawArc(cx, cy, radius, startAngle, endAngle)
		dc.LineTo(cx, cy)
		dc.ClosePath()
		dc.SetRGBA(1, 1, 1, alpha)
		dc.Fill()
	}
}

// drawDivineGlow draws a pulsing radial gradient glow (holy aura) behind the
// main potato using concentric semi-transparent circles.
func drawDivineGlow(dc *gg.Context, cx, cy float64, radius int, alpha float64) {
	rings := 12
	for i := rings; i >= 0; i-- {
		r := float64(radius) * float64(i) / float64(rings)
		// Alpha fades out from center; inner rings are brighter.
		ringAlpha := alpha * (1.0 - float64(i)/float64(rings)) * 0.8
		if ringAlpha < 0.01 {
			continue
		}
		// Gold/yellow color with fading alpha.
		dc.SetRGBA(1.0, 0.85, 0.2, ringAlpha)
		dc.DrawCircle(cx, cy, r)
		dc.Fill()
	}
}

// drawPotatoClones draws smaller potato copies at their computed positions.
func drawPotatoClones(dc *gg.Context, potatoImg image.Image, clones []PotatoClone) {
	for _, clone := range clones {
		cloneW := int(float64(canvasWidth) * clone.Scale)
		cloneH := scaleHeight(potatoImg, cloneW)
		if cloneW < 1 || cloneH < 1 {
			continue
		}
		scaledClone := scaleImage(potatoImg, cloneW, cloneH)

		drawX := clone.X
		drawY := clone.Y + clone.BounceY
		centerX := float64(drawX) + float64(cloneW)/2
		centerY := float64(drawY) + float64(cloneH)/2

		dc.Push()
		dc.RotateAbout(clone.Rotation, centerX, centerY)
		dc.DrawImage(scaledClone, drawX, drawY)
		dc.Pop()
	}
}

// drawComicBursts draws starburst shapes with comic text that flash on/off.
func drawComicBursts(dc *gg.Context, f *truetype.Font, bursts []ComicBurst) {
	for _, burst := range bursts {
		if !burst.Visible {
			continue
		}

		cx := float64(burst.X)
		cy := float64(burst.Y)
		outerR := 45.0 * burst.Scale
		innerR := 22.0 * burst.Scale
		points := 10

		dc.Push()
		dc.RotateAbout(burst.Rotation, cx, cy)

		// Draw yellow starburst shape.
		drawStarburst(dc, cx, cy, outerR, innerR, points, 0)
		dc.SetRGBA(1.0, 0.95, 0.0, 0.9) // bright yellow
		dc.Fill()

		// Draw starburst outline.
		drawStarburst(dc, cx, cy, outerR, innerR, points, 0)
		dc.SetRGBA(0.8, 0.0, 0.0, 1.0) // red outline
		dc.SetLineWidth(2)
		dc.Stroke()

		// Draw burst text.
		burstFontSize := 16.0 * burst.Scale
		face := truetype.NewFace(f, &truetype.Options{Size: burstFontSize})
		dc.SetFontFace(face)
		dc.SetRGBA(0.8, 0.0, 0.0, 1.0) // red text
		dc.DrawStringAnchored(burst.Text, cx, cy, 0.5, 0.5)

		dc.Pop()
	}
}

// drawStarburst traces a starburst polygon path with alternating long/short radii.
func drawStarburst(dc *gg.Context, cx, cy, outerR, innerR float64, points int, rotation float64) {
	for i := 0; i < points*2; i++ {
		angle := rotation + float64(i)*math.Pi/float64(points)
		r := outerR
		if i%2 == 1 {
			r = innerR
		}
		x := cx + r*math.Cos(angle)
		y := cy + r*math.Sin(angle)
		if i == 0 {
			dc.MoveTo(x, y)
		} else {
			dc.LineTo(x, y)
		}
	}
	dc.ClosePath()
}

// drawTicker draws a semi-transparent banner at the bottom with scrolling text.
func drawTicker(dc *gg.Context, f *truetype.Font, message string, tickerX float64) {
	bannerHeight := 30.0
	bannerY := float64(canvasHeight) - bannerHeight

	// Semi-transparent dark banner.
	dc.SetRGBA(0, 0, 0, 0.7)
	dc.DrawRectangle(0, bannerY, float64(canvasWidth), bannerHeight)
	dc.Fill()

	// Red accent line at top of banner.
	dc.SetRGBA(0.8, 0.0, 0.0, 0.9)
	dc.DrawRectangle(0, bannerY, float64(canvasWidth), 2)
	dc.Fill()

	// Ticker text in white.
	tickerFace := truetype.NewFace(f, &truetype.Options{Size: 18})
	dc.SetFontFace(tickerFace)
	dc.SetColor(color.White)

	textY := bannerY + bannerHeight/2

	// Draw the message; if it scrolls off the left, wrap it around.
	// Measure text width to know when to wrap.
	w, _ := dc.MeasureString(message)
	// Draw primary text.
	dc.DrawStringAnchored(message, tickerX, textY, 0, 0.5)
	// Draw wrapped copy so it seamlessly loops.
	dc.DrawStringAnchored(message, tickerX+w+100, textY, 0, 0.5)
}

// scaleImage renders src scaled to the given dimensions using bilinear interpolation.
func scaleImage(src image.Image, w, h int) *image.RGBA {
	scaled := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.BiLinear.Scale(scaled, scaled.Bounds(), src, src.Bounds(), draw.Over, nil)
	return scaled
}

// scaleHeight returns the height that preserves src's aspect ratio at the
// given width.
func scaleHeight(src image.Image, targetW int) int {
	b := src.Bounds()
	srcW, srcH := b.Dx(), b.Dy()
	if srcW == 0 {
		return targetW
	}
	return targetW * srcH / srcW
}

// drawMemeText renders text with a black outline and a colored fill, centered at
// (cx, cy). The outline is produced by drawing the text 8 times at small
// offsets in each cardinal and diagonal direction.
func drawMemeText(dc *gg.Context, face font.Face, text string, cx, cy float64, fillColor color.Color) {
	dc.SetFontFace(face)

	// Outline: draw in black at 8 surrounding offsets.
	dc.SetColor(color.Black)
	for dx := -outlineShift; dx <= outlineShift; dx += outlineShift {
		for dy := -outlineShift; dy <= outlineShift; dy += outlineShift {
			if dx == 0 && dy == 0 {
				continue
			}
			dc.DrawStringAnchored(text, cx+float64(dx), cy+float64(dy), 0.5, 0.5)
		}
	}

	// Fill: draw in the provided color on top.
	dc.SetColor(fillColor)
	dc.DrawStringAnchored(text, cx, cy, 0.5, 0.5)
}

// drawSparkle renders a 4-pointed star shape at the given position.
func drawSparkle(dc *gg.Context, x, y, size int, alpha float64) {
	c := color.RGBA{R: 255, G: 255, B: 200, A: uint8(alpha * 255)} // warm yellow-white
	dc.SetColor(c)
	dc.SetLineWidth(2)
	// Vertical line
	dc.DrawLine(float64(x), float64(y-size), float64(x), float64(y+size))
	dc.Stroke()
	// Horizontal line
	dc.DrawLine(float64(x-size), float64(y), float64(x+size), float64(y))
	dc.Stroke()
	// Diagonal lines (smaller)
	half := float64(size) * 0.6
	dc.DrawLine(float64(x)-half, float64(y)-half, float64(x)+half, float64(y)+half)
	dc.Stroke()
	dc.DrawLine(float64(x)+half, float64(y)-half, float64(x)-half, float64(y)+half)
	dc.Stroke()
}
