package app

import (
	"bytes"
	"context"
	"encoding/json"
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
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		for {
			select {
			case <-ctx.Done():
				wg.Done()
				return
			default:
				if err := a.monitorEvents(); err != nil {
					slog.Error("failed to monitor events", "error", err)
				}
				time.Sleep(10 * time.Second)
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				wg.Done()
				return
			default:
				if err := a.monitorMatches(); err != nil {
					slog.Error("failed to monitor matches", "error", err)
				}
				time.Sleep(10 * time.Second)
			}
		}
	}()
	slog.Info("Started app")
	wg.Wait()
	return nil
}

func (a *app) monitorMatches() error {
	ctx := context.Background()
	slog.Debug("getting matches from database")
	existingMatches, err := a.db.GetAllMatches(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get matches from database. %w", err)
	}
	slog.Debug("getting live matches from FIFA")
	matches, err := fifa.GetLiveMatches(ctx, a.fifa)
	if err != nil {
		return fmt.Errorf("failed to get live matches from FIFA. %w", err)
	}
	slog.Debug("found matches", "count", len(matches))
	for _, m := range matches {
		if len(a.CompetitionId) > 0 && m.CompetitionId != a.CompetitionId {
			slog.Debug("match is in wrong competition", "matchId", m.MatchId, "competitionId", m.CompetitionId, "desiredCompetitionId", a.CompetitionId)
			continue
		}
		if slices.Contains(existingMatches, m.MatchId) {
			slog.Debug("match already exists, skipping", "match", m.MatchId)
			continue
		}
		slog.Debug("adding match", "matchId", m.MatchId)
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
	matchAttr := slog.Attr{Key: "matchId", Value: slog.StringValue(matchID)}
	slog.Debug("getting match info", matchAttr)
	match, err := a.db.GetMatch(ctx, matchID)
	if err != nil {
		return fmt.Errorf("failed to get match %s from database. %w", matchID, err)
	}
	slog.Debug("getting match previous events", matchAttr)
	existingEvents, err := a.db.GetMatchEvents(ctx, matchID)
	if err != nil {
		return fmt.Errorf("failed to get events for match %s. %w", matchID, err)
	}
	matchOpts := models.Match{
		CompetitionId: match.CompetitionId,
		SeasonId:      match.SeasonId,
		StageId:       match.StageId,
		MatchId:       match.MatchId,
	}
	slog.Debug("getting live match events", matchAttr)
	matchData, err := fifa.GetMatchEvents(ctx, a.fifa, &matchOpts)
	if err != nil {
		return fmt.Errorf("failed to get match %s events from FIFA. %w", matchID, err)
	}

	slog.Debug("looking for new events", matchAttr)
	ids, messages := findNewEvents(ctx, existingEvents, matchData.NewEvents, &matchOpts)
	allEvents := append(existingEvents, ids...)
	if len(allEvents) > len(existingEvents) {
		slog.Debug("found new events", matchAttr, "oldEvents", len(existingEvents), "newEvents", len(allEvents))
		err = a.db.UpdateMatchEvents(ctx, matchID, allEvents)
		if err != nil {
			return fmt.Errorf("failed to save match %s events to the database. %w", matchID, err)
		}
	}

	slog.Debug("sending message to slack", matchAttr)
	err = a.sendEventsToSlack(ctx, messages)
	if err != nil {
		return fmt.Errorf("failed to send events to Slack. %w", err)
	}
	if !matchData.Done {
		slog.Debug("match not done, not deleting", matchAttr)
		return nil
	}
	slog.Debug("match done, deleting it", matchAttr)
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
			sentry.CaptureException(err)
			return err
		}
		_, err = http.Post(a.slackWebhookURL, "application/json", bytes.NewReader(b))
		if err != nil {
			sentry.CaptureException(err)
			return err
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
