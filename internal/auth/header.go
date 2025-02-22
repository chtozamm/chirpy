package auth

import (
	"fmt"
	"net/http"
	"strings"
)

// token := MakeJWT()
// bearer := fmt.Sprintf("Bearer: %s", token)
// headers.Set(bearer)

// GetBearerToken extracts the Bearer token from the Authorization header.
func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header not provided")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("malformed authorization header")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return "", fmt.Errorf("token is empty")
	}

	return token, nil
}
