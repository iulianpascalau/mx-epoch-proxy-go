package api

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHTTPServer_StartAndClose(t *testing.T) {
	t.Parallel()

	storage := &mockStorage{}
	h, _ := NewHandler(storage, &mockConfigProvider{config: map[string]interface{}{"key": "value"}}, &mockAccountHandler{})

	// Use port 0 to let the OS assign a free port
	server := NewHTTPServer(h, 0, "test-api-key")

	// Start the server
	err := server.Start()
	require.NoError(t, err)

	// The address should now be updated with the actual port
	addr := server.GetAddress()
	require.NotEmpty(t, addr)
	require.False(t, strings.HasSuffix(addr, ":0"))

	// Create a request to the ping endpoint
	client := &http.Client{}

	// Wait a bit for goroutine to pick up
	time.Sleep(100 * time.Millisecond)

	url := "http://" + addr + "/config"
	if strings.HasPrefix(addr, ":") {
		url = "http://localhost" + addr + "/config"
	} else if strings.HasPrefix(addr, "0.0.0.0") {
		url = strings.Replace(url, "0.0.0.0", "localhost", 1)
	}

	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Close the server
	err = server.Close()
	require.NoError(t, err)
}
