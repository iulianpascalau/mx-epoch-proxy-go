package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth(t *testing.T) {
	SetJwtKey("test_secret")

	t.Run("GenerateToken should return a token", func(t *testing.T) {
		token, err := GenerateToken("user1", true)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		claims, err := ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, "user1", claims.Username)
		assert.True(t, claims.IsAdmin)
	})

	t.Run("ValidateToken with invalid token should error", func(t *testing.T) {
		_, err := ValidateToken("invalid.token.string")
		assert.Error(t, err)
	})

	t.Run("CheckAuth with valid header", func(t *testing.T) {
		token, _ := GenerateToken("user2", false)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		claims, err := CheckAuth(req)
		require.NoError(t, err)
		assert.Equal(t, "user2", claims.Username)
		assert.False(t, claims.IsAdmin)
	})

	t.Run("CheckAuth missing header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		_, err := CheckAuth(req)
		assert.ErrorContains(t, err, "missing authorization header")
	})

	t.Run("CheckAuth invalid header format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		_, err := CheckAuth(req)
		assert.ErrorContains(t, err, "invalid authorization header format")
	})

	t.Run("Expired token should fail", func(t *testing.T) {
		// Manually create expired token
		expirationTime := time.Now().Add(-1 * time.Hour)
		claims := &Claims{
			Username: "expiredUser",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expirationTime),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString([]byte("test_secret"))

		_, err := ValidateToken(tokenString)
		assert.Error(t, err)
		assert.ErrorIs(t, err, jwt.ErrTokenExpired)
	})
}
