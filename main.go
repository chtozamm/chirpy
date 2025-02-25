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
	filepathRoot   string
	platform       string
	authSecret     string
	polkaKey       string
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	authSecret := os.Getenv("AUTH_SECRET")
	polkaKey := os.Getenv("POLKA_KEY")

	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())

	dbQueries := database.New(conn)

	const filepathRoot = "./static"
	const port = "8080"

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
		authSecret:     authSecret,
		polkaKey:       polkaKey,
		filepathRoot:   filepathRoot,
	}

	mux := getRouter(&apiCfg)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
