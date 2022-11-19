package queue

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/getsentry/sentry-go"
)

type MatchOptions struct {
	CompetitionId  string
	SeasonId       string
	StageId        string
	MatchId        string
	LastEvent      string
	HomeTeamName   string
	AwayTeamName   string
	HomeTeamAbbrev string
	AwayTeamAbbrev string
}

func MatchOptsFromSQS(ctx context.Context, attributes map[string]events.SQSMessageAttribute) MatchOptions {
	span := sentry.StartSpan(ctx, "sqs.MatchOptsFromSQS")
	defer span.Finish()

	opts := MatchOptions{
		CompetitionId: *attributes["CompetitionId"].StringValue,
		SeasonId:      *attributes["SeasonId"].StringValue,
		StageId:       *attributes["StageId"].StringValue,
		MatchId:       *attributes["MatchId"].StringValue,
		HomeTeamName:  *attributes["HomeTeamName"].StringValue,
		AwayTeamName:  *attributes["AwayTeamName"].StringValue,
		LastEvent:     *attributes["LastEvent"].StringValue,
	}

	if val, ok := attributes["AwayTeamAbbrev"]; ok {
		opts.AwayTeamAbbrev = *val.StringValue
	}

	if val, ok := attributes["HomeTeamAbbrev"]; ok {
		opts.HomeTeamAbbrev = *val.StringValue
	}

	return opts
}

func SendToQueue(ctx context.Context, queueURL string, opts *MatchOptions) error {
	span := sentry.StartSpan(ctx, "sqs.SendToQueue")
	defer span.Finish()

	ctx = span.Context()

	input := &sqs.SendMessageInput{
		MessageAttributes: map[string]types.MessageAttributeValue{
			"CompetitionId": {
				DataType:    aws.String("String"),
				StringValue: aws.String(opts.CompetitionId),
			},
			"SeasonId": {
				DataType:    aws.String("String"),
				StringValue: aws.String(opts.SeasonId),
			},
			"StageId": {
				DataType:    aws.String("String"),
				StringValue: aws.String(opts.StageId),
			},
			"MatchId": {
				DataType:    aws.String("String"),
				StringValue: aws.String(opts.MatchId),
			},
			"LastEvent": {
				DataType:    aws.String("String"),
				StringValue: aws.String(opts.LastEvent),
			},
			"HomeTeamName": {
				DataType:    aws.String("String"),
				StringValue: aws.String(opts.HomeTeamName),
			},
			"AwayTeamName": {
				DataType:    aws.String("String"),
				StringValue: aws.String(opts.AwayTeamName),
			},
			"TraceId": {
				DataType:    aws.String("String"),
				StringValue: aws.String(sentry.TransactionFromContext(ctx).ToSentryTrace()),
			},
		},
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String("match"),
	}

	if len(opts.AwayTeamAbbrev) > 0 {
		input.MessageAttributes["AwayTeamAbbrev"] = types.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(opts.AwayTeamAbbrev),
		}
	}

	if len(opts.HomeTeamAbbrev) > 0 {
		input.MessageAttributes["HomeTeamAbbrev"] = types.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(opts.HomeTeamAbbrev),
		}
	}

	if opts.LastEvent != "-1" {
		input.DelaySeconds = 60
	}
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		sentry.CaptureException(err)
		return fmt.Errorf("failed to load default config. %w", err)
	}

	client := sqs.NewFromConfig(cfg)
	_, err = client.SendMessage(ctx, input)
	if err != nil {
		sentry.CaptureException(err)
		return fmt.Errorf("failed to send message to queue. %w", err)
	}
	return nil
}
