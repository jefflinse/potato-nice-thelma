package server

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/gif"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/jefflinse/potato-nice-thelma/internal/cataas"
	"github.com/jefflinse/potato-nice-thelma/internal/meme"
	"github.com/jefflinse/potato-nice-thelma/internal/potato"
	"golang.org/x/sync/errgroup"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"
)

// Server is the HTTP server for the potato-cat meme service.
type Server struct {
	potato     potato.Searcher
	cataas     cataas.Fetcher
	meme       meme.Generator
	httpClient *http.Client
	router     *http.ServeMux
}

// NewServer creates a Server wired with the given dependencies and routes.
func NewServer(potatoClient potato.Searcher, cataasClient cataas.Fetcher, memeGen meme.Generator, httpClient *http.Client) *Server {
	s := &Server{
		potato:     potatoClient,
		cataas:     cataasClient,
		meme:       memeGen,
		httpClient: httpClient,
		router:     http.NewServeMux(),
	}

	s.router.HandleFunc("GET /meme", s.handleMeme)
	s.router.HandleFunc("GET /health", s.handleHealth)

	return s
}

// ServeHTTP delegates to the internal mux so Server implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleMeme(w http.ResponseWriter, r *http.Request) {
	topText := r.URL.Query().Get("top")
	bottomText := r.URL.Query().Get("bottom")

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	queries := []string{"weird potato", "funny potato", "potato fail", "potato meme", "ugly potato", "potato face"}
	query := queries[rand.IntN(len(queries))]

	var potatoImg, catImg image.Image

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		potatoURL, err := s.potato.SearchRandom(gctx, query)
		if err != nil {
			return fmt.Errorf("searching for potato image: %w", err)
		}

		req, err := http.NewRequestWithContext(gctx, http.MethodGet, potatoURL, nil)
		if err != nil {
			return fmt.Errorf("creating potato image request: %w", err)
		}

		resp, err := s.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("downloading potato image: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("potato image download returned status %d", resp.StatusCode)
		}

		img, _, err := image.Decode(resp.Body)
		if err != nil {
			return fmt.Errorf("decoding potato image: %w", err)
		}

		potatoImg = img
		return nil
	})

	g.Go(func() error {
		img, err := s.cataas.FetchRandomCat(gctx)
		if err != nil {
			return fmt.Errorf("fetching cat image: %w", err)
		}
		catImg = img
		return nil
	})

	if err := g.Wait(); err != nil {
		slog.Error("failed to fetch images", "error", err)
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	var result *gif.GIF
	var err error

	if topText != "" && bottomText != "" {
		result, err = s.meme.Generate(potatoImg, catImg, topText, bottomText)
	} else {
		result, err = s.meme.GenerateRandom(potatoImg, catImg)
	}

	if err != nil {
		slog.Error("failed to generate meme", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "image/gif")
	if err := gif.EncodeAll(w, result); err != nil {
		slog.Error("failed to encode meme as GIF", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
