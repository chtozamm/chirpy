package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

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
		w.WriteHeader(http.StatusInternalServerError)
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(newUser)
	if err != nil {
		log.Printf("Error marshalling user struct: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
}

func (cfg *apiConfig) handleAuthenticateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
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
			http.Error(w, "email already exists", http.StatusConflict)
			return
		}
		log.Printf("Error creating a user: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	type userResponse struct {
		ID        pgtype.UUID      `json:"id"`
		CreatedAt pgtype.Timestamp `json:"created_at"`
		UpdatedAt pgtype.Timestamp `json:"updated_at"`
		Email     string           `json:"email"`
	}

	userResp := userResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	resp, err := json.Marshal(userResp)
	if err != nil {
		log.Printf("Error marshalling user struct: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
}
