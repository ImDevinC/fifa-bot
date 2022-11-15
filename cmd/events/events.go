package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"

	go_fifa "github.com/ImDevinC/go-fifa"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/imdevinc/fifa-bot/pkg/fifa"
	"github.com/imdevinc/fifa-bot/pkg/queue"
	log "github.com/sirupsen/logrus"
)

type SlackMessage struct {
	Text string `json:"text"`
}

// sendEventsToSlack sends the payload to the webhookURL. This expects the message to
// be a raw string that will be sent as the `text: ` value in a slack message
func sendEventsToSlack(ctx context.Context, webhookURL string, events []string) error {
	for _, evt := range events {
		payload := SlackMessage{Text: evt}
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		_, err = http.Post(webhookURL, "application/json", bytes.NewReader(b))
		if err != nil {
			return err
		}
	}
	return nil
}

// isMatchActive checks the active matches to see if the match still exists by checking
// for the matchId, competitionId, stageId, and seasonId
func isMatchActive(ctx context.Context, client *go_fifa.Client, opts *queue.MatchOptions) (bool, error) {
	matches, err := client.GetCurrentMatches()
	if err != nil {
		return false, err
	}
	var matchFound bool = false
	for _, m := range matches {
		if m.Id == opts.MatchId && m.CompetitionId == opts.CompetitionId && m.StageId == opts.StageId && m.SeasonId == opts.SeasonId {
			matchFound = true
			break
		}
	}
	return matchFound, nil
}

func HandleRequest(ctx context.Context, event events.SQSEvent) error {
	log.SetFormatter(&log.JSONFormatter{})

	queueURL := os.Getenv("QUEUE_URL")
	if queueURL == "" {
		log.Error("Missing QUEUE_URL")
		return errors.New("missing QUEUE_URL")
	}

	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		log.Error("[ERROR] Missing SLACK_WEBHOOK_URL")
		return errors.New("missing SLACK_WEBHOOK_URL")
	}

	var errWrap []string
	for _, r := range event.Records {

		opts := queue.MatchOptions{
			CompetitionId:  *r.MessageAttributes["CompetitionId"].StringValue,
			SeasonId:       *r.MessageAttributes["SeasonId"].StringValue,
			StageId:        *r.MessageAttributes["StageId"].StringValue,
			MatchId:        *r.MessageAttributes["MatchId"].StringValue,
			HomeTeamName:   *r.MessageAttributes["HomeTeamName"].StringValue,
			AwayTeamName:   *r.MessageAttributes["AwayTeamName"].StringValue,
			HomeTeamAbbrev: *r.MessageAttributes["HomeTeamAbbrev"].StringValue,
			AwayTeamAbbrev: *r.MessageAttributes["AwayTeamAbbrev"].StringValue,
			LastEvent:      *r.MessageAttributes["LastEvent"].StringValue,
		}
		fifaClient := go_fifa.Client{}
		events, matchOver, err := fifa.GetMatchEvents(ctx, &fifaClient, &opts)
		if err != nil {
			log.WithField("error", err).Error("failed to get match events")
			errWrap = append(errWrap, err.Error())
			continue
		}
		err = sendEventsToSlack(ctx, webhookURL, events)
		if err != nil {
			log.WithField("error", err).Error("failed to send message to slack")
			errWrap = append(errWrap, err.Error())
			continue
		}
		if matchOver {
			continue
		}
		active, err := isMatchActive(ctx, &fifaClient, &opts)
		if err != nil {
			log.Error(err)
			errWrap = append(errWrap, err.Error())
			continue
		}
		if !active {
			log.WithFields(log.Fields{
				"competitionId": opts.CompetitionId,
				"seasonId":      opts.SeasonId,
				"stageId":       opts.StageId,
				"matchId":       opts.MatchId,
			}).Warn("match was not marked as completed, but is no longer live")
			continue
		}
		// err = queue.SendToQueue(ctx, queueURL, &opts)
		// if err != nil {
		// 	log.WithField("error", err).Error("failed to send message to queue")
		// 	errWrap = append(errWrap, err.Error())
		// 	continue
		// }
	}
	if len(errWrap) > 0 {
		return errors.New(strings.Join(errWrap, "\n"))
	}
	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
