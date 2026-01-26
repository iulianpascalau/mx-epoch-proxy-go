package framework

import (
	"fmt"
	"io"
	"sync/atomic"
)

type MockCaptchaWrapper struct {
	id uint64
}

// VerifyString returns true iof the id == digits
func (wrapper *MockCaptchaWrapper) VerifyString(id string, digits string) bool {
	return id == digits
}

// NewCaptcha returns the current id as a string. Then it increments the id
func (wrapper *MockCaptchaWrapper) NewCaptcha() string {
	val := atomic.AddUint64(&wrapper.id, 1)
	return fmt.Sprintf("%d", val)
}

// Reload does nothing
func (wrapper *MockCaptchaWrapper) Reload(_ string) {
}

// WriteNoError writes the id to the writer
func (wrapper *MockCaptchaWrapper) WriteNoError(w io.Writer, id string) {
	_, _ = w.Write([]byte(id))
}

// IsInterfaceNil -
func (wrapper *MockCaptchaWrapper) IsInterfaceNil() bool {
	return wrapper == nil
}
