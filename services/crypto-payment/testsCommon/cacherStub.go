package testsCommon

// CacherStub -
type CacherStub struct {
	GetHandler            func(key string) (interface{}, bool)
	SetHandler            func(key string, value interface{})
	CloseHandler          func()
	IsInterfaceNilHandler func() bool
}

// Get -
func (stub *CacherStub) Get(key string) (interface{}, bool) {
	if stub.GetHandler != nil {
		return stub.GetHandler(key)
	}
	return nil, false
}

// Set -
func (stub *CacherStub) Set(key string, value interface{}) {
	if stub.SetHandler != nil {
		stub.SetHandler(key, value)
	}
}

// Close -
func (stub *CacherStub) Close() {
	if stub.CloseHandler != nil {
		stub.CloseHandler()
	}
}

// IsInterfaceNil -
func (stub *CacherStub) IsInterfaceNil() bool {
	return stub == nil
}
