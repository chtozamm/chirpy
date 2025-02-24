package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/chtozamm/chirpy/internal/auth"
	"github.com/chtozamm/chirpy/internal/database"
	"github.com/jackc/pgx/v5/pgtype"
)

func (cfg *apiConfig) handleUpgradeUser(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil || apiKey != cfg.polkaKey {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if params.Event == "" {
		http.Error(w, "event field cannot be empty", http.StatusBadRequest)
		return
	}
	if params.Data.UserID == "" {
		http.Error(w, "data.user_id field cannot be empty", http.StatusBadRequest)
		return
	}
	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	userID := pgtype.UUID{}
	userID.Scan(params.Data.UserID)
	timestamp := pgtype.Timestamp{}
	timestamp.Scan(time.Now().UTC())

	err = cfg.db.UpgradeUser(context.Background(), database.UpgradeUserParams{
		UpdatedAt: timestamp,
		ID:        userID,
	})
	if err != nil {
		log.Printf("Error from database while upgrading user: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
