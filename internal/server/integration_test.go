//go:build integration

package server_test

import (
	"encoding/json"
	"image"
	_ "image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jefflinse/potato-nice-thelma/internal/cataas"
	"github.com/jefflinse/potato-nice-thelma/internal/meme"
	"github.com/jefflinse/potato-nice-thelma/internal/potato"
	"github.com/jefflinse/potato-nice-thelma/internal/server"
)

func TestIntegration_MemeEndpoint(t *testing.T) {
	httpClient := &http.Client{Timeout: 15 * time.Second}
	potatoClient := potato.NewRedditClient(httpClient)
	cataasClient := cataas.NewClient(httpClient)

	memeGen, err := meme.NewGenerator()
	if err != nil {
		t.Fatalf("failed to create meme generator: %v", err)
	}

	srv := server.NewServer(potatoClient, cataasClient, memeGen, httpClient)
	ts := httptest.NewServer(srv)
	defer ts.Close()

	// Use a client with a generous timeout for real API calls.
	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Get(ts.URL + "/meme")
	if err != nil {
		t.Fatalf("GET /meme failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("GET /meme: expected status 200, got %d; body: %s", resp.StatusCode, body)
	}

	ct := resp.Header.Get("Content-Type")
	if ct != "image/png" {
		t.Errorf("GET /meme: expected Content-Type image/png, got %q", ct)
	}

	img, format, err := image.Decode(resp.Body)
	if err != nil {
		t.Fatalf("GET /meme: response body is not a valid image: %v", err)
	}
	if format != "png" {
		t.Errorf("GET /meme: expected png format, got %q", format)
	}

	bounds := img.Bounds()
	if bounds.Dx() == 0 || bounds.Dy() == 0 {
		t.Errorf("GET /meme: decoded image has zero dimensions: %v", bounds)
	}
}

func TestIntegration_HealthEndpoint(t *testing.T) {
	httpClient := &http.Client{Timeout: 15 * time.Second}
	potatoClient := potato.NewRedditClient(httpClient)
	cataasClient := cataas.NewClient(httpClient)

	memeGen, err := meme.NewGenerator()
	if err != nil {
		t.Fatalf("failed to create meme generator: %v", err)
	}

	srv := server.NewServer(potatoClient, cataasClient, memeGen, httpClient)
	ts := httptest.NewServer(srv)
	defer ts.Close()

	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("GET /health: expected status 200, got %d; body: %s", resp.StatusCode, body)
	}

	ct := resp.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("GET /health: expected Content-Type application/json, got %q", ct)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("GET /health: failed to decode JSON response: %v", err)
	}

	if body["status"] != "ok" {
		t.Errorf("GET /health: expected status \"ok\", got %q", body["status"])
	}
}
