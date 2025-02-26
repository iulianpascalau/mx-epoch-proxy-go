package metrics

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewRequestMetrics(t *testing.T) {
	t.Parallel()

	t.Run("nil storer should error", func(t *testing.T) {
		t.Parallel()

		metrics, err := NewRequestMetrics(nil)
		assert.Nil(t, metrics)
		assert.Equal(t, errNilStorer, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		metrics, err := NewRequestMetrics(&testscommon.StorerStub{})
		assert.NotNil(t, metrics)
		assert.Nil(t, err)
	})
}

func TestRequestsMetrics_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var instance *requestsMetrics
	assert.True(t, instance.IsInterfaceNil())

	instance = &requestsMetrics{}
	assert.False(t, instance.IsInterfaceNil())
}

func TestRequestsMetrics_ProcessedResponse(t *testing.T) {
	t.Parallel()

	keyValue := make(map[string]int)
	metrics, _ := NewRequestMetrics(&testscommon.StorerStub{
		IncrementHandler: func(ctx context.Context, key string) error {
			keyValue[key]++

			return nil
		},
	})

	metrics.ProcessedResponse("alias1")
	assert.Equal(t, 1, keyValue["alias1_total"])
	assert.Equal(t, 1, keyValue["ALL_total"])

	metrics.ProcessedResponse("alias1")
	assert.Equal(t, 2, keyValue["alias1_total"])
	assert.Equal(t, 2, keyValue["ALL_total"])

	metrics.ProcessedResponse("alias2")
	assert.Equal(t, 2, keyValue["alias1_total"])
	assert.Equal(t, 1, keyValue["alias2_total"])
	assert.Equal(t, 3, keyValue["ALL_total"])

	metrics.ProcessedResponse("alias2")
	assert.Equal(t, 2, keyValue["alias1_total"])
	assert.Equal(t, 2, keyValue["alias2_total"])
	assert.Equal(t, 4, keyValue["ALL_total"])

	metrics.ProcessedResponse("ALL")
	assert.Equal(t, 2, keyValue["alias1_total"])
	assert.Equal(t, 2, keyValue["alias2_total"])
	assert.Equal(t, 5, keyValue["ALL_total"])

	metrics.ProcessedResponse("ALL")
	assert.Equal(t, 2, keyValue["alias1_total"])
	assert.Equal(t, 2, keyValue["alias2_total"])
	assert.Equal(t, 6, keyValue["ALL_total"])
}

func TestRequestsMetrics_GetAllKeyValues(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	t.Run("get all keys handler errors", func(t *testing.T) {
		t.Parallel()

		metrics, _ := NewRequestMetrics(&testscommon.StorerStub{
			GetAllKeysHandler: func(ctx context.Context) ([]string, error) {
				return nil, expectedErr
			},
			GetHandler: func(ctx context.Context, key string) (string, bool, error) {
				assert.Fail(t, "should have not called")
				return "", false, nil
			},
		})

		data := metrics.GetAllKeyValues()
		expectedData := []string{fmt.Sprintf("%s while getting all the keys", expectedErr.Error())}
		assert.Equal(t, expectedData, data)
	})
	t.Run("a random get errors", func(t *testing.T) {
		t.Parallel()

		metrics, _ := NewRequestMetrics(&testscommon.StorerStub{
			GetAllKeysHandler: func(ctx context.Context) ([]string, error) {
				return []string{"key1", "key2"}, nil
			},
			GetHandler: func(ctx context.Context, key string) (string, bool, error) {
				switch key {
				case "key1":
					return "value1", true, nil
				case "key2":
					return "", false, expectedErr
				default:
					return "", false, nil
				}
			},
		})

		data := metrics.GetAllKeyValues()
		expectedData := []string{
			"    key1: value1",
			"    key2: " + fmt.Sprintf("%s while getting data", expectedErr.Error()),
		}
		assert.Equal(t, expectedData, data)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		metrics, _ := NewRequestMetrics(&testscommon.StorerStub{
			GetAllKeysHandler: func(ctx context.Context) ([]string, error) {
				return []string{"key1", "key2"}, nil
			},
			GetHandler: func(ctx context.Context, key string) (string, bool, error) {
				switch key {
				case "key1":
					return "value1", true, nil
				case "key2":
					return "value2", true, nil
				default:
					return "", false, nil
				}
			},
		})

		data := metrics.GetAllKeyValues()
		expectedData := []string{
			"    key1: value1",
			"    key2: value2",
		}
		assert.Equal(t, expectedData, data)
	})
}
