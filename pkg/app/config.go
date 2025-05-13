package app

import "github.com/kelseyhightower/envconfig"

type Config struct {
	SlackWebhookURL string `envconfig:"SLACK_WEBHOOK_URL" required:"true"`
	CompetitionID   string `envconfig:"COMPETITION_ID"`
	Redis           struct {
		Address  string `envconfig:"REDIS_ADDRESS" required:"true"`
		Password string `envconfig:"REDIS_PASSWORD"`
		Database int    `envconfig:"REDIS_DB" required:"true"`
	}
}

func GetConfigFromEnv() (*Config, error) {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		return nil, err
	}
	return &config, nil
}
