package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/imdevinc/fifa-bot/pkg/database"
	"github.com/imdevinc/fifa-bot/pkg/fifa"
	"github.com/imdevinc/fifa-bot/pkg/models"
	go_fifa "github.com/imdevinc/go-fifa"
	"golang.org/x/sync/errgroup"
)

type app struct {
	db               database.Database
	fifa             *go_fifa.Client
	slackWebhookURL  string
	CompetitionId    string
	matches          map[string]models.Match
	sleepTimeSeconds time.Duration
	eventsToSkip     map[go_fifa.MatchEvent]bool
	matchMutex       *sync.Mutex
	sentryEnabled    bool
}

func New(db database.Database, fifa *go_fifa.Client, slackWebhookURL string, competitionId string, sleepTimeSeconds int, eventsToSkip map[go_fifa.MatchEvent]bool, sentryEnabled bool) *app {
	if eventsToSkip == nil {
		eventsToSkip = make(map[go_fifa.MatchEvent]bool)
	}
	return &app{
		db:               db,
		fifa:             fifa,
		slackWebhookURL:  slackWebhookURL,
		CompetitionId:    competitionId,
		matches:          map[string]models.Match{},
		sleepTimeSeconds: time.Duration(sleepTimeSeconds),
		eventsToSkip:     eventsToSkip,
		matchMutex:       &sync.Mutex{},
		sentryEnabled:    sentryEnabled,
	}
}

func (a *app) Run(ctx context.Context) error {
	matches, err := a.db.GetAllMatches(ctx)
	if err != nil {
		slog.Error("failed to get matches from database", "error", err)
	} else {
		a.matchMutex.Lock()
		for _, m := range matches {
			a.matches[m.MatchId] = m
		}
		a.matchMutex.Unlock()
	}
	slog.Debug("starting app loop")
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			ctx := context.Background()
			slog.Debug("getting matches")
			err := a.getMatches(ctx)
			if err != nil {
				slog.Error("failed to get matches", "error", err)
			}
			slog.Debug("getting events")
			err = a.monitorEvents(ctx)
			if err != nil {
				slog.Error("failed to update events", "error", err)
			}
			time.Sleep(a.sleepTimeSeconds * time.Second)
		}
	}
}

func (a *app) getMatches(ctx context.Context) error {
	slog.Debug("getting matches from FIFA")
	matches, err := fifa.GetLiveMatches(ctx, a.fifa)
	if err != nil {
		return fmt.Errorf("failed to get live matches from FIFA. %w", err)
	}
	for _, m := range matches {
		if len(a.CompetitionId) > 0 && m.CompetitionId != a.CompetitionId {
			continue
		}
		if _, exists := a.matches[m.MatchId]; exists {
			slog.Debug("match already exists in db, no need to add", "matchID", m.MatchId)
			continue
		}
		slog.Debug("adding match to database", "matchID", m.MatchId, "competitionId", m.CompetitionId, "seasonId", m.SeasonId, "stageId", m.StageId, "homeTeam", m.HomeTeamName, "awayTeam", m.AwayTeamName)
		err = a.db.AddMatch(ctx, m)
		if err != nil {
			return fmt.Errorf("failed to add match %s to database. %w", m.MatchId, err)
		}
		a.matches[m.MatchId] = m
	}
	return nil
}

func (a *app) monitorEvents(ctx context.Context) error {
	slog.Debug("starting event monitor")
	g, ctx := errgroup.WithContext(ctx)
	for _, match := range a.matches {
		g.Go(func() error {
			return a.processMatch(ctx, &match)
		})
	}
	err := g.Wait()
	if err != nil {
		return err
	}
	return nil
}

