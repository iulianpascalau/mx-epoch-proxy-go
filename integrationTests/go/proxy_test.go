package _go

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/integrationTests/go/framework"
	"github.com/stretchr/testify/require"
)

func TestCreateFreeUserAndCreateKeyAndTestRequestsAreThrottled(t *testing.T) {
	if !framework.IsChainSimulatorIsRunning() {
		t.Skip("No chain simulator instance running found. Skipping slow test")
	}
	proxyService := framework.NewProxyService(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	proxyService.Setup(ctx)
	defer proxyService.TearDown()

	proxyService.CreateService("") // no crypto payment test

	session := framework.NewTestSession(
		proxyService.Components,
		"testuser@example.com",
		"password123456",
	)

	log.Info("======== 1. Register and activate user")
	session.Register(t)
	session.Activate(t, proxyService.EmailSender)
	log.Info("Done ✓")

	log.Info("======== 2. Login")
	session.Login(t)
	log.Info("Done ✓")

	log.Info("======== 3. Create API key")
	key := "test_api_key_000000001"
	session.CreateKey(t, key)
	log.Info("Done ✓")

	// Test Throttling
	log.Info("======== 4. Do 10 requests, should work because they are not throttled")
	for i := 0; i < 10; i++ {
		resp, err := session.DoTestRequest(t, key)
		require.Nil(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode, "Call %d should succeed", i+1)
		_, _ = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
	}
	log.Info("Done ✓")

	log.Info("======== 5. Do the 11-th request, should be throttled")
	resp, err := session.DoTestRequest(t, key)
	require.Nil(t, err)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	log.Info("Done ✓")
}
