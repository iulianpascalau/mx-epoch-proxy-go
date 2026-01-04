package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/config"
	"github.com/iulianpascalau/mx-epoch-proxy-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testAppDomainsConfig = config.AppDomainsConfig{
	Backend:  "https://link.com",
	Frontend: "https://redirect.com",
}

func TestNewRegistrationHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil key access provider", func(t *testing.T) {
		t.Parallel()

		handler, err := NewRegistrationHandler(nil, &testscommon.EmailSenderStub{}, testAppDomainsConfig)
		assert.Equal(t, errNilKeyAccessChecker, err)
		assert.Nil(t, handler)
	})

	t.Run("nil email sender", func(t *testing.T) {
		t.Parallel()

		handler, err := NewRegistrationHandler(&testscommon.StorerStub{}, nil, testAppDomainsConfig)
		assert.Equal(t, errNilEmailSender, err)
		assert.Nil(t, handler)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		handler, err := NewRegistrationHandler(&testscommon.StorerStub{}, &testscommon.EmailSenderStub{}, testAppDomainsConfig)
		assert.Nil(t, err)
		assert.NotNil(t, handler)
	})
}

func TestRegistrationHandler_ServeHTTP_Register(t *testing.T) {
	t.Parallel()

	handler, _ := NewRegistrationHandler(&testscommon.StorerStub{}, &testscommon.EmailSenderStub{}, testAppDomainsConfig)

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/register", nil)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusMethodNotAllowed, resp.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBufferString("invalid"))
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("invalid email", func(t *testing.T) {
		reqBody := registerRequest{Username: "invalid", Password: "password123"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(body))
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Contains(t, resp.Body.String(), "Invalid email address")
	})

	t.Run("short password", func(t *testing.T) {
		reqBody := registerRequest{Username: "test@example.com", Password: "short"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(body))
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Contains(t, resp.Body.String(), "Password must be at least 8 characters long")
	})

	t.Run("db error", func(t *testing.T) {
		storer := &testscommon.StorerStub{
			AddUserHandler: func(username string, password string, isAdmin bool, maxRequests uint64, accountType string, isActive bool, activationToken string) error {
				return errors.New("db fail")
			},
		}
		h, _ := NewRegistrationHandler(storer, &testscommon.EmailSenderStub{}, testAppDomainsConfig)

		reqBody := registerRequest{Username: "test@example.com", Password: "password123"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(body))
		resp := httptest.NewRecorder()

		h.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})

	t.Run("success", func(t *testing.T) {
		var sentTo, sentSubject, sentBody string
		storer := &testscommon.StorerStub{
			AddUserHandler: func(username string, password string, isAdmin bool, maxRequests uint64, accountType string, isActive bool, activationToken string) error {
				assert.Equal(t, "test@example.com", username)
				assert.Equal(t, "password123", password)
				assert.False(t, isActive)
				assert.NotEmpty(t, activationToken)
				return nil
			},
		}
		emailSender := &testscommon.EmailSenderStub{
			SendEmailHandler: func(to string, subject string, body string) error {
				sentTo = to
				sentSubject = subject
				sentBody = body
				return nil
			},
		}
		h, _ := NewRegistrationHandler(storer, emailSender, testAppDomainsConfig)

		reqBody := registerRequest{Username: "test@example.com", Password: "password123"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(body))
		resp := httptest.NewRecorder()

		h.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "test@example.com", sentTo)
		assert.Contains(t, sentSubject, "Activate your account")
		assert.Contains(t, sentBody, "https://link.com/")
	})

	t.Run("email error (should NOT fail request)", func(t *testing.T) {
		storer := &testscommon.StorerStub{
			AddUserHandler: func(username string, password string, isAdmin bool, maxRequests uint64, accountType string, isActive bool, activationToken string) error {
				return nil
			},
		}
		emailSender := &testscommon.EmailSenderStub{
			SendEmailHandler: func(to string, subject string, body string) error {
				return errors.New("email fail")
			},
		}
		h, _ := NewRegistrationHandler(storer, emailSender, testAppDomainsConfig)

		reqBody := registerRequest{Username: "test@example.com", Password: "password123"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(body))
		resp := httptest.NewRecorder()

		h.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code) // Still 200 OK
	})
}

func TestRegistrationHandler_ServeHTTP_Activate(t *testing.T) {
	t.Parallel()

	handler, _ := NewRegistrationHandler(&testscommon.StorerStub{}, &testscommon.EmailSenderStub{}, testAppDomainsConfig)

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/activate", nil)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusMethodNotAllowed, resp.Code)
	})

	t.Run("missing token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/activate", nil)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Contains(t, resp.Body.String(), "Token is required")
	})

	t.Run("db error", func(t *testing.T) {
		storer := &testscommon.StorerStub{
			ActivateUserHandler: func(token string) error {
				return errors.New("db fail")
			},
		}
		h, _ := NewRegistrationHandler(storer, &testscommon.EmailSenderStub{}, testAppDomainsConfig)

		req := httptest.NewRequest(http.MethodGet, "/api/activate?token=val", nil)
		resp := httptest.NewRecorder()

		h.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("success", func(t *testing.T) {
		activatedToken := ""
		storer := &testscommon.StorerStub{
			ActivateUserHandler: func(token string) error {
				activatedToken = token
				return nil
			},
		}
		h, _ := NewRegistrationHandler(storer, &testscommon.EmailSenderStub{}, testAppDomainsConfig)

		req := httptest.NewRequest(http.MethodGet, "/api/activate?token=validToken", nil)
		resp := httptest.NewRecorder()

		h.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusFound, resp.Code)
		assert.Equal(t, "validToken", activatedToken)
		require.Equal(t, "https://redirect.com/#/login?activated=true", resp.Header().Get("Location"))
	})
}
