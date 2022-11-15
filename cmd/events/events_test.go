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
						StringValue: aws.String("2000000005"),
					},
					"SeasonId": {
						StringValue: aws.String("400250052"),
					},
					"StageId": {
						StringValue: aws.String("b1ayaoa4q68n6464fy4orklqs"),
					},
					"MatchId": {
						StringValue: aws.String("3y748w6ppuxciynnoonrt9jx0"),
					},
					"HomeTeamName": {
						StringValue: aws.String("Austria Wien"),
					},
					"AwayTeamName": {
						StringValue: aws.String("FAK"),
					},
					"HomeTeamAbbrev": {
						StringValue: aws.String("Wolfsberger AC"),
					},
					"AwayTeamAbbrev": {
						StringValue: aws.String("WAC"),
					},
					"LastEvent": {
						StringValue: aws.String("820"),
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
