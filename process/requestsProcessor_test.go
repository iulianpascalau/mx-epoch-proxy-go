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
			make([]string, 0),
		)
		assert.Nil(t, processor)
		assert.True(t, processor.IsInterfaceNil())
		assert.Equal(t, errNilHostsFinder, err)
	})
	t.Run("nil access checker should error", func(t *testing.T) {
		t.Parallel()

		processor, err := NewRequestsProcessor(
			&testscommon.HostsFinderStub{},
			nil,
			make([]string, 0),
		)
		assert.Nil(t, processor)
		assert.True(t, processor.IsInterfaceNil())
		assert.Equal(t, errNilAccessChecker, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		processor, err := NewRequestsProcessor(
			&testscommon.HostsFinderStub{},
			&testscommon.AccessCheckerStub{},
			make([]string, 0),
		)
		assert.NotNil(t, processor)
		assert.False(t, processor.IsInterfaceNil())
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

		processor, _ := NewRequestsProcessor(
			&testscommon.HostsFinderStub{},
			&testscommon.AccessCheckerStub{},
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
				ShouldProcessRequestHandler: func(header http.Header, requestURI string) (string, error) {
					return "", expectedErr
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

		processor, _ := NewRequestsProcessor(
			&testscommon.HostsFinderStub{
				FindHostCalled: func(urlValues map[string][]string) (config.GatewayConfig, error) {
					return config.GatewayConfig{}, expectedErr
				},
			},
			&testscommon.AccessCheckerStub{},
			make([]string, 0))

		request := httptest.NewRequest(http.MethodGet, "/test/aa", nil)
		request.URL.RawQuery = "a=b"
		recorder := httptest.NewRecorder()
		processor.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "expected error")
	})
	t.Run("can not assemble request, should error", func(t *testing.T) {
		t.Parallel()

		processor, _ := NewRequestsProcessor(
			&testscommon.HostsFinderStub{
				FindHostCalled: func(urlValues map[string][]string) (config.GatewayConfig, error) {
					return config.GatewayConfig{
						URL: "AAAA",
					}, nil
				},
			},
			&testscommon.AccessCheckerStub{},
			make([]string, 0),
		)

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

		processor, _ := NewRequestsProcessor(
			&testscommon.HostsFinderStub{
				FindHostCalled: func(urlValues map[string][]string) (config.GatewayConfig, error) {
					return config.GatewayConfig{
						URL: "unknown host",
					}, nil
				},
			},
			&testscommon.AccessCheckerStub{},
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
	})
	t.Run("should return page not found if closed endpoint", func(t *testing.T) {
		t.Parallel()

		testHttp := httptest.NewServer(&testscommon.HttpHandlerStub{
			ServeHTTPCalled: func(writer http.ResponseWriter, request *http.Request) {},
		})
		defer testHttp.Close()

		processor, _ := NewRequestsProcessor(
			&testscommon.HostsFinderStub{
				FindHostCalled: func(urlValues map[string][]string) (config.GatewayConfig, error) {
					return config.GatewayConfig{
						URL: testHttp.URL,
					}, nil
				},
			},
			&testscommon.AccessCheckerStub{},
			[]string{"/test/"},
		)

		request := httptest.NewRequest(http.MethodGet, "/test/aa", nil)
		request.Header["c"] = []string{"d"}
		request.URL.RawQuery = "a=b"
		recorder := httptest.NewRecorder()
		processor.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusNotFound, recorder.Code)
		assert.Equal(t, "404 page not found\n", recorder.Body.String())
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

		processor, _ := NewRequestsProcessor(
			&testscommon.HostsFinderStub{
				FindHostCalled: func(urlValues map[string][]string) (config.GatewayConfig, error) {
					return config.GatewayConfig{
						URL: testHttp.URL,
					}, nil
				},
			},
			&testscommon.AccessCheckerStub{},
			make([]string, 0),
		)

		request := httptest.NewRequest(http.MethodGet, "/test/aa", nil)
		request.Header["c"] = []string{"d"}
		request.URL.RawQuery = "a=b"
		recorder := httptest.NewRecorder()
		processor.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, string(expectedResponseMarshalled), recorder.Body.String())
	})
}
