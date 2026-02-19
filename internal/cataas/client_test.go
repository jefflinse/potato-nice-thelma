package cataas

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// redirectTransport rewrites outgoing requests to point at a local test server
// instead of the hardcoded baseURL. This lets us intercept all HTTP traffic
// without modifying the production code.
type redirectTransport struct {
	testServerURL string
}

func (t *redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	testURL, _ := url.Parse(t.testServerURL)
	req.URL.Scheme = testURL.Scheme
	req.URL.Host = testURL.Host
	return http.DefaultTransport.RoundTrip(req)
}

// newTestClient creates a Client whose HTTP traffic is redirected to the given
// test server URL.
func newTestClient(serverURL string) *Client {
	return NewClient(&http.Client{
		Transport: &redirectTransport{testServerURL: serverURL},
	})
}

// makeJPEG creates a small JPEG image in memory and returns the encoded bytes.
func makeJPEG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill with a recognisable colour so we can assert on pixel values if needed.
	for y := range height {
		for x := range width {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatalf("encoding test JPEG: %v", err)
	}
	return buf.Bytes()
}

// makePNG creates a small PNG image in memory and returns the encoded bytes.
func makePNG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			img.Set(x, y, color.RGBA{R: 0, G: 0, B: 255, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encoding test PNG: %v", err)
	}
	return buf.Bytes()
}

func TestNewClient(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil client", func(t *testing.T) {
		t.Parallel()
		c := NewClient(http.DefaultClient)
		if c == nil {
			t.Fatal("expected non-nil Client")
		}
	})

	t.Run("stores provided http client", func(t *testing.T) {
		t.Parallel()
		hc := &http.Client{}
		c := NewClient(hc)
		if c.httpClient != hc {
			t.Fatal("expected Client to store the provided http.Client")
		}
	})
}

func TestClient_FetchRandomCat(t *testing.T) {
	t.Parallel()

	t.Run("successful fetch with JPEG image", func(t *testing.T) {
		t.Parallel()

		const width, height = 100, 80
		jpegData := makeJPEG(t, width, height)

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/jpeg")
			w.WriteHeader(http.StatusOK)
			w.Write(jpegData)
		}))
		t.Cleanup(srv.Close)

		client := newTestClient(srv.URL)
		img, err := client.FetchRandomCat(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if img == nil {
			t.Fatal("expected non-nil image")
		}

		bounds := img.Bounds()
		gotWidth := bounds.Dx()
		gotHeight := bounds.Dy()
		if gotWidth != width || gotHeight != height {
			t.Errorf("image dimensions = %dx%d, want %dx%d", gotWidth, gotHeight, width, height)
		}
	})

	t.Run("successful fetch with PNG image", func(t *testing.T) {
		t.Parallel()

		const width, height = 64, 48
		pngData := makePNG(t, width, height)

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/png")
			w.WriteHeader(http.StatusOK)
			w.Write(pngData)
		}))
		t.Cleanup(srv.Close)

		client := newTestClient(srv.URL)
		img, err := client.FetchRandomCat(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if img == nil {
			t.Fatal("expected non-nil image")
		}

		bounds := img.Bounds()
		gotWidth := bounds.Dx()
		gotHeight := bounds.Dy()
		if gotWidth != width || gotHeight != height {
			t.Errorf("image dimensions = %dx%d, want %dx%d", gotWidth, gotHeight, width, height)
		}
	})

	t.Run("non-200 status code", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name       string
			statusCode int
			wantSubstr string
		}{
			{
				name:       "500 Internal Server Error",
				statusCode: http.StatusInternalServerError,
				wantSubstr: "cataas returned status 500",
			},
			{
				name:       "404 Not Found",
				statusCode: http.StatusNotFound,
				wantSubstr: "cataas returned status 404",
			},
			{
				name:       "503 Service Unavailable",
				statusCode: http.StatusServiceUnavailable,
				wantSubstr: "cataas returned status 503",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tt.statusCode)
				}))
				t.Cleanup(srv.Close)

				client := newTestClient(srv.URL)
				img, err := client.FetchRandomCat(context.Background())
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if img != nil {
					t.Errorf("expected nil image on error, got %v", img)
				}
				if !strings.Contains(err.Error(), tt.wantSubstr) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.wantSubstr)
				}
			})
		}
	})

	t.Run("invalid image data", func(t *testing.T) {
		t.Parallel()

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/jpeg")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("not an image"))
		}))
		t.Cleanup(srv.Close)

		client := newTestClient(srv.URL)
		img, err := client.FetchRandomCat(context.Background())
		if err == nil {
			t.Fatal("expected error for invalid image data, got nil")
		}
		if img != nil {
			t.Errorf("expected nil image on decode error, got %v", img)
		}
		if !strings.Contains(err.Error(), "failed to decode cat image") {
			t.Errorf("error %q does not contain %q", err.Error(), "failed to decode cat image")
		}
	})

	t.Run("empty response body", func(t *testing.T) {
		t.Parallel()

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			// Write nothing — empty body.
		}))
		t.Cleanup(srv.Close)

		client := newTestClient(srv.URL)
		img, err := client.FetchRandomCat(context.Background())
		if err == nil {
			t.Fatal("expected error for empty body, got nil")
		}
		if img != nil {
			t.Errorf("expected nil image on decode error, got %v", img)
		}
		if !strings.Contains(err.Error(), "failed to decode cat image") {
			t.Errorf("error %q does not contain %q", err.Error(), "failed to decode cat image")
		}
	})

	t.Run("context already cancelled", func(t *testing.T) {
		t.Parallel()

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("handler should not be called with cancelled context")
		}))
		t.Cleanup(srv.Close)

		client := newTestClient(srv.URL)
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately.

		img, err := client.FetchRandomCat(ctx)
		if err == nil {
			t.Fatal("expected error for cancelled context, got nil")
		}
		if img != nil {
			t.Errorf("expected nil image on context error, got %v", img)
		}
		if !strings.Contains(err.Error(), "fetching cat image") {
			t.Errorf("error %q does not contain %q", err.Error(), "fetching cat image")
		}
	})

	t.Run("request uses GET method and correct path", func(t *testing.T) {
		t.Parallel()

		jpegData := makeJPEG(t, 10, 10)

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("expected GET request, got %s", r.Method)
			}
			if r.URL.Path != "/cat" {
				t.Errorf("expected path /cat, got %s", r.URL.Path)
			}
			w.Header().Set("Content-Type", "image/jpeg")
			w.WriteHeader(http.StatusOK)
			w.Write(jpegData)
		}))
		t.Cleanup(srv.Close)

		client := newTestClient(srv.URL)
		_, err := client.FetchRandomCat(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

// TestFetcherInterface verifies that *Client satisfies the Fetcher interface
// at compile time.
var _ Fetcher = (*Client)(nil)
