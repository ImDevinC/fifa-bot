package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
)

type Client struct {
	client *sfn.Client
	sfnARN string
}

type StartMatchOptions struct {
	CompetitionId string `json:"competition_id"`
	SeasonId      string `json:"season_id"`
	StageId       string `json:"stage_id"`
	MatchId       string `json:"match_id"`
}

var sfnARN = os.Getenv("STATE_FUNCTION_MACHINE_ARN")

func createDefaultClient(ctx context.Context) (*sfn.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return sfn.New(sfn.Options{
		Credentials: cfg.Credentials,
		Region:      cfg.Region,
	}), nil
}

func (c *Client) StartMatch(ctx context.Context, opts *StartMatchOptions) error {
	input, err := json.Marshal(opts)
	if err != nil {
		return fmt.Errorf("failed to unmarshal match options. %w", err)
	}

	arn := sfnARN
	if c.sfnARN != "" {
		arn = c.sfnARN
	}

	if c.client == nil {
		client, err := createDefaultClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to create AWS client. %w", err)
		}
		c.client = client
	}

	_, err = c.client.StartExecution(ctx, &sfn.StartExecutionInput{
		StateMachineArn: aws.String(arn),
		Input:           aws.String(string(input)),
		Name:            aws.String(opts.MatchId),
	})
	if err != nil {
		return fmt.Errorf("failed to start execution. %w", err)
	}

	return nil
}
