package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/common"
	"github.com/iulianpascalau/mx-epoch-proxy-go/config"
	"github.com/iulianpascalau/mx-epoch-proxy-go/testscommon"
	"github.com/stretchr/testify/assert"
)

type emailSenderStub struct {
	SendEmailHandler func(to string, subject string, bodyObject interface{}, htmlTemplate string) error
}

func (e *emailSenderStub) SendEmail(to string, subject string, bodyObject interface{}, htmlTemplate string) error {
	if e.SendEmailHandler != nil {
		return e.SendEmailHandler(to, subject, bodyObject, htmlTemplate)
	}
	return nil
}

func (e *emailSenderStub) IsInterfaceNil() bool {
	return e == nil
}

func TestNewUserCredentialsHandler(t *testing.T) {
	t.Parallel()

	storer := &testscommon.StorerStub{}
	emailSender := &emailSenderStub{}
	cfg := config.AppDomainsConfig{Frontend: "http://frontend", Backend: "http://backend"}

	auth := NewJWTAuthenticator("test_key")

	t.Run("nil key access provider", func(t *testing.T) {
		t.Parallel()
		h, err := NewUserCredentialsHandler(nil, emailSender, cfg, "template", auth)
		assert.Nil(t, h)
		assert.Equal(t, errNilKeyAccessChecker, err)
	})

	t.Run("nil email sender", func(t *testing.T) {
		t.Parallel()
		h, err := NewUserCredentialsHandler(storer, nil, cfg, "template", auth)
		assert.Nil(t, h)
		assert.Equal(t, errNilEmailSender, err)
	})

	t.Run("empty html template", func(t *testing.T) {
		t.Parallel()
		h, err := NewUserCredentialsHandler(storer, emailSender, cfg, "", auth)
		assert.Nil(t, h)
		assert.Equal(t, errEmptyHTMLTemplate, err)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		h, err := NewUserCredentialsHandler(storer, emailSender, cfg, "template", auth)
		assert.NotNil(t, h)
		assert.Nil(t, err)
	})
}

