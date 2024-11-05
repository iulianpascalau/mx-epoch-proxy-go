package process

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/multiversx/mx-chain-core-go/core/check"
)

const origin = "Origin"

type requestsProcessor struct {
	hostFinder HostFinder
}

// NewRequestsProcessor creates a new requests processor
func NewRequestsProcessor(hostFinder HostFinder) (*requestsProcessor, error) {
	if check.IfNil(hostFinder) {
		return nil, errNilHostsFinder
	}

	return &requestsProcessor{
		hostFinder: hostFinder,
	}, nil
}

// ServeHTTP will serve the http requests
func (processor *requestsProcessor) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	values, err := url.ParseQuery(request.URL.RawQuery)
	if err != nil {
		RespondWithError(writer, fmt.Errorf("%w while parsing query", err), http.StatusBadRequest)
		return
	}

	newHost, err := processor.hostFinder.FindHost(values)
	if err != nil {
		RespondWithError(writer, err, http.StatusInternalServerError)
		return
	}

	urlPath, err := url.JoinPath(newHost.URL, request.RequestURI)
	if err != nil {
		RespondWithError(writer, err, http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequest(request.Method, urlPath, request.Body)
	if err != nil {
		log.Error("can not create request", "target host", newHost, "error", err)
		RespondWithError(writer, err, http.StatusInternalServerError)
		return
	}

	// pass through the header attributes
	for key, value := range request.Header {
		req.Header[key] = value
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("can not do request", "target host", newHost, "error", err)
		RespondWithError(writer, err, http.StatusInternalServerError)
		return
	}

	// pass through the response header attributes
	for key, value := range response.Header {
		writer.Header()[key] = value
	}
	writer.Header()[origin] = []string{newHost.Name}

	writer.WriteHeader(response.StatusCode)
	if response.ContentLength > 0 {
		_, _ = io.CopyN(writer, response.Body, response.ContentLength)
	}
}

// IsInterfaceNil returns true if the value under the interface is nil
func (processor *requestsProcessor) IsInterfaceNil() bool {
	return processor == nil
}
