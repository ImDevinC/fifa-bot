package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/imdevinc/fifa-bot/pkg/app"
	"github.com/imdevinc/fifa-bot/pkg/database"
	go_fifa "github.com/imdevinc/go-fifa"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(slog.LevelDebug)
	cfg, err := app.GetConfigFromEnv()
	if err != nil {
		logger.Error("failed to get config from env", "error", err)
		os.Exit(1)
	}
	db := database.NewRedisClient(cfg.Redis.Address, cfg.Redis.Password, cfg.Redis.Database)
	fc := go_fifa.Client{}
	server := app.New(db, &fc, cfg.SlackWebhookURL, cfg.CompetitionID)
	if err := server.Run(context.Background()); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
