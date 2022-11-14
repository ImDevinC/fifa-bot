package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	go_fifa "github.com/ImDevinC/go-fifa"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/imdevinc/fifa-bot/pkg/fifa"
	"github.com/imdevinc/fifa-bot/pkg/queue"
)

type SlackMessage struct {
	Text string `json:"text"`
}

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

func HandleRequest(ctx context.Context, event events.SQSEvent) error {
	queueURL := os.Getenv("QUEUE_URL")
	if queueURL == "" {
		log.Println("[ERROR] Missing QUEUE_URL")
		return errors.New("missing QUEUE_URL")
	}

	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		log.Println("[ERROR] Missing SLACK_WEBHOOK_URL")
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
			log.Printf("[ERROR] %s", err)
			errWrap = append(errWrap, err.Error())
			continue
		}
		err = sendEventsToSlack(ctx, webhookURL, events)
		if err != nil {
			log.Printf("[ERROR] %s", err)
			errWrap = append(errWrap, err.Error())
			continue
		}
		if matchOver {
			continue
		}
		err = queue.SendToQueue(ctx, queueURL, &opts)
		if err != nil {
			log.Printf("[ERROR] %s", err)
			errWrap = append(errWrap, err.Error())
			continue
		}
	}
	if len(errWrap) > 0 {
		return errors.New(strings.Join(errWrap, "\n"))
	}
	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
