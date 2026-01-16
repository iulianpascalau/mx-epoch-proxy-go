package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAuthenticationMiddleware(t *testing.T) {
	t.Parallel()

	expectedKey := "test-key"
	middleware := NewAuthenticationMiddleware(expectedKey)

	setupRouter := func() *gin.Engine {
		gin.SetMode(gin.TestMode)
		r := gin.New()
		r.Use(middleware.Middleware())
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})
		return r
	}

	t.Run("missing key", func(t *testing.T) {
		r := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid key", func(t *testing.T) {
		r := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Service-Api-Key", "wrong-key")
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("valid key", func(t *testing.T) {
		r := setupRouter()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Service-Api-Key", expectedKey)
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
	})
}
