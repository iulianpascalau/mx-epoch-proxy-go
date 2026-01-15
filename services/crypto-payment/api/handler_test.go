package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// mockStorage is a mock implementation of the Storage interface
type mockStorage struct {
	addCalled int
	addID     uint64
	addAddr   string
	addErr    error
}

func (m *mockStorage) Add() (uint64, string, error) {
	m.addCalled++
	return m.addID, m.addAddr, m.addErr
}

func (m *mockStorage) IsInterfaceNil() bool {
	return m == nil
}

func TestNewHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil storage", func(t *testing.T) {
		h, err := NewHandler(nil)
		require.EqualError(t, err, "nil storage")
		require.Nil(t, h)
	})

	t.Run("success", func(t *testing.T) {
		h, err := NewHandler(&mockStorage{})
		require.NoError(t, err)
		require.NotNil(t, h)
	})
}

func TestHandler_Ping(t *testing.T) {
	t.Parallel()

	storage := &mockStorage{}
	h, _ := NewHandler(storage)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	h.Ping(c)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	require.Equal(t, "alive", response["status"])
}

func TestHandler_CreateAddress(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		expectedID := uint64(123)
		expectedAddr := "erd1test"
		storage := &mockStorage{
			addID:   expectedID,
			addAddr: expectedAddr,
		}
		h, _ := NewHandler(storage)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		h.CreateAddress(c)

		require.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.Equal(t, float64(expectedID), response["id"]) // json unmarshal numbers as float64
		require.Equal(t, expectedAddr, response["address"])
		require.Equal(t, 1, storage.addCalled)
	})

	t.Run("storage error", func(t *testing.T) {
		expectedErr := errors.New("storage fail")
		storage := &mockStorage{
			addErr: expectedErr,
		}
		h, _ := NewHandler(storage)

		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		h.CreateAddress(c)

		require.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		require.Equal(t, expectedErr.Error(), response["error"])
		require.Equal(t, 1, storage.addCalled)
	})
}
