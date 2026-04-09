package app

import (
	"context"
	"net/http"
	"time"

	"github.com/imdevinc/fifa-bot/pkg/database"
)

const healthCheckTimeout = 2 * time.Second

func HealthzHandler(db database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), healthCheckTimeout)
		defer cancel()

		err := db.Ping(ctx)
		if err != nil {
			http.Error(w, "redis unavailable", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}
}
