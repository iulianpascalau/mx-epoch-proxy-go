package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCaptchaHandler struct {
	verifyStringCalled bool
	newCaptchaCalled   bool
	reloadCalled       bool
	writeNoErrorCalled bool

	verifyResult bool
	newCaptchaId string
	lastId       string
}

func (m *mockCaptchaHandler) VerifyString(id string, _ string) bool {
	m.verifyStringCalled = true
	m.lastId = id
	return m.verifyResult
}

func (m *mockCaptchaHandler) NewCaptcha() string {
	m.newCaptchaCalled = true
	return m.newCaptchaId
}

func (m *mockCaptchaHandler) Reload(id string) {
	m.reloadCalled = true
	m.lastId = id
}

func (m *mockCaptchaHandler) WriteNoError(w io.Writer, id string) {
	m.writeNoErrorCalled = true
	m.lastId = id
	_, _ = w.Write([]byte("captcha-image-content"))
}

func (m *mockCaptchaHandler) IsInterfaceNil() bool {
	return m == nil
}

func TestNewCaptchaHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil inner handler should error", func(t *testing.T) {
		handler, err := NewCaptchaHandler(nil)
		assert.Nil(t, handler)
		assert.Equal(t, errNilCaptchaHandler, err)
	})
	t.Run("should work", func(t *testing.T) {
		handler, err := NewCaptchaHandler(&mockCaptchaHandler{})
		assert.NotNil(t, handler)
		assert.Nil(t, err)
	})
}

func TestGenerateCaptchaHandler(t *testing.T) {
	t.Parallel()

	t.Run("wrong method", func(t *testing.T) {
		mock := &mockCaptchaHandler{}
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/captcha", nil)

		handler, _ := NewCaptchaHandler(mock)
		handler.GenerateCaptchaHandler(w, r)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("success", func(t *testing.T) {
		mock := &mockCaptchaHandler{
			newCaptchaId: "test-id",
		}
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, EndpointCaptchaSingle, nil)

		handler, _ := NewCaptchaHandler(mock)
		handler.GenerateCaptchaHandler(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, mock.newCaptchaCalled)

		var resp CaptchaResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "test-id", resp.CaptchaId)
	})
}

func TestServeCaptchaImageHandler(t *testing.T) {
	t.Parallel()

	t.Run("wrong method", func(t *testing.T) {
		mock := &mockCaptchaHandler{}
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/captcha/123.png", nil)

		handler, _ := NewCaptchaHandler(mock)
		handler.ServeCaptchaImageHandler(w, r)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("short id", func(t *testing.T) {
		mock := &mockCaptchaHandler{}
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/captcha/123", nil)

		handler, _ := NewCaptchaHandler(mock)
		handler.ServeCaptchaImageHandler(w, r)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("success with extension", func(t *testing.T) {
		mock := &mockCaptchaHandler{}
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/captcha/test-id.png", nil)

		handler, _ := NewCaptchaHandler(mock)
		handler.ServeCaptchaImageHandler(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, mock.writeNoErrorCalled)
		assert.Equal(t, "test-id", mock.lastId)
		assert.Equal(t, "image/png", w.Header().Get("Content-Type"))
	})

	t.Run("success without extension", func(t *testing.T) {
		mock := &mockCaptchaHandler{}
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/captcha/test-id", nil)

		handler, _ := NewCaptchaHandler(mock)
		handler.ServeCaptchaImageHandler(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, mock.writeNoErrorCalled)
		assert.Equal(t, "test-id", mock.lastId)
	})

	t.Run("reload", func(t *testing.T) {
		mock := &mockCaptchaHandler{}
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/captcha/test-id.png?reload=true", nil)

		handler, _ := NewCaptchaHandler(mock)
		handler.ServeCaptchaImageHandler(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, mock.reloadCalled)
		assert.True(t, mock.writeNoErrorCalled)
		assert.Equal(t, "test-id", mock.lastId)
	})
}
