package app

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/imdevinc/fifa-bot/pkg/models"
	"github.com/stretchr/testify/assert"
)

type fakeDatabase struct {
	healthCheckErr error
}

func (f *fakeDatabase) AddMatch(ctx context.Context, match models.Match) error {
	return nil
}

func (f *fakeDatabase) GetMatch(ctx context.Context, matchID string) (models.Match, error) {
	return models.Match{}, nil
}

func (f *fakeDatabase) DeleteMatch(ctx context.Context, matchID string) error {
	return nil
}

func (f *fakeDatabase) UpdateMatch(ctx context.Context, match models.Match) error {
	return nil
}

func (f *fakeDatabase) GetAllMatches(ctx context.Context) ([]models.Match, error) {
	return nil, nil
}

func (f *fakeDatabase) HealthCheck(ctx context.Context) error {
	return f.healthCheckErr
}

func TestNewHealthHandler(t *testing.T) {
	t.Run("returns status ok when redis is reachable", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rr := httptest.NewRecorder()

		handler := NewHealthHandler(&fakeDatabase{})
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("returns service unavailable when redis is unreachable", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rr := httptest.NewRecorder()

		handler := NewHealthHandler(&fakeDatabase{healthCheckErr: errors.New("redis down")})
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	})
}
