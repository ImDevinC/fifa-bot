package database_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/imdevinc/fifa-bot/pkg/database"
	"github.com/stretchr/testify/assert"
)

const existingMatchId = "1"
const missingMatchId = "2"

type TestDatabase struct {
	GetItemCalls    int
	PutItemCalls    int
	UpdateItemCalls int
}

func (d *TestDatabase) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	d.GetItemCalls += 1
	returnValue := &dynamodb.GetItemOutput{}
	var returnErr error
	matchId := params.Key["MatchId"].(*types.AttributeValueMemberS)
	if matchId.Value == existingMatchId {
		returnValue = &dynamodb.GetItemOutput{
			Item: map[string]types.AttributeValue{
				"MatchId": &types.AttributeValueMemberS{
					Value: "1",
				},
				"Events": &types.AttributeValueMemberSS{
					Value: []string{"1", "2", "3"},
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

func (d *TestDatabase) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	d.UpdateItemCalls += 1
	returnValue := &dynamodb.UpdateItemOutput{}
	var returnErr error

	return returnValue, returnErr
}

func (d *TestDatabase) DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optsFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	return &dynamodb.DeleteItemOutput{}, nil
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
			err := client.DoesMatchExist(context.TODO(), tc.matchId)
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
			err := client.AddMatch(context.TODO(), "")
			if err != nil {
				t.Error(err)
			}
			if tc.expectedCalls != db.PutItemCalls {
				t.Errorf("expected %d calls but got %d", tc.expectedCalls, db.PutItemCalls)
			}
		})
	}
}

func TestGetEvents(t *testing.T) {
	t.Parallel()
	db := &TestDatabase{}
	client := database.Client{
		Database: db,
	}
	resp, err := client.GetEvents(context.TODO(), "1")
	if ok := assert.NoError(t, err); !ok {
		t.FailNow()
	}
	if ok := assert.Len(t, resp, 3); !ok {
		t.Fail()
	}
}

func TestUpdateItem(t *testing.T) {
	t.Parallel()
	db := &TestDatabase{}
	client := database.Client{
		Database: db,
	}
	ctx := context.TODO()
	resp, err := client.GetEvents(ctx, "1")
	if ok := assert.NoError(t, err); !ok {
		t.FailNow()
	}
	resp = append(resp, "4")
	err = client.UpdateMatchEvents(ctx, "1", resp)
	if ok := assert.NoError(t, err); !ok {
		t.Fail()
	}
}
