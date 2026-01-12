package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/common"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAccessKeysHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil provider", func(t *testing.T) {
		handler, err := NewAccessKeysHandler(nil, &testscommon.AuthenticatorStub{})
		assert.Nil(t, handler)
		assert.Equal(t, errNilKeyAccessProvider, err)
	})

	t.Run("nil authenticator", func(t *testing.T) {
		handler, err := NewAccessKeysHandler(&testscommon.StorerStub{}, nil)
		assert.Nil(t, handler)
		assert.Equal(t, errNilAuthenticator, err)
	})

	t.Run("ok", func(t *testing.T) {
		handler, err := NewAccessKeysHandler(&testscommon.StorerStub{}, &testscommon.AuthenticatorStub{})
		assert.NotNil(t, handler)
		assert.Nil(t, err)
	})
}

func TestAccessKeysHandler_ServeHTTP(t *testing.T) {
	t.Parallel()

	auth := NewJWTAuthenticator("test_key")

	t.Run("method not allowed", func(t *testing.T) {
		username := "user1"
		token, _ := auth.GenerateToken(username, false)

		handler, _ := NewAccessKeysHandler(&testscommon.StorerStub{}, auth)
		req := httptest.NewRequest(http.MethodTrace, "/api/admin-access-keys", nil)
		resp := httptest.NewRecorder()

		req.Header.Set("Authorization", "Bearer "+token)

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusMethodNotAllowed, resp.Code)
	})

	t.Run("unauthorized - no token", func(t *testing.T) {
		handler, _ := NewAccessKeysHandler(&testscommon.StorerStub{}, auth)
		req := httptest.NewRequest(http.MethodGet, "/api/admin-access-keys", nil)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("authorized - get keys", func(t *testing.T) {
		username := "user1"
		token, _ := auth.GenerateToken(username, false)

		provider := &testscommon.StorerStub{
			GetAllKeysHandler: func(usr string) (map[string]common.AccessKeyDetails, error) {
				assert.Equal(t, username, usr)
				return map[string]common.AccessKeyDetails{
					"key1": {MaxRequests: 100},
				}, nil
			},
		}
		handler, _ := NewAccessKeysHandler(provider, auth)
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

	t.Run("authorized - get keys as admin", func(t *testing.T) {
		username := "user1"
		token, _ := auth.GenerateToken(username, true)

		provider := &testscommon.StorerStub{
			GetAllKeysHandler: func(usr string) (map[string]common.AccessKeyDetails, error) {
				assert.Equal(t, "", usr)
				return map[string]common.AccessKeyDetails{
					"key1": {MaxRequests: 100},
				}, nil
			},
		}
		handler, _ := NewAccessKeysHandler(provider, auth)
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

		expectedKey := "key1_longer_than_12_chars"
		expectedUsername := "admin"
		token, _ := auth.GenerateToken(expectedUsername, true)

		provider := &testscommon.StorerStub{
			AddKeyHandler: func(username string, key string) error {
				assert.Equal(t, expectedUsername, username)
				assert.Equal(t, expectedKey, key)
				return nil
			},
		}

		handler, err := NewAccessKeysHandler(provider, auth)
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

	t.Run("post as admin - success", func(t *testing.T) {
		t.Parallel()

		expectedKey := "key1_longer_than_12_chars"
		expectedUsername := "user1"
		admin := "admin"
		token, _ := auth.GenerateToken(admin, true)

		provider := &testscommon.StorerStub{
			AddKeyHandler: func(username string, key string) error {
				assert.Equal(t, expectedUsername, username)
				assert.Equal(t, expectedKey, key)
				return nil
			},
		}

		handler, err := NewAccessKeysHandler(provider, auth)
		require.Nil(t, err)

		reqBody := addKeyRequest{
			Key:      expectedKey,
			Username: expectedUsername,
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/admin-access-keys", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("post - empty key generates key", func(t *testing.T) {
		t.Parallel()

		expectedUsername := "admin"
		token, _ := auth.GenerateToken(expectedUsername, true)

		provider := &testscommon.StorerStub{
			AddKeyHandler: func(username string, key string) error {
				assert.Equal(t, expectedUsername, username)
				assert.Len(t, key, 32)
				return nil
			},
		}

		handler, err := NewAccessKeysHandler(provider, auth)
		require.Nil(t, err)

		reqBody := addKeyRequest{Key: ""}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/admin-access-keys", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("post - short key failure", func(t *testing.T) {
		t.Parallel()

		token, _ := auth.GenerateToken("admin", true)
		handler, _ := NewAccessKeysHandler(&testscommon.StorerStub{}, auth)

		reqBody := addKeyRequest{Key: "short"}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/admin-access-keys", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "key must be at least 12 characters long")
	})

	t.Run("delete fails if no key is provided", func(t *testing.T) {
		t.Parallel()

		username := "admin"
		token, _ := auth.GenerateToken(username, true)

		provider := &testscommon.StorerStub{
			RemoveKeyHandler: func(usr string, key string) error {
				assert.Fail(t, "should not be called")
				return nil
			},
		}

		handler, err := NewAccessKeysHandler(provider, auth)
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodDelete, "/api/admin-access-keys?key=", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("delete - success", func(t *testing.T) {
		t.Parallel()

		expectedKey := "key1"
		username := "admin"
		token, _ := auth.GenerateToken(username, true)

		provider := &testscommon.StorerStub{
			RemoveKeyHandler: func(usr string, key string) error {
				assert.Equal(t, strings.ToLower(expectedKey), key)
				assert.Equal(t, username, usr)
				return nil
			},
		}

		handler, err := NewAccessKeysHandler(provider, auth)
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodDelete, "/api/admin-access-keys?key="+expectedKey, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("delete as admin - success", func(t *testing.T) {
		t.Parallel()

		expectedKey := "key1"
		expectedUsername := "user1"
		admin := "admin"
		token, _ := auth.GenerateToken(admin, true)

		provider := &testscommon.StorerStub{
			RemoveKeyHandler: func(usr string, key string) error {
				assert.Equal(t, strings.ToLower(expectedKey), key)
				assert.Equal(t, expectedUsername, usr)
				return nil
			},
		}

		handler, err := NewAccessKeysHandler(provider, auth)
		require.Nil(t, err)

		req := httptest.NewRequest(http.MethodDelete, "/api/admin-access-keys?key="+expectedKey+"&username="+expectedUsername, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
