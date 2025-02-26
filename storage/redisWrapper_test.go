//go:build redis

package storage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const redisDockerURL = "127.0.0.1:6379"
const invalidURL = "sa"

func TestNewRedisWrapper(t *testing.T) {
	t.Parallel()

	t.Run("with an invalid connection URL should not error", func(t *testing.T) {
		t.Parallel()

		wrapper := NewRedisWrapper(invalidURL, "")
		assert.NotNil(t, wrapper)

		_ = wrapper.Close()
	})
	t.Run("with a correct connection URL should work", func(t *testing.T) {
		t.Parallel()

		wrapper := NewRedisWrapper(redisDockerURL, "")
		assert.NotNil(t, wrapper)

		_ = wrapper.Close()
	})
}

func TestRedisWrapper_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var instance *redisWrapper
	assert.True(t, instance.IsInterfaceNil())

	instance = &redisWrapper{}
	assert.False(t, instance.IsInterfaceNil())
}

func TestRedisWrapper_SetWithoutExpiryGetDelete(t *testing.T) {
	t.Parallel()

	t.Run("with an invalid connection URL should error", func(t *testing.T) {
		t.Parallel()

		wrapper := NewRedisWrapper(invalidURL, "")
		defer func() {
			err := wrapper.Close()
			assert.Nil(t, err)
		}()

		err := wrapper.SetWithoutExpiry(context.Background(), "key1", "value1")
		assert.NotNil(t, err)

		recoveredValue, found, err := wrapper.Get(context.Background(), "key1")
		assert.NotNil(t, err)
		assert.False(t, found)
		assert.Empty(t, recoveredValue)

		err = wrapper.Delete(context.Background(), "key1")
		assert.NotNil(t, err)
	})
	t.Run("with a correct connection URL should work", func(t *testing.T) {
		t.Parallel()

		wrapper := NewRedisWrapper(redisDockerURL, "")
		_ = wrapper.Delete(context.Background(), "key2")
		defer func() {
			err := wrapper.Close()
			assert.Nil(t, err)
		}()

		err := wrapper.SetWithoutExpiry(context.Background(), "key2", "value1")
		assert.Nil(t, err)

		recoveredValue, found, err := wrapper.Get(context.Background(), "key2")
		assert.Nil(t, err)
		assert.True(t, found)
		assert.Equal(t, "value1", recoveredValue)

		err = wrapper.Delete(context.Background(), "key2")
		assert.Nil(t, err)

		recoveredValue, found, err = wrapper.Get(context.Background(), "key2")
		assert.Nil(t, err)
		assert.False(t, found)
		assert.Empty(t, recoveredValue)
	})
}

func TestRedisWrapper_SetGetDelete(t *testing.T) {
	t.Parallel()

	t.Run("with an invalid connection URL should error", func(t *testing.T) {
		t.Parallel()

		wrapper := NewRedisWrapper(invalidURL, "")
		defer func() {
			err := wrapper.Close()
			assert.Nil(t, err)
		}()

		err := wrapper.Set(context.Background(), "key3", "value1", time.Second)
		assert.NotNil(t, err)

		recoveredValue, found, err := wrapper.Get(context.Background(), "key3")
		assert.NotNil(t, err)
		assert.False(t, found)
		assert.Empty(t, recoveredValue)
	})
	t.Run("with a correct connection URL should work", func(t *testing.T) {
		t.Parallel()

		wrapper := NewRedisWrapper(redisDockerURL, "")
		_ = wrapper.Delete(context.Background(), "key4")
		defer func() {
			err := wrapper.Close()
			assert.Nil(t, err)
		}()

		recoveredValue, found, err := wrapper.Get(context.Background(), "non-existing-key")
		assert.Nil(t, err)
		assert.False(t, found)
		assert.Empty(t, recoveredValue)

		err = wrapper.Set(context.Background(), "key4", "value1", time.Second)
		assert.Nil(t, err)

		recoveredValue, found, err = wrapper.Get(context.Background(), "key4")
		assert.Nil(t, err)
		assert.True(t, found)
		assert.Equal(t, "value1", recoveredValue)

		time.Sleep(time.Second * 3)

		recoveredValue, found, err = wrapper.Get(context.Background(), "key4")
		assert.Nil(t, err)
		assert.False(t, found)
		assert.Empty(t, recoveredValue)

		recoveredValue, found, err = wrapper.Get(context.Background(), "key4")
		assert.Nil(t, err)
		assert.False(t, found)
		assert.Empty(t, recoveredValue)
	})
}

func TestRedisWrapper_Increment(t *testing.T) {
	t.Parallel()

	wrapper := NewRedisWrapper(redisDockerURL, "")
	defer func() {
		_ = wrapper.Close()
	}()
	_ = wrapper.Delete(context.Background(), "key5")

	val, found, err := wrapper.Get(context.Background(), "key5")
	assert.Nil(t, err)
	assert.False(t, found)
	assert.Empty(t, val)

	err = wrapper.Increment(context.Background(), "key5")
	assert.Nil(t, err)

	val, found, err = wrapper.Get(context.Background(), "key5")
	assert.Nil(t, err)
	assert.True(t, found)
	assert.Equal(t, "1", val)

	err = wrapper.Increment(context.Background(), "key5")
	assert.Nil(t, err)

	val, found, err = wrapper.Get(context.Background(), "key5")
	assert.Nil(t, err)
	assert.True(t, found)
	assert.Equal(t, "2", val)

	for i := 0; i < 10; i++ {
		err = wrapper.Increment(context.Background(), "key5")
		assert.Nil(t, err)
	}

	val, found, err = wrapper.Get(context.Background(), "key5")
	assert.Nil(t, err)
	assert.True(t, found)
	assert.Equal(t, "12", val)
}

func TestRedisWrapper_GetAllKeys(t *testing.T) {
	t.Parallel()

	wrapper := NewRedisWrapper(redisDockerURL, "")
	defer func() {
		_ = wrapper.Close()
	}()
	_ = wrapper.Delete(context.Background(), "key6")
	_ = wrapper.Delete(context.Background(), "key7")
	_ = wrapper.Delete(context.Background(), "key8")

	_ = wrapper.SetWithoutExpiry(context.Background(), "key6", "value6")
	_ = wrapper.SetWithoutExpiry(context.Background(), "key7", "value7")
	_ = wrapper.SetWithoutExpiry(context.Background(), "key8", "value8")

	allKeys, err := wrapper.GetAllKeys(context.Background())
	assert.Nil(t, err)

	keysMap := map[string]struct{}{
		"key6": {},
		"key7": {},
		"key8": {},
	}
	for _, key := range allKeys {
		delete(keysMap, key)
	}

	assert.Empty(t, keysMap)
}
