package api

import "errors"

var errNilHandler = errors.New("nil http handler")
var errNilKeyAccessChecker = errors.New("nil key access checker")
var errNilEmailSender = errors.New("nil email sender")
var errEmptyHTMLTemplate = errors.New("empty HTML template")
var errNilCaptchaHandler = errors.New("nil captcha handler")
