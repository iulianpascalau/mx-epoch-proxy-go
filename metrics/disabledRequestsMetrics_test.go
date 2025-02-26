package metrics

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDisabledRequestsMetrics_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var instance *DisabledRequestsMetrics
	assert.True(t, instance.IsInterfaceNil())

	instance = &DisabledRequestsMetrics{}
	assert.False(t, instance.IsInterfaceNil())
}

func TestDisabledRequestsMetrics_AllMethodsShouldNotPanic(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panicked %v", r))
		}
	}()

	metrics := &DisabledRequestsMetrics{}
	metrics.ProcessedResponse("")

	keyValues := metrics.GetAllKeyValues()
	assert.Empty(t, keyValues)
}
