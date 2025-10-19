package user

import (
    "net/http"

    "github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router) {
    r.Route("/users", func(rr chi.Router) {
        rr.Get("/{id}", getUser)
    })
}

func getUser(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("user ok"))
}