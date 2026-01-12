package common

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewKeyCounter(t *testing.T) {
	t.Parallel()

	kc := NewKeyCounter()
	assert.NotNil(t, kc)
	assert.False(t, kc.IsInterfaceNil())
}

func TestKeyCounter_Add(t *testing.T) {
	t.Parallel()
	kc := NewKeyCounter()

	assert.Equal(t, uint64(1), kc.IncrementReturningCurrent("key1"))
	assert.Equal(t, uint64(2), kc.IncrementReturningCurrent("key1"))

	assert.Equal(t, uint64(1), kc.IncrementReturningCurrent("key2"))
	assert.Equal(t, uint64(2), kc.IncrementReturningCurrent("key2"))

	assert.Equal(t, uint64(3), kc.IncrementReturningCurrent("key1"))
	assert.Equal(t, uint64(4), kc.IncrementReturningCurrent("key1"))
}

func TestKeyCounter_Clear(t *testing.T) {
	t.Parallel()

	kc := NewKeyCounter()
	kc.Clear()

	assert.Equal(t, uint64(1), kc.IncrementReturningCurrent("key1"))
	assert.Equal(t, uint64(2), kc.IncrementReturningCurrent("key1"))

	assert.Equal(t, uint64(1), kc.IncrementReturningCurrent("key2"))
	assert.Equal(t, uint64(2), kc.IncrementReturningCurrent("key2"))

	assert.Equal(t, uint64(3), kc.IncrementReturningCurrent("key1"))
	assert.Equal(t, uint64(4), kc.IncrementReturningCurrent("key1"))

	kc.Clear()

	assert.Equal(t, uint64(1), kc.IncrementReturningCurrent("key1"))
	assert.Equal(t, uint64(1), kc.IncrementReturningCurrent("key2"))
	assert.Equal(t, uint64(2), kc.IncrementReturningCurrent("key1"))
}

func TestKeyCounter_Concurrency(t *testing.T) {
	t.Parallel()

	kc := NewKeyCounter()

	wg := sync.WaitGroup{}
	numIterations := 10000
	numKeys := 1000
	wg.Add(numIterations)
	for i := 0; i < numIterations; i++ {
		go func(cnt int) {
			time.Sleep(time.Millisecond * 100)

			key := fmt.Sprintf("key_%d", cnt%numKeys)
			_ = kc.IncrementReturningCurrent(key)
			wg.Done()
		}(i)
	}

	wg.Wait()
	// all keys should have the counter set to numIterations/numKeys + 1
	for i := 0; i < numKeys; i++ {
		key := fmt.Sprintf("key_%d", i)
		val := kc.IncrementReturningCurrent(key)
		assert.Equal(t, uint64(numIterations/numKeys+1), val)
	}
}

func BenchmarkKeyCounter_IncrementAndReturnCurrent(b *testing.B) {
	b.Skip("long run, should be run manually")

	kc := NewKeyCounter()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		key := fmt.Sprintf("key_%d", i%1000)
		b.StartTimer()

		_ = kc.IncrementReturningCurrent(key)
	}
}
