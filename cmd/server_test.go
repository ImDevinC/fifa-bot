package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/imdevinc/fifa-bot/pkg/models"
)

type stubDatabase struct {
	pingErr error
}

func (s stubDatabase) AddMatch(context.Context, models.Match) error { return nil }

func (s stubDatabase) GetMatch(context.Context, string) (models.Match, error) {
	return models.Match{}, nil
}

func (s stubDatabase) DeleteMatch(context.Context, string) error { return nil }

func (s stubDatabase) UpdateMatch(context.Context, models.Match) error { return nil }

func (s stubDatabase) GetAllMatches(context.Context) ([]models.Match, error) {
	return nil, nil
}

func (s stubDatabase) Ping(context.Context) error { return s.pingErr }

func TestHealthHandlerReturnsOKWhenRedisIsHealthy(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()

	handler := healthHandler(stubDatabase{})
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestHealthHandlerReturnsServiceUnavailableWhenRedisPingFails(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()

	handler := healthHandler(stubDatabase{pingErr: errors.New("down")})
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, rr.Code)
	}
}

func TestHealthHandlerRejectsNonGetMethods(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	rr := httptest.NewRecorder()

	handler := healthHandler(stubDatabase{})
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}
