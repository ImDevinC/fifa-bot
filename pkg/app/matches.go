package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	go_fifa "github.com/ImDevinC/go-fifa"
	"github.com/getsentry/sentry-go"
	"github.com/imdevinc/fifa-bot/pkg/database"
	"github.com/imdevinc/fifa-bot/pkg/fifa"
	"github.com/imdevinc/fifa-bot/pkg/queue"
	log "github.com/sirupsen/logrus"
)

type MatchConfig struct {
	LogLevel       log.Level
	HttpClient     *http.Client
	SentryConfig   SentryConfig
	DatabaseClient *database.Client
	QueueClient    *queue.Client
	FifaClient     *go_fifa.Client
	CompetitionId  string
}

func GetMatches(ctx context.Context, config *MatchConfig) error {
	initLogging(config.LogLevel)
	err := initSentry(config.SentryConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize sentry")
	}
	defer sentry.Flush(2 * time.Second)

	transaction := sentry.StartTransaction(ctx, "matches.HandleRequest", sentry.OpName("HandleRequest"))
	defer transaction.Finish()

	span := transaction.StartChild("matches.HandleRequest")
	defer span.Finish()

	ctx = span.Context()

	matches, err := fifa.GetLiveMatches(ctx, config.FifaClient)
	if err != nil {
		sentry.CaptureException(err)
		log.WithError(err).Error("failed to get live matches")
		return err
	}

	var errWrap []string
	for _, m := range matches {
		if len(config.CompetitionId) > 0 && m.CompetitionId != config.CompetitionId {
			continue
		}

		err := config.DatabaseClient.DoesMatchExist(ctx, &m)
		if err != nil && !errors.Is(err, database.ErrMatchNotFound) {
			sentry.CaptureException(err)
			log.WithError(err).Error("failed to get match info from database")
			errWrap = append(errWrap, err.Error())
			continue
		}
		if !errors.Is(err, database.ErrMatchNotFound) {
			continue
		}

		err = config.DatabaseClient.AddMatch(ctx, &m)
		if err != nil {
			sentry.CaptureException(err)
			log.WithError(err).Error("failed to save match to database")
			errWrap = append(errWrap, err.Error())
			continue
		}

		m.LastEvent = "-1"
		err = config.QueueClient.SendToQueue(ctx, &m)
		if err != nil {
			sentry.CaptureException(err)
			log.WithError(err).Error("failed to send message to queue")
			errWrap = append(errWrap, err.Error())
			continue
		}
	}
	if len(errWrap) > 0 {
		return errors.New(strings.Join(errWrap, "\n"))
	}
	return nil
}