func TestUserCredentialsHandler_ServeHTTP_ChangePassword(t *testing.T) {
	t.Parallel()

	cfg := config.AppDomainsConfig{}
	auth := NewJWTAuthenticator("test_key")

	t.Run("method not allowed", func(t *testing.T) {
		t.Parallel()
		storer := &testscommon.StorerStub{}
		emailSender := &emailSenderStub{}
		h, _ := NewUserCredentialsHandler(storer, emailSender, cfg, "template", auth)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/change-password", nil)
		h.ServeHTTP(w, r)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()
		storer := &testscommon.StorerStub{}
		emailSender := &emailSenderStub{}
		h, _ := NewUserCredentialsHandler(storer, emailSender, cfg, "template", auth)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/change-password", nil)
		h.ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		storer := &testscommon.StorerStub{}
		emailSender := &emailSenderStub{}
		h, _ := NewUserCredentialsHandler(storer, emailSender, cfg, "template", auth)

		storer.CheckUserCredentialsHandler = func(username, password string) (*common.UsersDetails, error) {
			return &common.UsersDetails{}, nil
		}
		storer.UpdatePasswordHandler = func(username, password string) error {
			return nil
		}

		reqBody := changePasswordRequest{OldPassword: "old", NewPassword: "newpassword"}
		reqBytes, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/change-password", bytes.NewReader(reqBytes))

		// Mock auth
		token, _ := createMockToken("user")
		r.Header.Set("Authorization", "Bearer "+token)

		h.ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid old password", func(t *testing.T) {
		t.Parallel()
		storer := &testscommon.StorerStub{}
		emailSender := &emailSenderStub{}
		h, _ := NewUserCredentialsHandler(storer, emailSender, cfg, "template", auth)

		storer.CheckUserCredentialsHandler = func(username, password string) (*common.UsersDetails, error) {
			return nil, errors.New("invalid")
		}

		reqBody := changePasswordRequest{OldPassword: "wrong", NewPassword: "newpassword"}
		reqBytes, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/change-password", bytes.NewReader(reqBytes))

		token, _ := createMockToken("user")
		r.Header.Set("Authorization", "Bearer "+token)

		h.ServeHTTP(w, r)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestUserCredentialsHandler_ServeHTTP_RequestEmailChange(t *testing.T) {
	t.Parallel()

	cfg := config.AppDomainsConfig{}
	auth := NewJWTAuthenticator("test_key")

	t.Run("method not allowed", func(t *testing.T) {
		t.Parallel()
		storer := &testscommon.StorerStub{}
		emailSender := &emailSenderStub{}
		h, _ := NewUserCredentialsHandler(storer, emailSender, cfg, "template", auth)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/request-email-change", nil)
		h.ServeHTTP(w, r)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("invalid email", func(t *testing.T) {
		t.Parallel()
		storer := &testscommon.StorerStub{}
		emailSender := &emailSenderStub{}
		h, _ := NewUserCredentialsHandler(storer, emailSender, cfg, "template", auth)

		reqBody := changeEmailRequest{OldPassword: "pass", NewEmail: "invalid"}
		reqBytes, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/request-email-change", bytes.NewReader(reqBytes))

		token, _ := createMockToken("user")
		r.Header.Set("Authorization", "Bearer "+token)

		h.ServeHTTP(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		storer := &testscommon.StorerStub{}
		emailSender := &emailSenderStub{}
		h, _ := NewUserCredentialsHandler(storer, emailSender, cfg, "template", auth)

		storer.CheckUserCredentialsHandler = func(username, password string) (*common.UsersDetails, error) {
			return &common.UsersDetails{}, nil
		}
		storer.RequestEmailChangeHandler = func(username, newEmail, token string) error {
			return nil
		}
		emailSender.SendEmailHandler = func(to, subject string, bodyObject interface{}, htmlTemplate string) error {
			return nil
		}

		reqBody := changeEmailRequest{OldPassword: "pass", NewEmail: "new@example.com"}
		reqBytes, _ := json.Marshal(reqBody)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/request-email-change", bytes.NewReader(reqBytes))

		token, _ := createMockToken("user")
		r.Header.Set("Authorization", "Bearer "+token)

		h.ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestUserCredentialsHandler_ServeHTTP_ConfirmEmailChange(t *testing.T) {
	t.Parallel()

	cfg := config.AppDomainsConfig{Frontend: "http://front", Backend: "http://back"}
	auth := NewJWTAuthenticator("test_key")

	t.Run("method not allowed", func(t *testing.T) {
		t.Parallel()
		storer := &testscommon.StorerStub{}
		emailSender := &emailSenderStub{}
		h, _ := NewUserCredentialsHandler(storer, emailSender, cfg, "template", auth)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/confirm-email-change", nil)
		h.ServeHTTP(w, r)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("missing token", func(t *testing.T) {
		t.Parallel()
		storer := &testscommon.StorerStub{}
		emailSender := &emailSenderStub{}
		h, _ := NewUserCredentialsHandler(storer, emailSender, cfg, "template", auth)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/confirm-email-change", nil)
		h.ServeHTTP(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("db error", func(t *testing.T) {
		t.Parallel()
		storer := &testscommon.StorerStub{}
		emailSender := &emailSenderStub{}
		h, _ := NewUserCredentialsHandler(storer, emailSender, cfg, "template", auth)
		storer.ConfirmEmailChangeHandler = func(token string) (string, error) {
			return "", errors.New("db error")
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/confirm-email-change?token=123", nil)
		h.ServeHTTP(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		storer := &testscommon.StorerStub{}
		emailSender := &emailSenderStub{}
		h, _ := NewUserCredentialsHandler(storer, emailSender, cfg, "template", auth)
		storer.ConfirmEmailChangeHandler = func(token string) (string, error) {
			return "new@example.com", nil
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/confirm-email-change?token=123", nil)
		h.ServeHTTP(w, r)

		assert.Equal(t, http.StatusFound, w.Code)
		assert.Contains(t, w.Header().Get("Location"), "email=new@example.com")
	})
}

// Helper to create mock token, assuming SetJwtKey was called in init or setup
func createMockToken(username string) (string, error) {
	auth := NewJWTAuthenticator("test_key")
	return auth.GenerateToken(username, false)
}
