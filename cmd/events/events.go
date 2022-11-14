package main

import (
	"context"
	"fmt"
	"log"

	go_fifa "github.com/ImDevinC/go-fifa"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/imdevinc/fifa-bot/pkg/executor"
	"github.com/imdevinc/fifa-bot/pkg/fifa"
)

func HandleRequest(ctx context.Context, opts executor.StartMatchOptions) error {
	fifaClient := go_fifa.Client{}
	events, err := fifa.GetMatchEvents(ctx, &fifaClient, &opts)
	if err != nil {
		log.Println(err)
		return err
	}
	for _, evt := range events {
		fmt.Println(evt)
	}
	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
