package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/chtozamm/chirpy/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	authSecret     string
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	authSecret := os.Getenv("AUTH_SECRET")

	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())

	dbQueries := database.New(conn)

	const filepathRoot = "."
	const port = "8080"

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
		authSecret:     authSecret,
	}

	mux := http.NewServeMux()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	mux.Handle("/app/", fsHandler)

	mux.HandleFunc("GET /api/chirps", apiCfg.handleGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handleGetChirp)
	mux.HandleFunc("POST /api/chirps", apiCfg.handleCreateChirp)
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.handleUpgradeUser)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handleDeleteChirp)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)

	mux.HandleFunc("POST /api/users", apiCfg.handleCreateUser)
	mux.HandleFunc("PUT /api/users", apiCfg.handleUpdateUser)
	mux.HandleFunc("POST /api/login", apiCfg.handleAuthenticateUser)
	mux.HandleFunc("POST /api/refresh", apiCfg.handleRefreshToken)
	mux.HandleFunc("POST /api/revoke", apiCfg.handleRevokeRefreshToken)

	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)

	mux.HandleFunc("GET /api/healthz", handlerReadiness)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
