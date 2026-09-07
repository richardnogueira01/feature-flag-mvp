package httpapi

import (
	"github.com/richardnogueira01/feature-flag-mvp/internal/snapshot"
	"net/http"
	"strings"
	"time"
)

type EvaluateObserver func(found, enabled bool, started time.Time)

func NewEvaluateHandler(store *snapshot.Store, observe EvaluateObserver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			writeError(w, 405, "method not allowed")
			return
		}
		key := strings.TrimPrefix(r.URL.Path, "/v1/evaluate/")
		if key == "" || strings.Contains(key, "/") {
			writeError(w, 400, "flag key is required")
			return
		}
		started := time.Now()
		value, found, revision := store.EvaluateValue(key)
		enabled, _, _ := store.Evaluate(key)
		if observe != nil {
			observe(found, enabled, started)
		}
		if !found {
			writeJSON(w, 404, map[string]any{"error": "flag not found", "revision": revision})
			return
		}
		if !enabled {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		writeJSON(w, 200, map[string]any{"value": value, "enabled": enabled, "revision": revision})
	})
}
