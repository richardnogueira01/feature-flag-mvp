package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/richardnogueira01/feature-flag-mvp/internal/control"
	"github.com/richardnogueira01/feature-flag-mvp/internal/httpapi"
	"github.com/richardnogueira01/feature-flag-mvp/internal/messaging"
	appmetrics "github.com/richardnogueira01/feature-flag-mvp/internal/metrics"
	"github.com/richardnogueira01/feature-flag-mvp/internal/persistence"
	"github.com/richardnogueira01/feature-flag-mvp/internal/snapshot"
	"github.com/richardnogueira01/feature-flag-mvp/internal/syncer"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	data := snapshot.NewStore(nil)
	var service httpapi.Service
	var database *persistence.Store
	var stream nats.JetStreamContext
	if u := os.Getenv("DATABASE_URL"); u != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		pool, e := pgxpool.New(ctx, u)
		if e != nil {
			log.Fatal(e)
		}
		if e = pool.Ping(ctx); e != nil {
			pool.Close()
			log.Fatal(e)
		}
		if os.Getenv("AUTO_MIGRATE") == "true" {
			if e = persistence.Migrate(ctx, pool); e != nil {
				pool.Close()
				log.Fatal(e)
			}
		}
		defer pool.Close()
		database = persistence.NewStore(pool)
		service = control.NewPersistentService(persistence.NewControlAdapter(database), data)
	} else {
		service = control.NewService(data)
	}
	var fetcher syncer.Fetcher
	if database != nil {
		fetcher = database.Snapshot
	}
	syncState := syncer.New(data, fetcher)
	syncState.SetObserver(operationalMetrics)
	if database != nil {
		initial, e := database.Snapshot(context.Background())
		if e != nil {
			log.Fatal(e)
		}
		if e = syncState.Initialize(initial); e != nil {
			log.Fatal(e)
		}
	}
	var publisher *messaging.Publisher
	if u := os.Getenv("NATS_URL"); u != "" {
		conn, e := nats.Connect(u)
		if e != nil {
			log.Fatal(e)
		}
		defer conn.Drain()
		stream, e = conn.JetStream()
		if e != nil {
			log.Fatal(e)
		}
		if _, e = messaging.NewSubscriber(stream).Subscribe(getenv("NATS_SUBJECT", "feature-flags.events"), getenv("NATS_DURABLE", "feature-flag-mvp"), syncState.Apply); e != nil {
			log.Fatal(e)
		}
		publisher = messaging.NewPublisher(stream, getenv("NATS_SUBJECT", "feature-flags.events"))
	}
	if database != nil && stream != nil {
		worker := persistence.NewWorkerWithObserver(database, publisher, 100, time.Second, operationalMetrics)
		go func() {
			if e := worker.Run(context.Background()); e != nil {
				log.Printf("outbox worker stopped: %v", e)
			}
		}()
	}
	m := appmetrics.New(prometheus.DefaultRegisterer)
	mux := http.NewServeMux()
	fh := m.Middleware(httpapi.NewHandler(service))
	mux.Handle("/v1/flags", fh)
	mux.Handle("/v1/flags/", fh)
	mux.Handle("/v1/evaluate/", httpapi.NewEvaluateHandler(data, m.ObserveEvaluation))
	status := httpapi.NewStatusHandler(syncState)
	mux.Handle("/healthz", status)
	mux.Handle("/readyz", status)
	mux.Handle("/internal/status", status)
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/swagger.yaml", httpapi.OpenAPISpecHandler())
	swagger := httpapi.SwaggerUIHandler()
	mux.Handle("/swagger", swagger)
	mux.Handle("/swagger/", swagger)
	mux.Handle("/test/large-payload", httpapi.LargePayloadTestHandler())
	log.Fatal(http.ListenAndServe(":8080", httpapi.Gzip(mux)))
}
func getenv(k, f string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return f
}
