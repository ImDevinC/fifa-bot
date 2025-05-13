package fifa_test

import (
	"context"
	"testing"

	"github.com/imdevinc/fifa-bot/pkg/fifa"
	go_fifa "github.com/imdevinc/go-fifa"
	"github.com/stretchr/testify/assert"
)

func TestLiveEvents(t *testing.T) {
	client := go_fifa.Client{}
	matchResp, err := fifa.GetLiveMatches(context.Background(), &client)
	if ok := assert.NoError(t, err); !ok {
		t.FailNow()
	}
	if ok := assert.Greater(t, len(matchResp), 0); !ok {
		t.FailNow()
	}
	for _, m := range matchResp {
		resp, err := fifa.GetMatchEvents(context.Background(), &client, &m)
		if ok := assert.NoError(t, err); !ok {
			t.Fail()
		}
		for _, e := range resp.NewEvents {
			msg := fifa.ProcessEvent(context.Background(), e, &m)
			if len(msg) == 0 {
				continue
			}
			t.Log(msg)
		}
	}
}
