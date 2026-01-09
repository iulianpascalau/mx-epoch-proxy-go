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
	"github.com/stretchr/testify/require"
)

func TestNewLoginHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil access provider should error", func(t *testing.T) {
		t.Parallel()

		handler, err := NewAccessKeysHandler(nil, &testscommon.AuthenticatorStub{})
		assert.Nil(t, handler)
		assert.Equal(t, errNilKeyAccessProvider, err)
	})

	t.Run("nil authenticator should error", func(t *testing.T) {
		t.Parallel()

		handler, err := NewAccessKeysHandler(&testscommon.StorerStub{}, nil)
		assert.Nil(t, handler)
		assert.Equal(t, errNilAuthenticator, err)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		handler, err := NewAccessKeysHandler(&testscommon.StorerStub{}, &testscommon.AuthenticatorStub{})
		assert.NotNil(t, handler)
		assert.Nil(t, err)
	})
}

func TestLoginHandler(t *testing.T) {
	auth := NewJWTAuthenticator("test_key")

	t.Run("ServeHTTP non-POST method", func(t *testing.T) {
		handler, _ := NewLoginHandler(&testscommon.StorerStub{}, auth)
		req := httptest.NewRequest(http.MethodGet, "/login", nil)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusMethodNotAllowed, resp.Code)
	})

	t.Run("ServeHTTP bad request body", func(t *testing.T) {
		handler, _ := NewLoginHandler(&testscommon.StorerStub{}, auth)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("invalid json"))
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("ServeHTTP invalid credentials", func(t *testing.T) {
		storer := &testscommon.StorerStub{
			CheckUserCredentialsHandler: func(username, password string) (*common.UsersDetails, error) {
				return nil, errors.New("invalid credentials")
			},
		}
		handler, _ := NewLoginHandler(storer, auth)

		creds := map[string]string{"username": "user", "password": "wrong"}
		body, _ := json.Marshal(creds)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("ServeHTTP inactive user should error", func(t *testing.T) {
		storer := &testscommon.StorerStub{
			CheckUserCredentialsHandler: func(username, password string) (*common.UsersDetails, error) {
				return &common.UsersDetails{
					Username: "user",
					IsAdmin:  true,
					IsActive: false,
				}, nil
			},
		}
		handler, _ := NewLoginHandler(storer, auth)

		creds := map[string]string{"username": "user", "password": "pass"}
		body, _ := json.Marshal(creds)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusForbidden, resp.Code)
		assert.Contains(t, resp.Body.String(), "Account not activated. Please check your email.")
	})

	t.Run("ServeHTTP success", func(t *testing.T) {
		storer := &testscommon.StorerStub{
			CheckUserCredentialsHandler: func(username, password string) (*common.UsersDetails, error) {
				return &common.UsersDetails{
					Username: "user",
					IsAdmin:  true,
					IsActive: true,
				}, nil
			},
		}
		handler, _ := NewLoginHandler(storer, auth)

		creds := map[string]string{"username": "user", "password": "pass"}
		body, _ := json.Marshal(creds)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)

		var respData map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&respData)
		require.NoError(t, err)
		assert.NotEmpty(t, respData["token"])
		assert.Equal(t, "user", respData["username"])
		assert.Equal(t, true, respData["is_admin"])
	})
}
