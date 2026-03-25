package app

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeHealthChecker struct {
	err error
}

func (f fakeHealthChecker) HealthCheck(ctx context.Context) error {
	return f.err
}

func TestHealthzHandler(t *testing.T) {
	t.Run("returns ok when redis is healthy", func(t *testing.T) {
		h := HealthzHandler(fakeHealthChecker{})
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		res := httptest.NewRecorder()

		h(res, req)

		if res.Code != http.StatusOK {
			t.Fatalf("expected status %d got %d", http.StatusOK, res.Code)
		}
	})

	t.Run("returns unavailable when redis is unhealthy", func(t *testing.T) {
		h := HealthzHandler(fakeHealthChecker{err: errors.New("redis down")})
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		res := httptest.NewRecorder()

		h(res, req)

		if res.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected status %d got %d", http.StatusServiceUnavailable, res.Code)
		}
	})

	t.Run("rejects non-get methods", func(t *testing.T) {
		h := HealthzHandler(fakeHealthChecker{})
		req := httptest.NewRequest(http.MethodPost, "/healthz", nil)
		res := httptest.NewRecorder()

		h(res, req)

		if res.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected status %d got %d", http.StatusMethodNotAllowed, res.Code)
		}
	})
}
