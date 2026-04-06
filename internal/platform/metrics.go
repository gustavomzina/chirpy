package platform

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type DatabaseResetter interface {
	DeleteAllUsers(ctx context.Context) error
}

type Handler struct {
	fileserverHits atomic.Int32
	Platform       string
	DBResetter     DatabaseResetter
}

func (s *Handler) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (s *Handler) HandleMetrics() http.Handler {
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
`, s.fileserverHits.Load())

		_, err := w.Write([]byte(message))
		if err != nil {
			log.Fatal(err)
		}
	}

	return http.HandlerFunc(handler)
}

func (s *Handler) HandleReset() http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "text/plain; charset=utf-8")
		if s.Platform != "dev" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		err := s.DBResetter.DeleteAllUsers(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		s.fileserverHits.Store(0)
	}

	return http.HandlerFunc(handler)
}
