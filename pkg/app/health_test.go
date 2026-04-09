package app

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/imdevinc/fifa-bot/pkg/models"
)

type mockDatabase struct {
	pingErr error
}

func (m *mockDatabase) AddMatch(ctx context.Context, match models.Match) error {
	return nil
}

func (m *mockDatabase) GetMatch(ctx context.Context, matchID string) (models.Match, error) {
	return models.Match{}, nil
}

func (m *mockDatabase) DeleteMatch(ctx context.Context, matchID string) error {
	return nil
}

func (m *mockDatabase) UpdateMatch(ctx context.Context, match models.Match) error {
	return nil
}

func (m *mockDatabase) GetAllMatches(ctx context.Context) ([]models.Match, error) {
	return nil, nil
}

func (m *mockDatabase) Ping(ctx context.Context) error {
	return m.pingErr
}

func TestHealthzHandlerReturnsOKWhenRedisIsReachable(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler := HealthzHandler(&mockDatabase{})

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Fatalf("expected body %q, got %q", "ok", rec.Body.String())
	}
}

func TestHealthzHandlerReturnsServiceUnavailableWhenRedisIsDown(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler := HealthzHandler(&mockDatabase{pingErr: errors.New("redis down")})

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}
}
