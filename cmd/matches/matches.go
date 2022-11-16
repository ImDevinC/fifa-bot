package main

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	go_fifa "github.com/ImDevinC/go-fifa"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/getsentry/sentry-go"
	"github.com/imdevinc/fifa-bot/pkg/database"
	"github.com/imdevinc/fifa-bot/pkg/fifa"
	"github.com/imdevinc/fifa-bot/pkg/queue"
	log "github.com/sirupsen/logrus"
)

var release string = "development"

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

func HandleRequest(ctx context.Context) error {
	initLogging()

	queueURL := os.Getenv("QUEUE_URL")
	if queueURL == "" {
		log.Error("missing QUEUE_URL")
		return errors.New("missing QUEUE_URL")
	}

	tableName := os.Getenv("TABLE_NAME")
	if queueURL == "" {
		log.Error("missing TABLE_NAME")
		return errors.New("missing TABLE_NAME")
	}

	initSentry()
	sentry.Flush(2 * time.Second)

	span := sentry.StartSpan(ctx, "matches.HandleRequest", sentry.TransactionName("matches.HandleRequest"))
	defer span.Finish()

	spanCtx := span.Context()

	dbClient, err := database.NewDynamoClient(spanCtx, tableName)
	if err != nil {
		sentry.CaptureException(err)
		log.WithError(err).Error("failed to connect to database")
		return err
	}

	fifaClient := go_fifa.Client{}
	matches, err := fifa.GetLiveMatches(spanCtx, &fifaClient)
	if err != nil {
		sentry.CaptureException(err)
		log.WithError(err).Error("failed to get live matches")
		return err
	}
	var errWrap []string
	for _, m := range matches {
		err := dbClient.DoesMatchExist(spanCtx, &m)
		if !errors.Is(err, database.ErrMatchNotFound) {
			continue
		}
		if err != nil && !errors.Is(err, database.ErrMatchNotFound) {
			sentry.CaptureException(err)
			log.WithError(err).Error("failed to get match info from database")
			errWrap = append(errWrap, err.Error())
			continue
		}
		err = dbClient.AddMatch(spanCtx, &m)
		if err != nil {
			sentry.CaptureException(err)
			log.WithError(err).Error("failed to save match to database")
			errWrap = append(errWrap, err.Error())
			continue
		}
		m.LastEvent = "-1"
		err = queue.SendToQueue(spanCtx, queueURL, &m)
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

func main() {
	if _, exists := os.LookupEnv("AWS_LAMBDA_RUNTIME_API"); exists {
		lambda.Start(HandleRequest)
	} else {
		HandleRequest(context.TODO())
	}
}
