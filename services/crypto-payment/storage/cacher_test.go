package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTimeCacher_SetGet(t *testing.T) {
	t.Parallel()

	tc := NewTimeCacher(time.Minute)
	defer tc.Close()

	key := "key"
	val := "value"

	tc.Set(key, val)

	res, found := tc.Get(key)
	require.True(t, found)
	require.Equal(t, val, res)
}

func TestTimeCacher_Expiration(t *testing.T) {
	t.Parallel()

	ttl := 50 * time.Millisecond
	tc := NewTimeCacher(ttl)
	defer tc.Close()

	key := "key"
	val := "value"

	tc.Set(key, val)

	// fast check
	res, found := tc.Get(key)
	require.True(t, found)
	require.Equal(t, val, res)

	// wait for expiration (flush)
	time.Sleep(ttl * 2)

	res, found = tc.Get(key)
	require.False(t, found)
	require.Nil(t, res)
}

func TestTimeCacher_Cleanup(t *testing.T) {
	t.Parallel()

	// Short TTL to trigger cleanup quickly
	ttl := 10 * time.Millisecond
	tc := NewTimeCacher(ttl)
	defer tc.Close()

	key := "key"
	val := "value"

	tc.Set(key, val)

	// Wait for cleanup loop to run: ttl + tiny buffer
	time.Sleep(ttl * 2)

	// key should have been removed from the map
	tc.mut.RLock()
	_, found := tc.data[key]
	tc.mut.RUnlock()
	require.False(t, found, "key should have been removed by background cleanup")
}

func TestTimeCacher_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	tc := NewTimeCacher(time.Millisecond * 10)
	defer tc.Close()

	key := "key"
	val := "value"

	// Start concurrent writes
	go func() {
		for i := 0; i < 100; i++ {
			tc.Set(key, val)
			time.Sleep(time.Millisecond)
		}
	}()

	// Start concurrent reads
	go func() {
		for i := 0; i < 100; i++ {
			tc.Get(key)
			time.Sleep(time.Millisecond)
		}
	}()
}
