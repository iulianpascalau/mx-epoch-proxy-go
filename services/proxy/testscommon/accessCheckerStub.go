package testscommon

import (
	"net/http"
)

// AccessCheckerStub -
type AccessCheckerStub struct {
	ShouldProcessRequestHandler func(header http.Header, requestURI string) (string, error)
}

// ShouldProcessRequest -
func (stub *AccessCheckerStub) ShouldProcessRequest(header http.Header, requestURI string) (string, error) {
	if stub.ShouldProcessRequestHandler == nil {
		return requestURI, nil
	}

	return stub.ShouldProcessRequestHandler(header, requestURI)
}

// IsInterfaceNil -
func (stub *AccessCheckerStub) IsInterfaceNil() bool {
	return stub == nil
}
