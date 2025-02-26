package process

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const origin = "Origin"

type requestsProcessor struct {
	hostFinder      HostFinder
	accessChecker   AccessChecker
	metrics         RequestMetrics
	closedEndpoints []string
}

// NewRequestsProcessor creates a new requests processor
func NewRequestsProcessor(
	hostFinder HostFinder,
	accessChecker AccessChecker,
	metrics RequestMetrics,
	closedEndpoints []string,
) (*requestsProcessor, error) {
	if check.IfNil(hostFinder) {
		return nil, errNilHostsFinder
	}
	if check.IfNil(accessChecker) {
		return nil, errNilAccessChecker
	}
	if check.IfNil(metrics) {
		return nil, errNilRequestMetrics
	}

	return &requestsProcessor{
		hostFinder:      hostFinder,
		accessChecker:   accessChecker,
		metrics:         metrics,
		closedEndpoints: closedEndpoints,
	}, nil
}

// ServeHTTP will serve the http requests
func (processor *requestsProcessor) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	values, err := url.ParseQuery(request.URL.RawQuery)
	if err != nil {
		RespondWithError(writer, fmt.Errorf("%w while parsing query", err), http.StatusBadRequest)
		return
	}

	requestID := createUniqueIdentifier()
	log.Trace("received request",
		"request ID", requestID,
		"URI", request.RequestURI,
		"query", parseStringMapsForLogger(values),
		"remote address", request.RemoteAddr,
		"header", parseStringMapsForLogger(request.Header),
	)

	newRequestURI, alias, err := processor.accessChecker.ShouldProcessRequest(request.Header, request.RequestURI)
	if err != nil {
		log.Trace("can not process request",
			"request ID", requestID,
			"error", err,
		)
		RespondWithError(writer, err, http.StatusUnauthorized)
		return
	}

	log.Trace("processing request",
		"request ID", requestID,
		"alias", alias,
	)
	processor.metrics.ProcessedResponse(alias)

	newHost, err := processor.hostFinder.FindHost(values)
	if err != nil {
		log.Trace("host not found",
			"request ID", requestID,
			"error", err,
		)
		RespondWithError(writer, err, http.StatusInternalServerError)
		return
	}

	urlPath := newHost.URL + newRequestURI

	if processor.isEndpointClosed(urlPath) {
		log.Trace("endpoint is closed",
			"request ID", requestID,
		)
		http.NotFound(writer, request)
		return
	}

	req, err := http.NewRequest(request.Method, urlPath, request.Body)
	if err != nil {
		log.Error("can not create request",
			"request ID", requestID,
			"target host", newHost,
			"URI", newRequestURI,
			"remote address", request.RemoteAddr,
			"error", err,
		)
		RespondWithError(writer, err, http.StatusInternalServerError)
		return
	}

	// pass through the header attributes
	for key, value := range request.Header {
		req.Header[key] = value
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("can not do request",
			"request ID", requestID,
			"target host", newHost,
			"URI", newRequestURI,
			"remote address", request.RemoteAddr,
			"error", err,
		)
		RespondWithError(writer, err, http.StatusInternalServerError)
		return
	}
	defer func() {
		_ = response.Body.Close()
	}()

	// pass through the response header attributes
	for key, value := range response.Header {
		writer.Header()[key] = value
	}
	writer.Header()[origin] = []string{newHost.Name}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		RespondWithError(writer, err, http.StatusInternalServerError)
		return
	}

	log.Trace("response generated",
		"request ID", requestID,
	)
	writer.WriteHeader(response.StatusCode)

	_, _ = writer.Write(bodyBytes)
}

func (processor *requestsProcessor) isEndpointClosed(url string) bool {
	for _, endoint := range processor.closedEndpoints {
		if strings.Contains(url, endoint) {
			return true
		}
	}

	return false
}

func createUniqueIdentifier() string {
	idLen := 10
	b := make([]byte, idLen)
	_, _ = rand.Read(b)

	return hex.EncodeToString(b)
}

func parseStringMapsForLogger(data map[string][]string) string {
	if logger.GetLoggerLogLevel(loggerName) != logger.LogTrace {
		// optimization, the log won't be written anyway
		return ""
	}

	result := "{"
	for key, values := range data {
		result += fmt.Sprintf("(%s=%v)", key, strings.Join(values, ", "))
	}

	return result + "}"
}

// IsInterfaceNil returns true if the value under the interface is nil
func (processor *requestsProcessor) IsInterfaceNil() bool {
	return processor == nil
}
