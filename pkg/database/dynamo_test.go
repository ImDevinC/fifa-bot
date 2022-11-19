package database_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/imdevinc/fifa-bot/pkg/database"
	"github.com/imdevinc/fifa-bot/pkg/queue"
)

const existingMatchId = "1"
const missingMatchId = "2"

type TestDatabase struct {
	GetItemCalls int
	PutItemCalls int
}

func (d *TestDatabase) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	d.GetItemCalls += 1
	returnValue := &dynamodb.GetItemOutput{}
	var returnErr error
	matchId := params.Key["MatchId"].(*types.AttributeValueMemberS)
	if matchId.Value == existingMatchId {
		returnValue = &dynamodb.GetItemOutput{
			Item: map[string]types.AttributeValue{
				"Value": &types.AttributeValueMemberS{
					Value: "Whatever",
				},
			},
		}
	} else if matchId.Value == missingMatchId {
		returnErr = database.ErrMatchNotFound
	}

	return returnValue, returnErr
}

func (d *TestDatabase) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	d.PutItemCalls += 1
	returnValue := &dynamodb.PutItemOutput{}
	var returnErr error

	return returnValue, returnErr
}

var _ database.Database = (*TestDatabase)(nil)

func TestDoesMatchExist(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		matchId     string
		expectedErr error
	}{
		{
			name:        "competition should exist",
			matchId:     existingMatchId,
			expectedErr: nil,
		},
		{
			name:        "competition should not exist",
			matchId:     missingMatchId,
			expectedErr: database.ErrMatchNotFound,
		},
	}
	db := &TestDatabase{}
	client := database.Client{
		Database: db,
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := client.DoesMatchExist(context.TODO(), &queue.MatchOptions{MatchId: tc.matchId})
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error %s, but got %s", tc.expectedErr, err)
			}
		})
	}
}

func TestAddMatchToDatabase(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name          string
		expectedCalls int
	}{
		{
			name:          "add match",
			expectedCalls: 1,
		},
	}
	db := TestDatabase{}
	client := database.Client{
		Database: &db,
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := client.AddMatch(context.TODO(), &queue.MatchOptions{})
			if err != nil {
				t.Error(err)
			}
			if tc.expectedCalls != db.PutItemCalls {
				t.Errorf("expected %d calls but got %d", tc.expectedCalls, db.PutItemCalls)
			}
		})
	}
}
