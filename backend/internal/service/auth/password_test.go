package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashAndVerifyPassword(t *testing.T) {
	password := "MySecureP@ssw0rd"

	hash, err := HashPassword(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)

	err = VerifyPassword(hash, password)
	assert.NoError(t, err)
}

func TestVerifyWrongPassword(t *testing.T) {
	hash, _ := HashPassword("correct-password")
	err := VerifyPassword(hash, "wrong-password")
	assert.Error(t, err)
}

func TestHashProducesDifferentHashes(t *testing.T) {
	h1, _ := HashPassword("same-password")
	h2, _ := HashPassword("same-password")
	// BCrypt 每次使用不同 salt
	assert.NotEqual(t, h1, h2)
}
