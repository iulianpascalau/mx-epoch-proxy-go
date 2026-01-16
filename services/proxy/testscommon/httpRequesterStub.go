package testscommon

// HttpRequesterStub -
type HttpRequesterStub struct {
	DoRequestHandler func(method string, url string, apiKey string, result any) error
}

// DoRequest -
func (stub *HttpRequesterStub) DoRequest(method string, url string, apiKey string, result any) error {
	if stub.DoRequestHandler != nil {
		return stub.DoRequestHandler(method, url, apiKey, result)
	}

	return nil
}

// IsInterfaceNil -
func (stub *HttpRequesterStub) IsInterfaceNil() bool {
	return stub == nil
}
