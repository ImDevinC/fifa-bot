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

type Database interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
}

type Client struct {
	Database  Database
	TableName string
}

var ErrMatchNotFound = errors.New("match not found")

func NewDynamoClient(ctx context.Context, tableName string) (Client, error) {
	span := sentry.StartSpan(ctx, "db.init")
	defer span.Finish()
	span.Description = "database.NewDynamoClient"

	ctx = span.Context()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		sentry.CaptureException(err)
		return Client{}, err
	}
	client := dynamodb.NewFromConfig(cfg)
	return Client{
		Database:  client,
		TableName: tableName,
	}, nil
}

func (d *Client) DoesMatchExist(ctx context.Context, opts *queue.MatchOptions) error {
	span := sentry.StartSpan(ctx, "db.query")
	defer span.Finish()
	span.Description = "database.DoesMatchExist"
	span.SetTag("matchId", opts.MatchId)

	ctx = span.Context()

	input := &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			"MatchId": &types.AttributeValueMemberS{
				Value: opts.MatchId,
			},
		},
		TableName: &d.TableName,
	}
	resp, err := d.Database.GetItem(ctx, input)
	if err != nil {
		sentry.CaptureException(err)
		return fmt.Errorf("failed to get item from dynamo. %w", err)
	}
	if resp.Item == nil {
		return ErrMatchNotFound
	}
	return nil
}

func (d *Client) AddMatch(ctx context.Context, opts *queue.MatchOptions) error {
	span := sentry.StartSpan(ctx, "db.query")
	defer span.Finish()
	span.Description = "database.AddMatch"
	span.SetTag("matchId", opts.MatchId)

	ctx = span.Context()

	ttl := time.Now().Add(time.Hour * 6)
	ttlValue := strconv.FormatInt(ttl.Unix(), 10)
	input := &dynamodb.PutItemInput{
		TableName: &d.TableName,
		Item: map[string]types.AttributeValue{
			"MatchId": &types.AttributeValueMemberS{
				Value: opts.MatchId,
			},
			"Expiration": &types.AttributeValueMemberN{
				Value: ttlValue,
			},
		},
	}
	_, err := d.Database.PutItem(ctx, input)
	if err != nil {
		sentry.CaptureException(err)
		return fmt.Errorf("failed to put item into dynamo. %w", err)
	}
	return nil
}
