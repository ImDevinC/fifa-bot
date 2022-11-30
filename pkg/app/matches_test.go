package app_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/imdevinc/fifa-bot/pkg/app"
	"github.com/imdevinc/fifa-bot/pkg/database"
	"github.com/imdevinc/fifa-bot/pkg/queue"
	go_fifa "github.com/imdevinc/go-fifa"
)

type TestDB struct {
	GetItemFn func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
}

func (d *TestDB) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	if d.GetItemFn != nil {
		return d.GetItemFn(ctx, params, optFns...)
	}
	return &dynamodb.GetItemOutput{}, nil
}

func (d *TestDB) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	return &dynamodb.PutItemOutput{}, nil
}

func (d *TestDB) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optsFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	return &dynamodb.UpdateItemOutput{}, nil
}

func (d *TestDB) DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optsFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	return &dynamodb.DeleteItemOutput{}, nil
}

var _ database.Database = (*TestDB)(nil)

type TestQueue struct{}

func (q *TestQueue) SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	return nil, nil
}

var _ queue.Queue = (*TestQueue)(nil)

func TestMatches(t *testing.T) {
	config := app.GetMatchesConfig{
		FifaClient: &go_fifa.Client{},
		DatabaseClient: &database.Client{Database: &TestDB{
			GetItemFn: func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
				return &dynamodb.GetItemOutput{
					Item: map[string]types.AttributeValue{},
				}, database.ErrMatchNotFound
			},
		}},
		QueueClient: &queue.Client{Queue: &TestQueue{}},
	}
	err := app.GetMatches(context.Background(), &config)
	if err != nil {
		t.Error(err)
	}
}
