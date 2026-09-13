package main

import (
	"context"
	"fmt"
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
	mux.Handle("/test/user-menu", agent.UserMenuHandler())
	// /healthz and /readyz allow cross-origin reads: they carry no sensitive data and the
	// admin dashboard (task 015) runs on the Control Plane's own origin, a different port.
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"UP"}`))
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		ready := a.Store().Revision() > 0
		status := http.StatusOK
		if !ready {
			status = http.StatusServiceUnavailable
		}
		w.WriteHeader(status)
		fmt.Fprintf(w, `{"status":%q,"ready":%t,"revision":%d}`, map[bool]string{true: "READY", false: "DEGRADED"}[ready], ready, a.Store().Revision())
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
