package app_test

// This test suite will communicate with dynamodb in an attempt to make new events
// make sure your config values are setup properly

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/imdevinc/fifa-bot/pkg/app"
	"github.com/imdevinc/fifa-bot/pkg/database"
	"github.com/imdevinc/fifa-bot/pkg/queue"
	go_fifa "github.com/imdevinc/go-fifa"
	"github.com/stretchr/testify/assert"
)

type TestClient struct{}

func (c *TestClient) Do(req *http.Request) (*http.Response, error) {
	tl := go_fifa.LiveResponse{
		Results: []go_fifa.LiveMatch{{
			MatchId:       "400235456",
			StageId:       "285063",
			SeasonId:      "255711",
			CompetitionId: "17",
			HomeTeam: go_fifa.Team{
				Name: []go_fifa.LocaleDescription{{
					Locale:      "en-GB",
					Description: "Unicorns",
				}},
				Abbreviation: "UNC",
			},
			AwayTeam: go_fifa.Team{
				Name: []go_fifa.LocaleDescription{{
					Locale:      "en-GB",
					Description: "Ligers",
				}},
				Abbreviation: "LGR",
			},
		}},
	}
	payload, _ := json.Marshal(tl)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Status:     http.StatusText(http.StatusOK),
		Body:       io.NopCloser(bytes.NewReader(payload)),
	}
	return resp, nil
}

var _ go_fifa.HTTPClient = (*TestClient)(nil)

func TestLiveMatch(t *testing.T) {
	if !liveTest {
		t.Skip("skip testing when not in livetest mode")
	}
	db, err := database.NewDynamoClient(context.Background(), "fifa-bot")
	if ok := assert.NoError(t, err); !ok {
		t.FailNow()
	}
	config := app.GetMatchesConfig{
		DatabaseClient: &db,
		QueueClient:    &queue.Client{Queue: &TestQueue{}},
		FifaClient:     &go_fifa.Client{Client: &TestClient{}},
	}
	err = app.GetMatches(context.TODO(), &config)
	if ok := assert.NoError(t, err); !ok {
		t.FailNow()
	}

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		t.Log(string(body))
		w.Write([]byte(`"OK`))
	}))
	defer s.Close()

	evtConfig := app.GetEventsConfig{
		DatabaseClient: &db,
		QueueClient:    &queue.Client{Queue: &TestQueue{}},
		FifaClient:     &go_fifa.Client{},
		WebhookURL:     s.URL,
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
				StringValue: aws.String("400235456"),
			},
			"HomeTeamName": {
				StringValue: aws.String("Iran"),
			},
			"AwayTeamName": {
				StringValue: aws.String("United States"),
			},
			"HomeTeamAbbrev": {
				StringValue: aws.String("IRN"),
			},
			"AwayTeamAbbrev": {
				StringValue: aws.String("USA"),
			},
			"LastEvent": {
				StringValue: aws.String("0"),
			},
		},
	}
	err = app.GetEvents(context.Background(), &evtConfig, event)
	if ok := assert.NoError(t, err); !ok {
		t.Fail()
	}

}
