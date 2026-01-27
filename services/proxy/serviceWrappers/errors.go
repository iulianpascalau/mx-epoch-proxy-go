package serviceWrappers

import "errors"

var errNilHTTPRequester = errors.New("nil http requester")
var errCryptoPaymentIsDisabled = errors.New("crypto payment service is disabled")
