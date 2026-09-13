package httpapi

import (
	"encoding/json"
	"fmt"
	"github.com/richardnogueira01/feature-flag-mvp/internal/snapshot"
	"net/http"
	"strings"
	"sync"
	"time"
)

type EvaluateObserver func(found, enabled bool, started time.Time)
type evaluationCacheEntry struct {
	revision uint64
	body     []byte
}
type evaluationCache struct {
	mu    sync.RWMutex
	items map[string]evaluationCacheEntry
}

func newEvaluationCache() *evaluationCache {
	return &evaluationCache{items: make(map[string]evaluationCacheEntry)}
}
func (c *evaluationCache) get(k string, rev uint64) ([]byte, bool) {
	c.mu.RLock()
	e, ok := c.items[k]
	c.mu.RUnlock()
	if !ok || e.revision != rev {
		return nil, false
	}
	return e.body, true
}
func (c *evaluationCache) set(k string, rev uint64, b []byte) {
	c.mu.Lock()
	c.items[k] = evaluationCacheEntry{revision: rev, body: append([]byte(nil), b...)}
	c.mu.Unlock()
}
func NewEvaluateHandler(store *snapshot.Store, observe EvaluateObserver) http.Handler {
	return newEvaluateHandler(store, observe, newEvaluationCache())
}
func newEvaluateHandler(store *snapshot.Store, observe EvaluateObserver, cache *evaluationCache) http.Handler {
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
		w.Header().Set("Cache-Control", "private, max-age=1")
		w.Header().Set("ETag", fmt.Sprintf(`"%d"`, revision))
		if r.Header.Get("If-None-Match") == fmt.Sprintf(`"%d"`, revision) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		if !enabled {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		body, ok := cache.get(key, revision)
		if !ok {
			body, _ = json.Marshal(map[string]any{"value": value, "enabled": enabled, "revision": revision})
			cache.set(key, revision, body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write(body)
	})
}
