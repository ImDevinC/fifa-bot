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
	matchMutex       *sync.Mutex
}

func New(db database.Database, fifa *go_fifa.Client, slackWebhookURL string, competitionId string, sleepTimeSeconds int) *app {
	return &app{
		db:               db,
		fifa:             fifa,
		slackWebhookURL:  slackWebhookURL,
		CompetitionId:    competitionId,
		matches:          map[string]models.Match{},
		sleepTimeSeconds: time.Duration(sleepTimeSeconds),
	}
}

func (a *app) Run(ctx context.Context) error {
	matches, err := a.db.GetAllMatches(ctx)
	if err != nil {
		slog.Error("failed to get matches from database", "error", err)
	} else {
		for _, m := range matches {
			a.matches[m.MatchId] = m
		}
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
		slog.Debug("adding match to database", "matchID", m.MatchId)
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
	ids, messages := findNewEvents(ctx, match.Events, matchData.NewEvents, match)
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

func findNewEvents(ctx context.Context, existingEvents []string, newEvents []go_fifa.TimelineEvent, opts *models.Match) ([]string, []string) {
	eventMsgs := []string{}
	eventIds := []string{}

	for _, event := range newEvents {
		eventFound := slices.Contains(existingEvents, event.Id)
		if eventFound {
			continue
		}
		slog.Debug("found new event", "eventId", event.Id, "message", fifa.ProcessEvent(ctx, event, opts))
		eventIds = append(eventIds, event.Id)
		eventMsgs = append(eventMsgs, fifa.ProcessEvent(ctx, event, opts))
	}
	return eventIds, eventMsgs
}
