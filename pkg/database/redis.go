package database

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/imdevinc/fifa-bot/pkg/models"
	"github.com/redis/go-redis/v9"
)

type redisClient struct {
	client *redis.Client
}

var _ Database = (*redisClient)(nil)

func NewRedisClient(address string, password string, db int) *redisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	})

	return &redisClient{
		client: rdb,
	}
}

func (r *redisClient) AddMatch(ctx context.Context, match models.Match) error {
	_, err := r.client.HMSet(ctx, getRedisMatchKey(match.MatchId), match).Result()
	if err != nil {
		return fmt.Errorf("failed to save match to redis. %w", err)
	}
	return nil
}

func (r *redisClient) GetMatch(ctx context.Context, matchID string) (models.Match, error) {
	key := getRedisMatchKey(matchID)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return models.Match{}, fmt.Errorf("failed to get match %s from database. %w", matchID, err)
	}
	if err != nil {
		return models.Match{}, ErrMatchNotFound
	}
	match := models.Match{}
	err = json.Unmarshal([]byte(val), &match)
	if err != nil {
		return models.Match{}, fmt.Errorf("failed to unmarshal match. %w", err)
	}
	return match, nil
}

func (r *redisClient) GetMatchEvents(ctx context.Context, matchID string) ([]string, error) {
	key := getRedisEventsKey(matchID)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return []string{}, fmt.Errorf("failed to get events for match %s. %w", matchID, err)
	}
	if err != nil {
		return []string{}, nil
	}
	results := []string{}
	err = json.Unmarshal([]byte(val), &results)
	if err != nil {
		return []string{}, fmt.Errorf("failed to unmarshal events. %w", err)
	}
	return results, nil
}

func (r *redisClient) DeleteMatch(ctx context.Context, matchID string) error {
	eventKey := getRedisMatchKey(matchID)
	matchKey := getRedisEventsKey(matchID)
	_, err := r.client.Del(ctx, eventKey, matchKey).Result()
	if err != nil {
		return fmt.Errorf("failed to delete from redis. %w", err)
	}
	return nil
}

func (r *redisClient) UpdateMatchEvents(ctx context.Context, matchID string, events []string) error {
	key := getRedisEventsKey(matchID)
	_, err := r.client.HMSet(ctx, key, events).Result()
	if err != nil {
		return fmt.Errorf("failed to update events in redis. %w", err)
	}
	return nil
}

func (r *redisClient) GetAllMatches(ctx context.Context) ([]string, error) {
	keys, err := r.client.Keys(ctx, "match:*").Result()
	if err != nil {
		return []string{}, fmt.Errorf("failed to get keys from redis. %w", err)
	}
	results := []string{}
	for _, k := range keys {
		results = append(results, strings.TrimPrefix(k, "match:"))
	}
	return results, nil
}

func getRedisMatchKey(matchID string) string {
	return "match:" + matchID
}

func getRedisEventsKey(matchID string) string {
	return "events:" + matchID
}
