package process

import (
	"encoding/json"
	"net/http"

	logger "github.com/multiversx/mx-chain-logger-go"
)

// UrlParameterBlockNonce represents the name of a URL parameter
const UrlParameterBlockNonce = "blockNonce"

// UrlParameterHintEpoch represents the name of a URL parameter
const UrlParameterHintEpoch = "hintEpoch"

// ReturnCode defines the type defines to identify return codes
type ReturnCode string

const (
	// ReturnCodeRequestError defines a request which hasn't been executed successfully due to a bad request received
	ReturnCodeRequestError ReturnCode = "bad_request"
)

var jsonContentType = []string{"application/json; charset=utf-8"}
var log = logger.GetOrCreate("process")

// GenericAPIResponse defines the structure of all responses on API endpoints
type GenericAPIResponse struct {
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
	Code  ReturnCode  `json:"code"`
}

// RespondWithError should be called when the request cannot be satisfied due to an internal error
func RespondWithError(writer http.ResponseWriter, err error, statusCode int) {
	writer.WriteHeader(statusCode)
	writeContentType(writer, jsonContentType)

	trySendResponse(writer, err)
}

func trySendResponse(writer http.ResponseWriter, err error) {
	response := &GenericAPIResponse{
		Error: err.Error(),
		Code:  ReturnCodeRequestError,
	}

	jsonBytes, errJson := json.Marshal(response)
	if errJson != nil {
		return
	}

	_, _ = writer.Write(jsonBytes)
}

func writeContentType(w http.ResponseWriter, value []string) {
	header := w.Header()
	val := header["Content-Type"]
	if len(val) == 0 {
		header["Content-Type"] = value
	}
}
