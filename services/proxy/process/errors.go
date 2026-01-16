package process

import (
	"errors"
)

var errNilHostsFinder = errors.New("nil hosts finder")
var errNoGatewayDefined = errors.New("no gateway defined")
var errBadGatewayInterval = errors.New("bad gateway interval")
var errUnexpectedIntervalStart = errors.New("unexpected start interval")
var errCanNotDetermineSuitableHost = errors.New("can not determine a suitable host")
var errNilMap = errors.New("nil URL values map")
var errMoreThanOneLatestDataGatewayFound = errors.New("more than one latest data gateway found")
var errMissingValue = errors.New("missing value")
var errNoLatestDataGatewayDefined = errors.New("no latest data gateway defined")
var errUnauthorized = errors.New("unauthorized")
var errNilAccessChecker = errors.New("nil access checker")
var errNilKeyAccessChecker = errors.New("nil key access checker")
var errNilKeyCounter = errors.New("nil key counter")
var errTooManyRequestsForFreeAccount = errors.New("too many requests for free account")
var errUnexpectedStatusCode = errors.New("unexpected status code")
