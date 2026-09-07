package main

import (
	"log"
	"net/http"

	"github.com/richardnogueira01/feature-flag-mvp/internal/control"
	"github.com/richardnogueira01/feature-flag-mvp/internal/httpapi"
	"github.com/richardnogueira01/feature-flag-mvp/internal/snapshot"
)

func main() {
	data := snapshot.NewStore(nil)
	service := control.NewService(data)
	log.Println("feature-flag-mvp listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", httpapi.NewHandler(service)))
}
