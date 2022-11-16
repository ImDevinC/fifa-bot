package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	go_fifa "github.com/ImDevinC/go-fifa"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/getsentry/sentry-go"
	"github.com/imdevinc/fifa-bot/pkg/fifa"
	"github.com/imdevinc/fifa-bot/pkg/queue"
	log "github.com/sirupsen/logrus"
)

type SlackMessage struct {
	Text string `json:"text"`
}

var release string = "development"

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

func initLogging() {
	log.SetFormatter(&log.JSONFormatter{})
	logLevel, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		log.Error("could not determine log level")
		logLevel = log.InfoLevel
	}
	log.SetLevel(logLevel)
}

func initSentry() error {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DSN"),
		Debug:            false,
		TracesSampleRate: 1.0,
		Release:          release,
	})
	if err != nil {
		log.WithError(err).Error("failed to initialize sentry")
		return err
	}
	return nil
}

func HandleRequest(ctx context.Context, event events.SQSEvent) error {
	initLogging()

	queueURL := os.Getenv("QUEUE_URL")
	if queueURL == "" {
		log.Error("Missing QUEUE_URL")
		return errors.New("missing QUEUE_URL")
	}

	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		log.Error("Missing SLACK_WEBHOOK_URL")
		return errors.New("missing SLACK_WEBHOOK_URL")
	}

	initSentry()
	defer sentry.Flush(2 * time.Second)

	rootSpan := sentry.StartSpan(ctx, "events.HandleRequest", sentry.TransactionName("events.HandleRequest"))
	defer rootSpan.Finish()

	var errWrap []string
	for _, r := range event.Records {
		span := sentry.StartSpan(rootSpan.Context(), "events.EventLoop")
		defer span.Finish()

		opts := queue.MatchOptions{
			CompetitionId:  *r.MessageAttributes["CompetitionId"].StringValue,
			SeasonId:       *r.MessageAttributes["SeasonId"].StringValue,
			StageId:        *r.MessageAttributes["StageId"].StringValue,
			MatchId:        *r.MessageAttributes["MatchId"].StringValue,
			HomeTeamName:   *r.MessageAttributes["HomeTeamName"].StringValue,
			AwayTeamName:   *r.MessageAttributes["AwayTeamName"].StringValue,
			HomeTeamAbbrev: *r.MessageAttributes["HomeTeamAbbrev"].StringValue,
			AwayTeamAbbrev: *r.MessageAttributes["AwayTeamAbbrev"].StringValue,
			LastEvent:      *r.MessageAttributes["LastEvent"].StringValue,
		}

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

		fifaClient := go_fifa.Client{}
		events, matchOver, err := fifa.GetMatchEvents(rootSpan.Context(), &fifaClient, &opts)
		if err != nil {
			sentry.CaptureException(err)
			log.WithField("error", err).Error("failed to get match events")
			errWrap = append(errWrap, err.Error())
			continue
		}
		fields["lastEvent"] = opts.LastEvent

		log.WithFields(fields).WithFields(log.Fields{"events": events, "matchOver": matchOver}).Debug("sending events to slack")
		err = sendEventsToSlack(rootSpan.Context(), webhookURL, events)
		if err != nil {
			sentry.CaptureException(err)
			log.WithField("error", err).Error("failed to send message to slack")
			errWrap = append(errWrap, err.Error())
			continue
		}
		if matchOver {
			continue
		}
		active, err := isMatchActive(rootSpan.Context(), &fifaClient, &opts)
		if err != nil {
			sentry.CaptureException(err)
			log.Error(err)
			errWrap = append(errWrap, err.Error())
			continue
		}
		if !active {
			log.WithFields(fields).Warn("match was not marked as completed, but is no longer live")
			continue
		}
		err = queue.SendToQueue(rootSpan.Context(), queueURL, &opts)
		if err != nil {
			sentry.CaptureException(err)
			log.WithField("error", err).Error("failed to send message to queue")
			errWrap = append(errWrap, err.Error())
			continue
		}
	}
	if len(errWrap) > 0 {
		return errors.New(strings.Join(errWrap, "\n"))
	}
	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
