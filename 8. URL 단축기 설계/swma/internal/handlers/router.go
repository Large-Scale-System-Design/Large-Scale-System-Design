package handlers

import (
	"database/sql"
	"net/http"

	"example.com/urlshortener/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(cfg config.Config, db *sql.DB) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	h := &Handler{Cfg: cfg, DB: db}

	r.Route("/api/v1", func(api chi.Router) {
		api.Post("/data/shorten", h.Shorten)
		api.Delete("/data/{code}", h.Delete)

		// Redirect endpoint (your chosen pattern)
		api.Get("/shortUrl/{code}", h.Redirect)
	})

	// Short URL for end users
	r.Get("/s/{code}", h.Redirect)

	// health
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	return r
}
