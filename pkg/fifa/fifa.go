package fifa

import (
	"context"
	"fmt"
	"strings"

	go_fifa "github.com/ImDevinC/go-fifa"
	"github.com/imdevinc/fifa-bot/pkg/queue"
)

func GetLiveMatches(ctx context.Context, fifaClient *go_fifa.Client) ([]queue.MatchOptions, error) {
	matches, err := fifaClient.GetCurrentMatches()
	if err != nil {
		return nil, err
	}
	var returnValue []queue.MatchOptions
	for _, m := range matches {
		returnValue = append(returnValue, queue.MatchOptions{
			CompetitionId:  m.CompetitionId,
			SeasonId:       m.SeasonId,
			StageId:        m.StageId,
			MatchId:        m.Id,
			LastEvent:      "",
			HomeTeamName:   m.HomeTeam.Name[0].Description,
			AwayTeamName:   m.AwayTeam.Name[0].Description,
			HomeTeamAbbrev: m.HomeTeam.Abbreviation,
			AwayTeamAbbrev: m.AwayTeam.Abbreviation,
		})
	}
	return returnValue, nil
}

func GetMatchEvents(ctx context.Context, fifaClient *go_fifa.Client, opts *queue.MatchOptions) ([]string, bool, error) {
	events, err := fifaClient.GetMatchEvents(&go_fifa.GetMatchEventOptions{
		CompetitionId: opts.CompetitionId,
		SeasonId:      opts.SeasonId,
		StageId:       opts.StageId,
		MatchId:       opts.MatchId,
	})
	if err != nil {
		return nil, false, fmt.Errorf("failed to get match events. %w", err)
	}
	var returnValue []string
	var lastEventFound = false
	var matchOver = false
	for _, evt := range events.Events {
		if evt.Type == go_fifa.MatchEnd {
			matchOver = true
		}
		if !lastEventFound {
			if evt.Id == opts.LastEvent || opts.LastEvent == "0" {
				lastEventFound = true
			}
			if opts.LastEvent != "0" {
				continue
			}
		}
		opts.LastEvent = evt.Id
		resp := processEvent(ctx, evt)
		if resp == "" {
			continue
		}
		returnValue = append(returnValue, resp)
	}
	return returnValue, matchOver, nil
}

func processEvent(ctx context.Context, evt go_fifa.EventResponse) string {
	if _, exists := eventsToSkip[evt.Type]; exists {
		return ""
	}
	if len(evt.EventDescription) == 0 {
		return fmt.Sprintf("[EVENTINFO] Need info for event type: %d", evt.Type)

	}
	prefix := ""
	suffix := ""
	switch evt.Type {
	case go_fifa.GoalScore,
		go_fifa.OwnGoal,
		go_fifa.PenaltyGoal:
		prefix = ":soccer:"
	case go_fifa.YellowCard,
		go_fifa.DoubleYellow:
		prefix = ":large_yellow_square:"
	case go_fifa.RedCard:
		prefix = ":large_red_square:"
	case go_fifa.Substitution:
		prefix = ":arrows_counterclockwise:"
	case go_fifa.MatchStart,
		go_fifa.MatchEnd:
		prefix = ":clock12:"
	case go_fifa.HalfEnd:
		prefix = ":clock1230:"
	case go_fifa.PenaltyMissed,
		go_fifa.PenaltyMissed2:
		prefix = ":no_entry_sign:"
	}
	fmt.Printf("[DEBUG] (%d) %s\n", evt.Type, evt.EventDescription[0].Description)
	msg := fmt.Sprintf("%s %s %s", prefix, evt.EventDescription[0].Description, suffix)
	return strings.TrimSpace(msg)
}
