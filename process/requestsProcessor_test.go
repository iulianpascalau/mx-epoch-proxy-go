package process

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewRequestsProcessor(t *testing.T) {
	t.Parallel()

	t.Run("nil hosts finder should error", func(t *testing.T) {
		t.Parallel()

		processor, err := NewRequestsProcessor(nil)
		assert.Nil(t, processor)
		assert.Equal(t, errNilHostsFinder, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		processor, err := NewRequestsProcessor(&testscommon.HostsFinderStub{})
		assert.NotNil(t, processor)
		assert.Nil(t, err)
	})
}

func TestRequestsProcessor_ServeHTTP(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	expectedResponse := &GenericAPIResponse{
		Data:  "data",
		Error: "",
		Code:  "ok",
	}
	expectedResponseMarshalled, _ := json.Marshal(expectedResponse)
	t.Run("parse query errors, should error", func(t *testing.T) {
		t.Parallel()

		processor, _ := NewRequestsProcessor(&testscommon.HostsFinderStub{})

		request := httptest.NewRequest(http.MethodGet, "/test/aa", nil)
		request.URL.RawQuery = "a=b;c=d"
		recorder := httptest.NewRecorder()
		processor.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "invalid semicolon separator in query while parsing query")
	})
	t.Run("hosts finder errors, should error", func(t *testing.T) {
		t.Parallel()

		processor, _ := NewRequestsProcessor(&testscommon.HostsFinderStub{
			FindHostCalled: func(urlValues map[string][]string) (string, error) {
				return "", expectedErr
			},
		})

		request := httptest.NewRequest(http.MethodGet, "/test/aa", nil)
		request.URL.RawQuery = "a=b"
		recorder := httptest.NewRecorder()
		processor.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "expected error")
	})
	t.Run("can not assemble request, should error", func(t *testing.T) {
		t.Parallel()

		processor, _ := NewRequestsProcessor(&testscommon.HostsFinderStub{
			FindHostCalled: func(urlValues map[string][]string) (string, error) {
				return "AAAA", nil
			},
		})

		request := httptest.NewRequest(http.MethodGet, "/test/aa", nil)
		request.Method = "invalid method"
		request.URL.RawQuery = "a=b"
		recorder := httptest.NewRecorder()
		processor.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "invalid method")
	})
	t.Run("request fails, should error", func(t *testing.T) {
		t.Parallel()

		processor, _ := NewRequestsProcessor(&testscommon.HostsFinderStub{
			FindHostCalled: func(urlValues map[string][]string) (string, error) {
				return "unknown host", nil
			},
		})

		request := httptest.NewRequest(http.MethodGet, "/test/aa", nil)
		request.Header["c"] = []string{"d"}
		request.URL.RawQuery = "a=b"
		recorder := httptest.NewRecorder()
		processor.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "unknown%20host")
		assert.Contains(t, recorder.Body.String(), "unsupported protocol scheme")
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		testHttp := httptest.NewServer(&testscommon.HttpHandlerStub{
			ServeHTTPCalled: func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusOK)
				buff, _ := json.Marshal(expectedResponse)

				// header was copied
				assert.Equal(t, []string{"d"}, request.Header["C"])

				_, _ = writer.Write(buff)
			},
		})
		defer testHttp.Close()

		processor, _ := NewRequestsProcessor(&testscommon.HostsFinderStub{
			FindHostCalled: func(urlValues map[string][]string) (string, error) {
				return testHttp.URL, nil
			},
		})

		request := httptest.NewRequest(http.MethodGet, "/test/aa", nil)
		request.Header["c"] = []string{"d"}
		request.URL.RawQuery = "a=b"
		recorder := httptest.NewRecorder()
		processor.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, string(expectedResponseMarshalled), recorder.Body.String())
	})
}