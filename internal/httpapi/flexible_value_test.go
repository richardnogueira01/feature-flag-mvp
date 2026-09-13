package httpapi

import (
	"github.com/richardnogueira01/feature-flag-mvp/internal/control"
	"github.com/richardnogueira01/feature-flag-mvp/internal/snapshot"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerSupportsJSONValues(t *testing.T) {
	server := httptest.NewServer(NewHandler(control.NewService(snapshot.NewStore(nil))))
	defer server.Close()
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/v1/flags", strings.NewReader(`{"key":"timeout_ms","value":1500}`))
	req.Header.Set("Content-Type", "application/json")
	res, e := http.DefaultClient.Do(req)
	if e != nil {
		t.Fatal(e)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusCreated || !strings.Contains(string(body), `"Value":1500`) {
		t.Fatalf("status=%d body=%s", res.StatusCode, body)
	}
}
