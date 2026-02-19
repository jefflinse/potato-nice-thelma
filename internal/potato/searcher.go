package potato

import "context"

// Searcher finds potato images on the internet.
type Searcher interface {
	SearchRandom(ctx context.Context, query string) (string, error)
}
