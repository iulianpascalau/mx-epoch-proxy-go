package integrationTests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/iulianpascalau/mx-epoch-proxy-go/api"
	"github.com/iulianpascalau/mx-epoch-proxy-go/process"
	"github.com/iulianpascalau/mx-epoch-proxy-go/testscommon"
	logger "github.com/multiversx/mx-chain-logger-go"
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
	handlerA := func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("A: %+v\n", r)
	}
	serverA := createTestHTTPServer(handlerA)
	defer serverA.Close()

	handlerB := func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("B: %+v\n", r)
	}
	serverB := createTestHTTPServer(handlerB)
	defer serverB.Close()

	hostsFinder := &testscommon.HostsFinderStub{
		FindHostCalled: func(urlValues map[string][]string) (string, error) {
			fmt.Printf("finder: %+v\n", urlValues)

			return serverA.URL, nil
		},
	}

	processor, err := process.NewRequestsProcessor(hostsFinder)
	require.Nil(t, err)

	engine, err := api.NewAPIEngine("localhost:0", processor)
	require.Nil(t, err)
	defer func() {
		_ = engine.Close()
	}()

	log.Info("API engine running", "interface", engine.Address())

	url := fmt.Sprintf("http://%s/transaction/8a64d0ad29f70595bf942c8d2e241a21a3988d9712ae268a9e33efbaffc16b3b?withResults=true&blockNonce=221001&hintEpoch=456", engine.Address())
	_, _ = http.DefaultClient.Get(url)

	time.Sleep(time.Second)
}
