package process

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewHttpRequester(t *testing.T) {
	t.Parallel()

	r := NewHttpRequester(time.Second)
	assert.NotNil(t, r)
	assert.NotNil(t, r.client)
	assert.Equal(t, time.Second, r.client.Timeout)
}

func TestHttpRequester_DoRequest(t *testing.T) {
	t.Parallel()

	type responseStruct struct {
		Field string `json:"field"`
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		expectedResp := responseStruct{Field: "value"}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "my-api-key", r.Header.Get("X-Service-Api-Key"))
			assert.Equal(t, "/test", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(expectedResp)
		}))
		defer server.Close()

		requester := NewHttpRequester(time.Second)
		var result responseStruct
		err := requester.DoRequest(http.MethodGet, server.URL+"/test", "my-api-key", &result)
		assert.NoError(t, err)
		assert.Equal(t, expectedResp, result)
	})

	t.Run("unexpected status code", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		requester := NewHttpRequester(time.Second)
		var result responseStruct
		err := requester.DoRequest(http.MethodGet, server.URL, "", &result)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, errUnexpectedStatusCode))
	})

	t.Run("network error", func(t *testing.T) {
		t.Parallel()

		// Create a requester with a very short timeout
		requester := NewHttpRequester(1 * time.Nanosecond)

		// Use a server that sleeps to force timeout, or just a closed port.
		// 1ns is usually enough to trigger timeout on any network call.
		// Or connect to invalid port.
		var result responseStruct
		// trying to reach an invalid URL
		err := requester.DoRequest(http.MethodGet, "not-a-valid-url", "", &result)
		assert.Error(t, err)
	})

	t.Run("nil result", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"field":"value"}`))
		}))
		defer server.Close()

		requester := NewHttpRequester(time.Second)
		err := requester.DoRequest(http.MethodGet, server.URL, "", nil)
		assert.NoError(t, err)
	})

	t.Run("invalid url", func(t *testing.T) {
		t.Parallel()

		requester := NewHttpRequester(time.Second)
		err := requester.DoRequest(http.MethodGet, ":invalid-url", "", nil)
		assert.Error(t, err)
	})

	t.Run("json decode error", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`invalid-json`))
		}))
		defer server.Close()

		requester := NewHttpRequester(time.Second)
		var result responseStruct
		err := requester.DoRequest(http.MethodGet, server.URL, "", &result)
		assert.Error(t, err)
	})
}

func TestHttpRequester_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var r *httpRequester
	assert.True(t, r.IsInterfaceNil())

	r = &httpRequester{}
	assert.False(t, r.IsInterfaceNil())
}
