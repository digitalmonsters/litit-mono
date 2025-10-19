package main

import (
	// ...
	"log"
	"net/http"

	"github.com/digitalmonsters/litit-mono/internal/configurator"
	"github.com/digitalmonsters/litit-mono/internal/notifications"
	"github.com/digitalmonsters/litit-mono/internal/user"
	"github.com/go-chi/chi/v5"
)

var Version = "dev"

func main() {
	r := chi.NewRouter()

	// health & version
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(Version))
	})

	// versioned API surface
	r.Route("/v1", func(v chi.Router) {
		user.RegisterRoutes(v)
		configurator.RegisterRoutes(v)
		notifications.RegisterRoutes(v)
	})

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
