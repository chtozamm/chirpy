package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/chtozamm/chirpy/internal/database"
	"github.com/jackc/pgx/v5/pgtype"
)

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   string `json:"body"`
		UserID string `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if params.Body == "" {
		http.Error(w, "body cannot be empty", http.StatusBadRequest)
		return
	}

	if params.UserID == "" {
		http.Error(w, "user_id cannot be empty", http.StatusBadRequest)
		return
	}

	pgUUID := pgtype.UUID{}
	err = pgUUID.Scan(params.UserID)
	if err != nil {
		log.Printf("Error scanning user_id from request into pgtype.UUID: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !pgUUID.Valid {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	newChirp, err := cfg.db.CreateChirp(context.Background(), database.CreateChirpParams{Body: params.Body, UserID: pgUUID})
	if err != nil {
		log.Printf("Error creating a chirp: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(newChirp)
	if err != nil {
		log.Printf("Error marshalling chirp struct: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
}
