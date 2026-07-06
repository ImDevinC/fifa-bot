package fifa_test

import (
	"context"
	"testing"

	"github.com/imdevinc/fifa-bot/pkg/fifa"
	"github.com/imdevinc/fifa-bot/pkg/models"
	go_fifa "github.com/imdevinc/go-fifa"
	"github.com/stretchr/testify/assert"
)

func TestLiveEvents(t *testing.T) {
	client := go_fifa.Client{}
	//	matchResp, err := fifa.GetLiveMatches(context.Background(), &client)
	//	if ok := assert.NoError(t, err); !ok {
	//		t.FailNow()
	//	}
	//	if ok := assert.Greater(t, len(matchResp), 0); !ok {
	//		t.FailNow()
	//	}
	m := models.Match{
		CompetitionId: "17",
		SeasonId:      "285023",
		StageId:       "289287",
		MatchId:       "400021512",
	}
	resp, err := fifa.GetMatchEvents(context.Background(), &client, &m)
	if ok := assert.NoError(t, err); !ok {
		t.Fail()
	}
	emptySkipSet := make(map[go_fifa.MatchEvent]bool)
	for _, e := range resp.NewEvents {
		msg := fifa.ProcessEvent(context.Background(), e, &m, emptySkipSet)
		if len(msg) == 0 {
			continue
		}
		t.Log(msg)
	}
}
