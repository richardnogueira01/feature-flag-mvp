package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type statusProvider struct {
	ready    bool
	revision uint64
}

func (p statusProvider) Ready() bool      { return p.ready }
func (p statusProvider) Revision() uint64 { return p.revision }

func TestStatusEndpoints(t *testing.T) {
	handler := NewStatusHandler(statusProvider{ready: false, revision: 0})
	for _, path := range []string{"/healthz", "/readyz", "/internal/status"} {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, path, nil)
		handler.ServeHTTP(response, request)
		want := http.StatusOK
		if path != "/healthz" {
			want = http.StatusServiceUnavailable
		}
		if response.Code != want {
			t.Fatalf("%s status=%d want=%d", path, response.Code, want)
		}
	}
}
