package healthz

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/imdevinc/fifa-bot/pkg/database"
)

// Handler creates an HTTP handler for the /healthz endpoint that validates
// the Redis connection is healthy
func Handler(db database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create a context with timeout for the health check
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			slog.Error("health check failed", "error", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("unhealthy: redis connection failed"))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}
}
