package storage

import "errors"

var errKeyIsEmpty = errors.New("key is empty")
var errNilCountersCache = errors.New("nil counters cache")
var errInvalidTTL = errors.New("invalid TTL")
