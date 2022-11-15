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
						StringValue: aws.String("dvtl8sf1262pd2aqgu641qa7u"),
					},
					"SeasonId": {
						StringValue: aws.String("91jifxei9sjpv0afezbdhobo4"),
					},
					"StageId": {
						StringValue: aws.String("avxfk1rxjckv9qc3ahpx9fod0"),
					},
					"MatchId": {
						StringValue: aws.String("83deossj5098pa1h7aeelv7ro"),
					},
					"HomeTeamName": {
						StringValue: aws.String("Portgual"),
					},
					"AwayTeamName": {
						StringValue: aws.String("Costa Rica"),
					},
					"HomeTeamAbbrev": {
						StringValue: aws.String("POR"),
					},
					"AwayTeamAbbrev": {
						StringValue: aws.String("CRC"),
					},
					"LastEvent": {
						StringValue: aws.String("0"),
					},
				},
			},
		},
	}
	return event
}

func TestProcess(t *testing.T) {
	os.Setenv("QUEUE_URL", "TEST")
	os.Setenv("SLACK_WEBHOOK_URL", "TEST")
	event := buildEvent()
	err := HandleRequest(context.TODO(), event)
	if err != nil {
		t.Error(err)
	}
}
