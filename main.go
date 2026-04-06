package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gustavomzina/chirpy/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	dbUrl := os.Getenv("CHIRPY_DB_DSN")
	if dbUrl == "" {
		log.Fatal("CHIRPY_DB_DSN must be set")
	}
	platform := os.Getenv("CHIRPY_PLATFORM")
	if platform == "" {
		log.Fatal("CHIRPY_PLATFORM must be set")
	}

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)

	apiConfig := &apiConfig{queries: dbQueries, platform: platform}

	serveMux := http.NewServeMux()

	fileserverHandler := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))
	serveMux.Handle("GET /app/", apiConfig.middlewareMetricsInc(fileserverHandler))

	serveMux.HandleFunc("GET /api/healthz", handlerReadiness)
	serveMux.Handle("GET /admin/metrics", apiConfig.handlerMetrics())
	serveMux.Handle("POST /admin/reset", apiConfig.handlerResetMetrics())
	serveMux.HandleFunc("POST /api/validate_chirp", handlerChirpValidator)

	serveMux.Handle("POST /api/users", handlerAddUser(apiConfig))

	server := http.Server{
		Addr:    ":" + port,
		Handler: serveMux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}

func handlerAddUser(apiConfig *apiConfig) http.Handler {
	type parameters struct {
		Email string `json:"email"`
	}

	type returnVals struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		in := parameters{}
		err := decoder.Decode(&in)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
			return
		}

		user, err := apiConfig.queries.CreateUser(r.Context(), in.Email)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
			return
		}

		ret := returnVals{
			Id:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		}

		respondWithJson(w, http.StatusCreated, ret)
	}

	return http.HandlerFunc(handler)
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
