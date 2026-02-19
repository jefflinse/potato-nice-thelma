package potato

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Compile-time check: RedditClient must implement Searcher.
var _ Searcher = (*RedditClient)(nil)

func TestSearchRandom_FallsBackOnRedditFailure(t *testing.T) {
	// Server that always returns 500.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	rc := &RedditClient{
		httpClient: srv.Client(),
		subreddits: []string{"potato"},
	}
	// Override the subreddit fetch to hit our test server by using a transport
	// that redirects all requests to the test server.
	rc.httpClient.Transport = &rewriteTransport{base: srv.URL}

	url, err := rc.SearchRandom(context.Background(), "potato")
	if err != nil {
		t.Fatalf("expected fallback, got error: %v", err)
	}
	if url == "" {
		t.Fatal("expected non-empty fallback URL")
	}
}

func TestSearchRandom_PropagatesContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	rc := NewRedditClient(&http.Client{})
	_, err := rc.SearchRandom(ctx, "potato")
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}

func TestSearchRandom_FiltersCorrectly(t *testing.T) {
	listing := redditListing{}
	listing.Data.Children = []struct {
		Data struct {
			URL      string `json:"url"`
			PostHint string `json:"post_hint"`
			IsVideo  bool   `json:"is_video"`
			Over18   bool   `json:"over_18"`
		} `json:"data"`
	}{
		{Data: struct {
			URL      string `json:"url"`
			PostHint string `json:"post_hint"`
			IsVideo  bool   `json:"is_video"`
			Over18   bool   `json:"over_18"`
		}{URL: "https://i.redd.it/good.jpg", PostHint: "image", IsVideo: false, Over18: false}},
		{Data: struct {
			URL      string `json:"url"`
			PostHint string `json:"post_hint"`
			IsVideo  bool   `json:"is_video"`
			Over18   bool   `json:"over_18"`
		}{URL: "https://i.redd.it/nsfw.jpg", PostHint: "image", IsVideo: false, Over18: true}},
		{Data: struct {
			URL      string `json:"url"`
			PostHint string `json:"post_hint"`
			IsVideo  bool   `json:"is_video"`
			Over18   bool   `json:"over_18"`
		}{URL: "https://v.redd.it/video.mp4", PostHint: "hosted:video", IsVideo: true, Over18: false}},
		{Data: struct {
			URL      string `json:"url"`
			PostHint string `json:"post_hint"`
			IsVideo  bool   `json:"is_video"`
			Over18   bool   `json:"over_18"`
		}{URL: "https://reddit.com/gallery/abc", PostHint: "image", IsVideo: false, Over18: false}},
	}

	body, _ := json.Marshal(listing)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify User-Agent header is set.
		if ua := r.Header.Get("User-Agent"); ua != "potato-nice-thelma/1.0" {
			t.Errorf("expected User-Agent 'potato-nice-thelma/1.0', got %q", ua)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	}))
	defer srv.Close()

	rc := &RedditClient{
		httpClient: &http.Client{Transport: &rewriteTransport{base: srv.URL}},
		subreddits: []string{"potato"},
	}

	url, err := rc.SearchRandom(context.Background(), "potato")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "https://i.redd.it/good.jpg" {
		t.Fatalf("expected 'https://i.redd.it/good.jpg', got %q", url)
	}
}

func TestIsImageURL(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"https://i.redd.it/abc.jpg", true},
		{"https://i.redd.it/abc.JPEG", true},
		{"https://i.redd.it/abc.png", true},
		{"https://i.redd.it/abc.gif", true},
		{"https://i.redd.it/abc.mp4", false},
		{"https://reddit.com/gallery/abc", false},
		{"https://i.redd.it/abc.JPG", true},
	}
	for _, tt := range tests {
		if got := isImageURL(tt.url); got != tt.want {
			t.Errorf("isImageURL(%q) = %v, want %v", tt.url, got, tt.want)
		}
	}
}

// rewriteTransport redirects all requests to a test server URL.
type rewriteTransport struct {
	base string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.URL.Scheme = "http"
	req.URL.Host = t.base[len("http://"):]
	return http.DefaultTransport.RoundTrip(req)
}
