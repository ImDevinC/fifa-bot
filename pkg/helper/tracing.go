package helper

import (
	"github.com/getsentry/sentry-go"
	log "github.com/sirupsen/logrus"
)

type SentryConfig struct {
	DSN             string
	TraceSampleRate float64
	Release         string
	Debug           bool
}

func InitLogging(logLevel log.Level) {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(logLevel)
}

func InitSentry(config SentryConfig) error {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              config.DSN,
		Debug:            config.Debug,
		TracesSampleRate: config.TraceSampleRate,
		Release:          config.Release,
	})
	if err != nil {
		log.WithError(err).Error("failed to initialize sentry")
		return err
	}
	return nil
}
