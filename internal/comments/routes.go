package comments
import ("net/http"; "github.com/go-chi/chi/v5")
func RegisterRoutes(r chi.Router) {
  r.Route("/comments", func(rr chi.Router) {
    rr.Get("/ping", func(w http.ResponseWriter, r *http.Request){ w.Write([]byte("comments ok")) })
  })
}
