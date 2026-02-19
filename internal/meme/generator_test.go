package meme

import (
	"image"
	"image/color"
	"testing"
)

// newTestImage creates a solid-color RGBA image of the given size.
func newTestImage(w, h int, c color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, c)
		}
	}
	return img
}

func TestNewGenerator(t *testing.T) {
	g, err := NewGenerator()
	if err != nil {
		t.Fatalf("NewGenerator() error: %v", err)
	}
	if g == nil {
		t.Fatal("NewGenerator() returned nil generator")
	}
	if g.font == nil {
		t.Fatal("NewGenerator() font is nil")
	}
}

func TestGenerate_ValidInputs(t *testing.T) {
	g, err := NewGenerator()
	if err != nil {
		t.Fatalf("NewGenerator() error: %v", err)
	}

	potato := newTestImage(200, 200, color.RGBA{R: 255, G: 200, B: 100, A: 255})
	cat := newTestImage(640, 480, color.RGBA{R: 100, G: 100, B: 100, A: 255})

	img, err := g.Generate(potato, cat, "top text", "bottom text")
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	if img == nil {
		t.Fatal("Generate() returned nil image")
	}

	bounds := img.Bounds()
	if bounds.Dx() != canvasWidth || bounds.Dy() != canvasHeight {
		t.Errorf("Generate() image size = %dx%d, want %dx%d",
			bounds.Dx(), bounds.Dy(), canvasWidth, canvasHeight)
	}
}

func TestGenerate_NilPotatoImage(t *testing.T) {
	g, err := NewGenerator()
	if err != nil {
		t.Fatalf("NewGenerator() error: %v", err)
	}

	cat := newTestImage(640, 480, color.RGBA{R: 100, G: 100, B: 100, A: 255})

	_, err = g.Generate(nil, cat, "top", "bottom")
	if err == nil {
		t.Fatal("Generate() with nil potato image should return error")
	}
	if err.Error() != "potato image is required" {
		t.Errorf("Generate() error = %q, want %q", err.Error(), "potato image is required")
	}
}

func TestGenerate_NilCatImage(t *testing.T) {
	g, err := NewGenerator()
	if err != nil {
		t.Fatalf("NewGenerator() error: %v", err)
	}

	potato := newTestImage(200, 200, color.RGBA{R: 255, G: 200, B: 100, A: 255})

	_, err = g.Generate(potato, nil, "top", "bottom")
	if err == nil {
		t.Fatal("Generate() with nil cat image should return error")
	}
	if err.Error() != "cat image is required" {
		t.Errorf("Generate() error = %q, want %q", err.Error(), "cat image is required")
	}
}

func TestGenerateRandom(t *testing.T) {
	g, err := NewGenerator()
	if err != nil {
		t.Fatalf("NewGenerator() error: %v", err)
	}

	potato := newTestImage(200, 200, color.RGBA{R: 255, G: 200, B: 100, A: 255})
	cat := newTestImage(640, 480, color.RGBA{R: 100, G: 100, B: 100, A: 255})

	img, err := g.GenerateRandom(potato, cat)
	if err != nil {
		t.Fatalf("GenerateRandom() error: %v", err)
	}
	if img == nil {
		t.Fatal("GenerateRandom() returned nil image")
	}

	bounds := img.Bounds()
	if bounds.Dx() != canvasWidth || bounds.Dy() != canvasHeight {
		t.Errorf("GenerateRandom() image size = %dx%d, want %dx%d",
			bounds.Dx(), bounds.Dy(), canvasWidth, canvasHeight)
	}
}

func TestGenerateRandom_NilInputs(t *testing.T) {
	g, err := NewGenerator()
	if err != nil {
		t.Fatalf("NewGenerator() error: %v", err)
	}

	cat := newTestImage(640, 480, color.RGBA{R: 100, G: 100, B: 100, A: 255})
	potato := newTestImage(200, 200, color.RGBA{R: 255, G: 200, B: 100, A: 255})

	if _, err := g.GenerateRandom(nil, cat); err == nil {
		t.Error("GenerateRandom() with nil potato should return error")
	}
	if _, err := g.GenerateRandom(potato, nil); err == nil {
		t.Error("GenerateRandom() with nil cat should return error")
	}
}

func TestMemeGenerator_ImplementsGenerator(t *testing.T) {
	var _ Generator = (*MemeGenerator)(nil)
}

func TestScaleHeight(t *testing.T) {
	// 400x200 image scaled to width 200 should give height 100.
	img := newTestImage(400, 200, color.Black)
	got := scaleHeight(img, 200)
	if got != 100 {
		t.Errorf("scaleHeight(400x200, 200) = %d, want 100", got)
	}
}
