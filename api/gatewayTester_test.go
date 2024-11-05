package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/config"
	"github.com/iulianpascalau/mx-epoch-proxy-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewGatewayTester(t *testing.T) {
	t.Parallel()

	instance := NewGatewayTester()
	assert.NotNil(t, instance)
}

func TestGatewayTester_TestGateways(t *testing.T) {
	t.Parallel()

	tester := NewGatewayTester()
	t.Run("wrong URL should error", func(t *testing.T) {
		t.Parallel()

		err := tester.TestGateways([]config.GatewayConfig{
			{
				URL: string([]byte{0x7f}),
			},
		})
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid control character in URL")
	})
	t.Run("empty URL should error", func(t *testing.T) {
		t.Parallel()

		err := tester.TestGateways([]config.GatewayConfig{
			{
				URL: "",
			},
		})
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "unsupported protocol")
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		handlerAWasCalled := false
		handlerA := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			handlerAWasCalled = true
		}
		serverA := createTestHTTPServer(handlerA)
		defer serverA.Close()

		handlerBWasCalled := false
		handlerB := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			handlerBWasCalled = true
		}
		serverB := createTestHTTPServer(handlerB)
		defer serverB.Close()

		gateways := []config.GatewayConfig{
			{
				URL: serverA.URL,
			},
			{
				URL: serverB.URL,
			},
		}

		err := tester.TestGateways(gateways)
		assert.Nil(t, err)

		assert.True(t, handlerAWasCalled)
		assert.True(t, handlerBWasCalled)
	})
}

func createTestHTTPServer(handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	server := httptest.NewServer(&testscommon.HttpHandlerStub{
		ServeHTTPCalled: handler,
	})

	return server
}
