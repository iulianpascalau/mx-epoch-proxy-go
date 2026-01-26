package _go

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/iulianpascalau/mx-epoch-proxy-go/integrationTests/go/framework"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestCreateFreeUserAndCreateKeyAndTestRequestsAreThrottledThenSwitchToPremium(t *testing.T) {
	if !framework.IsChainSimulatorIsRunning() {
		t.Skip("No chain simulator instance running found. Skipping slow test")
	}
	proxyService := framework.NewProxyService(t)
	cryptoPaymentService := framework.NewCryptoPaymentService(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	proxyService.Setup(ctx)
	cryptoPaymentService.Setup(ctx, 100)

	defer proxyService.TearDown()
	defer cryptoPaymentService.TearDown()

	cryptoPaymentService.CreateService()
	cryptoPaymentService.Components.StartCronJobs(ctx)
	cryptoPaymentServiceURL := "http://" + cryptoPaymentService.Components.GetHTTPServer().GetAddress()
	proxyService.CreateService(cryptoPaymentServiceURL)
	proxyService.Components.StartCronJobs(ctx)

	time.Sleep(time.Second)

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

	log.Info("======== 3. Check the crypto-payment serrvice")
	session.CheckCryptoPaymentService(t)
	log.Info("Running ✓")

	log.Info("======== 4. Create API key")
	key := "test_api_key_000000001"
	session.CreateKey(t, key)
	log.Info("Done ✓")

	// Test Throttling
	log.Info("======== 5. Do 10 requests, should work because they are not throttled")
	for i := 0; i < 10; i++ {
		resp, err := session.DoTestRequest(t, key)
		require.Nil(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode, "Call %d should succeed", i+1)
		_, _ = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
	}
	log.Info("Done ✓")

	log.Info("======== 6. Do the 11-th request, should be throttled")
	resp, err := session.DoTestRequest(t, key)
	require.Nil(t, err)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	log.Info("Done ✓")

	log.Info("======== 7. Obtain an address to deposit 0.5 egld (50 credits)")
	session.ObtainDepositAddress(t)
	log.Info(fmt.Sprintf("Done: %s ✓", session.GetDepositAddress()))

	log.Info("======== 8. Use the Owner key to send funds to the deposit address")
	senderSk := cryptoPaymentService.Keys.OwnerKeys.MvxSk
	// Value: 0.5 EGLD = 0.5 * 10^18 = 500000000000000000
	value := "500000000000000000"

	receiverAddr := framework.NewMvxAddressFromBech32(t, session.GetDepositAddress())

	// Send Tx
	hash, _, status := cryptoPaymentService.ChainSimulator.SendTx(ctx, senderSk, receiverAddr, value, 500000, []byte{})
	require.Equal(t, transaction.TxStatusSuccess, status)
	t.Logf("Sent funds tx: %s to %s", hash, session.GetDepositAddress())

	cryptoPaymentService.ChainSimulator.GenerateBlocksUntilTxProcessed(ctx, hash)
	log.Info("Done ✓")

	log.Info("======== 9. Wait until credits are given")
	// Wait until credits are given
	require.Eventually(t, func() bool {
		// Generate a block to ensure any pending txs (from balanceProcessor) are processed
		cryptoPaymentService.ChainSimulator.GenerateBlocks(ctx, 1)

		reqs := session.GetNumberOfRequests(t)
		return reqs >= 50
	}, 2*time.Minute, 1*time.Second)
	log.Info("Done ✓")

	log.Info("======== 10. Wait until the user is unthrottled (DB sync)")
	require.Eventually(t, func() bool {
		resp, err = session.DoTestRequest(t, key)
		if err != nil {
			return false
		}
		defer func() {
			_ = resp.Body.Close()
		}()
		return resp.StatusCode == http.StatusOK
	}, 10*time.Second, 100*time.Millisecond)
	log.Info("Done ✓")

	log.Info("======== 11. Spend 49 credits (since 1 was used in Eventually)")
	for i := 0; i < 49; i++ {
		resp, err = session.DoTestRequest(t, key)
		require.Nil(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode, "Call %d after topup should succeed", i+1)
		_, _ = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
	}
	log.Info("Done ✓")

	log.Info("======== 12. 51st call should fail because when switching to free trial, the token already consumed the free allocated credits per quota")
	resp, err = session.DoTestRequest(t, key)
	require.Nil(t, err)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	log.Info("Done ✓")
}
