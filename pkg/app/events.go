package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/getsentry/sentry-go"
	"github.com/imdevinc/fifa-bot/pkg/database"
	"github.com/imdevinc/fifa-bot/pkg/fifa"
	"github.com/imdevinc/fifa-bot/pkg/queue"
	go_fifa "github.com/imdevinc/go-fifa"
	log "github.com/sirupsen/logrus"
)

type GetEventsConfig struct {
	QueueClient    *queue.Client
	FifaClient     *go_fifa.Client
	DatabaseClient *database.Client
	WebhookURL     string
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

	existingEvents, err := config.DatabaseClient.GetEvents(ctx, opts.MatchId)
	if errors.Is(err, database.ErrMatchNotFound) {
		return fmt.Errorf("match %s does not exist. %w", opts.MatchId, err)
	}

	matchData, err := fifa.GetMatchEvents(ctx, config.FifaClient, &opts)
	if err != nil {
		return fmt.Errorf("failed to get match events. %w", err)
	}

	fields["lastEvent"] = opts.LastEvent

	ids, messages := findNewEvents(ctx, existingEvents, matchData.NewEvents, &opts)
	allEvents := append(existingEvents, ids...)

	if len(allEvents) > len(existingEvents) {
		err = config.DatabaseClient.UpdateMatchEvents(ctx, opts.MatchId, allEvents)
		if err != nil {
			return fmt.Errorf("failed to save events to database. %w", err)
		}
	}

	err = sendEventsToSlack(ctx, config.WebhookURL, messages)
	if err != nil {
		return fmt.Errorf("failed to send events to Slack. %w", err)
	}
	if matchData.Done && !matchData.PendingEventFound {
		return nil
	}

	err = config.QueueClient.SendToQueue(ctx, &opts)
	if err != nil {
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
		if strings.TrimSpace(evt) == "" {
			continue
		}
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

func findNewEvents(ctx context.Context, existingEvents []string, newEvents []go_fifa.TimelineEvent, opts *queue.MatchOptions) ([]string, []string) {
	span := sentry.StartSpan(ctx, "function")
	defer span.Finish()
	span.Description = "events.findNewEvents"

	eventMsgs := []string{}
	eventIds := []string{}

	for _, event := range newEvents {
		eventFound := false
		for _, ex := range existingEvents {
			if event.Id == ex {
				eventFound = true
				break
			}
		}
		if eventFound {
			continue
		}
		eventIds = append(eventIds, event.Id)
		eventMsgs = append(eventMsgs, fifa.ProcessEvent(ctx, event, opts))
	}
	return eventIds, eventMsgs
}
