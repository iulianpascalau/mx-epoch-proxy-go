package process

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/config"
	"github.com/iulianpascalau/mx-epoch-proxy-go/testscommon"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRequestsProcessor(t *testing.T) {
	t.Parallel()

	t.Run("nil hosts finder should error", func(t *testing.T) {
		t.Parallel()

		processor, err := NewRequestsProcessor(
			nil,
			&testscommon.AccessCheckerStub{},
			&testscommon.RequestsMetricsStub{},
			make([]string, 0),
		)
		assert.Nil(t, processor)
		assert.Equal(t, errNilHostsFinder, err)
	})
	t.Run("nil access checker should error", func(t *testing.T) {
		t.Parallel()

		processor, err := NewRequestsProcessor(
			&testscommon.HostsFinderStub{},
			nil,
			&testscommon.RequestsMetricsStub{},
			make([]string, 0),
		)
		assert.Nil(t, processor)
		assert.Equal(t, errNilAccessChecker, err)
	})
	t.Run("nil requests metrics should error", func(t *testing.T) {
		t.Parallel()

		processor, err := NewRequestsProcessor(
			&testscommon.HostsFinderStub{},
			&testscommon.AccessCheckerStub{},
			nil,
			make([]string, 0),
		)
		assert.Nil(t, processor)
		assert.Equal(t, errNilRequestMetrics, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		processor, err := NewRequestsProcessor(
			&testscommon.HostsFinderStub{},
			&testscommon.AccessCheckerStub{},
			&testscommon.RequestsMetricsStub{},
			make([]string, 0),
		)
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

		processor, _ := NewRequestsProcessor(
			&testscommon.HostsFinderStub{},
			&testscommon.AccessCheckerStub{},
			&testscommon.RequestsMetricsStub{
				ProcessedResponseHandler: func(alias string) {
					assert.Fail(t, "should have not called this handler")
				},
			},
			make([]string, 0),
		)

		request := httptest.NewRequest(http.MethodGet, "/test/aa", nil)
		request.URL.RawQuery = "a=b;c=d"
		recorder := httptest.NewRecorder()
		processor.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "invalid semicolon separator in query while parsing query")
	})
	t.Run("access checker errors, should error", func(t *testing.T) {
		t.Parallel()

		log.SetLevel(logger.LogTrace)

		processor, _ := NewRequestsProcessor(
			&testscommon.HostsFinderStub{
				FindHostCalled: func(urlValues map[string][]string) (config.GatewayConfig, error) {
					require.Fail(t, "should have not called the host finder")
					return config.GatewayConfig{}, nil
				},
			},
			&testscommon.AccessCheckerStub{
				ShouldProcessRequestHandler: func(header http.Header, requestURI string) (string, string, error) {
					return "", "", expectedErr
				},
			},
			&testscommon.RequestsMetricsStub{
				ProcessedResponseHandler: func(alias string) {
					assert.Fail(t, "should have not called this handler")
				},
			},
			make([]string, 0))

		request := httptest.NewRequest(http.MethodGet, "/test/aa", nil)
		request.URL.RawQuery = "a=b"
		request.Header["api"] = []string{"token"}
		recorder := httptest.NewRecorder()
		processor.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "expected error")
	})
	t.Run("hosts finder errors, should error", func(t *testing.T) {
		t.Parallel()

		requestsProcessed := make(map[string]int)
		processor, _ := NewRequestsProcessor(
			&testscommon.HostsFinderStub{
				FindHostCalled: func(urlValues map[string][]string) (config.GatewayConfig, error) {
					return config.GatewayConfig{}, expectedErr
				},
			},
			&testscommon.AccessCheckerStub{},
			&testscommon.RequestsMetricsStub{
				ProcessedResponseHandler: func(alias string) {
					requestsProcessed[alias]++
				},
			},
			make([]string, 0))

		request := httptest.NewRequest(http.MethodGet, "/test/aa", nil)
		request.URL.RawQuery = "a=b"
		recorder := httptest.NewRecorder()
		processor.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "expected error")
		assert.Equal(t, map[string]int{"ALL": 1}, requestsProcessed)
	})
	t.Run("can not assemble request, should error", func(t *testing.T) {
		t.Parallel()

		requestsProcessed := make(map[string]int)
		processor, _ := NewRequestsProcessor(
			&testscommon.HostsFinderStub{
				FindHostCalled: func(urlValues map[string][]string) (config.GatewayConfig, error) {
					return config.GatewayConfig{
						URL: "AAAA",
					}, nil
				},
			},
			&testscommon.AccessCheckerStub{},
			&testscommon.RequestsMetricsStub{
				ProcessedResponseHandler: func(alias string) {
					requestsProcessed[alias]++
				},
			},
			make([]string, 0),
		)

		request := httptest.NewRequest(http.MethodGet, "/test/aa", nil)
		request.Method = "invalid method"
		request.URL.RawQuery = "a=b"
		recorder := httptest.NewRecorder()
		processor.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "invalid method")
		assert.Equal(t, map[string]int{"ALL": 1}, requestsProcessed)
	})
	t.Run("request fails, should error", func(t *testing.T) {
		t.Parallel()

		requestsProcessed := make(map[string]int)
		processor, _ := NewRequestsProcessor(
			&testscommon.HostsFinderStub{
				FindHostCalled: func(urlValues map[string][]string) (config.GatewayConfig, error) {
					return config.GatewayConfig{
						URL: "unknown host",
					}, nil
				},
			},
			&testscommon.AccessCheckerStub{},
			&testscommon.RequestsMetricsStub{
				ProcessedResponseHandler: func(alias string) {
					requestsProcessed[alias]++
				},
			},
			make([]string, 0),
		)

		request := httptest.NewRequest(http.MethodGet, "/test/aa", nil)
		request.Header["c"] = []string{"d"}
		request.URL.RawQuery = "a=b"
		recorder := httptest.NewRecorder()
		processor.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "unknown%20host")
		assert.Contains(t, recorder.Body.String(), "unsupported protocol scheme")
		assert.Equal(t, map[string]int{"ALL": 1}, requestsProcessed)
	})
	t.Run("should return page not found if closed endpoint", func(t *testing.T) {
		t.Parallel()

		testHttp := httptest.NewServer(&testscommon.HttpHandlerStub{
			ServeHTTPCalled: func(writer http.ResponseWriter, request *http.Request) {},
		})
		defer testHttp.Close()

		requestsProcessed := make(map[string]int)
		processor, _ := NewRequestsProcessor(
			&testscommon.HostsFinderStub{
				FindHostCalled: func(urlValues map[string][]string) (config.GatewayConfig, error) {
					return config.GatewayConfig{
						URL: testHttp.URL,
					}, nil
				},
			},
			&testscommon.AccessCheckerStub{},
			&testscommon.RequestsMetricsStub{
				ProcessedResponseHandler: func(alias string) {
					requestsProcessed[alias]++
				},
			},
			[]string{"/test/"},
		)

		request := httptest.NewRequest(http.MethodGet, "/test/aa", nil)
		request.Header["c"] = []string{"d"}
		request.URL.RawQuery = "a=b"
		recorder := httptest.NewRecorder()
		processor.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusNotFound, recorder.Code)
		assert.Equal(t, "404 page not found\n", recorder.Body.String())
		assert.Equal(t, map[string]int{"ALL": 1}, requestsProcessed)
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

		requestsProcessed := make(map[string]int)
		processor, _ := NewRequestsProcessor(
			&testscommon.HostsFinderStub{
				FindHostCalled: func(urlValues map[string][]string) (config.GatewayConfig, error) {
					return config.GatewayConfig{
						URL: testHttp.URL,
					}, nil
				},
			},
			&testscommon.AccessCheckerStub{},
			&testscommon.RequestsMetricsStub{
				ProcessedResponseHandler: func(alias string) {
					requestsProcessed[alias]++
				},
			},
			make([]string, 0),
		)

		request := httptest.NewRequest(http.MethodGet, "/test/aa", nil)
		request.Header["c"] = []string{"d"}
		request.URL.RawQuery = "a=b"
		recorder := httptest.NewRecorder()
		processor.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, string(expectedResponseMarshalled), recorder.Body.String())
		assert.Equal(t, map[string]int{"ALL": 1}, requestsProcessed)
	})
}
