package process

import (
	"net/http"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core/check"
)

const headerApiKey = "X-Api-Key"
const uriSeparator = "/"

var allowedVersions = []string{"v1"}

type accessChecker struct {
	keyAccessProvider KeyAccessProvider
}

// NewAccessChecker creates a new instance of type access checker
func NewAccessChecker(keyAccessProvider KeyAccessProvider) (*accessChecker, error) {
	if check.IfNil(keyAccessProvider) {
		return nil, errNilKeyAccessChecker
	}

	return &accessChecker{
		keyAccessProvider: keyAccessProvider,
	}, nil
}

// ShouldProcessRequest returns true if the request is allowed to be processed
func (checker *accessChecker) ShouldProcessRequest(header http.Header, requestURI string) (string, error) {
	accessKey, processedRequestURI := processRequestURI(requestURI)
	err := checker.keyAccessProvider.IsKeyAllowed(accessKey)
	if err == nil {
		// authorized, useless to check the header
		return processedRequestURI, nil
	}

	accessKey = parseHeaderForAccessKey(header)
	err = checker.keyAccessProvider.IsKeyAllowed(accessKey)
	if err == nil {
		// authorized, useless to check the header
		return processedRequestURI, nil
	}

	return "", errUnauthorized
}

func processRequestURI(inputRequestURI string) (string, string) {
	splt := strings.Split(inputRequestURI, uriSeparator)
	if len(splt) < 4 {
		// token was not provided in the URL
		return "", inputRequestURI
	}
	if !checkVersion(splt[1]) {
		// token was not provided in the URL
		return "", inputRequestURI
	}

	return strings.ToLower(splt[2]), uriSeparator + strings.Join(splt[3:], uriSeparator)
}

func checkVersion(version string) bool {
	for _, vers := range allowedVersions {
		if vers == strings.ToLower(version) {
			return true
		}
	}

	return false
}

func parseHeaderForAccessKey(header http.Header) string {
	val := header.Get(headerApiKey)
	if len(val) == 0 {
		return ""
	}

	return strings.ToLower(val)
}

// IsInterfaceNil returns true if the value under the interface is nil
func (checker *accessChecker) IsInterfaceNil() bool {
	return checker == nil
}
