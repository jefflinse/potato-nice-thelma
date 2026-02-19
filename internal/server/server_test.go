package server

import (
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/color/palette"
	"image/gif"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type mockSearcher struct {
	url string
	err error
}

func (m *mockSearcher) SearchRandom(_ context.Context, _ string) (string, error) {
	return m.url, m.err
}

type mockFetcher struct {
	img image.Image
	err error
}

func (m *mockFetcher) FetchRandomCat(_ context.Context) (image.Image, error) {
	return m.img, m.err
}

type mockGenerator struct {
	gif            *gif.GIF
	err            error
	generateCalled bool
	randomCalled   bool
}

func (m *mockGenerator) Generate(_, _ image.Image, _, _ string) (*gif.GIF, error) {
	m.generateCalled = true
	return m.gif, m.err
}

func (m *mockGenerator) GenerateRandom(_, _ image.Image) (*gif.GIF, error) {
	m.randomCalled = true
	return m.gif, m.err
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// testImage creates a minimal 1x1 RGBA image for use in tests.
func testImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	return img
}

// testGIF creates a minimal 1-frame GIF for use in tests.
func testGIF() *gif.GIF {
	frame := image.NewPaletted(image.Rect(0, 0, 1, 1), palette.Plan9)
	return &gif.GIF{
		Image: []*image.Paletted{frame},
		Delay: []int{8},
	}
}

// pngServer returns an httptest.Server that serves a valid PNG at any path.
func pngServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		if err := png.Encode(w, testImage()); err != nil {
			t.Fatalf("pngServer: failed to encode PNG: %v", err)
		}
	}))
}

// errorServer returns an httptest.Server that always responds with the given status code.
func errorServer(status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
	}))
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestHandleHealth(t *testing.T) {
	t.Parallel()

	srv := NewServer(&mockSearcher{}, &mockFetcher{}, &mockGenerator{}, http.DefaultClient)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %q", body["status"])
	}
}

