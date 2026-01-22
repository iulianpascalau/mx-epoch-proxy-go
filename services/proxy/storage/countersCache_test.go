package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewCountersCache(t *testing.T) {
	t.Parallel()

	cc, err := NewCountersCache(time.Nanosecond)
	assert.Equal(t, errInvalidTTL, err)
	assert.True(t, cc.IsInterfaceNil())

	cc, err = NewCountersCache(time.Second)
	assert.Nil(t, err)
	assert.False(t, cc.IsInterfaceNil())
}

func TestCountersCache_SetGet(t *testing.T) {
	t.Parallel()

	cc, _ := NewCountersCache(time.Second)

	cc.Set("key1", 100)
	val := cc.Get("key1")
	assert.Equal(t, uint64(100), val)

	val = cc.Get("key2")
	assert.Equal(t, uint64(0), val)
}

func TestCountersCache_GetFromMissing(t *testing.T) {
	t.Parallel()

	cc, _ := NewCountersCache(time.Second)
	val := cc.Get("key1")
	assert.Equal(t, uint64(0), val)
}

func TestCountersCache_Remove(t *testing.T) {
	t.Parallel()

	cc, _ := NewCountersCache(time.Second)

	cc.Set("key1", 100)
	cc.Remove("key1")

	val := cc.Get("key1")
	assert.Equal(t, uint64(0), val)
}

func TestCountersCache_Sweep(t *testing.T) {
	t.Parallel()

	cc, _ := NewCountersCache(time.Millisecond * 100)

	cc.Set("key1", 100)
	time.Sleep(time.Millisecond * 150)
	cc.Set("key2", 200)

	cc.Sweep()

	assert.Equal(t, uint64(0), cc.Get("key1"))
	assert.Equal(t, uint64(200), cc.Get("key2"))
}

func TestCountersCache_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var cc *countersCache
	assert.True(t, cc.IsInterfaceNil())

	cc, _ = NewCountersCache(time.Second)
	assert.False(t, cc.IsInterfaceNil())
}
