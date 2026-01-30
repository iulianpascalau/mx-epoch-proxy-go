package api

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
)

const serviceApiKeyHeader = "X-Service-Api-Key"

// AuthenticationMiddleware handles the authentication for the API
type AuthenticationMiddleware struct {
	expectedApiKey string
}

// NewAuthenticationMiddleware creates a new AuthenticationMiddleware instance
func NewAuthenticationMiddleware(expectedApiKey string) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{
		expectedApiKey: expectedApiKey,
	}
}

// Middleware returns the gin middleware function
func (am *AuthenticationMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader(serviceApiKeyHeader)
		if len(apiKey) == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing API key"})
			return
		}

		// Constant time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(apiKey), []byte(am.expectedApiKey)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
			return
		}

		c.Next()
	}
}
