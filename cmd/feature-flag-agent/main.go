package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/richardnogueira01/feature-flag-mvp/internal/agent"
	"github.com/richardnogueira01/feature-flag-mvp/internal/httpapi"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	a := agent.New(agent.Env("AGENT_UPSTREAM_URL", "http://app:8080"))
	interval := agent.ParseDuration(os.Getenv("AGENT_POLL_INTERVAL"), 15*time.Second)
	go a.Run(ctx, interval)
	mux := http.NewServeMux()
	mux.Handle("/v1/evaluate/", httpapi.NewEvaluateHandler(a.Store(), nil))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		if a.Store().Revision() == 0 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	server := &http.Server{Addr: agent.ParsePort(agent.Env("AGENT_ADDR", "2772"), ":2772"), Handler: httpapi.Gzip(mux)}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = server.Shutdown(shutdown)
		cancel()
	}()
	log.Printf("feature flag agent listening on %s, polling every %s", server.Addr, interval)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
