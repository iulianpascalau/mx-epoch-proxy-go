package storage

import (
	"sync"
	"time"
)

type timeCacher struct {
	mut  sync.RWMutex
	data map[string]interface{}
	ttl  time.Duration
	stop chan struct{}
}

// NewTimeCacher creates a new time cacher. Concurrent safe component.
func NewTimeCacher(ttl time.Duration) *timeCacher {
	tc := &timeCacher{
		data: make(map[string]interface{}),
		ttl:  ttl,
		stop: make(chan struct{}),
	}

	go tc.cleanupLoop()

	return tc
}

// Get will try to get the value for the provided key, returning it and a boolean indicating whether the value was found
func (tc *timeCacher) Get(key string) (interface{}, bool) {
	tc.mut.RLock()
	defer tc.mut.RUnlock()

	rec, found := tc.data[key]

	return rec, found
}

// Set will set the value for the provided key, overwriting any existing value
func (tc *timeCacher) Set(key string, value interface{}) {
	tc.mut.Lock()
	defer tc.mut.Unlock()

	tc.data[key] = value
}

// Close stops the cleanup loop
func (tc *timeCacher) Close() {
	close(tc.stop)
}

func (tc *timeCacher) cleanupLoop() {
	ticker := time.NewTicker(tc.ttl)
	defer ticker.Stop()

	for {
		select {
		case <-tc.stop:
			return
		case <-ticker.C:
			tc.cleanup()
		}
	}
}

// cleanup clears the entire cache
func (tc *timeCacher) cleanup() {
	tc.mut.Lock()
	defer tc.mut.Unlock()

	tc.data = make(map[string]interface{})
}

// IsInterfaceNil returns true if the value under the interface is nil
func (tc *timeCacher) IsInterfaceNil() bool {
	return tc == nil
}
