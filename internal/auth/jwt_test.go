package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMakeJWT(t *testing.T) {
	tokenSecret := "my_secret_key"
	userID := uuid.New()

	t.Run("Valid JWT", func(t *testing.T) {
		token, err := MakeJWT(userID, tokenSecret, time.Hour)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
			return []byte(tokenSecret), nil
		})
		assert.NoError(t, err)
		assert.True(t, parsedToken.Valid)

		id, err := parsedToken.Claims.GetSubject()
		assert.NoError(t, err)
		assert.Equal(t, userID.String(), id)
	})
}

func TestValidateJWT(t *testing.T) {
	tokenSecret := "my_secret_key"
	userID := uuid.New()

	validToken, err := MakeJWT(userID, tokenSecret, time.Hour)
	if err != nil {
		t.Fatalf("Failed to create valid token: %v", err)
	}

	t.Run("Valid Token", func(t *testing.T) {
		id, err := ValidateJWT(validToken, tokenSecret)
		assert.NoError(t, err)
		assert.Equal(t, userID, id)
	})

	t.Run("Invalid Token - Malformed User ID", func(t *testing.T) {
		claims := jwt.RegisteredClaims{
			Issuer:    "chirpy",
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour)),
			Subject:   "invalid-uuid",
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		malformedToken, err := token.SignedString([]byte(tokenSecret))
		if err != nil {
			t.Fatalf("Failed to create malformed token: %v", err)
		}

		id, err := ValidateJWT(malformedToken, tokenSecret)
		assert.Error(t, err)
		assert.Equal(t, uuid.Nil, id)
	})

	t.Run("Expired Token", func(t *testing.T) {
		claims := jwt.RegisteredClaims{
			Issuer:    "chirpy",
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(-time.Hour)),
			Subject:   userID.String(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		expiredToken, err := token.SignedString([]byte(tokenSecret))
		if err != nil {
			t.Fatalf("Failed to create expired token: %v", err)
		}

		id, err := ValidateJWT(expiredToken, tokenSecret)
		assert.Error(t, err)
		assert.Equal(t, uuid.Nil, id)
	})
}
