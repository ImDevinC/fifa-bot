package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/imdevinc/fifa-bot/pkg/app"
	"github.com/imdevinc/fifa-bot/pkg/database"
	go_fifa "github.com/imdevinc/go-fifa"
	_ "net/http/pprof"
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

	if cfg.EnableProfiling {
		go func() {
			logger.Info("starting pprof server", "port", cfg.ProfilingPort)
			http.HandleFunc("/", handle)
			if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.ProfilingPort), nil); err != nil {
				logger.Error("profiling server failed", "error", err)
			}
		}()
	}

	server := app.New(db, &fc, cfg.SlackWebhookURL, cfg.CompetitionID, cfg.SleepTimeSeconds)

	// Start health check endpoint
	go func() {
		http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := server.Health(ctx); err != nil {
				logger.Error("health check failed", "error", err)
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		logger.Info("starting health check server", "port", cfg.HealthPort)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.HealthPort), nil); err != nil {
			logger.Error("health check server failed", "error", err)
		}
	}()

	if err := server.Run(context.Background()); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func handle(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
