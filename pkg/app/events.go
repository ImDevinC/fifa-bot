package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/getsentry/sentry-go"
	"github.com/imdevinc/fifa-bot/pkg/fifa"
	"github.com/imdevinc/fifa-bot/pkg/queue"
	go_fifa "github.com/imdevinc/go-fifa"
	log "github.com/sirupsen/logrus"
)

type GetEventsConfig struct {
	QueueClient *queue.Client
	FifaClient  *go_fifa.Client
	WebhookURL  string
}

func GetEvents(ctx context.Context, config *GetEventsConfig, event events.SQSMessage) error {
	span := sentry.StartSpan(ctx, "function")
	span.Description = "events.GetEvents"
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

	matchData, err := fifa.GetMatchEvents(ctx, config.FifaClient, &opts)
	if err != nil {
		sentry.CaptureException(err)
		return fmt.Errorf("failed to get match events. %w", err)
	}

	fields["lastEvent"] = opts.LastEvent
	err = sendEventsToSlack(ctx, config.WebhookURL, matchData.NewEvents)
	if err != nil {
		sentry.CaptureException(err)
		return fmt.Errorf("failed to send events to Slack. %w", err)
	}
	if matchData.Done && !matchData.PendingEventFound {
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
	span := sentry.StartSpan(ctx, "function")
	defer span.Finish()
	span.Description = "events.sendEventsToSlack"

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
