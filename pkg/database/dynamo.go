package database

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/getsentry/sentry-go"
	"github.com/imdevinc/fifa-bot/pkg/queue"
)

type DatabaseClient struct {
	ddbClient *dynamodb.Client
	tableName string
}

var ErrMatchNotFound = errors.New("match not found")

func NewDynamoClient(ctx context.Context, tableName string) (DatabaseClient, error) {
	span := sentry.StartSpan(ctx, "dynamo.NewDynamoClient")
	defer span.Finish()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		sentry.CaptureException(err)
		return DatabaseClient{}, err
	}
	client := dynamodb.NewFromConfig(cfg)
	return DatabaseClient{
		ddbClient: client,
		tableName: tableName,
	}, nil
}

func (d *DatabaseClient) DoesMatchExist(ctx context.Context, opts *queue.MatchOptions) error {
	span := sentry.StartSpan(ctx, "dynamo.DoesMatchExist")
	defer span.Finish()

	input := &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			"MatchId": &types.AttributeValueMemberS{
				Value: opts.MatchId,
			},
		},
		TableName: &d.tableName,
	}
	resp, err := d.ddbClient.GetItem(ctx, input)
	if err != nil {
		sentry.CaptureException(err)
		return fmt.Errorf("failed to get item from dynamo. %w", err)
	}
	if resp.Item == nil {
		return ErrMatchNotFound
	}
	return nil
}

func (d *DatabaseClient) AddMatch(ctx context.Context, opts *queue.MatchOptions) error {
	span := sentry.StartSpan(ctx, "dynamo.AddMatch")
	defer span.Finish()

	ttl := time.Now().Add(time.Hour * 6)
	ttlValue := strconv.FormatInt(ttl.Unix(), 10)
	input := &dynamodb.PutItemInput{
		TableName: &d.tableName,
		Item: map[string]types.AttributeValue{
			"MatchId": &types.AttributeValueMemberS{
				Value: opts.MatchId,
			},
			"Expiration": &types.AttributeValueMemberN{
				Value: ttlValue,
			},
		},
	}
	_, err := d.ddbClient.PutItem(ctx, input)
	if err != nil {
		sentry.CaptureException(err)
		return fmt.Errorf("failed to put item into dynamo. %w", err)
	}
	return nil
}

func (d *DatabaseClient) DeleteMatch(ctx context.Context, opts *queue.MatchOptions) error {
	span := sentry.StartSpan(ctx, "dynamo.DeleteMatch")
	defer span.Finish()

	input := &dynamodb.DeleteItemInput{
		TableName: &d.tableName,
		Key: map[string]types.AttributeValue{
			"MatchId": &types.AttributeValueMemberS{
				Value: opts.MatchId,
			},
		},
	}
	_, err := d.ddbClient.DeleteItem(ctx, input)
	if err != nil {
		sentry.CaptureException(err)
		return fmt.Errorf("failed to delete item from dynamo. %w", err)
	}
	return nil
}
