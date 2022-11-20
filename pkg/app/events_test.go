package app_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	go_fifa "github.com/ImDevinC/go-fifa"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/getsentry/sentry-go"
	"github.com/imdevinc/fifa-bot/pkg/app"
	"github.com/imdevinc/fifa-bot/pkg/helper"
	"github.com/imdevinc/fifa-bot/pkg/queue"
	"github.com/joho/godotenv"
)

func TestGetEvents(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	err = helper.InitSentry(helper.SentryConfig{
		DSN:             os.Getenv("SENTRY_DSN"),
		TraceSampleRate: .5,
		Release:         "development",
		Debug:           true,
	})

	ctx := context.TODO()
	transaction := sentry.StartTransaction(ctx, "events.HandleRequest", sentry.OpName("HandleRequest"))

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	defer transaction.Finish()
	span := transaction.StartChild("events.HandleRequest")
	defer span.Finish()

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`"OK`))
	}))
	defer s.Close()

	config := app.GetEventsConfig{
		FifaClient:  &go_fifa.Client{},
		QueueClient: &queue.Client{Queue: &TestQueue{}},
		WebhookURL:  s.URL,
	}
	event := events.SQSMessage{
		MessageAttributes: map[string]events.SQSMessageAttribute{
			"CompetitionId": {
				StringValue: aws.String("cesdwwnxbc5fmajgroc0hqzy2"),
			},
			"SeasonId": {
				StringValue: aws.String("40sncpbsyexdrmedcwjz1j0gk"),
			},
			"StageId": {
				StringValue: aws.String("5w0vi7wp50objhjfn51o5ck5w"),
			},
			"MatchId": {
				StringValue: aws.String("3qxv1fe65nezrara3zsm5s55g"),
			},
			"HomeTeamName": {
				StringValue: aws.String("Albania"),
			},
			"AwayTeamName": {
				StringValue: aws.String("Italy"),
			},
			"HomeTeamAbbrev": {
				StringValue: aws.String("ALB"),
			},
			"AwayTeamAbbrev": {
				StringValue: aws.String("ITA"),
			},
			"LastEvent": {
				StringValue: aws.String("0"),
			},
		},
	}
	err = app.GetEvents(span.Context(), &config, event)
	if err != nil {
		t.Error(err)
	}
}
