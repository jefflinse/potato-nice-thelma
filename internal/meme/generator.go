package meme

import (
	_ "embed"
	"errors"
	"fmt"
	"image"
	"image/color"
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
	canvasWidth  = 800
	canvasHeight = 600
	fontSize     = 60
	outlineShift = 2
	topMargin    = 50
	bottomMargin = 550
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
	Generate(potatoImg, catImg image.Image, topText, bottomText string) (image.Image, error)
	GenerateRandom(potatoImg, catImg image.Image) (image.Image, error)
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
// lower-right area, and renders topText/bottomText in classic meme style.
func (g *MemeGenerator) Generate(potatoImg, catImg image.Image, topText, bottomText string) (image.Image, error) {
	if potatoImg == nil {
		return nil, errors.New("potato image is required")
	}
	if catImg == nil {
		return nil, errors.New("cat image is required")
	}

	dc := gg.NewContext(canvasWidth, canvasHeight)

	// Draw cat image scaled to fill the entire canvas.
	drawScaled(dc, catImg, canvasWidth, canvasHeight, 0, 0)

	// Draw potato image at ~40% canvas width, positioned in the lower-right.
	potatoW := int(float64(canvasWidth) * potatoScale)
	potatoH := scaleHeight(potatoImg, potatoW)
	drawScaled(dc, potatoImg, potatoW, potatoH, canvasWidth-potatoW-20, canvasHeight-potatoH-60)

	// Render meme text.
	face := truetype.NewFace(g.font, &truetype.Options{Size: fontSize})
	drawMemeText(dc, face, strings.ToUpper(topText), canvasWidth/2, topMargin)
	drawMemeText(dc, face, strings.ToUpper(bottomText), canvasWidth/2, bottomMargin)

	return dc.Image(), nil
}

// GenerateRandom picks a random predefined text pair and calls Generate.
func (g *MemeGenerator) GenerateRandom(potatoImg, catImg image.Image) (image.Image, error) {
	pair := memeTexts[rand.IntN(len(memeTexts))]
	return g.Generate(potatoImg, catImg, pair.Top, pair.Bottom)
}

// drawScaled renders src into dc at the given position and dimensions using
// bilinear interpolation.
func drawScaled(dc *gg.Context, src image.Image, w, h, x, y int) {
	scaled := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.BiLinear.Scale(scaled, scaled.Bounds(), src, src.Bounds(), draw.Over, nil)
	dc.DrawImage(scaled, x, y)
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

// drawMemeText renders text with a black outline and white fill, centered at
// (cx, cy). The outline is produced by drawing the text 8 times at small
// offsets in each cardinal and diagonal direction.
func drawMemeText(dc *gg.Context, face font.Face, text string, cx, cy float64) {
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

	// Fill: draw in white on top.
	dc.SetColor(color.White)
	dc.DrawStringAnchored(text, cx, cy, 0.5, 0.5)
}
