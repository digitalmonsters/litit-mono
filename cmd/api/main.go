package main

import (
    "log"
    "net/http"

    "github.com/go-chi/chi/v5"

    // import the user module from this repo
    "github.com/digitalmonsters/litit-mono/internal/user"
)

func main() {
    r := chi.NewRouter()

    // health
    r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("ok"))
    })

    // versioned API
    r.Route("/v1", func(v chi.Router) {
        user.RegisterRoutes(v)
        // auth.RegisterRoutes(v)
        // content.RegisterRoutes(v)
        // comments.RegisterRoutes(v)
        // ads.RegisterRoutes(v)
        // notifications.RegisterRoutes(v)
        // music.RegisterRoutes(v)
        // tokenomics.RegisterRoutes(v)
        // configurator.RegisterRoutes(v)
    })

    log.Println("listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}