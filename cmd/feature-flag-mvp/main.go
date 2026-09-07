package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/richardnogueira01/feature-flag-mvp/internal/control"
	"github.com/richardnogueira01/feature-flag-mvp/internal/httpapi"
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
		defer pool.Close()
		service = control.NewPersistentService(persistence.NewControlAdapter(persistence.NewStore(pool)), data)
		log.Println("using PostgreSQL control plane")
	} else {
		service = control.NewService(data)
		log.Println("using in-memory control plane; set DATABASE_URL for PostgreSQL")
	}

	syncState := syncer.New(data, nil)
	mux := http.NewServeMux()
	mux.Handle("/v1/flags", httpapi.NewHandler(service))
	status := httpapi.NewStatusHandler(syncState)
	mux.Handle("/healthz", status)
	mux.Handle("/readyz", status)
	mux.Handle("/internal/status", status)
	log.Println("feature-flag-mvp listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
