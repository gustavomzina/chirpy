package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
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
		w.Header().Add("content-type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		message := fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())
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
		w.WriteHeader(http.StatusOK)
		cfg.fileserverHits.Store(0)
	}

	return http.HandlerFunc(handler)
}
