package app

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	SlackWebhookURL  string `mapstructure:"slack_webhook_url"`
	CompetitionID    string `mapstructure:"competition_id"`
	SleepTimeSeconds int    `mapstructure:"sleep_time_seconds"`
	Redis            struct {
		Address  string `mapstructure:"address"`
		Password string `mapstructure:"password"`
		Database int    `mapstructure:"database"`
	} `mapstructure:"redis"`
	LogLevel        string   `mapstructure:"log_level"`
	EnableProfiling bool     `mapstructure:"enable_profiling"`
	ProfilingPort   int      `mapstructure:"profiling_port"`
	SkipEvents      []string `mapstructure:"skip_events"`
	SentryDSN       string   `mapstructure:"sentry_dsn"`
}

func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	v.SetDefault("sleep_time_seconds", 60)
	v.SetDefault("log_level", "WARN")
	v.SetDefault("enable_profiling", false)
	v.SetDefault("profiling_port", 8080)

	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	var missing []string
	if cfg.SlackWebhookURL == "" {
		missing = append(missing, "slack_webhook_url")
	}
	if cfg.Redis.Address == "" {
		missing = append(missing, "redis.address")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("required config fields are missing: %s", strings.Join(missing, ", "))
	}

	return &cfg, nil
}
