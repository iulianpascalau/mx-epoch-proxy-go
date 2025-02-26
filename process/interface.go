package process

import (
	"net/http"

	"github.com/iulianpascalau/mx-epoch-proxy-go/config"
)

// HostFinder is able to return a valid host based on a search criteria
type HostFinder interface {
	FindHost(urlValues map[string][]string) (config.GatewayConfig, error)
	IsInterfaceNil() bool
}

// AccessChecker is able to check if the request should be processed or not
type AccessChecker interface {
	ShouldProcessRequest(header http.Header, requestURI string) (string, string, error)
	IsInterfaceNil() bool
}

// RequestMetrics is able to handle the metrics of an alias
type RequestMetrics interface {
	ProcessedResponse(alias string)
	IsInterfaceNil() bool
}
