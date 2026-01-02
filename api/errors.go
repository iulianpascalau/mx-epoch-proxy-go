package api

import "errors"

var errNilHandler = errors.New("nil http handler")
var errNilKeyAccessChecker = errors.New("nil key access checker")
