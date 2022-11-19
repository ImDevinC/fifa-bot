package app_test

import (
	"context"
	"testing"

	go_fifa "github.com/ImDevinC/go-fifa"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/imdevinc/fifa-bot/pkg/app"
	"github.com/imdevinc/fifa-bot/pkg/database"
	"github.com/imdevinc/fifa-bot/pkg/queue"
)

type TestDB struct {
	GetItemFn func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
}

func (d *TestDB) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	if d.GetItemFn != nil {
		return d.GetItemFn(ctx, params, optFns...)
	}
	return nil, nil
}

func (d *TestDB) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	return nil, nil
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
	err := app.GetMatches(context.TODO(), &config)
	if err != nil {
		t.Error(err)
	}
}
