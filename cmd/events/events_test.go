package main

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
)

func buildEvent() events.SQSEvent {
	event := events.SQSEvent{
		Records: []events.SQSMessage{
			{
				MessageAttributes: map[string]events.SQSMessageAttribute{
					"CompetitionId": {
						StringValue: aws.String("cesdwwnxbc5fmajgroc0hqzy2"),
					},
					"SeasonId": {
						StringValue: aws.String("40sncpbsyexdrmedcwjz1j0gk"),
					},
					"StageId": {
						StringValue: aws.String("5w0vi7wp50objhjfn51o5ck5w"),
					},
					"MatchId": {
						StringValue: aws.String("3qxv1fe65nezrara3zsm5s55g"),
					},
					"HomeTeamName": {
						StringValue: aws.String("Albania"),
					},
					"AwayTeamName": {
						StringValue: aws.String("Italy"),
					},
					"HomeTeamAbbrev": {
						StringValue: aws.String("ALB"),
					},
					"AwayTeamAbbrev": {
						StringValue: aws.String("ITA"),
					},
					"LastEvent": {
						StringValue: aws.String("0"),
					},
					// "TraceId": {
					// 	StringValue: aws.String("5275bf3ebd698b81b3e225089f0d9c07-e34858c90ac8a076-1"),
					// },
				},
			},
		},
	}
	return event
}

func TestProcess(t *testing.T) {
	os.Setenv("QUEUE_URL", "TEST")
	os.Setenv("SLACK_WEBHOOK_URL", "http://localhost:8000")
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("SENTRY_DSN", "https://799c42ee5136461b8547028121aefd4d@o4504167699382272.ingest.sentry.io/4504167701479424")
	event := buildEvent()
	err := HandleRequest(context.TODO(), event)
	if err != nil {
		t.Error(err)
	}
}
