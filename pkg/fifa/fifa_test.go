package fifa_test

import (
	"context"
	"testing"

	go_fifa "github.com/ImDevinC/go-fifa"
	"github.com/imdevinc/fifa-bot/pkg/executor"
	"github.com/imdevinc/fifa-bot/pkg/fifa"
)

func TestGetMatchEvents(t *testing.T) {
	client := go_fifa.Client{}
	resp, err := fifa.GetMatchEvents(context.TODO(), &client, &executor.StartMatchOptions{
		CompetitionId: "2000000005",
		SeasonId:      "400250052",
		StageId:       "b1ayaoa4q68n6464fy4orklqs",
		MatchId:       "3y748w6ppuxciynnoonrt9jx0",
	})
	if err != nil {
		t.Error(err)
	}
	if len(resp) == 0 {
		t.Error("no events found")
	}
}
