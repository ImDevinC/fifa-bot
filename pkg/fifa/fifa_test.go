package fifa_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/imdevinc/fifa-bot/pkg/fifa"
	"github.com/imdevinc/fifa-bot/pkg/models"
	go_fifa "github.com/imdevinc/go-fifa"
	"github.com/stretchr/testify/assert"
)

func TestLiveEvents(t *testing.T) {
	client := go_fifa.Client{}
	m := models.Match{
		CompetitionId:  "17",
		SeasonId:       "285023",
		StageId:        "289287",
		MatchId:        "400021528",
		HomeTeamID:     "43971",
		AwayTeamID:     "43926",
		HomeTeamAbbrev: "ARG",
		AwayTeamAbbrev: "EGY",
	}
	resp, err := fifa.GetMatchEvents(context.Background(), &client, &m)
	if ok := assert.NoError(t, err); !ok {
		t.Fail()
	}
	evts, _ := json.Marshal(resp.NewEvents)
	os.WriteFile("events.json", evts, 0644)
	emptySkipSet := make(map[go_fifa.MatchEvent]bool)
	for _, e := range resp.NewEvents {
		result := fifa.ProcessEvent(context.Background(), e, &m, emptySkipSet)
		if len(result.SlackMessage) == 0 {
			continue
		}
		t.Log(result.SlackMessage)
	}
}

func TestCustomEvents(t *testing.T) {
	match := models.Match{
		CompetitionId:  "17",
		SeasonId:       "285023",
		StageId:        "289287",
		MatchId:        "400021535",
		HomeTeamID:     "43971",
		AwayTeamID:     "43926",
		HomeTeamAbbrev: "SUI",
		AwayTeamAbbrev: "COL",
	}
	emptySkipSet := make(map[go_fifa.MatchEvent]bool)
	events := []go_fifa.TimelineEvent{
		{
			Type:        go_fifa.GoalScore,
			Period:      go_fifa.ShootoutPeriod,
			TeamId:      match.AwayTeamID,
			Description: []go_fifa.LocaleDescription{{Locale: "en-GB", Description: "Player one scores!"}},
		},
		{
			Type:        go_fifa.PenaltyMissed,
			Period:      go_fifa.ShootoutPeriod,
			TeamId:      match.HomeTeamID,
			Description: []go_fifa.LocaleDescription{{Locale: "en-GB", Description: "Player two missed"}},
		},
		{
			Type:        go_fifa.GoalScore,
			Period:      go_fifa.ShootoutPeriod,
			TeamId:      match.AwayTeamID,
			Description: []go_fifa.LocaleDescription{{Locale: "en-GB", Description: "Player one scores!"}},
		},
		{
			Type:        go_fifa.PenaltyMissed,
			Period:      go_fifa.ShootoutPeriod,
			TeamId:      match.HomeTeamID,
			Description: []go_fifa.LocaleDescription{{Locale: "en-GB", Description: "Player two missed"}},
		},
		{
			Type:        go_fifa.GoalScore,
			Period:      go_fifa.ShootoutPeriod,
			TeamId:      match.AwayTeamID,
			Description: []go_fifa.LocaleDescription{{Locale: "en-GB", Description: "Player one scores!"}},
		},
		{
			Type:        go_fifa.PenaltyMissed,
			Period:      go_fifa.ShootoutPeriod,
			TeamId:      match.HomeTeamID,
			Description: []go_fifa.LocaleDescription{{Locale: "en-GB", Description: "Player two missed"}},
		},
		{
			Type:        go_fifa.GoalScore,
			Period:      go_fifa.ShootoutPeriod,
			TeamId:      match.AwayTeamID,
			Description: []go_fifa.LocaleDescription{{Locale: "en-GB", Description: "Player one scores!"}},
		},
		{
			Type:        go_fifa.PenaltyMissed,
			Period:      go_fifa.ShootoutPeriod,
			TeamId:      match.HomeTeamID,
			Description: []go_fifa.LocaleDescription{{Locale: "en-GB", Description: "Player two missed"}},
		},
		{
			Type:        go_fifa.GoalScore,
			Period:      go_fifa.ShootoutPeriod,
			TeamId:      match.AwayTeamID,
			Description: []go_fifa.LocaleDescription{{Locale: "en-GB", Description: "Player one scores!"}},
		},
		{
			Type:        go_fifa.PenaltyMissed,
			Period:      go_fifa.ShootoutPeriod,
			TeamId:      match.HomeTeamID,
			Description: []go_fifa.LocaleDescription{{Locale: "en-GB", Description: "Player two missed"}},
		},
		{
			Type:        go_fifa.PenaltyMissed2,
			Period:      go_fifa.ShootoutPeriod,
			TeamId:      match.AwayTeamID,
			Description: []go_fifa.LocaleDescription{{Locale: "en-GB", Description: "Player one misses"}},
		},
		{
			Type:        go_fifa.GoalScore,
			Period:      go_fifa.ShootoutPeriod,
			TeamId:      match.HomeTeamID,
			Description: []go_fifa.LocaleDescription{{Locale: "en-GB", Description: "Player two scores"}},
		},
		{
			Type:        go_fifa.GoalScore,
			Period:      go_fifa.ShootoutPeriod,
			TeamId:      match.AwayTeamID,
			Description: []go_fifa.LocaleDescription{{Locale: "en-GB", Description: "Player one scores!"}},
		},
		{
			Type:        go_fifa.PenaltyMissed,
			Period:      go_fifa.ShootoutPeriod,
			TeamId:      match.HomeTeamID,
			Description: []go_fifa.LocaleDescription{{Locale: "en-GB", Description: "Player two missed"}},
		},
	}
	for _, evt := range events {
		result := fifa.ProcessEvent(context.Background(), evt, &match, emptySkipSet)
		if len(result.SlackMessage) == 0 {
			continue
		}
		t.Log(result.SlackMessage)
	}
}
