package process

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const xServiceApiKey = "X-Service-Api-Key"

type httpRequester struct {
	client *http.Client
}

// NewHttpRequester creates a new instance of httpRequester
func NewHttpRequester(timeout time.Duration) *httpRequester {
	return &httpRequester{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// DoRequest executes a request
func (r *httpRequester) DoRequest(method string, url string, apiKey string, result any) error {
	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}

	if len(apiKey) > 0 {
		request.Header.Add(xServiceApiKey, apiKey)
	}

	resp, err := r.client.Do(request)
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %d", errUnexpectedStatusCode, resp.StatusCode)
	}

	if result == nil {
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(&result)
}

// IsInterfaceNil returns true if the value under the interface is nil
func (r *httpRequester) IsInterfaceNil() bool {
	return r == nil
}
