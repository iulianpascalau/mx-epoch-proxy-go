package process

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/common"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

const headerApiKey = "X-Api-Key"
const uriSeparator = "/"

var allowedVersions = []string{"v1"}

type accessChecker struct {
	keyAccessProvider         KeyAccessProvider
	counter                   KeyCounter
	maxNumCallsForFreeAccount uint64
}

// NewAccessChecker creates a new instance of type access checker
func NewAccessChecker(
	keyAccessProvider KeyAccessProvider,
	counter KeyCounter,
	maxNumCallsForFreeAccount uint64,
) (*accessChecker, error) {
	if check.IfNil(keyAccessProvider) {
		return nil, errNilKeyAccessChecker
	}
	if check.IfNil(counter) {
		return nil, errNilKeyCounter
	}

	return &accessChecker{
		keyAccessProvider:         keyAccessProvider,
		counter:                   counter,
		maxNumCallsForFreeAccount: maxNumCallsForFreeAccount,
	}, nil
}

// ShouldProcessRequest returns true if the request is allowed to be processed
func (checker *accessChecker) ShouldProcessRequest(header http.Header, requestURI string) (string, error) {
	accessKeyFromURI, processedRequestURI := processRequestURI(requestURI)
	accessKeyFromHeader := parseHeaderForAccessKey(header)

	accessKeys := common.NewKeysQueue(
		accessKeyFromURI,
		accessKeyFromHeader,
	)

	err := checker.atLeastOneKeyIsAllowed(accessKeys.Get())
	if err != nil {
		return "", err
	}

	return processedRequestURI, nil
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

func (checker *accessChecker) atLeastOneKeyIsAllowed(keys []string) error {
	if len(keys) == 0 {
		return fmt.Errorf("%w: no key provided", errUnauthorized)
	}

	var lastErr error
	for _, key := range keys {
		err := checker.isKeyAllowed(key)
		if err == nil {
			return nil
		}

		lastErr = err
	}

	return lastErr
}

func (checker *accessChecker) isKeyAllowed(key string) error {
	username, accountType, err := checker.keyAccessProvider.IsKeyAllowed(key)
	if err != nil {
		// error determining if the key is allowed, we should return false
		return fmt.Errorf("%w: %s", errUnauthorized, err.Error())
	}

	if accountType == common.PremiumAccountType {
		// the account is premium, no further throttling
		return nil
	}

	return checker.isNotThrottled(username)
}

func (checker *accessChecker) isNotThrottled(username string) error {
	currentCounter := checker.counter.IncrementReturningCurrent(username)

	if currentCounter <= checker.maxNumCallsForFreeAccount {
		return nil
	}

	return fmt.Errorf("%w: %s: current counter: %d, maximum per quota: %d",
		errUnauthorized, errTooManyRequestsForFreeAccount.Error(), currentCounter, checker.maxNumCallsForFreeAccount)
}

// IsInterfaceNil returns true if the value under the interface is nil
func (checker *accessChecker) IsInterfaceNil() bool {
	return checker == nil
}
