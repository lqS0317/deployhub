package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAndValidateJWT(t *testing.T) {
	svc := NewJWTService("test-secret-key-for-jwt-signing")

	token, err := svc.GenerateToken(1, "alice", "admin")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := svc.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, uint(1), claims.UserID)
	assert.Equal(t, "alice", claims.Username)
	assert.Equal(t, "admin", claims.Role)
}

func TestExpiredToken(t *testing.T) {
	svc := &JWTService{
		secret: []byte("test-secret"),
		expiry: -1 * time.Hour,
	}

	token, err := svc.GenerateToken(1, "alice", "admin")
	require.NoError(t, err)

	_, err = svc.ValidateToken(token)
	assert.Error(t, err)
}

func TestInvalidToken(t *testing.T) {
	svc := NewJWTService("test-secret")

	_, err := svc.ValidateToken("invalid-token-string")
	assert.Error(t, err)
}

func TestWrongSecretToken(t *testing.T) {
	svc1 := NewJWTService("secret-1")
	svc2 := NewJWTService("secret-2")

	token, _ := svc1.GenerateToken(1, "alice", "admin")
	_, err := svc2.ValidateToken(token)
	assert.Error(t, err)
}
