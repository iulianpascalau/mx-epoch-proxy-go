package storage

// CountersCache is the interface for the counters cache
type CountersCache interface {
	Get(key string) uint64
	Set(key string, counter uint64)
	Remove(key string)
	Sweep()
	IsInterfaceNil() bool
}
