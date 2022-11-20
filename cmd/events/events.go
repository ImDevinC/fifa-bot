package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	go_fifa "github.com/ImDevinC/go-fifa"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/getsentry/sentry-go"
	"github.com/imdevinc/fifa-bot/pkg/app"
	"github.com/imdevinc/fifa-bot/pkg/helper"
	"github.com/imdevinc/fifa-bot/pkg/queue"
	log "github.com/sirupsen/logrus"
)

var release string = "development"

func HandleRequest(ctx context.Context, event events.SQSEvent) error {
	logLevel, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		log.WithError(err).Warn("failed to parse LOG_LEVEL")
		logLevel = log.InfoLevel
	}
	helper.InitLogging(logLevel)
	queueURL := os.Getenv("QUEUE_URL")
	if len(queueURL) == 0 {
		log.Error("missing QUEUE_URL")
		return errors.New("missing QUEUE_URL")
	}

	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if len(webhookURL) == 0 {
		log.Error("missing SLACK_WEBHOOK_URL")
		return errors.New("missing SLACK_WEBHOOK_URL")
	}

	err = helper.InitSentry(helper.SentryConfig{
		DSN:             os.Getenv("SENTRY_DSN"),
		TraceSampleRate: 1,
		Release:         release,
		Debug:           logLevel == log.DebugLevel,
	})

	if err != nil {
		log.WithError(err).Error("failed to initialize Sentry")
		return err
	}

	defer sentry.Flush(2 * time.Second)

	var initialTrace string
	var transaction *sentry.Span
	if len(event.Records) > 0 {
		if val, exists := event.Records[0].MessageAttributes["TraceId"]; exists {
			initialTrace = *val.StringValue
		}
	}

	if len(initialTrace) > 0 {
		log.WithField("sentry-trace", initialTrace).Debug("continuing trace")
		transaction = sentry.StartTransaction(ctx, "events.HandleRequest", sentry.ContinueFromTrace(initialTrace), sentry.OpName("HandleRequest"))
	} else {
		log.Debug("new transaction")
		transaction = sentry.StartTransaction(ctx, "events.HandleRequest", sentry.OpName("HandleRequest"))
	}
	defer transaction.Finish()

	span := transaction.StartChild("events.HandleRequest")
	defer span.Finish()

	ctx = span.Context()

	sqsClient, err := queue.NewSQSClient(ctx, queueURL)
	if err != nil {
		log.WithError(err).Error("failed to create SQS client")
		return err
	}

	config := app.GetEventsConfig{
		FifaClient: &go_fifa.Client{
			Client: &http.Client{
				Timeout: 5 * time.Second,
			},
		},
		QueueClient: &sqsClient,
		WebhookURL:  webhookURL,
	}

	for _, record := range event.Records {
		return app.GetEvents(ctx, &config, record)
	}
	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
