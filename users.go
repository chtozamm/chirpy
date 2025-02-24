package main

import (
	"context"
	"database/sql"
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

	type response struct {
		ID          pgtype.UUID      `json:"id"`
		CreatedAt   pgtype.Timestamp `json:"created_at"`
		UpdatedAt   pgtype.Timestamp `json:"updated_at"`
		Email       string           `json:"email"`
		IsChirpyRed bool             `json:"is_chirpy_red"`
	}

	resp, err := json.Marshal(response{
		ID:          newUser.ID,
		CreatedAt:   newUser.CreatedAt,
		UpdatedAt:   newUser.UpdatedAt,
		Email:       newUser.Email,
		IsChirpyRed: newUser.IsChirpyRed,
	})
	if err != nil {
		log.Printf("Error marshalling response struct: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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

	token, err := auth.MakeJWT(user.ID.Bytes, cfg.authSecret, time.Hour)
	if err != nil {
		log.Printf("Failed to make JWT: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	type response struct {
		ID           pgtype.UUID      `json:"id"`
		CreatedAt    pgtype.Timestamp `json:"created_at"`
		UpdatedAt    pgtype.Timestamp `json:"updated_at"`
		Email        string           `json:"email"`
		IsChirpyRed  bool             `json:"is_chirpy_red"`
		Token        string           `json:"token"`
		RefreshToken string           `json:"refresh_token"`
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("Failed to make refresh token: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	_, err = cfg.db.CreateRefreshToken(context.Background(), database.CreateRefreshTokenParams{
		Token:  refreshToken,
		UserID: user.ID,
	})
	if err != nil {
		log.Printf("Failed to add refresh token to database: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	userResponse := response{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		IsChirpyRed:  user.IsChirpyRed,
		Token:        token,
		RefreshToken: refreshToken,
	}

	resp, err := json.Marshal(userResponse)
	if err != nil {
		log.Printf("Error marshalling response struct: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Write(resp)
}

func (cfg *apiConfig) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, "Authorization header should contain valid refresh token", http.StatusBadRequest)
		return
	}

	refreshToken, err := cfg.db.GetRefreshToken(context.Background(), token)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("Error getting refresh token: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Refresh token is not found", http.StatusUnauthorized)
		return
	}
	if time.Now().Compare(refreshToken.ExpiresAt.Time) > 0 {
		http.Error(w, "Refresh token is expired", http.StatusUnauthorized)
		return
	}
	if refreshToken.RevokedAt.Valid {
		http.Error(w, "Refresh token has been revoked", http.StatusUnauthorized)
		return
	}

	accessToken, err := auth.MakeJWT(refreshToken.UserID.Bytes, cfg.authSecret, time.Hour)
	if err != nil {
		log.Printf("Failed to make JWT: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	type response struct {
		Token string `json:"token"`
	}

	tokenResponse, err := json.Marshal(response{
		Token: accessToken,
	})
	if err != nil {
		log.Printf("Error marshalling response struct: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Write(tokenResponse)
}

func (cfg *apiConfig) handleRevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, "Authorization header should contain valid refresh token", http.StatusBadRequest)
		return
	}

	refreshToken, err := cfg.db.GetRefreshToken(context.Background(), token)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("Error getting refresh token: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Refresh token is not found", http.StatusNotFound)
		return
	}
	if refreshToken.RevokedAt.Valid {
		w.Write([]byte("Refresh token has already bene revoked"))
		return
	}

	timestamp := pgtype.Timestamp{}
	timestamp.Scan(time.Now().UTC())

	err = cfg.db.RevokeRefreshToken(context.Background(), database.RevokeRefreshTokenParams{
		RevokedAt: timestamp,
		UpdatedAt: timestamp,
		Token:     token,
	})
	if err != nil {
		http.Error(w, "Authorization header should contain valid refresh token", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	// Parse request body
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

	if params.Email == "" && params.Password == "" {
		http.Error(w, "No data to update provided", http.StatusBadRequest)
		return
	}

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

	user, err := cfg.db.GetUserByID(context.Background(), userID)
	if err != nil {
		log.Printf("Error getting user from database: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if params.Email != "" {
		user.Email = params.Email
	}

	if params.Password != "" {
		hashedPassword, err := auth.HashPassword(params.Password)
		if err != nil {
			log.Printf("Error hashing password: %v\n", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		user.HashedPassword = hashedPassword
	}

	timestamp := pgtype.Timestamp{}
	timestamp.Scan(time.Now().UTC())

	updatedUser, err := cfg.db.UpdateUser(context.Background(), database.UpdateUserParams{
		ID:             userID,
		Email:          user.Email,
		HashedPassword: user.HashedPassword,
		UpdatedAt:      timestamp,
	})
	if err != nil {
		log.Printf("Error updating user in database: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	type response struct {
		ID        pgtype.UUID      `json:"id"`
		CreatedAt pgtype.Timestamp `json:"created_at"`
		UpdatedAt pgtype.Timestamp `json:"updated_at"`
		Email     string           `json:"email"`
	}

	updatedUserResponse, err := json.Marshal(response{
		ID:        updatedUser.ID,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
		Email:     updatedUser.Email,
	})
	if err != nil {
		log.Printf("Error marshalling response struct: %v\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Write(updatedUserResponse)
}
