package process

import (
	"net/http"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/common"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/config"
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

// PerformanceMonitor is able to store performance metrics
type PerformanceMonitor interface {
	AddPerformanceMetricAsync(label string)
	IsInterfaceNil() bool
}
