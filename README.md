# potato-nice-thelma

A Go web service that scrapes the internet for weird potato images and combines them with random cat photos to create awful memes. No API keys. No sign-ups. Just memes.

Hit the `/meme` endpoint, get back an animated GIF of a potato-cat meme with rainbow text, bouncing potatoes, sparkles, and screen shake. Maximum chaos. That's it. That's the whole thing.

## How It Works

1. **Potato acquisition** — Scrapes Reddit (r/potato, r/PotatoesAreFunny, r/potatoes) for weird potato images. Falls back to a curated list of potato images if Reddit is unavailable.
2. **Cat acquisition** — Fetches a random cat image from [CATAAS](https://cataas.com) (Cat as a Service — yes, that's a real thing)
3. **Meme assembly** — Composites the potato onto the cat image with chaotic effects: rainbow color-cycling text, bouncing/wobbling potato, sparkle overlays, and screen shake. Rendered frame-by-frame using the [Anton](https://fonts.google.com/specimen/Anton) font
4. **Delivery** — Returns the masterpiece as an animated GIF (16 frames, ~1.3 second loop)

Both images are fetched concurrently because we respect your time, even if we don't respect your taste in memes.

## Prerequisites

- **Go 1.22+**

That's it. No API keys. No accounts. Nothing.

## Quick Start

```bash
# Clone the repo
git clone https://github.com/jefflinse/potato-nice-thelma.git
cd potato-nice-thelma

# Build and run
make run

# Generate a meme
curl http://localhost:8080/meme > meme.gif
open meme.gif
```

## API Endpoints

### `GET /meme`

Generate a random animated potato-cat meme. Returns an `image/gif` response with effects (rainbow text, bouncing potato, sparkles, screen shake).

**Optional query parameters:**

| Parameter | Description |
|-----------|-------------|
| `top`     | Custom top text (default: random from built-in list) |
| `bottom`  | Custom bottom text (default: random from built-in list) |

Both `top` and `bottom` must be provided together to use custom text. If either is omitted, a random predefined text pair is used instead.

**Examples:**

```bash
# Random meme text
curl http://localhost:8080/meme > meme.gif

# Custom text
curl "http://localhost:8080/meme?top=when+you+realize&bottom=you+are+a+potato" > meme.gif
```

### `GET /health`

Health check endpoint. Returns JSON:

```json
{"status": "ok"}
```

## Configuration

All configuration is via environment variables:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | `8080` | HTTP listen port |

Yep, that's the only config. Zero required environment variables.

## Docker

```bash
# Build the image
make docker-build

# Run it
make docker-run
```

The Docker image uses a multi-stage build (Go Alpine builder + distroless runtime) so the final image is tiny.

## Running Tests

```bash
make test                    # Unit tests (with race detector)
make test-integration        # Integration tests (hits real Reddit + CATAAS)
make lint                    # Lint with golangci-lint
```

## Project Structure

```
potato-nice-thelma/
├── cmd/
│   └── server/
│       └── main.go              # Entrypoint — wires up dependencies, starts HTTP server
├── internal/
│   ├── cataas/
│   │   ├── client.go            # CATAAS client (fetches random cat images)
│   │   └── client_test.go
│   ├── config/
│   │   ├── config.go            # Environment variable configuration
│   │   └── config_test.go
│   ├── potato/
│   │   ├── searcher.go          # Searcher interface
│   │   ├── reddit.go            # Reddit scraper (finds potato images)
│   │   ├── reddit_test.go
│   │   └── fallback.go          # Hardcoded fallback potato image URLs
│   ├── meme/
│   │   ├── Anton-Regular.ttf    # Embedded meme font
│   │   ├── generator.go         # Image compositing and meme text rendering
│   │   └── generator_test.go
│   └── server/
│       ├── server.go            # HTTP handlers and routing
│       ├── server_test.go
│       └── integration_test.go  # Integration tests (build-tagged)
├── Dockerfile                   # Multi-stage build (Alpine builder + distroless)
├── Makefile                     # Build, test, lint, Docker targets
├── go.mod
└── go.sum
```

## Credits

- [Reddit](https://reddit.com) — for the potato subreddits
- [CATAAS](https://cataas.com) — Cat as a Service (the internet is a beautiful place)
- [fogleman/gg](https://github.com/fogleman/gg) — 2D graphics library for Go
- [Anton](https://fonts.google.com/specimen/Anton) — the meme font, from Google Fonts
