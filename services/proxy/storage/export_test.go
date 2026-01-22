package storage

// GetCacheCounterForUser returns the counter for the provided username
func (wrapper *sqliteWrapper) GetCacheCounterForUser(username string) uint64 {
	return wrapper.counters.Get(username)
}
