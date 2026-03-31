package healthz_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/imdevinc/fifa-bot/pkg/healthz"
	"github.com/imdevinc/fifa-bot/pkg/models"
)

// mockDatabase is a test double that implements the database.Database interface
type mockDatabase struct {
	pingErr error
}

func (m *mockDatabase) Ping(ctx context.Context) error {
	return m.pingErr
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

func TestHandler_Healthy(t *testing.T) {
	db := &mockDatabase{pingErr: nil}
	handler := healthz.Handler(db)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Body.String() != "ok" {
		t.Errorf("expected body 'ok', got '%s'", rec.Body.String())
	}
}

func TestHandler_Unhealthy(t *testing.T) {
	db := &mockDatabase{pingErr: errors.New("redis connection failed")}
	handler := healthz.Handler(db)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}

	expected := "unhealthy: redis connection failed"
	if rec.Body.String() != expected {
		t.Errorf("expected body '%s', got '%s'", expected, rec.Body.String())
	}
}
