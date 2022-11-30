package app

import (
	"context"
	"errors"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/imdevinc/fifa-bot/pkg/database"
	"github.com/imdevinc/fifa-bot/pkg/fifa"
	"github.com/imdevinc/fifa-bot/pkg/queue"
	go_fifa "github.com/imdevinc/go-fifa"
	log "github.com/sirupsen/logrus"
)

type GetMatchesConfig struct {
	DatabaseClient *database.Client
	QueueClient    *queue.Client
	FifaClient     *go_fifa.Client
	CompetitionId  string
}

func GetMatches(ctx context.Context, config *GetMatchesConfig) error {
	span := sentry.StartSpan(ctx, "function")
	defer span.Finish()
	span.Description = "matches.GetMatches"

	ctx = span.Context()

	log.Debug("starting GetMatches")
	matches, err := fifa.GetLiveMatches(ctx, config.FifaClient)
	if err != nil {
		log.WithError(err).Error("failed to get live matches")
		return err
	}

	var errWrap []string
	for _, m := range matches {
		if len(config.CompetitionId) > 0 && m.CompetitionId != config.CompetitionId {
			log.WithFields(log.Fields{
				"wantedCompetitionId": config.CompetitionId,
				"competitionId":       m.CompetitionId,
			}).Debug("competitionId doesn't match, skipping")
			continue
		}

		log.WithField("matchId", m.MatchId).Debug("checking if match exists")
		err := config.DatabaseClient.DoesMatchExist(ctx, m.MatchId)
		if err != nil && !errors.Is(err, database.ErrMatchNotFound) {
			log.WithError(err).Error("failed to get match info from database")
			errWrap = append(errWrap, err.Error())
			continue
		}
		if !errors.Is(err, database.ErrMatchNotFound) {
			log.Debug("match exists, skipping")
			continue
		}

		log.Debug("adding match to database")
		err = config.DatabaseClient.AddMatch(ctx, m.MatchId)
		if err != nil {
			log.WithError(err).Error("failed to save match to database")
			errWrap = append(errWrap, err.Error())
			continue
		}

		log.Debug("sending match to queue")
		m.LastEvent = "-1"
		err = config.QueueClient.SendToQueue(ctx, &m)
		if err != nil {
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
