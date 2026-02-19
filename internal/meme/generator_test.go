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

	result, err := g.Generate(potato, cat, "top text", "bottom text")
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	if result == nil {
		t.Fatal("Generate() returned nil GIF")
	}

	if len(result.Image) != TotalFrames {
		t.Errorf("Generate() frame count = %d, want %d", len(result.Image), TotalFrames)
	}

	if len(result.Delay) != TotalFrames {
		t.Errorf("Generate() delay count = %d, want %d", len(result.Delay), TotalFrames)
	}

	for i, frame := range result.Image {
		bounds := frame.Bounds()
		if bounds.Dx() != canvasWidth || bounds.Dy() != canvasHeight {
			t.Errorf("Generate() frame %d size = %dx%d, want %dx%d",
				i, bounds.Dx(), bounds.Dy(), canvasWidth, canvasHeight)
		}
	}

	for i, d := range result.Delay {
		if d != FrameDelay {
			t.Errorf("Generate() frame %d delay = %d, want %d", i, d, FrameDelay)
		}
	}

	if result.LoopCount != 0 {
		t.Errorf("Generate() LoopCount = %d, want 0 (infinite)", result.LoopCount)
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

	result, err := g.GenerateRandom(potato, cat)
	if err != nil {
		t.Fatalf("GenerateRandom() error: %v", err)
	}
	if result == nil {
		t.Fatal("GenerateRandom() returned nil GIF")
	}

	if len(result.Image) != TotalFrames {
		t.Errorf("GenerateRandom() frame count = %d, want %d", len(result.Image), TotalFrames)
	}

	for i, frame := range result.Image {
		bounds := frame.Bounds()
		if bounds.Dx() != canvasWidth || bounds.Dy() != canvasHeight {
			t.Errorf("GenerateRandom() frame %d size = %dx%d, want %dx%d",
				i, bounds.Dx(), bounds.Dy(), canvasWidth, canvasHeight)
		}
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
