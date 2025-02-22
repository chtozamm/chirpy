package auth

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBearerToken(t *testing.T) {
	t.Run("Valid Authorization Header", func(t *testing.T) {
		tokenTarget := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.KMUFsIDTnFmyG3nMiGM6H9FNFUROf3wh7SmqJp-QV30"

		r := httptest.NewRequest("GET", "http://localhost", nil)
		r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenTarget))

		token, err := GetBearerToken(r.Header)

		assert.NoError(t, err)
		assert.Equal(t, token, tokenTarget)
	})

	t.Run("Malformed Authorization Header", func(t *testing.T) {
		r := httptest.NewRequest("GET", "http://localhost", nil)
		r.Header.Set("Authorization", "")

		token, err := GetBearerToken(r.Header)

		assert.Error(t, err)
		assert.Equal(t, token, "")
	})
}
