package queue_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/go-cmp/cmp"
	"github.com/imdevinc/fifa-bot/pkg/queue"
)

type TestMsgAttributes struct {
	competitionId  string
	seasonId       string
	stageId        string
	matchId        string
	homeTeamName   string
	homeTeamAbbrev string
	awayTeamName   string
	awayTeamAbbrev string
	lastEvent      string
}

type TestQueue struct {
	ExpectedInput *sqs.SendMessageInput
}

func (q *TestQueue) SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	diff := cmp.Diff(params, q.ExpectedInput, cmp.Comparer(attrComparer))
	if len(diff) > 0 {
		return nil, errors.New(diff)
	}
	return nil, nil
}

var _ queue.Queue = (*TestQueue)(nil)

func TestMatchOptsFromSQS(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		attrs TestMsgAttributes
		opts  queue.MatchOptions
	}{
		{
			name: "all values are propagated",
			attrs: TestMsgAttributes{
				competitionId:  "1",
				seasonId:       "2",
				stageId:        "3",
				matchId:        "4",
				homeTeamName:   "Unicorns",
				homeTeamAbbrev: "UNC",
				awayTeamName:   "Ligers",
				awayTeamAbbrev: "LGR",
				lastEvent:      "1234",
			},
			opts: queue.MatchOptions{
				CompetitionId:  "1",
				SeasonId:       "2",
				StageId:        "3",
				MatchId:        "4",
				HomeTeamName:   "Unicorns",
				HomeTeamAbbrev: "UNC",
				AwayTeamName:   "Ligers",
				AwayTeamAbbrev: "LGR",
				LastEvent:      "1234",
			},
		},
		{
			name: "missing abbrev doesn't cause crash",
			attrs: TestMsgAttributes{
				competitionId: "1",
				seasonId:      "2",
				stageId:       "3",
				matchId:       "4",
				homeTeamName:  "Unicorns",
				awayTeamName:  "Ligers",
				lastEvent:     "1234",
			},
			opts: queue.MatchOptions{
				CompetitionId: "1",
				SeasonId:      "2",
				StageId:       "3",
				MatchId:       "4",
				HomeTeamName:  "Unicorns",
				AwayTeamName:  "Ligers",
				LastEvent:     "1234",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			attrs := buildSQSRecordFromAttrs(t, tc.attrs)
			opts := queue.MatchOptsFromSQS(context.TODO(), attrs)
			diff := cmp.Diff(opts, tc.opts)
			if len(diff) > 0 {
				t.Errorf("expected no diffs, got %s", diff)
			}
		})
	}
}

func TestSendToQueue(t *testing.T) {
	opts := queue.MatchOptions{
		CompetitionId:  "1",
		SeasonId:       "2",
		StageId:        "3",
		MatchId:        "4",
		HomeTeamName:   "Unicorns",
		AwayTeamName:   "Ligers",
		HomeTeamAbbrev: "UNC",
		AwayTeamAbbrev: "LGR",
		LastEvent:      "1234",
	}
	client := queue.Client{
		QueueURL: "http://test-queue",
		Queue: &TestQueue{ExpectedInput: &sqs.SendMessageInput{
			MessageBody:  aws.String("match"),
			QueueUrl:     aws.String("http://test-queue"),
			DelaySeconds: 60,
			MessageAttributes: map[string]types.MessageAttributeValue{
				"AwayTeamAbbrev": {DataType: aws.String("String"), StringValue: aws.String("LGR")},
				"AwayTeamName":   {DataType: aws.String("String"), StringValue: aws.String("Ligers")},
				"CompetitionId":  {DataType: aws.String("String"), StringValue: aws.String("1")},
				"HomeTeamAbbrev": {DataType: aws.String("String"), StringValue: aws.String("UNC")},
				"HomeTeamName":   {DataType: aws.String("String"), StringValue: aws.String("Unicorns")},
				"LastEvent":      {DataType: aws.String("String"), StringValue: aws.String("1234")},
				"MatchId":        {DataType: aws.String("String"), StringValue: aws.String("4")},
				"SeasonId":       {DataType: aws.String("String"), StringValue: aws.String("2")},
				"StageId":        {DataType: aws.String("String"), StringValue: aws.String("3")},
			},
		}},
	}
	err := client.SendToQueue(context.TODO(), &opts)
	if err != nil {
		t.Error(err)
	}
}

func attrComparer(input *sqs.SendMessageInput, expected *sqs.SendMessageInput) bool {
	if input.DelaySeconds != expected.DelaySeconds {
		return false
	}
	if *input.MessageBody != *expected.MessageBody {
		return false
	}
	if *input.QueueUrl != *expected.QueueUrl {
		return false
	}
	if *input.MessageAttributes["CompetitionId"].StringValue != *expected.MessageAttributes["CompetitionId"].StringValue {
		return false
	}
	if *input.MessageAttributes["SeasonId"].StringValue != *expected.MessageAttributes["SeasonId"].StringValue {
		return false
	}
	if *input.MessageAttributes["StageId"].StringValue != *expected.MessageAttributes["StageId"].StringValue {
		return false
	}
	if *input.MessageAttributes["MatchId"].StringValue != *expected.MessageAttributes["MatchId"].StringValue {
		return false
	}
	if *input.MessageAttributes["LastEvent"].StringValue != *expected.MessageAttributes["LastEvent"].StringValue {
		return false
	}
	if *input.MessageAttributes["HomeTeamName"].StringValue != *expected.MessageAttributes["HomeTeamName"].StringValue {
		return false
	}
	if *input.MessageAttributes["AwayTeamName"].StringValue != *expected.MessageAttributes["AwayTeamName"].StringValue {
		return false
	}
	if val, ok := input.MessageAttributes["AwayTeamAbbrev"]; ok {
		if val2, ok := expected.MessageAttributes["AwayTeamAbbrev"]; ok {
			if *val2.StringValue != *val.StringValue {
				return false
			}
		}
	}
	if val, ok := input.MessageAttributes["HomeTeamAbbrev"]; ok {
		if val2, ok := expected.MessageAttributes["HomeTeamAbbrev"]; ok {
			if *val2.StringValue != *val.StringValue {
				return false
			}
		}
	}

	return true
}

func buildSQSRecordFromAttrs(t *testing.T, attr TestMsgAttributes) map[string]events.SQSMessageAttribute {
	t.Helper()

	returnValue := map[string]events.SQSMessageAttribute{
		"CompetitionId": {
			StringValue: &attr.competitionId,
		},
		"SeasonId": {
			StringValue: &attr.seasonId,
		},
		"StageId": {
			StringValue: &attr.stageId,
		},
		"MatchId": {
			StringValue: &attr.matchId,
		},
		"HomeTeamName": {
			StringValue: &attr.homeTeamName,
		},
		"AwayTeamName": {
			StringValue: &attr.awayTeamName,
		},
		"LastEvent": {
			StringValue: &attr.lastEvent,
		},
	}

	if len(attr.homeTeamAbbrev) > 0 {
		returnValue["HomeTeamAbbrev"] = events.SQSMessageAttribute{
			StringValue: &attr.homeTeamAbbrev,
		}
	}

	if len(attr.awayTeamAbbrev) > 0 {
		returnValue["AwayTeamAbbrev"] = events.SQSMessageAttribute{
			StringValue: &attr.awayTeamAbbrev,
		}
	}

	return returnValue
}
