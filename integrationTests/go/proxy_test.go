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

	"github.com/iulianpascalau/mx-epoch-proxy-go/integrationTests/go/framework"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/api"
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

	client := &http.Client{}
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
}
