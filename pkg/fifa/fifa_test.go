package fifa_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/imdevinc/fifa-bot/pkg/fifa"
	"github.com/imdevinc/fifa-bot/pkg/queue"
	go_fifa "github.com/imdevinc/go-fifa"
	"github.com/stretchr/testify/assert"
)

type TestClient struct{}

const eventList = `
{
	"IdStage": "1",
	"IdMatch": "2",
	"IdCompetition": "3",
	"IdSeason": "4",
	"IdGroup": "5",
	"Event": [
		{
			"EventId": "18185700000871",
			"IdTeam": "43960",
			"IdPlayer": "336098",
			"Timestamp": "2022-11-29T15:47:58.618Z",
			"MatchMinute": "45'+3'",
			"Period": 3,
			"HomeGoals": 1,
			"AwayGoals": 0,
			"Type": 12,
			"Qualifiers": [],
			"TypeLocalized": [
				{
					"Locale": "en-GB",
					"Description": "Attempt at Goal"
				}
			],
			"PositionX": 0.5585157,
			"PositionY": -0.24816446,
			"GoalGatePositionY": -0.21117647,
			"GoalGatePositionZ": 0.008,
			"HomePenaltyGoals": 0,
			"AwayPenaltyGoals": 0,
			"EventDescription": [
				{
					"Locale": "en-GB",
					"Description": "MEMPHIS (Netherlands) attempts an effort on goal."
				}
			]
		},
		{
			"EventId": "18185700000875",
			"Timestamp": "2022-11-29T15:48:06.788Z",
			"MatchMinute": "45'+4'",
			"Period": 3,
			"HomeGoals": 1,
			"AwayGoals": 0,
			"Type": 8,
			"Qualifiers": [],
			"TypeLocalized": [
				{
					"Locale": "en-GB",
					"Description": "End Time"
				}
			],
			"HomePenaltyGoals": 0,
			"AwayPenaltyGoals": 0,
			"EventDescription": [
				{
					"Locale": "en-GB",
					"Description": "The referee brings the first period to an end."
				}
			]
		},
		{
			"EventId": "18185700000873",
			"IdTeam": "43960",
			"IdPlayer": "336098",
			"Timestamp": "2022-11-29T15:47:58.618Z",
			"MatchMinute": "45'+3'",
			"Period": 3,
			"HomeGoals": 1,
			"AwayGoals": 0,
			"Type": 12,
			"Qualifiers": [],
			"TypeLocalized": [
				{
					"Locale": "en-GB",
					"Description": "Attempt at Goal"
				}
			],
			"PositionX": 0.5585157,
			"PositionY": -0.24816446,
			"GoalGatePositionY": -0.21117647,
			"GoalGatePositionZ": 0.008,
			"HomePenaltyGoals": 0,
			"AwayPenaltyGoals": 0,
			"EventDescription": [
				{
					"Locale": "en-GB",
					"Description": "MEMPHIS (Netherlands) attempts an effort on goal."
				}
			]
		}
	]
}
`

func (c *TestClient) Do(req *http.Request) (*http.Response, error) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Status:     http.StatusText(http.StatusOK),
		Body:       io.NopCloser(strings.NewReader(eventList)),
	}
	return resp, nil
}

var _ go_fifa.HTTPClient = (*TestClient)(nil)

func TestLastEvent(t *testing.T) {
	client := go_fifa.Client{
		Client: &TestClient{},
	}
	for i := 0; i < 100; i++ {
		opts := queue.MatchOptions{
			CompetitionId:  "3",
			SeasonId:       "4",
			StageId:        "1",
			MatchId:        "2",
			LastEvent:      "18185700000871",
			HomeTeamName:   "Netherlands",
			AwayTeamName:   "Qatar",
			HomeTeamAbbrev: "NED",
			AwayTeamAbbrev: "QAT",
		}

		_, err := fifa.GetMatchEvents(context.Background(), &client, &opts)
		if ok := assert.NoError(t, err); !ok {
			t.Fail()
		}
		if ok := assert.Equal(t, opts.LastEvent, "18185700000875"); !ok {
			t.Fail()
		}
	}
}

func TestLiveEvents(t *testing.T) {
	client := go_fifa.Client{}
	resp, err := fifa.GetMatchEvents(context.Background(), &client, &queue.MatchOptions{
		CompetitionId:  "17",
		SeasonId:       "255711",
		StageId:        "285063",
		MatchId:        "400235450",
		LastEvent:      "18185702501692",
		HomeTeamName:   "Netherlands",
		AwayTeamName:   "Qatar",
		HomeTeamAbbrev: "NED",
		AwayTeamAbbrev: "QAT",
	})
	if ok := assert.NoError(t, err); !ok {
		t.Fail()
	}
	if ok := assert.Len(t, resp.NewEvents, 150); !ok {
		t.Fail()
	}
}
