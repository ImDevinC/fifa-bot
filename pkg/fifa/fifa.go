package fifa

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/imdevinc/fifa-bot/pkg/queue"
	go_fifa "github.com/imdevinc/go-fifa"
)

func GetLiveMatches(ctx context.Context, fifaClient *go_fifa.Client) ([]queue.MatchOptions, error) {
	span := sentry.StartSpan(ctx, "function")
	defer span.Finish()
	span.Description = "fifa.GetLiveMatches"

	childSpan := sentry.StartSpan(ctx, "http")
	span.Description = "go-fifa.GetCurrentMatches"
	matches, err := fifaClient.GetCurrentMatches()
	if err != nil {
		sentry.CaptureException(err)
		childSpan.Finish()
		return nil, err
	}
	childSpan.Finish()
	var returnValue []queue.MatchOptions
	for _, m := range matches {
		returnValue = append(returnValue, queue.MatchOptions{
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
	NewEvents         []string
	Done              bool
	PendingEventFound bool
}

func GetMatchEvents(ctx context.Context, fifaClient *go_fifa.Client, opts *queue.MatchOptions) (MatchData, error) {
	span := sentry.StartSpan(ctx, "function")
	defer span.Finish()
	span.Description = "fifa.GetMatchEvents"
	span.SetTag("competitionId", opts.CompetitionId)
	span.SetTag("seasonId", opts.SeasonId)
	span.SetTag("stageId", opts.StageId)
	span.SetTag("matchId", opts.MatchId)
	span.SetTag("lastEvent", opts.LastEvent)

	ctx = span.Context()

	childSpan := sentry.StartSpan(ctx, "http")
	childSpan.Description = "go-fifa.GetMatchEvents"
	childSpan.SetTag("competitionId", opts.CompetitionId)
	childSpan.SetTag("seasonId", opts.SeasonId)
	childSpan.SetTag("stageId", opts.StageId)
	childSpan.SetTag("matchId", opts.MatchId)

	returnData := MatchData{
		PendingEventFound: false,
		Done:              false,
		NewEvents:         []string{},
	}

	events, err := fifaClient.GetMatchEvents(&go_fifa.GetMatchEventOptions{
		CompetitionId: opts.CompetitionId,
		SeasonId:      opts.SeasonId,
		StageId:       opts.StageId,
		MatchId:       opts.MatchId,
	})
	if err != nil {
		sentry.CaptureException(err)
		childSpan.Finish()
		return returnData, fmt.Errorf("failed to get match events. %w", err)
	}
	childSpan.Finish()
	var lastEventFound = false

	// Sort events by event ID
	sort.SliceStable(events.Events, func(i, j int) bool {
		firstId, err := strconv.Atoi(events.Events[i].Id)
		if err != nil {
			return true
		}
		secondId, err := strconv.Atoi(events.Events[j].Id)
		if err != nil {
			return true
		}
		return firstId < secondId
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
		if !lastEventFound {
			if evt.Id == opts.LastEvent || opts.LastEvent == "0" {
				lastEventFound = true
			}
			if opts.LastEvent != "0" {
				continue
			}
		}
		opts.LastEvent = evt.Id
		resp := processEvent(ctx, fifaClient, evt, opts)
		if resp == "" {
			continue
		}
		returnData.NewEvents = append(returnData.NewEvents, resp)
	}
	// If an event gets deleted, we may not find it above. In that case,
	// let's just reset to the most recent event
	if !returnData.PendingEventFound && opts.LastEvent != "0" && !lastEventFound && len(events.Events) > 0 {
		opts.LastEvent = events.Events[len(events.Events)-1].Id
	}
	return returnData, nil
}

func processEvent(ctx context.Context, fifaClient *go_fifa.Client, evt go_fifa.TimelineEvent, opts *queue.MatchOptions) string {
	span := sentry.StartSpan(ctx, "function")
	defer span.Finish()
	span.Description = "fifa.processEvents"
	span.SetTag("eventId", evt.Id)
	span.SetTag("eventType", fmt.Sprintf("%d", int(evt.Type)))

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
		suffix = fmt.Sprintf("%d %s %s : %s %s %d", evt.HomeGoals, opts.HomeTeamAbbrev, homeTeamFlag, awayTeamFlag, opts.AwayTeamAbbrev, evt.AwayGoals)
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
		suffix = fmt.Sprintf("%d %s %s : %s %s %d", evt.HomeGoals, opts.HomeTeamAbbrev, homeTeamFlag, awayTeamFlag, opts.AwayTeamAbbrev, evt.AwayGoals)
	case go_fifa.HalfEnd:
		prefix = ":clock1230:"
		suffix = fmt.Sprintf("%d %s %s : %s %s %d", evt.HomeGoals, opts.HomeTeamName, homeTeamFlag, awayTeamFlag, opts.AwayTeamName, evt.AwayGoals)
	case go_fifa.PenaltyMissed,
		go_fifa.PenaltyMissed2:
		prefix = ":no_entry_sign:"
	case go_fifa.PenaltyAwarded:
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
