package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/imdevinc/fifa-bot/pkg/database"
)

func NewHealthHandler(db database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := db.HealthCheck(ctx); err != nil {
			slog.Error("health check failed", "error", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