func (a *app) processMatch(ctx context.Context, match *models.Match) error {
	slog.Debug("getting match", "matchId", match.MatchId)
	matchData, err := fifa.GetMatchEvents(ctx, a.fifa, match)
	if err != nil {
		return fmt.Errorf("failed to get match %s events from FIFA. %w", match.MatchId, err)
	}
	ids, messages := a.findNewEvents(ctx, match.Events, matchData.NewEvents, match)
	existingEvents := match.Events
	allEvents := append(existingEvents, ids...)
	if len(allEvents) > len(existingEvents) {
		match.Events = allEvents
		err = a.db.UpdateMatch(ctx, *match)
		if err != nil {
			return fmt.Errorf("failed to save match %s events to the database. %w", match.MatchId, err)
		}
		a.matchMutex.Lock()
		a.matches[match.MatchId] = *match
		a.matchMutex.Unlock()
	}

	if len(messages) > 0 {
		slog.Debug("sending messages for match", "matchId", match.MatchId)
	}
	err = a.sendEventsToSlack(ctx, messages)
	if err != nil {
		slog.Error("failed to send message to slack", "error", err)
		return fmt.Errorf("failed to send events to Slack. %w", err)
	}
	if !matchData.Done {
		return nil
	}
	slog.Debug("match is done", "matchId", match.MatchId)
	err = a.db.DeleteMatch(ctx, match.MatchId)
	if err != nil {
		return fmt.Errorf("failed to delete match %s. %w", match.MatchId, err)
	}
	a.matchMutex.Lock()
	delete(a.matches, match.MatchId)
	a.matchMutex.Unlock()
	return nil
}

func (a *app) sendEventsToSlack(ctx context.Context, events []string) error {
	for _, evt := range events {
		if strings.TrimSpace(evt) == "" {
			continue
		}
		slog.Debug("sending message to slack", "message", evt)
		payload := models.SlackMessage{Text: evt}
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		resp, err := http.Post(a.slackWebhookURL, "application/json", bytes.NewReader(b))
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return errors.New(resp.Status)
		}
	}
	return nil
}

func (a *app) findNewEvents(ctx context.Context, existingEvents []string, newEvents []go_fifa.TimelineEvent, opts *models.Match) ([]string, []string) {
	eventMsgs := []string{}
	eventIds := []string{}

	for _, event := range newEvents {
		eventFound := slices.Contains(existingEvents, event.Id)
		if eventFound {
			continue
		}
		result := fifa.ProcessEvent(ctx, event, opts, a.eventsToSkip)

		// Unknown event types are captured to Sentry instead of sent to Slack
		if result.IsUnknown {
			slog.Warn("unknown event type detected", "eventId", event.Id, "eventType", event.Type, "matchId", opts.MatchId)
			if a.sentryEnabled {
				a.captureUnknownEvent(event, opts)
			}
			eventIds = append(eventIds, event.Id)
			continue
		}

		slog.Debug("found new event", "eventId", event.Id, "message", result.SlackMessage)
		eventIds = append(eventIds, event.Id)
		eventMsgs = append(eventMsgs, result.SlackMessage)
	}
	return eventIds, eventMsgs
}

func (a *app) captureUnknownEvent(evt go_fifa.TimelineEvent, match *models.Match) {
	eventJSON, err := json.Marshal(evt)
	if err != nil {
		slog.Error("failed to marshal unknown event to JSON", "error", err)
		return
	}

	matchJSON, err := json.Marshal(match)
	if err != nil {
		slog.Error("failed to marshal match info to JSON", "error", err)
		return
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetTags(map[string]string{
			"event_type":     fmt.Sprintf("%d", evt.Type),
			"match_id":       match.MatchId,
			"stage_id":       match.StageId,
			"season_id":      match.SeasonId,
			"competition_id": match.CompetitionId,
			"home_team":      match.HomeTeamAbbrev,
			"away_team":      match.AwayTeamAbbrev,
		})
		scope.SetExtra("full_event_json", string(eventJSON))
		scope.SetExtra("match_info_json", string(matchJSON))
		sentry.CaptureMessage(fmt.Sprintf("Unknown event type: %d in match %s (%s vs %s)",
			evt.Type, match.MatchId, match.HomeTeamAbbrev, match.AwayTeamAbbrev))
	})
}
