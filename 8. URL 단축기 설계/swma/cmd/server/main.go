package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/urlshortener/internal/config"
	"example.com/urlshortener/internal/db"
	"example.com/urlshortener/internal/handlers"
	"example.com/urlshortener/internal/migrate"
)

func main() {
	cfg := config.FromEnv()

	database, err := db.OpenWithRetry(cfg, 60*time.Second)
	if err != nil {
		log.Fatalf("db connect failed: %v", err)
	}
	defer database.Close()

	if err := migrate.EnsureSchema(database); err != nil {
		log.Fatalf("schema migration failed: %v", err)
	}

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handlers.NewRouter(cfg, database),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
