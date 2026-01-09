package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/iulianpascalau/mx-epoch-proxy-go/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth(t *testing.T) {
	auth := NewJWTAuthenticator("test_key")

	t.Run("GenerateToken should return a token", func(t *testing.T) {
		token, err := auth.GenerateToken("user1", true)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		claims, err := auth.ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, "user1", claims.Username)
		assert.True(t, claims.IsAdmin)
	})

	t.Run("ValidateToken with invalid token should error", func(t *testing.T) {
		_, err := auth.ValidateToken("invalid.token.string")
		assert.Error(t, err)
	})

	t.Run("CheckAuth with valid header", func(t *testing.T) {
		token, _ := auth.GenerateToken("user2", false)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		claims, err := auth.CheckAuth(req)
		require.NoError(t, err)
		assert.Equal(t, "user2", claims.Username)
		assert.False(t, claims.IsAdmin)
	})

	t.Run("CheckAuth missing header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		_, err := auth.CheckAuth(req)
		assert.ErrorContains(t, err, "missing authorization header")
	})

	t.Run("CheckAuth invalid header format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		_, err := auth.CheckAuth(req)
		assert.ErrorContains(t, err, "invalid authorization header format")
	})

	t.Run("Expired token should fail", func(t *testing.T) {
		// Manually create expired token
		expirationTime := time.Now().Add(-1 * time.Hour)
		claims := &common.Claims{
			Username: "expiredUser",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expirationTime),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString([]byte("test_key"))

		_, err := auth.ValidateToken(tokenString)
		assert.Error(t, err)
		assert.ErrorIs(t, err, jwt.ErrTokenExpired)
	})
}
