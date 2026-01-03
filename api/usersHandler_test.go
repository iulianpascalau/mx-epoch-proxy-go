package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/common"
	"github.com/iulianpascalau/mx-epoch-proxy-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewUsersHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil key access provider", func(t *testing.T) {
		t.Parallel()

		handler, err := NewUsersHandler(nil)
		assert.Equal(t, errNilKeyAccessChecker, err)
		assert.Nil(t, handler)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		provider := &testscommon.StorerStub{}
		handler, err := NewUsersHandler(provider)
		assert.Nil(t, err)
		assert.NotNil(t, handler)
	})
}

func TestUsersHandler_ServeHTTP(t *testing.T) {
	t.Parallel()

	SetJwtKey("test_key")

	t.Run("method not allowed", func(t *testing.T) {
		t.Parallel()

		token, _ := GenerateToken("admin", true)

		handler, _ := NewUsersHandler(&testscommon.StorerStub{})
		req := httptest.NewRequest(http.MethodTrace, "/api/admin-users", nil)
		resp := httptest.NewRecorder()
		req.Header.Set("Authorization", "Bearer "+token)

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusMethodNotAllowed, resp.Code)
	})

	t.Run("unauthorized - missing token", func(t *testing.T) {
		t.Parallel()

		handler, _ := NewUsersHandler(&testscommon.StorerStub{})
		req := httptest.NewRequest(http.MethodGet, "/api/admin-users", nil)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("forbidden - not admin", func(t *testing.T) {
		t.Parallel()

		token, _ := GenerateToken("user", false)

		provider := &testscommon.StorerStub{}
		handler, _ := NewUsersHandler(provider)
		req := httptest.NewRequest(http.MethodGet, "/api/admin-users", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})

	t.Run("authorized - get users", func(t *testing.T) {
		t.Parallel()

		token, _ := GenerateToken("admin", true)

		provider := &testscommon.StorerStub{
			GetAllUsersHandler: func() (map[string]common.UsersDetails, error) {
				return map[string]common.UsersDetails{
					"user1": {Username: "user1", IsAdmin: false},
				}, nil
			},
		}
		handler, _ := NewUsersHandler(provider)
		req := httptest.NewRequest(http.MethodGet, "/api/admin-users", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)

		var users map[string]common.UsersDetails
		err := json.NewDecoder(resp.Body).Decode(&users)
		assert.Nil(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, "user1", users["user1"].Username)
	})

	t.Run("error getting users", func(t *testing.T) {
		t.Parallel()

		token, _ := GenerateToken("admin", true)

		provider := &testscommon.StorerStub{
			GetAllUsersHandler: func() (map[string]common.UsersDetails, error) {
				return nil, errors.New("db error")
			},
		}
		handler, _ := NewUsersHandler(provider)
		req := httptest.NewRequest(http.MethodGet, "/api/admin-users", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})

	t.Run("post - success", func(t *testing.T) {
		t.Parallel()

		token, _ := GenerateToken("admin", true)

		expectedUsername := "user2"
		expectedPassword := "password"
		expectedMaxRequests := uint64(500)

		provider := &testscommon.StorerStub{
			AddUserHandler: func(username string, password string, isAdmin bool, maxRequests uint64) error {
				assert.Equal(t, expectedUsername, username)
				assert.Equal(t, expectedPassword, password)
				assert.True(t, isAdmin)
				assert.Equal(t, expectedMaxRequests, maxRequests)
				return nil
			},
		}
		handler, _ := NewUsersHandler(provider)

		reqBody := addUserRequest{
			Username:    expectedUsername,
			Password:    expectedPassword,
			IsAdmin:     true,
			MaxRequests: expectedMaxRequests,
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/admin-users", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
	})

	t.Run("post - invalid request body", func(t *testing.T) {
		t.Parallel()

		token, _ := GenerateToken("admin", true)

		provider := &testscommon.StorerStub{}
		handler, _ := NewUsersHandler(provider)

		req := httptest.NewRequest(http.MethodPost, "/api/admin-users", bytes.NewBufferString("invalid json"))
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("post - missing username", func(t *testing.T) {
		t.Parallel()

		token, _ := GenerateToken("admin", true)

		provider := &testscommon.StorerStub{}
		handler, _ := NewUsersHandler(provider)

		reqBody := addUserRequest{
			Password: "pass",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/admin-users", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("put fails if no username is provided", func(t *testing.T) {
		t.Parallel()

		token, _ := GenerateToken("admin", true)

		expectedPassword := "newpassword"
		expectedMaxRequests := uint64(1000)

		provider := &testscommon.StorerStub{
			UpdateUserHandler: func(username string, password string, isAdmin bool, maxRequests uint64) error {
				assert.Fail(t, "should have not called this")
				return nil
			},
		}
		handler, _ := NewUsersHandler(provider)

		reqBody := addUserRequest{
			Username:    "",
			Password:    expectedPassword,
			IsAdmin:     false,
			MaxRequests: expectedMaxRequests,
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/api/admin-users", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("put - success", func(t *testing.T) {
		t.Parallel()

		token, _ := GenerateToken("admin", true)

		expectedUsername := "user2"
		expectedPassword := "newpassword"
		expectedMaxRequests := uint64(1000)

		provider := &testscommon.StorerStub{
			UpdateUserHandler: func(username string, password string, isAdmin bool, maxRequests uint64) error {
				assert.Equal(t, expectedUsername, username)
				assert.Equal(t, expectedPassword, password)
				assert.False(t, isAdmin)
				assert.Equal(t, expectedMaxRequests, maxRequests)
				return nil
			},
		}
		handler, _ := NewUsersHandler(provider)

		reqBody := addUserRequest{
			Username:    expectedUsername,
			Password:    expectedPassword,
			IsAdmin:     false,
			MaxRequests: expectedMaxRequests,
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/api/admin-users", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
	})

	t.Run("delete fails if no username is provided", func(t *testing.T) {
		t.Parallel()

		token, _ := GenerateToken("admin", true)

		provider := &testscommon.StorerStub{
			RemoveUserHandler: func(username string) error {
				assert.Fail(t, "should have not called this")
				return nil
			},
		}
		handler, _ := NewUsersHandler(provider)

		req := httptest.NewRequest(http.MethodDelete, "/api/admin-users?username=", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("delete - success", func(t *testing.T) {
		t.Parallel()

		token, _ := GenerateToken("admin", true)

		expectedUsername := "user2"

		provider := &testscommon.StorerStub{
			RemoveUserHandler: func(username string) error {
				assert.Equal(t, expectedUsername, username)
				return nil
			},
		}
		handler, _ := NewUsersHandler(provider)

		req := httptest.NewRequest(http.MethodDelete, "/api/admin-users?username="+expectedUsername, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
	})
}
