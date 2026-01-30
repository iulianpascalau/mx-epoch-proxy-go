package framework

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/api"
	"github.com/stretchr/testify/require"
)

// testSession is a struct manipulating session data
type testSession struct {
	baseAddress    string
	proxyService   ProxyComponentsHandler
	username       string
	password       string
	jwtToken       string
	depositAddress string
}

// NewTestSession creates a new test session
func NewTestSession(
	proxyService ProxyComponentsHandler,
	username string,
	password string,
) *testSession {
	baseAddress := "http://" + proxyService.GetAPIEngine().Address()

	return &testSession{
		proxyService: proxyService,
		username:     username,
		password:     password,
		baseAddress:  baseAddress,
	}
}

// Register initiates the registration process
func (session *testSession) Register(tb testing.TB) {
	regReq := map[string]string{
		"username":        session.username,
		"password":        session.password,
		"captchaId":       "1",
		"captchaSolution": "1",
	}
	reqBody, _ := json.Marshal(regReq)
	resp, err := http.Post(session.baseAddress+api.EndpointApiRegister, "application/json", bytes.NewBuffer(reqBody))
	require.Nil(tb, err)
	require.Equal(tb, http.StatusOK, resp.StatusCode)

	_ = resp.Body.Close()
}

// Activate activates the account using the token from the email
func (session *testSession) Activate(tb testing.TB, mockEmailSender *MockEmailSender) {
	// Get token from mock email sender
	require.NotEmpty(tb, mockEmailSender.LastBody)
	bodyStr := fmt.Sprintf("%v", mockEmailSender.LastBody)
	// Extract token using regex or string manipulation. Token starts with EMAILTOKEN
	re := regexp.MustCompile("token=(EMAILTOKEN[A-Za-z0-9]+)")
	matches := re.FindStringSubmatch(bodyStr)
	require.Len(tb, matches, 2)
	token := matches[1]

	clientNoRedirect := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := clientNoRedirect.Get(session.baseAddress + api.EndpointApiActivate + "?token=" + token)
	require.Nil(tb, err)
	require.Equal(tb, http.StatusFound, resp.StatusCode)
	require.Contains(tb, resp.Header.Get("Location"), "activated=true")

	_ = resp.Body.Close()
}

// Login logs the user in and saves the JWT token
func (session *testSession) Login(tb testing.TB) {
	loginReq := map[string]string{
		"username": session.username,
		"password": session.password,
	}
	reqBody, _ := json.Marshal(loginReq)
	resp, err := http.Post(session.baseAddress+api.EndpointApiLogin, "application/json", bytes.NewBuffer(reqBody))
	require.Nil(tb, err)
	require.Equal(tb, http.StatusOK, resp.StatusCode)

	var loginResp map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&loginResp)
	_ = resp.Body.Close()
	session.jwtToken = loginResp["token"]
	require.NotEmpty(tb, session.jwtToken)
}

// CreateKey creates a new API key
func (session *testSession) CreateKey(tb testing.TB, key string) {
	createKeyReq := map[string]string{
		"key": key,
	}
	reqBody, _ := json.Marshal(createKeyReq)
	req, err := http.NewRequest(http.MethodPost, session.baseAddress+api.EndpointApiAccessKeys, bytes.NewBuffer(reqBody))
	require.Nil(tb, err)
	req.Header.Set("Authorization", "Bearer "+session.jwtToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.Nil(tb, err)
	require.Equal(tb, http.StatusOK, resp.StatusCode)

	_ = resp.Body.Close()
}

// DoTestRequest does a test request using the provided key token
func (session *testSession) DoTestRequest(tb testing.TB, key string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, session.baseAddress+"/network/config", nil)
	require.Nil(tb, err)
	req.Header.Set("X-Api-Key", key)

	client := &http.Client{}
	return client.Do(req)
}

// CheckCryptoPaymentService checks if the crypto payment service is up and running
func (session *testSession) CheckCryptoPaymentService(tb testing.TB) {
	// Check connection to crypto payment service
	reqConf, err := http.NewRequest(http.MethodGet, session.baseAddress+api.EndpointApiCryptoPaymentConfig, nil)
	require.Nil(tb, err)
	reqConf.Header.Set("Authorization", "Bearer "+session.jwtToken)
	client := &http.Client{}
	resp, err := client.Do(reqConf)
	require.Nil(tb, err)
	require.Equal(tb, http.StatusOK, resp.StatusCode)

	var cryptoConfig map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&cryptoConfig)
	_ = resp.Body.Close()
	require.True(tb, cryptoConfig["isAvailable"].(bool))
}

func (session *testSession) ObtainDepositAddress(tb testing.TB) {
	resp, err := session.InvokeCryptoPaymentCreateAddress(tb)
	require.Nil(tb, err)
	require.Equal(tb, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	resp, err = session.InvokeCryptoPaymentAccount(tb)
	require.Nil(tb, err)
	require.Equal(tb, http.StatusOK, resp.StatusCode)

	var accountResp map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&accountResp)
	_ = resp.Body.Close()
	session.depositAddress = accountResp["address"].(string)
	require.NotEmpty(tb, session.depositAddress)
}

func (session *testSession) InvokeCryptoPaymentCreateAddress(tb testing.TB) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, session.baseAddress+api.EndpointApiCryptoPaymentCreateAddress, nil)
	require.Nil(tb, err)
	req.Header.Set("Authorization", "Bearer "+session.jwtToken)
	client := &http.Client{}
	return client.Do(req)
}

func (session *testSession) InvokeCryptoPaymentAccount(tb testing.TB) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, session.baseAddress+api.EndpointApiCryptoPaymentAccount, nil)
	require.Nil(tb, err)
	req.Header.Set("Authorization", "Bearer "+session.jwtToken)
	client := &http.Client{}
	return client.Do(req)
}

// GetDepositAddress returns the deposit address
func (session *testSession) GetDepositAddress() string {
	return session.depositAddress
}

func (session *testSession) GetNumberOfRequests() int {
	req, _ := http.NewRequest(http.MethodGet, session.baseAddress+api.EndpointApiCryptoPaymentAccount, nil)
	req.Header.Set("Authorization", "Bearer "+session.jwtToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return -1
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return -1
	}
	var acc map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&acc)

	// Check numberOfRequests. Should be around 50 (0.5 EGLD * 100 requests/EGLD)
	reqs, ok := acc["numberOfRequests"].(float64)
	if !ok {
		return -1
	}

	return int(reqs)
}
