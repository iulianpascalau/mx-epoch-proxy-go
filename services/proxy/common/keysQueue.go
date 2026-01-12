package common

import "strings"

const initialQueueSize = 10

type keysQueue struct {
	keys []string
}

// NewKeysQueue will hold a unique keys queue
func NewKeysQueue(keys ...string) *keysQueue {
	kq := &keysQueue{
		keys: make([]string, 0, initialQueueSize),
	}

	for _, k := range keys {
		kq.Add(k)
	}

	return kq
}

// Add will try to add a non-empty, unique key
func (q *keysQueue) Add(key string) {
	key = strings.TrimSpace(key)
	if len(key) == 0 {
		return
	}

	q.addUnique(key)
}

func (q *keysQueue) addUnique(key string) {
	for _, k := range q.keys {
		if k == key {
			return
		}
	}

	q.keys = append(q.keys, key)
}

// Get returns the inner queue
func (q *keysQueue) Get() []string {
	return q.keys
}
