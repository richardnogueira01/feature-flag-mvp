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
		}
		defer pool.Close()
		database = persistence.NewStore(pool)
		service = control.NewPersistentService(persistence.NewControlAdapter(database), data)
	} else {
		service = control.NewService(data)
	}
	syncState := syncer.New(data, nil)
	syncState.SetObserver(operationalMetrics)
	if database != nil {
		initial, err := database.Snapshot(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		if err := syncState.Initialize(initial); err != nil {
			log.Fatal(err)
		}
	}
	var eventPublisher *messaging.Publisher
	if natsURL := os.Getenv("NATS_URL"); natsURL != "" {
		conn, err := nats.Connect(natsURL)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Drain()
		stream, err = conn.JetStream()
		if err != nil {
			log.Fatal(err)
		}
		subject := getenv("NATS_SUBJECT", "feature-flags.events")
		durable := getenv("NATS_DURABLE", "feature-flag-mvp")
		if _, err := messaging.NewSubscriber(stream).Subscribe(subject, durable, syncState.Apply); err != nil {
			log.Fatal(err)
		}
		eventPublisher = messaging.NewPublisher(stream, subject)
	}
	if database != nil && stream != nil {
		worker := persistence.NewWorkerWithObserver(database, eventPublisher, 100, time.Second, operationalMetrics)
		go func() {
			if err := worker.Run(context.Background()); err != nil {
				log.Printf("outbox worker stopped: %v", err)
			}
		}()
		log.Println("outbox worker enabled")
	}
	applicationMetrics := appmetrics.New(prometheus.DefaultRegisterer)
	mux := http.NewServeMux()
	mux.Handle("/v1/flags", applicationMetrics.Middleware(httpapi.NewHandler(service)))
	mux.Handle("/v1/evaluate/", httpapi.NewEvaluateHandler(data, applicationMetrics.ObserveEvaluation))
	status := httpapi.NewStatusHandler(syncState)
	mux.Handle("/healthz", status)
	mux.Handle("/readyz", status)
	mux.Handle("/internal/status", status)
	mux.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", mux))
}
func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
