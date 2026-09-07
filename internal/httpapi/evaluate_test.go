package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/richardnogueira01/feature-flag-mvp/internal/snapshot"
)

func TestEvaluateUsesMemorySnapshot(t *testing.T) {
	store := snapshot.NewStore(&snapshot.Snapshot{Revision: 7, Flags: map[string]snapshot.Flag{
		"checkout": {Key: "checkout", Enabled: true},
	}})
	handler := NewEvaluateHandler(store, nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/evaluate/checkout", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d", response.Code)
	}
}

func TestEvaluateMissingFlag(t *testing.T) {
	handler := NewEvaluateHandler(snapshot.NewStore(nil), nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/evaluate/missing", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("status=%d", response.Code)
	}
}
