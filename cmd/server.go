package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"time"

	"github.com/imdevinc/fifa-bot/pkg/app"
	"github.com/imdevinc/fifa-bot/pkg/database"
	go_fifa "github.com/imdevinc/go-fifa"
)

func main() {
	cfg, err := app.GetConfigFromEnv()
	if err != nil {
		log.Fatal("failed to get config from env", "error", err)
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

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthzHandler(db))

	go func() {
		logger.Info("starting health server", "port", cfg.HealthPort)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.HealthPort), mux); err != nil {
			logger.Error("health server failed", "error", err)
		}
	}()

	if cfg.EnableProfiling {
		profilingMux := http.NewServeMux()
		registerProfilingHandlers(profilingMux)
		go func() {
			logger.Info("starting pprof server", "port", cfg.ProfilingPort)
			if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.ProfilingPort), profilingMux); err != nil {
				logger.Error("profiling server failed", "error", err)
			}
		}()
	}

	server := app.New(db, &fc, cfg.SlackWebhookURL, cfg.CompetitionID, cfg.SleepTimeSeconds)
	if err := server.Run(context.Background()); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func healthzHandler(db database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			http.Error(w, "redis unavailable", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func registerProfilingHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
}
