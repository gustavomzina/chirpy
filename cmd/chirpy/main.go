package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gustavomzina/chirpy/internal/chirps"
	"github.com/gustavomzina/chirpy/internal/database"
	"github.com/gustavomzina/chirpy/internal/platform"
	"github.com/gustavomzina/chirpy/internal/users"
	_ "github.com/lib/pq"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	// 1. Carregamento de configurações
	dbUrl := os.Getenv("CHIRPY_DB_DSN")
	if dbUrl == "" {
		log.Fatal("CHIRPY_DB_DSN must be set")
	}
	platformName := os.Getenv("CHIRPY_PLATFORM")
	if platformName == "" {
		log.Fatal("CHIRPY_PLATFORM must be set")
	}

	// 2. Inicialização do Banco de Dados
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)

	// 3. Inicialização dos Handlers de domínio
	userHandler := users.Handler{DB: dbQueries}
	chirpHandler := chirps.Handler{DB: dbQueries}

	metricsHandler := platform.Handler{Platform: platformName, DBResetter: dbQueries}

	// 4. Roteamento
	serveMux := http.NewServeMux()

	// Arquivos estáticos e métricas
	fileserverHandler := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))
	serveMux.Handle("GET /app/", metricsHandler.MiddlewareMetricsInc(fileserverHandler))

	serveMux.HandleFunc("GET /api/healthz", handlerReadiness)
	serveMux.Handle("GET /admin/metrics", metricsHandler.HandleMetrics())
	serveMux.Handle("POST /admin/reset", metricsHandler.HandleReset())

	// Rotas de domínio de Chirps
	serveMux.HandleFunc("POST /api/chirps", chirpHandler.HandleCreate)

	// Rotas de domínio de Usuários
	serveMux.HandleFunc("POST /api/users", userHandler.HandleCreate)

	// 5. Inicialização do Servidor
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
