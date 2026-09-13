package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/richardnogueira01/feature-flag-mvp/internal/control"
	"github.com/richardnogueira01/feature-flag-mvp/internal/snapshot"
)

func TestCRUDHTTPContract(t *testing.T) {
	handler := NewHandler(control.NewService(snapshot.NewStore(nil)))
	server := httptest.NewServer(handler)
	defer server.Close()

	response := request(t, server, http.MethodPost, "/v1/flags", jsonBody(t, map[string]any{"key": "checkout", "enabled": true}))
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("POST status=%d", response.StatusCode)
	}
	response.Body.Close()

	response = request(t, server, http.MethodPut, "/v1/flags/checkout", jsonBody(t, map[string]any{"enabled": false}))
	if response.StatusCode != http.StatusOK {
		t.Fatalf("PUT status=%d", response.StatusCode)
	}
	response.Body.Close()

	response = request(t, server, http.MethodGet, "/v1/flags/checkout", nil)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("GET status=%d", response.StatusCode)
	}
	response.Body.Close()

	response = request(t, server, http.MethodDelete, "/v1/flags/checkout", nil)
	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("DELETE status=%d", response.StatusCode)
	}
	response.Body.Close()
}

func TestHTTPValidationAndNotFound(t *testing.T) {
	handler := NewHandler(control.NewService(nil))
	server := httptest.NewServer(handler)
	defer server.Close()

	response := request(t, server, http.MethodPost, "/v1/flags", bytes.NewReader([]byte("bad")))
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid JSON status=%d", response.StatusCode)
	}
	response.Body.Close()

	response = request(t, server, http.MethodGet, "/v1/flags/missing", nil)
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("missing status=%d", response.StatusCode)
	}
	response.Body.Close()
}

func jsonBody(t *testing.T, value any) io.Reader {
	t.Helper()
	body, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return bytes.NewReader(body)
}

func request(t *testing.T, server *httptest.Server, method, path string, body io.Reader) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, server.URL+path, body)
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return response
}
