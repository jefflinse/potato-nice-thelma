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
// across multiple frames to produce an animated GIF.
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

	anim := &gif.GIF{
		LoopCount: 0, // infinite loop
	}

	for i := range TotalFrames {
		params := ComputeFrameParams(i, TotalFrames, canvasWidth, canvasHeight)

		dc := gg.NewContext(canvasWidth, canvasHeight)

		// Draw cat background with screen shake offset.
		dc.DrawImage(scaledCat, params.ShakeDX, params.ShakeDY)

		// Draw potato with bounce and rotation.
		potatoDrawX := potatoBaseX
		potatoDrawY := potatoBaseY + params.PotatoBounceY
		potatoCenterX := float64(potatoDrawX) + float64(potatoW)/2
		potatoCenterY := float64(potatoDrawY) + float64(potatoH)/2

		dc.Push()
		dc.RotateAbout(params.PotatoRotation, potatoCenterX, potatoCenterY)
		dc.DrawImage(scaledPotato, potatoDrawX, potatoDrawY)
		dc.Pop()

		// Render meme text with animated color and size.
		scaledFontSize := fontSize * params.FontScale
		face := truetype.NewFace(g.font, &truetype.Options{Size: scaledFontSize})
		drawMemeText(dc, face, topTextUpper, canvasWidth/2, topMargin, params.TextColor)
		drawMemeText(dc, face, bottomTextUpper, canvasWidth/2, bottomMargin, params.TextColor)

		// Draw sparkles.
		for _, sp := range params.Sparkles {
			drawSparkle(dc, sp.X, sp.Y, sp.Size, sp.Alpha)
		}

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
