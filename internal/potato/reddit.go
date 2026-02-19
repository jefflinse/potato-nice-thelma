package potato

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"strings"
)

var subreddits = []string{"potato", "PotatoesAreFunny", "potatoes"}

// RedditClient fetches potato images from Reddit's public JSON API.
// It requires no API key — only a descriptive User-Agent header.
type RedditClient struct {
	httpClient *http.Client
	subreddits []string
}

// NewRedditClient returns a RedditClient that uses the provided HTTP client
// for all outbound requests.
func NewRedditClient(httpClient *http.Client) *RedditClient {
	return &RedditClient{
		httpClient: httpClient,
		subreddits: subreddits,
	}
}

// redditListing models only the fields we need from Reddit's listing endpoint.
type redditListing struct {
	Data struct {
		Children []struct {
			Data struct {
				URL      string `json:"url"`
				PostHint string `json:"post_hint"`
				IsVideo  bool   `json:"is_video"`
				Over18   bool   `json:"over_18"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

// SearchRandom returns the URL of a random potato image sourced from Reddit.
// The query parameter is accepted for interface compatibility but ignored —
// images come from potato-specific subreddits.
//
// On any failure other than context cancellation, a random URL from the
// hardcoded fallback list is returned instead.
func (rc *RedditClient) SearchRandom(ctx context.Context, _ string) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	url, err := rc.fetchFromReddit(ctx)
	if err != nil {
		// Context cancellation/expiry: propagate, don't fall back.
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		return pickFallback(), nil
	}

	return url, nil
}

// fetchFromReddit picks a random subreddit, fetches its hot posts, filters
// for qualifying image posts, and returns a random image URL.
func (rc *RedditClient) fetchFromReddit(ctx context.Context) (string, error) {
	sub := rc.subreddits[rand.IntN(len(rc.subreddits))]
	endpoint := fmt.Sprintf("https://www.reddit.com/r/%s/hot.json?limit=50", sub)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("creating reddit request: %w", err)
	}
	req.Header.Set("User-Agent", "potato-nice-thelma/1.0")

	resp, err := rc.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("executing reddit request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("reddit returned status %d", resp.StatusCode)
	}

	var listing redditListing
	if err := json.NewDecoder(resp.Body).Decode(&listing); err != nil {
		return "", fmt.Errorf("decoding reddit response: %w", err)
	}

	var candidates []string
	for _, child := range listing.Data.Children {
		post := child.Data
		if post.PostHint != "image" {
			continue
		}
		if post.IsVideo {
			continue
		}
		if post.Over18 {
			continue
		}
		if !isImageURL(post.URL) {
			continue
		}
		candidates = append(candidates, post.URL)
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("no qualifying image posts found in r/%s", sub)
	}

	return candidates[rand.IntN(len(candidates))], nil
}

// isImageURL reports whether the URL ends with a common image extension.
func isImageURL(u string) bool {
	lower := strings.ToLower(u)
	return strings.HasSuffix(lower, ".jpg") ||
		strings.HasSuffix(lower, ".jpeg") ||
		strings.HasSuffix(lower, ".png") ||
		strings.HasSuffix(lower, ".gif")
}

// pickFallback returns a random URL from the hardcoded fallback list.
func pickFallback() string {
	return fallbackURLs[rand.IntN(len(fallbackURLs))]
}
