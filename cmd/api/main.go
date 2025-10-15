package main

import (
    "log"
    "net/http"
    "github.com/go-chi/chi/v5"
)

func main() {
    r := chi.NewRouter()
    r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("ok"))
    })
    log.Println("listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}