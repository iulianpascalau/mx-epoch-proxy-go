package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/common"
	"github.com/iulianpascalau/mx-epoch-proxy-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAccessKeysHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil provider", func(t *testing.T) {
		handler, err := NewAccessKeysHandler(nil)
		assert.Nil(t, handler)
		assert.Equal(t, errNilKeyAccessChecker, err)
	})

	t.Run("ok", func(t *testing.T) {
		handler, err := NewAccessKeysHandler(&testscommon.StorerStub{})
		assert.NotNil(t, handler)
		assert.Nil(t, err)
	})
}

func TestAccessKeysHandler_ServeHTTP(t *testing.T) {
	t.Parallel()

	SetJwtKey("test_key")

	t.Run("unauthorized - no token", func(t *testing.T) {
		handler, _ := NewAccessKeysHandler(&testscommon.StorerStub{})
		req := httptest.NewRequest(http.MethodGet, "/api/admin-access-keys", nil)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("authorized - get keys", func(t *testing.T) {
		username := "user1"
		token, _ := GenerateToken(username, false)

		provider := &testscommon.StorerStub{
			GetAllKeysHandler: func(usr string) (map[string]common.AccessKeyDetails, error) {
				assert.Equal(t, username, usr)
				return map[string]common.AccessKeyDetails{
					"key1": {MaxRequests: 100},
				}, nil
			},
		}
		handler, _ := NewAccessKeysHandler(provider)
		req := httptest.NewRequest(http.MethodGet, "/api/admin-access-keys", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)

		var keys map[string]common.AccessKeyDetails
		err := json.NewDecoder(resp.Body).Decode(&keys)
		assert.Nil(t, err)
		assert.Len(t, keys, 1)
	})

	t.Run("post - success", func(t *testing.T) {
		t.Parallel()

		expectedKey := "key1"
		expectedUsername := "admin"
		token, _ := GenerateToken(expectedUsername, true)

		provider := &testscommon.StorerStub{
			AddKeyHandler: func(username string, key string) error {
				assert.Equal(t, expectedUsername, username)
				assert.Equal(t, expectedKey, key)
				return nil
			},
		}

		handler, err := NewAccessKeysHandler(provider)
		require.Nil(t, err)

		reqBody := addKeyRequest{
			Key: expectedKey,
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/admin-access-keys", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("delete - success", func(t *testing.T) {
		t.Parallel()

		expectedKey := "key1"
		username := "admin"
		token, _ := GenerateToken(username, true)

		provider := &testscommon.StorerStub{
			RemoveKeyHandler: func(usr string, key string) error {
				assert.Equal(t, strings.ToLower(expectedKey), key)
				assert.Equal(t, username, usr)
				return nil
			},
		}

		handler, err := NewAccessKeysHandler(provider)
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodDelete, "/api/admin-access-keys?key="+expectedKey, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
