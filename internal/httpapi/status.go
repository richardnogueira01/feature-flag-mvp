package httpapi

import (
	"encoding/json"
	"net/http"
)

type StatusProvider interface {
	Ready() bool
	Revision() uint64
}

func NewStatusHandler(provider StatusProvider) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "UP"})
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		if !provider.Ready() {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "NOT_READY"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "READY"})
	})
	mux.HandleFunc("/internal/status", func(w http.ResponseWriter, _ *http.Request) {
		status := http.StatusOK
		if !provider.Ready() {
			status = http.StatusServiceUnavailable
		}
		writeStatusJSON(w, status, provider)
	})
	return mux
}

func writeStatusJSON(w http.ResponseWriter, status int, provider StatusProvider) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":       map[bool]string{true: "UP", false: "DEGRADED"}[provider.Ready()],
		"revision":     provider.Revision(),
		"synchronized": provider.Ready(),
	})
}
