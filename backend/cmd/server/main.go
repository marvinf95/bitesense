// Command server starts the BiteSense HTTP API.
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/marvinf95/bitesense/backend/internal/api"
	"github.com/marvinf95/bitesense/backend/internal/auth"
	"github.com/marvinf95/bitesense/backend/internal/config"
	"github.com/marvinf95/bitesense/backend/internal/correlation"
	"github.com/marvinf95/bitesense/backend/internal/db"
	"github.com/marvinf95/bitesense/backend/internal/foodfacts"
	"github.com/marvinf95/bitesense/backend/internal/pdf"
	"github.com/marvinf95/bitesense/backend/internal/vision"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("load config")
	}

	level, _ := zerolog.ParseLevel(cfg.LogLevel)
	zerolog.SetGlobalLevel(level)
	if cfg.LogFormat == "console" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	conn, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatal().Err(err).Msg("open db")
	}
	defer conn.Close()
	if err := db.Migrate(conn); err != nil {
		log.Fatal().Err(err).Msg("migrate db")
	}

	if err := os.MkdirAll(cfg.UploadDir, 0o750); err != nil {
		log.Fatal().Err(err).Msg("uploads dir")
	}

	authSvc := auth.NewService(conn, cfg.JWTSecret, cfg.AccessTTL, cfg.RefreshTTL)

	off := foodfacts.New(cfg.OFFUserAgent)
	visionA := &vision.Analyzer{
		DB:                conn,
		Gemini:            vision.NewGeminiClient(cfg.GeminiAPIKey),
		Claude:            vision.NewClaudeClient(cfg.AnthropicAPIKey),
		OFF:               off,
		UploadDir:         cfg.UploadDir,
		FallbackThreshold: cfg.VisionFallback,
	}
	corrA := &correlation.Analyzer{DB: conn}
	pdfE := &pdf.Exporter{DB: conn, Corr: corrA}

	router := api.NewRouter(api.Deps{
		Cfg:    cfg,
		DB:     conn,
		Auth:   authSvc,
		Vision: visionA,
		PDF:    pdfE,
		Corr:   corrA,
	})

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      90 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}

	go func() {
		log.Info().Str("addr", cfg.Addr).Msg("listening")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("listen")
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Info().Msg("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
