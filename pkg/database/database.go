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
	DeleteMatch(ctx context.Context, matchID string) error
	UpdateMatch(ctx context.Context, match models.Match) error
	GetAllMatches(ctx context.Context) ([]string, error)
}
