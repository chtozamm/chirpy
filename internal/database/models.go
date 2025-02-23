// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package database

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Chirp struct {
	ID        pgtype.UUID      `json:"id"`
	CreatedAt pgtype.Timestamp `json:"created_at"`
	UpdatedAt pgtype.Timestamp `json:"updated_at"`
	Body      string           `json:"body"`
	UserID    pgtype.UUID      `json:"user_id"`
}

type RefreshToken struct {
	Token     string           `json:"token"`
	CreatedAt pgtype.Timestamp `json:"created_at"`
	UpdatedAt pgtype.Timestamp `json:"updated_at"`
	UserID    pgtype.UUID      `json:"user_id"`
	ExpiresAt pgtype.Timestamp `json:"expires_at"`
	RevokedAt pgtype.Timestamp `json:"revoked_at"`
}

type User struct {
	ID             pgtype.UUID      `json:"id"`
	CreatedAt      pgtype.Timestamp `json:"created_at"`
	UpdatedAt      pgtype.Timestamp `json:"updated_at"`
	Email          string           `json:"email"`
	HashedPassword string           `json:"hashed_password"`
}
