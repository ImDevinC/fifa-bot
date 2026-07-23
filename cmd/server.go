package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/imdevinc/fifa-bot/pkg/app"
	"github.com/imdevinc/fifa-bot/pkg/database"
	"github.com/imdevinc/fifa-bot/pkg/fifa"
	go_fifa "github.com/imdevinc/go-fifa"
	_ "net/http/pprof"
)

func main() {
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = "config.yaml"
	}
	cfg, err := app.LoadConfig(configFile)
	if err != nil {
		log.Fatal("failed to load config", "error", err)
		os.Exit(1)
	}
	var level slog.Level
	err = level.UnmarshalText([]byte(cfg.LogLevel))
	if err != nil {
		log.Printf("unexpected log level: %s", cfg.LogLevel)
		level = slog.LevelInfo
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)
	db := database.NewRedisClient(cfg.Redis.Address, cfg.Redis.Password, cfg.Redis.Database)
	fc := go_fifa.Client{}

	go func() {
		http.HandleFunc("/healthz", healthHandler(db))

		if cfg.EnableProfiling {
			logger.Info("starting pprof and health check server", "port", cfg.ProfilingPort)
			http.HandleFunc("/", handle)
		} else {
			logger.Info("starting health check server", "port", cfg.ProfilingPort)
		}

		if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.ProfilingPort), nil); err != nil {
			logger.Error("HTTP server failed", "error", err)
		}
	}()

	skipSet, err := fifa.ParseEventNames(cfg.SkipEvents)
	if err != nil {
		logger.Warn("invalid skip_events entry, skipping unknown names", "error", err)
	}

	sentryEnabled := false
	if cfg.SentryDSN != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn: cfg.SentryDSN,
		})
		if err != nil {
			logger.Error("failed to initialize sentry", "error", err)
		} else {
			sentryEnabled = true
			logger.Info("sentry initialized", "dsn", cfg.SentryDSN)
			defer sentry.Flush(2 * time.Second)
		}
	}

	server := app.New(db, &fc, cfg.SlackWebhookURL, cfg.CompetitionID, cfg.SleepTimeSeconds, skipSet, sentryEnabled)
	if err := server.Run(context.Background()); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}

type healthResponse struct {
	Status string `json:"status"`
	Redis  string `json:"redis"`
}

func healthHandler(db database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		w.Header().Set("Content-Type", "application/json")

		resp := healthResponse{}

		if err := db.Ping(ctx); err != nil {
			resp.Status = "error"
			resp.Redis = "disconnected"
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(resp)
			return
		}

		resp.Status = "ok"
		resp.Redis = "connected"
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}

func handle(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
