package integrationTests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/iulianpascalau/mx-epoch-proxy-go/api"
	"github.com/iulianpascalau/mx-epoch-proxy-go/config"
	"github.com/iulianpascalau/mx-epoch-proxy-go/process"
	"github.com/iulianpascalau/mx-epoch-proxy-go/testscommon"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var log = logger.GetOrCreate("integrationTests")

func createTestHTTPServer(handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	server := httptest.NewServer(&testscommon.HttpHandlerStub{
		ServeHTTPCalled: handler,
	})

	return server
}

func TestRequestsArePassedCorrectly(t *testing.T) {
	handlerAValues := make([]string, 0)
	handlerA := func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("A: %+v\n", r)
		handlerAValues = append(handlerAValues, r.RequestURI)
	}
	serverA := createTestHTTPServer(handlerA)
	defer serverA.Close()

	handlerBValues := make([]string, 0)
	handlerB := func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("B: %+v\n", r)
		handlerBValues = append(handlerBValues, r.RequestURI)
	}
	serverB := createTestHTTPServer(handlerB)
	defer serverB.Close()

	gateways := []config.GatewayConfig{
		{
			URL:        serverA.URL,
			EpochStart: "0",
			EpochEnd:   "99",
			NonceStart: "0",
			NonceEnd:   "9999",
		},
		{
			URL:        serverB.URL,
			EpochStart: "100",
			EpochEnd:   "latest",
			NonceStart: "10000",
			NonceEnd:   "latest",
		},
	}

	hostsFinder, err := process.NewHostsFinder(gateways)
	require.Nil(t, err)

	processor, err := process.NewRequestsProcessor(hostsFinder)
	require.Nil(t, err)

	engine, err := api.NewAPIEngine("localhost:0", processor)
	require.Nil(t, err)
	defer func() {
		_ = engine.Close()
	}()

	log.Info("API engine running", "interface", engine.Address())

	url := fmt.Sprintf("http://%s/transaction/8a64d0ad29f70595bf942c8d2e241a21a3988d9712ae268a9e33efbaffc16b3b?withResults=true&blockNonce=100&hintEpoch=456", engine.Address())
	_, _ = http.DefaultClient.Get(url)

	url = fmt.Sprintf("http://%s/transaction/8a64d0ad29f70595bf942c8d2e241a21a3988d9712ae268a9e33efbaffc16b3b?withResults=true&hintEpoch=99", engine.Address())
	_, _ = http.DefaultClient.Get(url)

	url = fmt.Sprintf("http://%s/transaction/8a64d0ad29f70595bf942c8d2e241a21a3988d9712ae268a9e33efbaffc16b3b?withResults=true", engine.Address())
	_, _ = http.DefaultClient.Get(url)

	url = fmt.Sprintf("http://%s/transaction/8a64d0ad29f70595bf942c8d2e241a21a3988d9712ae268a9e33efbaffc16b3b?withResults=true&blockNonce=10000", engine.Address())
	_, _ = http.DefaultClient.Get(url)

	expectedHandlerAValues := []string{
		"//transaction/8a64d0ad29f70595bf942c8d2e241a21a3988d9712ae268a9e33efbaffc16b3b?withResults=true&blockNonce=100&hintEpoch=456",
		"//transaction/8a64d0ad29f70595bf942c8d2e241a21a3988d9712ae268a9e33efbaffc16b3b?withResults=true&hintEpoch=99",
	}
	assert.Equal(t, expectedHandlerAValues, handlerAValues)

	expectedHandlerBValues := []string{
		"//transaction/8a64d0ad29f70595bf942c8d2e241a21a3988d9712ae268a9e33efbaffc16b3b?withResults=true",
		"//transaction/8a64d0ad29f70595bf942c8d2e241a21a3988d9712ae268a9e33efbaffc16b3b?withResults=true&blockNonce=10000",
	}
	assert.Equal(t, expectedHandlerBValues, handlerBValues)

	time.Sleep(time.Second)
}
