package main

import (
	"context"
	"log"

	go_fifa "github.com/ImDevinC/go-fifa"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/imdevinc/fifa-bot/pkg/executor"
	"github.com/imdevinc/fifa-bot/pkg/fifa"
)

func HandleRequest(ctx context.Context, event events.CloudWatchEvent) error {
	fifaClient := go_fifa.Client{}
	sfnClient := executor.Client{}
	err := fifa.GetNewMatches(ctx, &fifaClient, &sfnClient)
	if err != nil {
		log.Println(err)
	}
	return err
}

func main() {
	lambda.Start(HandleRequest)
}
