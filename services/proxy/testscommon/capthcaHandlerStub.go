package testscommon

import (
	"io"
)

// CaptchaHandlerStub -
type CaptchaHandlerStub struct {
	VerifyStringHandler func(id string, digits string) bool
	NewCaptchaHandler   func() string
	ReloadHandler       func(id string)
	WriteNoErrorHandler func(w io.Writer, id string)
}

// VerifyString -
func (stub *CaptchaHandlerStub) VerifyString(id string, digits string) bool {
	if stub.VerifyStringHandler != nil {
		return stub.VerifyStringHandler(id, digits)
	}

	return true
}

// NewCaptcha -
func (stub *CaptchaHandlerStub) NewCaptcha() string {
	if stub.NewCaptchaHandler != nil {
		return stub.NewCaptchaHandler()
	}

	return ""
}

// Reload -
func (stub *CaptchaHandlerStub) Reload(id string) {
	if stub.ReloadHandler != nil {
		stub.ReloadHandler(id)
	}
}

// WriteNoError -
func (stub *CaptchaHandlerStub) WriteNoError(w io.Writer, id string) {
	if stub.WriteNoErrorHandler != nil {
		stub.WriteNoErrorHandler(w, id)
	}
}

// IsInterfaceNil -
func (stub *CaptchaHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
