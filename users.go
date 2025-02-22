package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/chtozamm/chirpy/internal/auth"
	"github.com/chtozamm/chirpy/internal/database"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if params.Email == "" {
		http.Error(w, "email field cannot be empty", http.StatusBadRequest)
		return
	}
	if params.Password == "" {
		http.Error(w, "password field cannot be empty", http.StatusBadRequest)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("Error hashing password: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	newUser, err := cfg.db.CreateUser(context.Background(), database.CreateUserParams{Email: params.Email, HashedPassword: hashedPassword})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			http.Error(w, "email already exists", http.StatusConflict)
			return
		}
		log.Printf("Error creating a user: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(newUser)
	if err != nil {
		log.Printf("Error marshalling user struct: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
}

func (cfg *apiConfig) handleAuthenticateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if params.Email == "" {
		http.Error(w, "email field cannot be empty", http.StatusBadRequest)
		return
	}
	if params.Password == "" {
		http.Error(w, "password field cannot be empty", http.StatusBadRequest)
		return
	}

	user, err := cfg.db.GetUserByEmail(context.Background(), params.Email)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			http.Error(w, "Email already exists", http.StatusConflict)
			return
		}
		log.Printf("Error creating a user: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var expires time.Duration
	if params.ExpiresInSeconds > 3600 || params.ExpiresInSeconds == 0 {
		expires = time.Hour
	} else {
		expires = time.Second * time.Duration(params.ExpiresInSeconds)
	}

	token, err := auth.MakeJWT(user.ID.Bytes, cfg.authSecret, expires)
	if err != nil {
		log.Printf("Failed to make JWT: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	type userResponse struct {
		ID        pgtype.UUID      `json:"id"`
		CreatedAt pgtype.Timestamp `json:"created_at"`
		UpdatedAt pgtype.Timestamp `json:"updated_at"`
		Email     string           `json:"email"`
		Token     string           `json:"token"`
	}

	userResp := userResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     token,
	}

	resp, err := json.Marshal(userResp)
	if err != nil {
		log.Printf("Error marshalling user struct: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
}
