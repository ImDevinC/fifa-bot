package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	go_fifa "github.com/ImDevinC/go-fifa"
	"github.com/aws/aws-lambda-go/events"
	"github.com/getsentry/sentry-go"
	"github.com/imdevinc/fifa-bot/pkg/fifa"
	"github.com/imdevinc/fifa-bot/pkg/queue"
	log "github.com/sirupsen/logrus"
)

type GetEventsConfig struct {
	QueueClient *queue.Client
	FifaClient  *go_fifa.Client
	WebhookURL  string
}

func GetEvents(ctx context.Context, config *GetEventsConfig, event events.SQSMessage) error {
	span := sentry.StartSpan(ctx, "events.GetEvents")
	defer span.Finish()

	ctx = span.Context()

	opts := queue.MatchOptsFromSQS(ctx, event.MessageAttributes)

	span.SetTag("competitionId", opts.CompetitionId)
	span.SetTag("seasonId", opts.SeasonId)
	span.SetTag("stageId", opts.StageId)
	span.SetTag("matchId", opts.MatchId)

	fields := log.Fields{
		"competitionId":  opts.CompetitionId,
		"seasonId":       opts.SeasonId,
		"stageId":        opts.StageId,
		"matchId":        opts.MatchId,
		"homeTeamName":   opts.HomeTeamName,
		"homeTeamAbbrev": opts.HomeTeamAbbrev,
		"awayTeamName":   opts.AwayTeamName,
		"awayteamAbbrev": opts.AwayTeamAbbrev,
		"lastEvent":      opts.LastEvent,
	}
	log.WithFields(fields).Debug("checking for events")

	active, err := isMatchActive(ctx, config.FifaClient, &opts)
	if err != nil {
		sentry.CaptureException(err)
		return fmt.Errorf("failed to determine if match is active. %w", err)
	}

	events, matchOver, err := fifa.GetMatchEvents(ctx, config.FifaClient, &opts)
	if err != nil {
		sentry.CaptureException(err)
		return fmt.Errorf("failed to get match events. %w", err)
	}

	fields["lastEvent"] = opts.LastEvent
	err = sendEventsToSlack(ctx, config.WebhookURL, events)
	if err != nil {
		sentry.CaptureException(err)
		return fmt.Errorf("failed to send events to Slack. %w", err)
	}
	if matchOver {
		return nil
	}

	if !active {
		log.WithFields(fields).Warn("match was not marked as completed, but is no longer live")
		return nil
	}

	err = config.QueueClient.SendToQueue(ctx, &opts)
	if err != nil {
		sentry.CaptureException(err)
		return fmt.Errorf("failed to send message to queue. %w", err)
	}

	return nil
}

type SlackMessage struct {
	Text string `json:"text"`
}

// sendEventsToSlack sends the payload to the webhookURL. This expects the message to
// be a raw string that will be sent as the `text: ` value in a slack message
func sendEventsToSlack(ctx context.Context, webhookURL string, events []string) error {
	span := sentry.StartSpan(ctx, "events.sendEventsToSlack")
	defer span.Finish()
	for _, evt := range events {
		payload := SlackMessage{Text: evt}
		b, err := json.Marshal(payload)
		if err != nil {
			sentry.CaptureException(err)
			return err
		}
		_, err = http.Post(webhookURL, "application/json", bytes.NewReader(b))
		if err != nil {
			sentry.CaptureException(err)
			return err
		}
	}
	return nil
}

// isMatchActive checks the active matches to see if the match still exists by checking
// for the matchId, competitionId, stageId, and seasonId
func isMatchActive(ctx context.Context, client *go_fifa.Client, opts *queue.MatchOptions) (bool, error) {
	span := sentry.StartSpan(ctx, "events.isMatchActive")
	defer span.Finish()

	matches, err := client.GetCurrentMatches()
	if err != nil {
		sentry.CaptureException(err)
		return false, err
	}
	var matchFound bool = false
	for _, m := range matches {
		if m.Id == opts.MatchId && m.CompetitionId == opts.CompetitionId && m.StageId == opts.StageId && m.SeasonId == opts.SeasonId {
			matchFound = true
			break
		}
	}
	return matchFound, nil
}
