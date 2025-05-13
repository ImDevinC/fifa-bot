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
	"time"

	"github.com/imdevinc/fifa-bot/pkg/database"
	"github.com/imdevinc/fifa-bot/pkg/fifa"
	"github.com/imdevinc/fifa-bot/pkg/models"
	go_fifa "github.com/imdevinc/go-fifa"
	"golang.org/x/sync/errgroup"
)

type app struct {
	db              database.Database
	fifa            *go_fifa.Client
	slackWebhookURL string
	CompetitionId   string
}

func New(db database.Database, fifa *go_fifa.Client, slackWebhookURL string, competitionId string) *app {
	return &app{
		db:              db,
		fifa:            fifa,
		slackWebhookURL: slackWebhookURL,
		CompetitionId:   competitionId,
	}
}

func (a *app) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			slog.Debug("getting matches")
			err := a.getMatches()
			if err != nil {
				slog.Error("failed to get matches", "error", err)
			}
			slog.Debug("getting events")
			err = a.monitorEvents()
			if err != nil {
				slog.Error("failed to update events", "error", err)
			}
			time.Sleep(10 * time.Second)
		}
	}
}

func (a *app) getMatches() error {
	ctx := context.Background()
	existingMatches, err := a.db.GetAllMatches(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get matches from database. %w", err)
	}
	matches, err := fifa.GetLiveMatches(ctx, a.fifa)
	if err != nil {
		return fmt.Errorf("failed to get live matches from FIFA. %w", err)
	}
	for _, m := range matches {
		if len(a.CompetitionId) > 0 && m.CompetitionId != a.CompetitionId {
			continue
		}
		if slices.Contains(existingMatches, m.MatchId) {
			continue
		}
		err = a.db.AddMatch(ctx, m)
		if err != nil {
			return fmt.Errorf("failed to add match %s to database. %w", m.MatchId, err)
		}
	}
	return nil
}

func (a *app) monitorEvents() error {
	ctx := context.Background()
	matches, err := a.db.GetAllMatches(ctx)
	if err != nil {
		return fmt.Errorf("failed to get matches. %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)
	for _, m := range matches {
		g.Go(func() error {
			return a.processMatch(ctx, m)
		})
	}
	err = g.Wait()
	if err != nil {
		return err
	}
	return nil
}

func (a *app) processMatch(ctx context.Context, matchID string) error {
	slog.Debug("getting match", "matchId", matchID)
	match, err := a.db.GetMatch(ctx, matchID)
	if err != nil {
		return fmt.Errorf("failed to get match %s from database. %w", matchID, err)
	}
	matchData, err := fifa.GetMatchEvents(ctx, a.fifa, &match)
	if err != nil {
		return fmt.Errorf("failed to get match %s events from FIFA. %w", matchID, err)
	}

	existingEvents := match.Events
	ids, messages := findNewEvents(ctx, match.Events, matchData.NewEvents, &match)
	allEvents := append(existingEvents, ids...)
	if len(allEvents) > len(existingEvents) {
		match.Events = allEvents
		err = a.db.UpdateMatch(ctx, match)
		if err != nil {
			return fmt.Errorf("failed to save match %s events to the database. %w", matchID, err)
		}
	}

	slog.Debug("sending messages for match", "matchId", matchID)
	err = a.sendEventsToSlack(ctx, messages)
	if err != nil {
		slog.Error("failed to send message to slack", "error", err)
		return fmt.Errorf("failed to send events to Slack. %w", err)
	}
	if !matchData.Done {
		return nil
	}
	err = a.db.DeleteMatch(ctx, matchID)
	if err != nil {
		return fmt.Errorf("failed to delete match %s. %w", matchID, err)
	}
	return nil
}

func (a *app) sendEventsToSlack(ctx context.Context, events []string) error {
	for _, evt := range events {
		if strings.TrimSpace(evt) == "" {
			continue
		}
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
		eventIds = append(eventIds, event.Id)
		eventMsgs = append(eventMsgs, fifa.ProcessEvent(ctx, event, opts))
	}
	return eventIds, eventMsgs
}
