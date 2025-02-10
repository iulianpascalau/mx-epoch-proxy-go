package process

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/config"
	"github.com/iulianpascalau/mx-epoch-proxy-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewRequestsProcessor(t *testing.T) {
	t.Parallel()

	t.Run("nil hosts finder should error", func(t *testing.T) {
		t.Parallel()

		processor, err := NewRequestsProcessor(nil, make([]string, 0))
		assert.Nil(t, processor)
		assert.Equal(t, errNilHostsFinder, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		processor, err := NewRequestsProcessor(&testscommon.HostsFinderStub{}, make([]string, 0))
		assert.NotNil(t, processor)
		assert.Nil(t, err)
	})
}

func TestRequestsProcessor_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var instance *requestsProcessor
	assert.True(t, instance.IsInterfaceNil())

	instance = &requestsProcessor{}
	assert.False(t, instance.IsInterfaceNil())
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

		processor, _ := NewRequestsProcessor(&testscommon.HostsFinderStub{}, make([]string, 0))

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
			FindHostCalled: func(urlValues map[string][]string) (config.GatewayConfig, error) {
				return config.GatewayConfig{}, expectedErr
			},
		}, make([]string, 0))

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
			FindHostCalled: func(urlValues map[string][]string) (config.GatewayConfig, error) {
				return config.GatewayConfig{
					URL: "AAAA",
				}, nil
			},
		}, make([]string, 0))

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
			FindHostCalled: func(urlValues map[string][]string) (config.GatewayConfig, error) {
				return config.GatewayConfig{
					URL: "unknown host",
				}, nil
			},
		}, make([]string, 0))

		request := httptest.NewRequest(http.MethodGet, "/test/aa", nil)
		request.Header["c"] = []string{"d"}
		request.URL.RawQuery = "a=b"
		recorder := httptest.NewRecorder()
		processor.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "unknown%20host")
		assert.Contains(t, recorder.Body.String(), "unsupported protocol scheme")
	})
	t.Run("should return page not found if closed endpoint", func(t *testing.T) {
		t.Parallel()

		testHttp := httptest.NewServer(&testscommon.HttpHandlerStub{
			ServeHTTPCalled: func(writer http.ResponseWriter, request *http.Request) {},
		})
		defer testHttp.Close()

		processor, _ := NewRequestsProcessor(&testscommon.HostsFinderStub{
			FindHostCalled: func(urlValues map[string][]string) (config.GatewayConfig, error) {
				return config.GatewayConfig{
					URL: testHttp.URL,
				}, nil
			},
		}, []string{"/test/"})

		request := httptest.NewRequest(http.MethodGet, "/test/aa", nil)
		request.Header["c"] = []string{"d"}
		request.URL.RawQuery = "a=b"
		recorder := httptest.NewRecorder()
		processor.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusNotFound, recorder.Code)
		assert.Empty(t, recorder.Body.String())
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
			FindHostCalled: func(urlValues map[string][]string) (config.GatewayConfig, error) {
				return config.GatewayConfig{
					URL: testHttp.URL,
				}, nil
			},
		}, make([]string, 0))

		request := httptest.NewRequest(http.MethodGet, "/test/aa", nil)
		request.Header["c"] = []string{"d"}
		request.URL.RawQuery = "a=b"
		recorder := httptest.NewRecorder()
		processor.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, string(expectedResponseMarshalled), recorder.Body.String())
	})
}
