package integrationTests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/process"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestRoutingLogic(t *testing.T) {
	// Mock the requests processor
	mockProcessor := &testscommon.HttpHandlerStub{
		ServeHTTPCalled: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("processed by requestsProcessor"))
		},
	}

	// Replicate main.go handler structure
	handlers := map[string]http.Handler{
		"/swagger/": http.StripPrefix("/swagger/", http.FileServer(http.Dir("."))), // Use current dir for test
		"/":         http.RedirectHandler("/swagger/", http.StatusFound),
		"*":         mockProcessor,
	}

	demuxer := process.NewDemuxer(handlers, nil)

	t.Run("root redirects to swagger", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		resp := httptest.NewRecorder()

		demuxer.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusFound, resp.Code)
		assert.Equal(t, "/swagger/", resp.Header().Get("Location"))
	})

	t.Run("swagger prefix is served by file server", func(t *testing.T) {
		// We are determining that it hits the file server if it returns 404 (since file doesn't exist)
		// or 200 if we point to an existing file.
		// Since we pointed to ".", let's try to get a known file if possible, or just check it doesn't hit mockProcessor.
		// Actually, if we ask for /swagger/nonexistent, FileServer returns 404.
		// If it fell through to * handler, it would return 200 "processed by...".

		req := httptest.NewRequest(http.MethodGet, "/swagger/nonexistent", nil)
		resp := httptest.NewRecorder()

		demuxer.ServeHTTP(resp, req)

		assert.NotEqual(t, "processed by requestsProcessor", resp.Body.String())
		assert.Equal(t, http.StatusNotFound, resp.Code) // FileServer returns 404
	})

	t.Run("v1 request falls through to processor", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/token/endpoint", nil)
		resp := httptest.NewRecorder()

		demuxer.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "processed by requestsProcessor", resp.Body.String())
	})

	t.Run("random request falls through to processor", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/random/endpoint", nil)
		resp := httptest.NewRecorder()

		demuxer.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "processed by requestsProcessor", resp.Body.String())
	})
}
