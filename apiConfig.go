package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/gustavomzina/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	queries        *database.Queries
	platform       string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(handler)
}

func (cfg *apiConfig) handlerMetrics() http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		message := fmt.Sprintf(`
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
`, cfg.fileserverHits.Load())

		_, err := w.Write([]byte(message))
		if err != nil {
			log.Fatal(err)
		}
	}

	return http.HandlerFunc(handler)
}

func (cfg *apiConfig) handlerResetMetrics() http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "text/plain; charset=utf-8")
		if cfg.platform != "dev" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		err := cfg.queries.DeleteAllUsers(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		cfg.fileserverHits.Store(0)
	}

	return http.HandlerFunc(handler)
}
