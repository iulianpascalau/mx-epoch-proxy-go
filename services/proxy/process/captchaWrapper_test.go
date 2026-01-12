package process

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCaptchaWrapper(t *testing.T) {
	t.Parallel()

	instance := NewCaptchaWrapper()
	assert.NotNil(t, instance)
	assert.False(t, instance.IsInterfaceNil())
}

func TestCaptchaWrapper_NewCaptcha(t *testing.T) {
	t.Parallel()

	instance := NewCaptchaWrapper()
	id := instance.NewCaptcha()
	fmt.Println(id)
	assert.NotEmpty(t, id)
}

func TestCaptchaWrapper_Reload(t *testing.T) {
	t.Parallel()

	instance := NewCaptchaWrapper()
	id := instance.NewCaptcha()
	fmt.Println(id)

	instance.Reload(id)

	assert.NotEmpty(t, id)
}

func TestCaptchaWrapper_WriteNoError(t *testing.T) {
	t.Parallel()

	t.Run("valid id", func(t *testing.T) {
		instance := NewCaptchaWrapper()
		id := instance.NewCaptcha()

		buff := bytes.NewBuffer(make([]byte, 0))
		instance.WriteNoError(buff, id)
		assert.NotEmpty(t, buff.String())
	})
	t.Run("not a valid id should not write to buffer", func(t *testing.T) {
		instance := NewCaptchaWrapper()

		buff := bytes.NewBuffer(make([]byte, 0))
		instance.WriteNoError(buff, "missing id")
		assert.Empty(t, buff.String())
	})
}

func TestCaptchaWrapper_Validate(t *testing.T) {
	t.Parallel()

	// limited testing
	instance := NewCaptchaWrapper()
	assert.False(t, instance.VerifyString("missing id", "dummy solution"))
}
