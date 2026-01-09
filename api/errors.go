package api

import "errors"

var errNilHandler = errors.New("nil http handler")
var errNilKeyAccessProvider = errors.New("nil key access provider")
var errNilEmailSender = errors.New("nil email sender")
var errEmptyHTMLTemplate = errors.New("empty HTML template")
var errNilCaptchaHandler = errors.New("nil captcha handler")
var errNilAuthenticator = errors.New("nil authenticator")
