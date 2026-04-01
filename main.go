package main

import (
	"log"
	"net/http"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	apiConfig := &apiConfig{}
	serveMux := http.NewServeMux()

	fileserverHandler := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))
	serveMux.Handle("GET /app/", apiConfig.middlewareMetricsInc(fileserverHandler))

	serveMux.HandleFunc("GET /api/healthz", handlerReadiness)
	serveMux.Handle("GET /admin/metrics", apiConfig.handlerMetrics())
	serveMux.Handle("POST /admin/reset", apiConfig.handlerResetMetrics())

	server := http.Server{
		Addr:    ":" + port,
		Handler: serveMux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte(http.StatusText(http.StatusOK)))
	if err != nil {
		log.Fatal(err)
	}
}
