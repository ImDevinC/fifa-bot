package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

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
	if err := server.Run(context.Background()); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func handle(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
