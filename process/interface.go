package process

import (
	"net/http"

	"github.com/iulianpascalau/mx-epoch-proxy-go/common"
	"github.com/iulianpascalau/mx-epoch-proxy-go/config"
)

// HostFinder is able to return a valid host based on a search criteria
type HostFinder interface {
	FindHost(urlValues map[string][]string) (config.GatewayConfig, error)
	IsInterfaceNil() bool
}

// AccessChecker is able to check if the request should be processed or not
type AccessChecker interface {
	ShouldProcessRequest(header http.Header, requestURI string) (string, error)
	IsInterfaceNil() bool
}

// KeyAccessProvider can decide if a provided key has or not query access
type KeyAccessProvider interface {
	IsKeyAllowed(key string) (string, common.AccountType, error)
	IsInterfaceNil() bool
}

// KeyCounter is able to keep track of how many increments a particular key has
type KeyCounter interface {
	IncrementReturningCurrent(key string) uint64
	Clear()
	IsInterfaceNil() bool
}
