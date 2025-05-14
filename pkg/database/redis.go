package database

import (
	"context"
	"errors"
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
	data, err := match.GetMap()
	if err != nil {
		return fmt.Errorf("failed to format match. %w", err)
	}
	_, err = r.client.HSet(ctx, getRedisMatchKey(match.MatchId), data).Result()
	if err != nil {
		return fmt.Errorf("failed to save match to redis. %w", err)
	}
	return nil
}

func (r *redisClient) GetMatch(ctx context.Context, matchID string) (models.Match, error) {
	key := getRedisMatchKey(matchID)
	val, err := r.client.HGetAll(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return models.Match{}, fmt.Errorf("failed to get match %s from database. %w", matchID, err)
	}
	if err != nil {
		return models.Match{}, ErrMatchNotFound
	}
	match, err := models.MatchFromRedis(val)
	if err != nil {
		return models.Match{}, fmt.Errorf("failed to unmarshal match. %w", err)
	}
	return match, nil
}

func (r *redisClient) GetMatchEvents(ctx context.Context, matchID string) (models.Match, error) {
	key := getRedisMatchKey(matchID)
	val, err := r.client.HGetAll(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return models.Match{}, fmt.Errorf("failed to get events for match %s. %w", matchID, err)
	}
	if err != nil {
		return models.Match{}, nil
	}
	match, err := models.MatchFromRedis(val)
	if err != nil {
		return models.Match{}, fmt.Errorf("failed to unmarshal events. %w", err)
	}
	return match, nil
}

func (r *redisClient) DeleteMatch(ctx context.Context, matchID string) error {
	eventKey := getRedisMatchKey(matchID)
	_, err := r.client.Del(ctx, eventKey).Result()
	if err != nil {
		return fmt.Errorf("failed to delete from redis. %w", err)
	}
	return nil
}

func (r *redisClient) UpdateMatch(ctx context.Context, match models.Match) error {
	data, err := match.GetMap()
	if err != nil {
		return fmt.Errorf("failed to format match. %w", err)
	}
	key := getRedisMatchKey(match.MatchId)
	_, err = r.client.HSet(ctx, key, data).Result()
	if err != nil {
		return fmt.Errorf("failed to update events in redis. %w", err)
	}
	return nil
}

func (r *redisClient) GetAllMatches(ctx context.Context) ([]models.Match, error) {
	keys, err := r.client.Keys(ctx, "match:*").Result()
	if err != nil {
		return []models.Match{}, fmt.Errorf("failed to get keys from redis. %w", err)
	}
	matches := []models.Match{}
	for _, k := range keys {
		matchID := strings.TrimPrefix(k, "match:")
		match, err := r.GetMatch(ctx, matchID)
		if err != nil && !errors.Is(err, ErrMatchNotFound) {
			return []models.Match{}, fmt.Errorf("failed to get match from the database. %w", err)
		}
		matches = append(matches, match)
	}
	return matches, nil
}

func getRedisMatchKey(matchID string) string {
	return "match:" + matchID
}
