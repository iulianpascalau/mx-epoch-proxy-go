package testscommon

import (
	"net/http"

	"github.com/iulianpascalau/mx-epoch-proxy-go/common"
)

// AccessCheckerStub -
type AccessCheckerStub struct {
	ShouldProcessRequestHandler func(header http.Header, requestURI string) (string, string, error)
}

// ShouldProcessRequest -
func (stub *AccessCheckerStub) ShouldProcessRequest(header http.Header, requestURI string) (string, string, error) {
	if stub.ShouldProcessRequestHandler == nil {
		return requestURI, common.AllAliases, nil
	}

	return stub.ShouldProcessRequestHandler(header, requestURI)
}

// IsInterfaceNil -
func (stub *AccessCheckerStub) IsInterfaceNil() bool {
	return stub == nil
}
