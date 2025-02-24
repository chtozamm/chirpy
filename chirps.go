package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"slices"

	"github.com/chtozamm/chirpy/internal/auth"
	"github.com/chtozamm/chirpy/internal/database"
	"github.com/jackc/pgx/v5/pgtype"
)

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.authSecret)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if params.Body == "" {
		http.Error(w, "body cannot be empty", http.StatusBadRequest)
		return
	}

	pgUUID := pgtype.UUID{}
	err = pgUUID.Scan(userID.String())
	if err != nil {
		log.Printf("Error scanning user_id from request into pgtype.UUID: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if !pgUUID.Valid {
		log.Printf("Failed to validate pgUUID: %v is invalid\n", pgUUID.String())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	newChirp, err := cfg.db.CreateChirp(context.Background(), database.CreateChirpParams{Body: params.Body, UserID: pgUUID})
	if err != nil {
		log.Printf("Error creating a chirp: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(newChirp)
	if err != nil {
		log.Printf("Error marshalling chirp struct: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
}

func (cfg *apiConfig) handleGetChirps(w http.ResponseWriter, r *http.Request) {
	authorID := r.URL.Query().Get("author_id")
	sortOrder := r.URL.Query().Get("sort")

	var chirps []database.Chirp
	var err error

	if authorID != "" {
		authorUUID := pgtype.UUID{}
		authorUUID.Scan(authorID)

		chirps, err = cfg.db.GetChirpsFromAuthor(context.Background(), authorUUID)
		if err != nil {
			log.Printf("Error getting chirps from db: %v\n", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	} else {
		chirps, err = cfg.db.GetChirps(context.Background())
		if err != nil {
			log.Printf("Error getting chirps from db: %v\n", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	if sortOrder == "desc" {
		slices.SortFunc(chirps, func(a, b database.Chirp) int {
			return b.CreatedAt.Time.Compare(a.CreatedAt.Time)
		})
	}

	resp, err := json.Marshal(chirps)
	if err != nil {
		log.Printf("Error marshalling chirps struct: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Write(resp)
}

func (cfg *apiConfig) handleGetChirp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	pgUUID := pgtype.UUID{}
	err := pgUUID.Scan(r.PathValue("chirpID"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	chirps, err := cfg.db.GetChirp(context.Background(), pgUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		log.Printf("Error getting chirp from db: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(chirps)
	if err != nil {
		log.Printf("Error marshalling chirp struct: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Write(resp)
}

func (cfg *apiConfig) handleDeleteChirp(w http.ResponseWriter, r *http.Request) {
	// Validate access token
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, "Access token is missing or invalid in the Authorization header", http.StatusUnauthorized)
		return
	}

	id, err := auth.ValidateJWT(accessToken, cfg.authSecret)
	if err != nil {
		http.Error(w, "Invalid access token", http.StatusUnauthorized)
		return
	}

	userID := pgtype.UUID{}
	userID.Scan(id.String())

	// Parse chirp ID from the path
	pgUUID := pgtype.UUID{}
	err = pgUUID.Scan(r.PathValue("chirpID"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	// Get chirp from database to compare user IDs
	chirp, err := cfg.db.GetChirp(context.Background(), pgUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		log.Printf("Error deleting chirp from db: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if chirp.UserID != userID {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
	}

	// Remove chirp from database
	err = cfg.db.DeleteChirp(context.Background(), pgUUID)
	if err != nil {
		log.Printf("Error deleting chirp from db: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
