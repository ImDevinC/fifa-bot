package database

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/getsentry/sentry-go"
)

type Database interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
}

type Client struct {
	Database  Database
	TableName string
}

type Match struct {
	Id         string   `dynamodbav:"MatchId"`
	Events     []string `dynamodbav:"Events"`
	Expiration int      `dynamodbav:"Expiration"`
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

func (d *Client) DoesMatchExist(ctx context.Context, matchId string) error {
	span := sentry.StartSpan(ctx, "db.query")
	defer span.Finish()
	span.Description = "database.DoesMatchExist"
	span.SetTag("matchId", matchId)

	ctx = span.Context()

	input := &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			"MatchId": &types.AttributeValueMemberS{
				Value: matchId,
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

func (d *Client) AddMatch(ctx context.Context, matchId string) error {
	span := sentry.StartSpan(ctx, "db.query")
	defer span.Finish()
	span.Description = "database.AddMatch"
	span.SetTag("matchId", matchId)

	ctx = span.Context()

	ttl := time.Now().Add(time.Hour * 6)
	ttlValue := strconv.FormatInt(ttl.Unix(), 10)
	input := &dynamodb.PutItemInput{
		TableName: &d.TableName,
		Item: map[string]types.AttributeValue{
			"MatchId": &types.AttributeValueMemberS{
				Value: matchId,
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

func (d *Client) GetEvents(ctx context.Context, matchId string) ([]string, error) {
	span := sentry.StartSpan(ctx, "db.query")
	defer span.Finish()
	span.Description = "database.GetEvents"
	span.SetTag("matchId", matchId)

	ctx = span.Context()

	input := &dynamodb.GetItemInput{
		TableName: &d.TableName,
		Key: map[string]types.AttributeValue{
			"MatchId": &types.AttributeValueMemberS{
				Value: matchId,
			},
		},
	}

	resp, err := d.Database.GetItem(ctx, input)
	if err != nil {
		sentry.CaptureException(err)
		return []string{}, fmt.Errorf("failed to get item from dynamo. %w", err)
	}

	if resp.Item == nil {
		return []string{}, ErrMatchNotFound
	}

	var match Match
	err = attributevalue.UnmarshalMap(resp.Item, &match)
	if err != nil {
		sentry.CaptureException(err)
		return []string{}, fmt.Errorf("failed to unmarshal item from table. %w", err)
	}
	return match.Events, nil
}

func (d *Client) UpdateMatchEvents(ctx context.Context, matchId string, events []string) error {
	span := sentry.StartSpan(ctx, "db.update")
	defer span.Finish()
	span.Description = "database.UpdateMatchEvents"
	span.SetTag("matchId", matchId)

	ctx = span.Context()

	if len(events) == 0 {
		return nil
	}

	input := &dynamodb.UpdateItemInput{
		TableName: &d.TableName,
		Key: map[string]types.AttributeValue{
			"MatchId": &types.AttributeValueMemberS{
				Value: matchId,
			},
		},
		UpdateExpression: aws.String("SET Events = :events"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":events": &types.AttributeValueMemberSS{
				Value: events,
			},
		},
	}

	_, err := d.Database.UpdateItem(ctx, input)
	if err != nil {
		sentry.CaptureException(err)
		return err
	}
	return nil
}

func (d *Client) DeleteMatch(ctx context.Context, matchId string) error {
	span := sentry.StartSpan(ctx, "db.update")
	defer span.Finish()
	span.Description = "database.UpdateMatchEvents"
	span.SetTag("matchId", matchId)

	ctx = span.Context()

	input := &dynamodb.DeleteItemInput{
		Key: map[string]types.AttributeValue{
			"MatchId": &types.AttributeValueMemberS{
				Value: matchId,
			},
		},
	}

	_, err := d.Database.DeleteItem(ctx, input)
	if err != nil {
		sentry.CaptureException(err)
		return err
	}
	return nil
}
