package storage

import (
	"sync"
	"time"
)

type item struct {
	counter   uint64
	timestamp time.Time
}

type countersCache struct {
	mut  sync.RWMutex
	data map[string]*item
	ttl  time.Duration
}

// NewCountersCache	creates a new instance of type counters cache
func NewCountersCache(ttl time.Duration) (*countersCache, error) {
	if ttl < time.Millisecond {
		return nil, errInvalidTTL
	}

	return &countersCache{
		data: make(map[string]*item),
		ttl:  ttl,
	}, nil
}

// Get returns the counter value for the provided key, 0 if the key is not found
func (cc *countersCache) Get(key string) uint64 {
	cc.mut.RLock()
	defer cc.mut.RUnlock()

	value, found := cc.data[key]
	if !found {
		return 0
	}

	return value.counter
}

// Set sets the counter value for the provided key
func (cc *countersCache) Set(key string, counter uint64) {
	cc.mut.Lock()
	defer cc.mut.Unlock()

	cc.data[key] = &item{
		counter:   counter,
		timestamp: time.Now(),
	}
}

// Remove removes the counter value for the provided key
func (cc *countersCache) Remove(key string) {
	cc.mut.Lock()
	defer cc.mut.Unlock()

	delete(cc.data, key)
}

// Sweep removes all expired items from the cache
func (cc *countersCache) Sweep() {
	cc.mut.Lock()
	defer cc.mut.Unlock()

	for key, value := range cc.data {
		if time.Since(value.timestamp) > cc.ttl {
			delete(cc.data, key)
		}
	}
}

// IsInterfaceNil returns true if the value under the interface is nil
func (cc *countersCache) IsInterfaceNil() bool {
	return cc == nil
}
