package testscommon

import (
	"net/http"
)

// HttpHandlerStub -
type HttpHandlerStub struct {
	ServeHTTPCalled func(writer http.ResponseWriter, request *http.Request)
}

// ServeHTTP -
func (stub *HttpHandlerStub) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if stub.ServeHTTPCalled != nil {
		stub.ServeHTTPCalled(writer, request)
	}
}
