package testscommon

// KeyCounterStub -
type KeyCounterStub struct {
	IncrementReturningCurrentHandler func(key string) uint64
	ClearHandler                     func()
}

// IncrementReturningCurrent -
func (stub *KeyCounterStub) IncrementReturningCurrent(key string) uint64 {
	if stub.IncrementReturningCurrentHandler != nil {
		return stub.IncrementReturningCurrentHandler(key)
	}

	return 0
}

// Clear -
func (stub *KeyCounterStub) Clear() {
	if stub.ClearHandler != nil {
		stub.ClearHandler()
	}
}

// IsInterfaceNil -
func (stub *KeyCounterStub) IsInterfaceNil() bool {
	return stub == nil
}
