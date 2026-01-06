package process

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/iulianpascalau/mx-epoch-proxy-go/common"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const origin = "Origin"

type requestsProcessor struct {
	hostFinder         HostFinder
	accessChecker      AccessChecker
	performanceMonitor PerformanceMonitor
	closedEndpoints    []string
}

// NewRequestsProcessor creates a new requests processor
func NewRequestsProcessor(
	hostFinder HostFinder,
	accessChecker AccessChecker,
	performanceMonitor PerformanceMonitor,
	closedEndpoints []string,
) (*requestsProcessor, error) {
	if check.IfNil(hostFinder) {
		return nil, errNilHostsFinder
	}
	if check.IfNil(accessChecker) {
		return nil, errNilAccessChecker
	}
	if check.IfNil(performanceMonitor) {
		return nil, fmt.Errorf("nil performance monitor")
	}

	return &requestsProcessor{
		hostFinder:         hostFinder,
		accessChecker:      accessChecker,
		performanceMonitor: performanceMonitor,
		closedEndpoints:    closedEndpoints,
	}, nil
}

// ServeHTTP will serve the http requests
func (processor *requestsProcessor) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	values, err := url.ParseQuery(request.URL.RawQuery)
	if err != nil {
		RespondWithError(writer, fmt.Errorf("%w while parsing query", err), http.StatusBadRequest)
		return
	}

	log.Trace("received request",
		"URI", request.RequestURI,
		"query", parseStringMapsForLogger(values),
		"remote address", request.RemoteAddr,
		"header", parseStringMapsForLogger(request.Header),
	)

	start := time.Now()
	newRequestURI, err := processor.accessChecker.ShouldProcessRequest(request.Header, request.RequestURI)
	if err != nil {
		log.Trace("can not process request",
			"error", err,
		)
		RespondWithError(writer, err, http.StatusUnauthorized)
		return
	}

	newHost, err := processor.hostFinder.FindHost(values)
	if err != nil {
		log.Trace("host not found",
			"error", err,
		)
		RespondWithError(writer, err, http.StatusInternalServerError)
		return
	}

	urlPath := newHost.URL + newRequestURI

	if processor.isEndpointClosed(urlPath) {
		log.Trace("endpoint is closed")
		http.NotFound(writer, request)
		return
	}

	req, err := http.NewRequest(request.Method, urlPath, request.Body)
	if err != nil {
		log.Error("can not create request",
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
	duration := time.Since(start)

	if err != nil {
		log.Error("can not do request",
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

	processor.updatePerformanceMetrics(duration)

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

	log.Trace("response generated")
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

func (processor *requestsProcessor) updatePerformanceMetrics(duration time.Duration) {
	label := common.ConvertTimeToInterval(duration)
	err := processor.performanceMonitor.AddPerformanceMetric(label)
	if err != nil {
		log.Error("failed to add performance metric", "error", err)
	}
}

// IsInterfaceNil returns true if the value under the interface is nil
func (processor *requestsProcessor) IsInterfaceNil() bool {
	return processor == nil
}
