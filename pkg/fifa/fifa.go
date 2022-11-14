package fifa

import (
	"context"
	"errors"
	"fmt"
	"strings"

	go_fifa "github.com/ImDevinC/go-fifa"
	"github.com/imdevinc/fifa-bot/pkg/executor"
)

var ErrMatchNotFound = errors.New("match not found")

func GetNewMatches(ctx context.Context, fifaClient *go_fifa.Client, sfnClient *executor.Client) error {
	matches, err := fifaClient.GetCurrentMatches()
	if err != nil {
		return err
	}
	for _, m := range matches {
		err = checkForExistingMatch(ctx, &m)
		if err != nil && !errors.Is(err, ErrMatchNotFound) {
			return fmt.Errorf("failed checking for existing match. %w", err)
		}
		err = startWatchingMatch(ctx, &m, sfnClient)
		if err != nil {
			return fmt.Errorf("failed trying to watch match. %w", err)
		}
	}
	return nil
}

func GetMatchEvents(ctx context.Context, fifaClient *go_fifa.Client, opts *executor.StartMatchOptions) ([]string, error) {
	events, err := fifaClient.GetMatchEvents(&go_fifa.GetMatchEventOptions{
		CompetitionId: opts.CompetitionId,
		SeasonId:      opts.SeasonId,
		StageId:       opts.StageId,
		MatchId:       opts.MatchId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get match events. %w", err)
	}
	var returnValue []string
	for _, evt := range events.Events {
		resp := processEvent(ctx, evt)
		if resp == "" {
			continue
		}
		returnValue = append(returnValue, resp)
	}
	return returnValue, nil
}

func checkForExistingMatch(ctx context.Context, m *go_fifa.MatchResponse) error {
	return ErrMatchNotFound
}

func startWatchingMatch(ctx context.Context, m *go_fifa.MatchResponse, sfnClient *executor.Client) error {
	err := sfnClient.StartMatch(ctx, &executor.StartMatchOptions{
		CompetitionId: m.CompetitionId,
		SeasonId:      m.SeasonId,
		StageId:       m.StageId,
		MatchId:       m.Id,
	})
	if err != nil {
		return err
	}
	return nil
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
	case go_fifa.YellowCard:
		prefix = ":yellowcard:"
	case go_fifa.RedCard,
		go_fifa.DoubleYellow:
		prefix = ":redcard:"
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
	msg := fmt.Sprintf("%s %s %s", prefix, evt.EventDescription[0].Description, suffix)
	return strings.TrimSpace(msg)
}
