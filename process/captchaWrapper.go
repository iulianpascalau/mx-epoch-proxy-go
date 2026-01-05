package process

import (
	"io"

	"github.com/dchest/captcha"
)

type captchaWrapper struct {
}

// NewCaptchaWrapper creates a new captcha wrapper
func NewCaptchaWrapper() *captchaWrapper {
	return &captchaWrapper{}
}

// VerifyString calls the captcha VerifyString method
func (wrapper *captchaWrapper) VerifyString(id string, digits string) bool {
	return captcha.VerifyString(id, digits)
}

// NewCaptcha calls the captcha NewCaptcha method and returns the id
func (wrapper *captchaWrapper) NewCaptcha() string {
	return captcha.New()
}

// Reload calls the captcha Reload method
func (wrapper *captchaWrapper) Reload(id string) {
	captcha.Reload(id)
}

// WriteNoError calls the captcha WriteImage method and ignores any errors
func (wrapper *captchaWrapper) WriteNoError(w io.Writer, id string) {
	_ = captcha.WriteImage(w, id, captcha.StdWidth, captcha.StdHeight)
}

// IsInterfaceNil returns true if there is no value under the interface
func (wrapper *captchaWrapper) IsInterfaceNil() bool {
	return wrapper == nil
}
