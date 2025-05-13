package fifa

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/imdevinc/fifa-bot/pkg/models"
	go_fifa "github.com/imdevinc/go-fifa"
)

func GetLiveMatches(ctx context.Context, fifaClient *go_fifa.Client) ([]models.Match, error) {
	matches, err := fifaClient.GetCurrentMatches()
	if err != nil {
		return nil, err
	}
	returnValue := []models.Match{}
	for _, m := range matches {
		returnValue = append(returnValue, models.Match{
			Events:         []string{},
			CompetitionId:  m.CompetitionId,
			SeasonId:       m.SeasonId,
			StageId:        m.StageId,
			MatchId:        m.MatchId,
			LastEvent:      "",
			HomeTeamName:   m.HomeTeam.Name[0].Description,
			AwayTeamName:   m.AwayTeam.Name[0].Description,
			HomeTeamAbbrev: m.HomeTeam.Abbreviation,
			AwayTeamAbbrev: m.AwayTeam.Abbreviation,
		})
	}
	return returnValue, nil
}

type MatchData struct {
	NewEvents         []go_fifa.TimelineEvent
	Done              bool
	PendingEventFound bool
}

func GetMatchEvents(ctx context.Context, fifaClient *go_fifa.Client, opts *models.Match) (MatchData, error) {
	returnData := MatchData{
		PendingEventFound: false,
		Done:              false,
		NewEvents:         []go_fifa.TimelineEvent{},
	}

	events, err := fifaClient.GetMatchEvents(&go_fifa.GetMatchEventOptions{
		CompetitionId: opts.CompetitionId,
		SeasonId:      opts.SeasonId,
		StageId:       opts.StageId,
		MatchId:       opts.MatchId,
	})
	if err != nil {
		return returnData, fmt.Errorf("failed to get match events. %w", err)
	}
	var lastEventFound = false

	// Sort events by event ID
	sort.SliceStable(events.Events, func(i, j int) bool {
		return events.Events[i].Timestamp.Before(events.Events[j].Timestamp)
	})

	// -1 means the event just came over from the match watcher
	if opts.LastEvent == "-1" {
		opts.LastEvent = "0"
	}
	for _, evt := range events.Events {
		if evt.Type == go_fifa.Pending {
			returnData.PendingEventFound = true
			break // Found a pending type, and we want to wait for it to update
		}
		if evt.Type == go_fifa.MatchEnd {
			returnData.Done = true
		}
		opts.LastEvent = evt.Id
		returnData.NewEvents = append(returnData.NewEvents, evt)
	}
	// If an event gets deleted, we may not find it above. In that case,
	// let's just reset to the most recent event
	if !returnData.PendingEventFound && opts.LastEvent != "0" && !lastEventFound && len(events.Events) > 0 {
		opts.LastEvent = events.Events[len(events.Events)-1].Id
	}
	return returnData, nil
}

func ProcessEvent(ctx context.Context, evt go_fifa.TimelineEvent, opts *models.Match) string {
	if _, exists := eventsToSkip[evt.Type]; exists {
		return ""
	}

	prefix := ""
	suffix := ""
	homeTeamFlag := flagEmojis[opts.HomeTeamAbbrev]
	awayTeamFlag := flagEmojis[opts.AwayTeamAbbrev]
	switch evt.Type {
	case go_fifa.GoalScore,
		go_fifa.OwnGoal,
		go_fifa.PenaltyGoal:
		prefix = ":soccer:"
		if evt.Period == go_fifa.ShootoutPeriod {
			suffix = fmt.Sprintf("%d (%d) %s %s : %s %s (%d) %d", evt.HomeGoals, evt.HomePenaltyGoals, opts.HomeTeamAbbrev, homeTeamFlag, awayTeamFlag, opts.AwayTeamAbbrev, evt.AwayPenaltyGoals, evt.AwayGoals)
		} else {
			suffix = fmt.Sprintf("%d %s %s : %s %s %d", evt.HomeGoals, opts.HomeTeamAbbrev, homeTeamFlag, awayTeamFlag, opts.AwayTeamAbbrev, evt.AwayGoals)
		}
	case go_fifa.YellowCard,
		go_fifa.DoubleYellow:
		prefix = ":large_yellow_square:"
	case go_fifa.RedCard:
		prefix = ":large_red_square:"
	case go_fifa.Substitution:
		prefix = ":arrows_counterclockwise:"
	case go_fifa.MatchStart:
		prefix = ":clock12:"
		suffix = fmt.Sprintf("%s %s vs %s %s", opts.HomeTeamName, homeTeamFlag, awayTeamFlag, opts.AwayTeamName)
	case go_fifa.MatchEnd:
		prefix = ":clock12:"
		if evt.Period == go_fifa.ShootoutPeriod {
			suffix = fmt.Sprintf("%d (%d) %s %s : %s %s (%d) %d", evt.HomeGoals, evt.HomePenaltyGoals, opts.HomeTeamAbbrev, homeTeamFlag, awayTeamFlag, opts.AwayTeamAbbrev, evt.AwayPenaltyGoals, evt.AwayGoals)
		} else {
			suffix = fmt.Sprintf("%d %s %s : %s %s %d", evt.HomeGoals, opts.HomeTeamAbbrev, homeTeamFlag, awayTeamFlag, opts.AwayTeamAbbrev, evt.AwayGoals)
		}
	case go_fifa.HalfEnd:
		prefix = ":clock1230:"
		if evt.Period == go_fifa.ShootoutPeriod {
			suffix = fmt.Sprintf("%d (%d) %s %s : %s %s (%d) %d", evt.HomeGoals, evt.HomePenaltyGoals, opts.HomeTeamAbbrev, homeTeamFlag, awayTeamFlag, opts.AwayTeamAbbrev, evt.AwayPenaltyGoals, evt.AwayGoals)
		} else {
			suffix = fmt.Sprintf("%d %s %s : %s %s %d", evt.HomeGoals, opts.HomeTeamAbbrev, homeTeamFlag, awayTeamFlag, opts.AwayTeamAbbrev, evt.AwayGoals)
		}
	case go_fifa.PenaltyMissed,
		go_fifa.PenaltyMissed2:
		prefix = ":no_entry_sign:"
	case go_fifa.PenaltyAwarded:
		// This causes some spam messaging, so skip during shootouts
		if evt.Period == go_fifa.ShootoutPeriod {
			return ""
		}
		prefix = "Penalty awarded!"
	}
	var msg string
	if len(evt.Description) > 0 {
		msg = fmt.Sprintf("%s %s %s", prefix, evt.Description[0].Description, suffix)
	} else {
		msg = fmt.Sprintf("%s %s", prefix, suffix)
	}

	msg = strings.TrimSpace(msg)

	if len(msg) == 0 {
		msg = fmt.Sprintf("[EVENTINFO] Need info for event type: %d", evt.Type)
	} else {
		msg = fmt.Sprintf("%s %s", evt.MatchMinute, msg)
	}

	return msg
}
