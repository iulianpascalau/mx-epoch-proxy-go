package _go

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/iulianpascalau/mx-epoch-proxy-go/integrationTests/go/framework"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/api"
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

	baseAddress := "http://" + proxyService.Components.GetAPIEngine().Address()

	// Register
	username := "testuser@example.com"
	password := "password123456"
	regReq := map[string]string{
		"username":        username,
		"password":        password,
		"captchaId":       "1",
		"captchaSolution": "1",
	}
	reqBody, _ := json.Marshal(regReq)
	resp, err := http.Post(baseAddress+api.EndpointApiRegister, "application/json", bytes.NewBuffer(reqBody))
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Activate
	// Get token from mock email sender
	require.NotEmpty(t, proxyService.EmailSender.LastBody)
	bodyStr := fmt.Sprintf("%v", proxyService.EmailSender.LastBody)
	// Extract token using regex or string manipulation. Token starts with EMAILTOKEN
	re := regexp.MustCompile("token=(EMAILTOKEN[A-Za-z0-9]+)")
	matches := re.FindStringSubmatch(bodyStr)
	require.Len(t, matches, 2)
	token := matches[1]

	clientNoRedirect := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err = clientNoRedirect.Get(baseAddress + api.EndpointApiActivate + "?token=" + token)
	require.Nil(t, err)
	require.Equal(t, http.StatusFound, resp.StatusCode)
	require.Contains(t, resp.Header.Get("Location"), "activated=true")

	// Login
	loginReq := map[string]string{
		"username": username,
		"password": password,
	}
	reqBody, _ = json.Marshal(loginReq)
	resp, err = http.Post(baseAddress+api.EndpointApiLogin, "application/json", bytes.NewBuffer(reqBody))
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var loginResp map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&loginResp)
	jwtToken := loginResp["token"]
	require.NotEmpty(t, jwtToken)

	client := &http.Client{}

	// Check connection to crypto payment service
	reqConf, err := http.NewRequest(http.MethodGet, baseAddress+api.EndpointApiCryptoPaymentConfig, nil)
	require.Nil(t, err)
	reqConf.Header.Set("Authorization", "Bearer "+jwtToken)
	resp, err = client.Do(reqConf)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var cryptoConfig map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&cryptoConfig)
	require.True(t, cryptoConfig["isAvailable"].(bool))

	// Create API Key
	apiKey := "test_api_key_000000001"
	createKeyReq := map[string]string{
		"key": apiKey,
	}
	reqBody, _ = json.Marshal(createKeyReq)
	req, err := http.NewRequest(http.MethodPost, baseAddress+api.EndpointApiAccessKeys, bytes.NewBuffer(reqBody))
	require.Nil(t, err)
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Test Throttling
	// Max calls is 10.
	for i := 0; i < 10; i++ {
		req, err = http.NewRequest(http.MethodGet, baseAddress+"/network/config", nil)
		require.Nil(t, err)
		req.Header.Set("X-Api-Key", apiKey)

		resp, err = client.Do(req)
		require.Nil(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode, "Call %d should succeed", i+1)
		_, _ = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
	}

	// 11th call should fail
	req, err = http.NewRequest(http.MethodGet, baseAddress+"/network/config", nil)
	require.Nil(t, err)
	req.Header.Set("X-Api-Key", apiKey)
	resp, err = client.Do(req)
	require.Nil(t, err)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Obtain an address to deposit 0.5 egld
	req, err = http.NewRequest(http.MethodPost, baseAddress+api.EndpointApiCryptoPaymentCreateAddress, nil)
	require.Nil(t, err)
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	resp, err = client.Do(req)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	req, err = http.NewRequest(http.MethodGet, baseAddress+api.EndpointApiCryptoPaymentAccount, nil)
	require.Nil(t, err)
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	resp, err = client.Do(req)
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var accountResp map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&accountResp)
	depositAddress := accountResp["address"].(string)
	require.NotEmpty(t, depositAddress)

	// Send 0.5 EGLD
	// Use OwnerKeys to send funds
	senderSk := cryptoPaymentService.Keys.OwnerKeys.MvxSk
	// Value: 0.5 EGLD = 0.5 * 10^18 = 500000000000000000
	value := "500000000000000000"

	receiverAddr := framework.NewMvxAddressFromBech32(t, depositAddress)

	// Send Tx
	hash, _, status := cryptoPaymentService.ChainSimulator.SendTx(ctx, senderSk, receiverAddr, value, 500000, []byte{})
	require.Equal(t, transaction.TxStatusSuccess, status)
	t.Logf("Sent funds tx: %s to %s", hash, depositAddress)

	cryptoPaymentService.ChainSimulator.GenerateBlocksUntilTxProcessed(ctx, hash)

	// Wait until credits are given
	require.Eventually(t, func() bool {
		// Generate a block to ensure any pending txs (from balanceProcessor) are processed
		cryptoPaymentService.ChainSimulator.GenerateBlocks(ctx, 1)

		req, _ = http.NewRequest(http.MethodGet, baseAddress+api.EndpointApiCryptoPaymentAccount, nil)
		req.Header.Set("Authorization", "Bearer "+jwtToken)
		resp, err = client.Do(req)
		if err != nil {
			return false
		}
		defer func() {
			_ = resp.Body.Close()
		}()
		if resp.StatusCode != http.StatusOK {
			return false
		}
		var acc map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&acc)

		// Check numberOfRequests. Should be around 50 (0.5 EGLD * 100 requests/EGLD)
		reqs, ok := acc["numberOfRequests"].(float64)
		if !ok {
			return false
		}
		return reqs >= 50
	}, 2*time.Minute, 1*time.Second)

	// Wait until the user is unthrottled (DB sync)
	require.Eventually(t, func() bool {
		req, _ = http.NewRequest(http.MethodGet, baseAddress+"/network/config", nil)
		req.Header.Set("X-Api-Key", apiKey)
		resp, err = client.Do(req)
		if err != nil {
			return false
		}
		defer func() {
			_ = resp.Body.Close()
		}()
		return resp.StatusCode == http.StatusOK
	}, 10*time.Second, 100*time.Millisecond)

	// Spend 49 credits (since 1 was used in Eventually)
	for i := 0; i < 49; i++ {
		req, err = http.NewRequest(http.MethodGet, baseAddress+"/network/config", nil)
		require.Nil(t, err)
		req.Header.Set("X-Api-Key", apiKey)

		resp, err = client.Do(req)
		require.Nil(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode, "Call %d after topup should succeed", i+1)
		_, _ = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
	}

	// 51st call should fail because when switching to free trial, the token already consumed the free
	// allocated credits per quota
	req, err = http.NewRequest(http.MethodGet, baseAddress+"/network/config", nil)
	require.Nil(t, err)
	req.Header.Set("X-Api-Key", apiKey)
	resp, err = client.Do(req)
	require.Nil(t, err)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
