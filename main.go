package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gustavomzina/chirpy/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	dbUrl := os.Getenv("CHIRPY_DB_DSN")
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)

	apiConfig := &apiConfig{queries: dbQueries}

	serveMux := http.NewServeMux()

	fileserverHandler := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))
	serveMux.Handle("GET /app/", apiConfig.middlewareMetricsInc(fileserverHandler))

	serveMux.HandleFunc("GET /api/healthz", handlerReadiness)
	serveMux.Handle("GET /admin/metrics", apiConfig.handlerMetrics())
	serveMux.Handle("POST /admin/reset", apiConfig.handlerResetMetrics())
	serveMux.HandleFunc("POST /api/validate_chirp", handlerChirpValidator)

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

func handlerChirpValidator(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type returnVals struct {
		Valid        bool   `json:"valid"`
		Cleaned_body string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	in := parameters{}
	err := decoder.Decode(&in)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	const maxChirpLength = 140
	if len(in.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", err)
		return
	}

	cleanedBody := filterBadWords(in.Body)

	respondWithJson(w, http.StatusOK, returnVals{Cleaned_body: cleanedBody, Valid: true})
}

func filterBadWords(s string) string {
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	words := strings.Fields(s)

	for i, word := range words {
		_, isBadWord := badWords[strings.ToLower(word)]
		if isBadWord {
			words[i] = "****"
		}
	}

	return strings.Join(words, " ")
}
