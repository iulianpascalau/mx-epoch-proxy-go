package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPerformanceHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil provider", func(t *testing.T) {
		handler, err := NewPerformanceHandler(nil)
		assert.Equal(t, errNilKeyAccessChecker, err)
		assert.Nil(t, handler)
	})

	t.Run("success", func(t *testing.T) {
		handler, err := NewPerformanceHandler(&testscommon.StorerStub{})
		assert.Nil(t, err)
		assert.NotNil(t, handler)
	})
}

func TestPerformanceHandler_ServeHTTP(t *testing.T) {
	t.Parallel()

	t.Run("unauthorized - no token", func(t *testing.T) {
		handler, _ := NewPerformanceHandler(&testscommon.StorerStub{})
		req := httptest.NewRequest(http.MethodGet, "/api/performance", nil)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("forbidden - not admin", func(t *testing.T) {
		SetJwtKey("testkey")
		token, err := GenerateToken("user", false)
		require.Nil(t, err)

		handler, _ := NewPerformanceHandler(&testscommon.StorerStub{})
		req := httptest.NewRequest(http.MethodGet, "/api/performance", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})

	t.Run("success - admin", func(t *testing.T) {
		SetJwtKey("testkey")
		token, err := GenerateToken("admin", true)
		require.Nil(t, err)

		dummyMetrics := map[string]uint64{
			"0-5ms": 100,
		}
		storer := &testscommon.StorerStub{
			GetPerformanceMetricsHandler: func() (map[string]uint64, error) {
				return dummyMetrics, nil
			},
		}

		handler, _ := NewPerformanceHandler(storer)
		req := httptest.NewRequest(http.MethodGet, "/api/performance", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)

		var decoded struct {
			Metrics map[string]uint64 `json:"metrics"`
			Labels  []string          `json:"labels"`
		}
		err = json.NewDecoder(resp.Body).Decode(&decoded)
		assert.Nil(t, err)
		assert.Equal(t, dummyMetrics, decoded.Metrics)
		assert.NotEmpty(t, decoded.Labels)
	})

	t.Run("method not allowed", func(t *testing.T) {
		SetJwtKey("testkey")
		token, err := GenerateToken("admin", true)
		require.Nil(t, err)

		handler, _ := NewPerformanceHandler(&testscommon.StorerStub{})
		req := httptest.NewRequest(http.MethodPost, "/api/performance", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusMethodNotAllowed, resp.Code)
	})

	t.Run("storage error", func(t *testing.T) {
		SetJwtKey("testkey")
		token, err := GenerateToken("admin", true)
		require.Nil(t, err)

		storer := &testscommon.StorerStub{
			GetPerformanceMetricsHandler: func() (map[string]uint64, error) {
				return nil, errors.New("db error")
			},
		}

		handler, _ := NewPerformanceHandler(storer)
		req := httptest.NewRequest(http.MethodGet, "/api/performance", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})
}
