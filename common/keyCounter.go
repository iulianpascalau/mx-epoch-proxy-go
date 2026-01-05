package common

import (
	"sync"
	"sync/atomic"
)

type keyCounter struct {
	mut  sync.RWMutex
	keys map[string]*uint64
}

// NewKeyCounter creates a new instance of type key counter
func NewKeyCounter() *keyCounter {
	return &keyCounter{
		keys: make(map[string]*uint64),
	}
}

// IncrementReturningCurrent will increment the counter of the provided key, returning the current value
func (kc *keyCounter) IncrementReturningCurrent(key string) uint64 {
	// we attempt to get the current counter under the RLock as to not exclusively lock the map
	kc.mut.RLock()
	counter, ok := kc.keys[key]
	if ok {
		newVal := atomic.AddUint64(counter, 1)
		kc.mut.RUnlock()

		return newVal
	}
	kc.mut.RUnlock()

	// the key is missing, we exclusively lock the map
	kc.mut.Lock()
	counter, ok = kc.keys[key]
	if ok {
		// the key is present by a previous exclusively lock, use the existing counter
		newVal := atomic.AddUint64(counter, 1)
		kc.mut.Unlock()
		return newVal
	}

	// missing counter, create one
	var one = uint64(1)
	kc.keys[key] = &one
	kc.mut.Unlock()

	return 1
}

// Clear will clear the inner map and remove everything
func (kc *keyCounter) Clear() {
	kc.mut.Lock()
	kc.keys = make(map[string]*uint64)
	kc.mut.Unlock()
}

// IsInterfaceNil returns true if the value under the interface is nil
func (kc *keyCounter) IsInterfaceNil() bool {
	return kc == nil
}
