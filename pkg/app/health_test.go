package app

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/imdevinc/fifa-bot/pkg/models"
)

type fakeHealthDB struct {
	healthErr error
}

func (f *fakeHealthDB) AddMatch(ctx context.Context, match models.Match) error {
	return nil
}

func (f *fakeHealthDB) GetMatch(ctx context.Context, matchID string) (models.Match, error) {
	return models.Match{}, nil
}

func (f *fakeHealthDB) DeleteMatch(ctx context.Context, matchID string) error {
	return nil
}

func (f *fakeHealthDB) UpdateMatch(ctx context.Context, match models.Match) error {
	return nil
}

func (f *fakeHealthDB) GetAllMatches(ctx context.Context) ([]models.Match, error) {
	return []models.Match{}, nil
}

func (f *fakeHealthDB) HealthCheck(ctx context.Context) error {
	return f.healthErr
}

func TestHealthHandlerHealthy(t *testing.T) {
	h := NewHealthHandler(&fakeHealthDB{})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if rr.Body.String() != "ok" {
		t.Fatalf("expected body ok, got %q", rr.Body.String())
	}
}

func TestHealthHandlerUnhealthy(t *testing.T) {
	h := NewHealthHandler(&fakeHealthDB{healthErr: errors.New("redis down")})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, rr.Code)
	}
	if rr.Body.String() != "unhealthy" {
		t.Fatalf("expected body unhealthy, got %q", rr.Body.String())
	}
}

func TestHealthHandlerMethodNotAllowed(t *testing.T) {
	h := NewHealthHandler(&fakeHealthDB{})
	req := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}
