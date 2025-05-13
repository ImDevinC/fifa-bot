package database

import (
	"context"
	"errors"

	"github.com/imdevinc/fifa-bot/pkg/models"
)

var ErrMatchNotFound = errors.New("match not found")

type Database interface {
	AddMatch(ctx context.Context, match models.Match) error
	GetMatch(ctx context.Context, matchID string) (models.Match, error)
	GetMatchEvents(ctx context.Context, matchID string) ([]string, error)
	DeleteMatch(ctx context.Context, matchID string) error
	UpdateMatchEvents(ctx context.Context, matchID string, events []string) error
	GetAllMatches(ctx context.Context) ([]string, error)
}
