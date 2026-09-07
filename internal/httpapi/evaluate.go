package httpapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/richardnogueira01/feature-flag-mvp/internal/snapshot"
)

type EvaluateObserver func(found, enabled bool, started time.Time)

func NewEvaluateHandler(store *snapshot.Store, observe EvaluateObserver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		key := strings.TrimPrefix(r.URL.Path, "/v1/evaluate/")
		if key == "" || strings.Contains(key, "/") {
			writeError(w, http.StatusBadRequest, "flag key is required")
			return
		}
		started := time.Now()
		enabled, found, revision := store.Evaluate(key)
		if observe != nil {
			observe(found, enabled, started)
		}
		if !found {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "flag not found", "revision": revision})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"enabled": enabled, "revision": revision})
	})
}
