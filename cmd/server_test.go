package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockPinger struct {
	err error
}

func (m mockPinger) Ping(ctx context.Context) error {
	return m.err
}

func TestHealthzHandler(t *testing.T) {
	t.Run("returns 200 when redis responds", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		resp := httptest.NewRecorder()

		handler := healthzHandler(mockPinger{})
		handler(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "ok", resp.Body.String())
	})

	t.Run("returns 503 when redis ping fails", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		resp := httptest.NewRecorder()

		handler := healthzHandler(mockPinger{err: errors.New("redis down")})
		handler(resp, req)

		assert.Equal(t, http.StatusServiceUnavailable, resp.Code)
		assert.Contains(t, resp.Body.String(), "redis unavailable")
	})
}
