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
var errDuplicatedAccessKeyAlias = errors.New("duplicated access key alias")
var errEmptyAlias = errors.New("empty alias")
var errEmptyKey = errors.New("empty key")
var errNilAccessChecker = errors.New("nil access checker")
var errNilRequestMetrics = errors.New("nil request metrics")
