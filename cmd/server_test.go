package main

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

func (m *mockDatabase) AddMatch(context.Context, models.Match) error {
	return nil
}

func (m *mockDatabase) GetMatch(context.Context, string) (models.Match, error) {
	return models.Match{}, nil
}

func (m *mockDatabase) DeleteMatch(context.Context, string) error {
	return nil
}

func (m *mockDatabase) UpdateMatch(context.Context, models.Match) error {
	return nil
}

func (m *mockDatabase) GetAllMatches(context.Context) ([]models.Match, error) {
	return nil, nil
}

func (m *mockDatabase) Ping(context.Context) error {
	return m.pingErr
}

func TestHealthzHandlerReturnsOKWhenRedisIsHealthy(t *testing.T) {
	handler := healthzHandler(&mockDatabase{})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestHealthzHandlerReturnsUnavailableWhenRedisPingFails(t *testing.T) {
	handler := healthzHandler(&mockDatabase{pingErr: errors.New("redis down")})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, rr.Code)
	}
}
