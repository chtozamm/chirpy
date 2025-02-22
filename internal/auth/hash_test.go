package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckPasswordHash(t *testing.T) {
	t.Run("Valid Hash", func(t *testing.T) {
		password := "hellothere"
		hash := "$2y$10$UG5KuRMZ1qgRN0mwdFPItuQC1ei.a2RJsnqefo6bvblbqJLLvbkeu"

		err := CheckPasswordHash(password, hash)
		assert.NoError(t, err)
	})

	t.Run("Invalid Hash", func(t *testing.T) {
		password := "hellothere"
		hash := "invalid-hash"

		err := CheckPasswordHash(password, hash)
		assert.Error(t, err)
	})
}
