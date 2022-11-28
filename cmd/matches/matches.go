package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/getsentry/sentry-go"
	"github.com/imdevinc/fifa-bot/pkg/app"
	"github.com/imdevinc/fifa-bot/pkg/database"
	"github.com/imdevinc/fifa-bot/pkg/helper"
	"github.com/imdevinc/fifa-bot/pkg/queue"
	go_fifa "github.com/imdevinc/go-fifa"
	log "github.com/sirupsen/logrus"
)

var release string = "development"

func HandleRequest(ctx context.Context) error {
	logLevel, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		log.WithError(err).Warn("failed to parse LOG_LEVEL")
		logLevel = log.InfoLevel
	}
	helper.InitLogging(logLevel)
	tableName := os.Getenv("TABLE_NAME")
	if len(tableName) == 0 {
		log.Error("missing TABLE_NAME")
		return errors.New("missing TABLE_NAME")
	}

	queueURL := os.Getenv("QUEUE_URL")
	if len(queueURL) == 0 {
		log.Error("missing QUEUE_URL")
		return errors.New("missing QUEUE_URL")
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

	span := sentry.StartSpan(ctx, "function.aws", sentry.TransactionName("HandleRequest"))
	defer span.Finish()
	span.Description = "matches.HandleRequest"

	ctx = span.Context()

	sqsClient, err := queue.NewSQSClient(ctx, queueURL)
	if err != nil {
		log.WithError(err).Error("failed to create SQS client")
		return err
	}

	dynamoClient, err := database.NewDynamoClient(ctx, tableName)
	if err != nil {
		log.WithError(err).Error("failed to create dynamo client")
		return err
	}

	config := app.GetMatchesConfig{
		FifaClient: &go_fifa.Client{
			Client: &http.Client{
				Timeout: 5 * time.Second,
			},
		},
		QueueClient:    &sqsClient,
		DatabaseClient: &dynamoClient,
	}

	competitionId := os.Getenv("WATCH_COMPETITION")
	if len(competitionId) > 0 {
		config.CompetitionId = competitionId
	}

	defer sentry.Recover()
	return app.GetMatches(ctx, &config)
}

func main() {
	if _, exists := os.LookupEnv("AWS_LAMBDA_RUNTIME_API"); exists {
		lambda.Start(HandleRequest)
	} else {
		HandleRequest(context.Background())
	}
}
