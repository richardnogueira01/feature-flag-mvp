package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/richardnogueira01/feature-flag-mvp/internal/snapshot"
)

func BenchmarkEvaluateHTTPParallel(b *testing.B) {
	handler := NewEvaluateHandler(
		snapshot.NewStore(&snapshot.Snapshot{Revision: 1, Flags: map[string]snapshot.Flag{
			"checkout": {Key: "checkout", Enabled: true},
		}}),
		nil,
	)
	server := httptest.NewServer(handler)
	defer server.Close()
	client := server.Client()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			response, err := client.Get(server.URL + "/v1/evaluate/checkout")
			if err != nil {
				b.Error(err)
				continue
			}
			response.Body.Close()
			if response.StatusCode != http.StatusOK {
				b.Errorf("status=%d", response.StatusCode)
			}
		}
	})
}