func TestHandleMeme_RandomText(t *testing.T) {
	t.Parallel()

	imgSrv := pngServer(t)
	defer imgSrv.Close()

	gen := &mockGenerator{gif: testGIF()}
	srv := NewServer(
		&mockSearcher{url: imgSrv.URL + "/potato.png"},
		&mockFetcher{img: testImage()},
		gen,
		imgSrv.Client(),
	)

	req := httptest.NewRequest(http.MethodGet, "/meme", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d; body: %s", rec.Code, rec.Body.String())
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "image/gif" {
		t.Errorf("expected Content-Type image/gif, got %q", ct)
	}

	// Verify the body is a valid GIF.
	if _, err := gif.DecodeAll(rec.Body); err != nil {
		t.Errorf("response body is not a valid GIF: %v", err)
	}

	if !gen.randomCalled {
		t.Error("expected GenerateRandom to be called")
	}
	if gen.generateCalled {
		t.Error("did not expect Generate to be called")
	}
}

func TestHandleMeme_CustomText(t *testing.T) {
	t.Parallel()

	imgSrv := pngServer(t)
	defer imgSrv.Close()

	gen := &mockGenerator{gif: testGIF()}
	srv := NewServer(
		&mockSearcher{url: imgSrv.URL + "/potato.png"},
		&mockFetcher{img: testImage()},
		gen,
		imgSrv.Client(),
	)

	req := httptest.NewRequest(http.MethodGet, "/meme?top=hello&bottom=world", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d; body: %s", rec.Code, rec.Body.String())
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "image/gif" {
		t.Errorf("expected Content-Type image/gif, got %q", ct)
	}

	if _, err := gif.DecodeAll(rec.Body); err != nil {
		t.Errorf("response body is not a valid GIF: %v", err)
	}

	if !gen.generateCalled {
		t.Error("expected Generate to be called")
	}
	if gen.randomCalled {
		t.Error("did not expect GenerateRandom to be called")
	}
}

func TestHandleMeme_PartialCustomText_UsesRandom(t *testing.T) {
	t.Parallel()

	// When only one of top/bottom is provided, GenerateRandom should be used.
	imgSrv := pngServer(t)
	defer imgSrv.Close()

	gen := &mockGenerator{gif: testGIF()}
	srv := NewServer(
		&mockSearcher{url: imgSrv.URL + "/potato.png"},
		&mockFetcher{img: testImage()},
		gen,
		imgSrv.Client(),
	)

	req := httptest.NewRequest(http.MethodGet, "/meme?top=hello", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d; body: %s", rec.Code, rec.Body.String())
	}

	if !gen.randomCalled {
		t.Error("expected GenerateRandom when only top text is provided")
	}
	if gen.generateCalled {
		t.Error("did not expect Generate when only top text is provided")
	}
}

func TestHandleMeme_GiphyFailure(t *testing.T) {
	t.Parallel()

	srv := NewServer(
		&mockSearcher{err: errors.New("potato search failed")},
		&mockFetcher{img: testImage()},
		&mockGenerator{gif: testGIF()},
		http.DefaultClient,
	)

	req := httptest.NewRequest(http.MethodGet, "/meme", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected status 502, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if body["error"] == "" {
		t.Error("expected non-empty error message in response body")
	}
}

func TestHandleMeme_CataasFailure(t *testing.T) {
	t.Parallel()

	imgSrv := pngServer(t)
	defer imgSrv.Close()

	srv := NewServer(
		&mockSearcher{url: imgSrv.URL + "/potato.png"},
		&mockFetcher{err: errors.New("cataas is down")},
		&mockGenerator{gif: testGIF()},
		imgSrv.Client(),
	)

	req := httptest.NewRequest(http.MethodGet, "/meme", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected status 502, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if body["error"] == "" {
		t.Error("expected non-empty error message in response body")
	}
}

func TestHandleMeme_MemeGenerationFailure(t *testing.T) {
	t.Parallel()

	imgSrv := pngServer(t)
	defer imgSrv.Close()

	srv := NewServer(
		&mockSearcher{url: imgSrv.URL + "/potato.png"},
		&mockFetcher{img: testImage()},
		&mockGenerator{err: errors.New("meme generation failed")},
		imgSrv.Client(),
	)

	req := httptest.NewRequest(http.MethodGet, "/meme", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if body["error"] == "" {
		t.Error("expected non-empty error message in response body")
	}
}

func TestHandleMeme_PotatoImageDownloadFailure(t *testing.T) {
	t.Parallel()

	errSrv := errorServer(http.StatusInternalServerError)
	defer errSrv.Close()

	srv := NewServer(
		&mockSearcher{url: errSrv.URL + "/potato.png"},
		&mockFetcher{img: testImage()},
		&mockGenerator{gif: testGIF()},
		errSrv.Client(),
	)

	req := httptest.NewRequest(http.MethodGet, "/meme", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected status 502, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if body["error"] == "" {
		t.Error("expected non-empty error message in response body")
	}
}

func TestWriteError(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	writeError(rec, http.StatusTeapot, "i'm a teapot")

	if rec.Code != http.StatusTeapot {
		t.Fatalf("expected status %d, got %d", http.StatusTeapot, rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["error"] != "i'm a teapot" {
		t.Errorf("expected error message %q, got %q", "i'm a teapot", body["error"])
	}
}

func TestNewServer_RoutesRegistered(t *testing.T) {
	t.Parallel()

	srv := NewServer(&mockSearcher{}, &mockFetcher{}, &mockGenerator{}, http.DefaultClient)

	// Verify that unknown routes return 404.
	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for unknown route, got %d", rec.Code)
	}

	// Verify POST to /health is not allowed (method not allowed).
	req = httptest.NewRequest(http.MethodPost, "/health", nil)
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for POST /health, got %d", rec.Code)
	}
}
