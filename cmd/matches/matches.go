package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"

	go_fifa "github.com/ImDevinC/go-fifa"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/imdevinc/fifa-bot/pkg/database"
	"github.com/imdevinc/fifa-bot/pkg/fifa"
	"github.com/imdevinc/fifa-bot/pkg/queue"
)

func HandleRequest(ctx context.Context) error {
	queueURL := os.Getenv("QUEUE_URL")
	if queueURL == "" {
		log.Println("[ERROR] Missing QUEUE_URL")
		return errors.New("missing QUEUE_URL")
	}

	tableName := os.Getenv("TABLE_NAME")
	if queueURL == "" {
		log.Println("[ERROR] Missing TABLE_NAME")
		return errors.New("missing TABLE_NAME")
	}

	dbClient, err := database.NewDynamoClient(ctx, tableName)
	if err != nil {
		log.Println(err)
		return err
	}

	fifaClient := go_fifa.Client{}
	matches, err := fifa.GetLiveMatches(ctx, &fifaClient)
	if err != nil {
		log.Println(err)
		return err
	}
	var errWrap []string
	for _, m := range matches {
		err := dbClient.DoesMatchExist(ctx, &m)
		if !errors.Is(err, database.ErrMatchNotFound) {
			continue
		}
		if err != nil && !errors.Is(err, database.ErrMatchNotFound) {
			log.Printf("failed to get match info from database. %s", err)
			errWrap = append(errWrap, err.Error())
			continue
		}
		err = dbClient.AddMatch(ctx, &m)
		if err != nil {
			log.Printf("failed to save match to database. %s", err)
			errWrap = append(errWrap, err.Error())
			continue
		}
		m.LastEvent = "0"
		err = queue.SendToQueue(ctx, queueURL, &m)
		if err != nil {
			log.Printf("failed to send message to queue. %s", err)
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
