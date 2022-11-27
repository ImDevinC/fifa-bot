package app_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	go_fifa "github.com/ImDevinC/go-fifa"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/imdevinc/fifa-bot/pkg/app"
	"github.com/imdevinc/fifa-bot/pkg/queue"
)

func TestGetEvents(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fmt.Println(string(body))
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
				StringValue: aws.String("17"),
			},
			"SeasonId": {
				StringValue: aws.String("255711"),
			},
			"StageId": {
				StringValue: aws.String("285063"),
			},
			"MatchId": {
				StringValue: aws.String("400235480"),
			},
			"HomeTeamName": {
				StringValue: aws.String("Belgium"),
			},
			"AwayTeamName": {
				StringValue: aws.String("Morocco"),
			},
			"HomeTeamAbbrev": {
				StringValue: aws.String("BEL"),
			},
			"AwayTeamAbbrev": {
				StringValue: aws.String("MOC"),
			},
			"LastEvent": {
				StringValue: aws.String("0"),
			},
		},
	}
	err := app.GetEvents(context.TODO(), &config, event)
	if err != nil {
		t.Error(err)
	}
}
