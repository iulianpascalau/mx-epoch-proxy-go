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

func TestLoginHandler(t *testing.T) {
	SetJwtKey("test_secret")

	t.Run("ServeHTTP non-POST method", func(t *testing.T) {
		handler := NewLoginHandler(&testscommon.StorerStub{})
		req := httptest.NewRequest(http.MethodGet, "/login", nil)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusMethodNotAllowed, resp.Code)
	})

	t.Run("ServeHTTP bad request body", func(t *testing.T) {
		handler := NewLoginHandler(&testscommon.StorerStub{})
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
		handler := NewLoginHandler(storer)

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
		handler := NewLoginHandler(storer)

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
		handler := NewLoginHandler(storer)

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
