package cataas

import (
	"context"
	"fmt"
	"image"
	"net/http"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

const baseURL = "https://cataas.com/cat"

// Fetcher retrieves cat images from CATAAS.
type Fetcher interface {
	FetchRandomCat(ctx context.Context) (image.Image, error)
}

// Client is an HTTP client for the CATAAS API.
type Client struct {
	httpClient *http.Client
}

// NewClient returns a new CATAAS client that uses the provided HTTP client.
func NewClient(httpClient *http.Client) *Client {
	return &Client{httpClient: httpClient}
}

// FetchRandomCat fetches a random cat image from CATAAS.
func (c *Client) FetchRandomCat(ctx context.Context) (image.Image, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching cat image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cataas returned status %d", resp.StatusCode)
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cat image: %w", err)
	}

	return img, nil
}
