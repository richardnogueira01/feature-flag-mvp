package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

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
)

func main() {
	data := snapshot.NewStore(nil)
	var service httpapi.Service
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		pool, err := pgxpool.New(ctx, databaseURL)
		if err != nil {
			log.Fatal(err)
		}
		if err := pool.Ping(ctx); err != nil {
			pool.Close()
			log.Fatal(err)
		}
		if os.Getenv("AUTO_MIGRATE") == "true" {
			if err := persistence.Migrate(ctx, pool); err != nil {
				pool.Close()
				log.Fatal(err)
			}
			log.Println("database migrations applied")
		}
		defer pool.Close()
		service = control.NewPersistentService(persistence.NewControlAdapter(persistence.NewStore(pool)), data)
		log.Println("using PostgreSQL control plane")
	} else {
		service = control.NewService(data)
		log.Println("using in-memory control plane; set DATABASE_URL for PostgreSQL")
	}

	syncState := syncer.New(data, nil)
	if natsURL := os.Getenv("NATS_URL"); natsURL != "" {
		conn, err := nats.Connect(natsURL)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Drain()
		stream, err := conn.JetStream()
		if err != nil {
			log.Fatal(err)
		}
		if _, err := messaging.NewSubscriber(stream).Subscribe(getenv("NATS_SUBJECT", "feature-flags.events"), getenv("NATS_DURABLE", "feature-flag-mvp"), syncState.Apply); err != nil {
			log.Fatal(err)
		}
		log.Println("NATS subscriber enabled")
	}

	applicationMetrics := appmetrics.New(prometheus.DefaultRegisterer)
	api := applicationMetrics.Middleware(httpapi.NewHandler(service))
	status := httpapi.NewStatusHandler(syncState)
	mux := http.NewServeMux()
	mux.Handle("/v1/flags", api)
	mux.Handle("/healthz", status)
	mux.Handle("/readyz", status)
	mux.Handle("/internal/status", status)
	mux.Handle("/metrics", promhttp.Handler())
	log.Println("feature-flag-mvp listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
