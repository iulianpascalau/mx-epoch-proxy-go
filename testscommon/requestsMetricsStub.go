package testscommon

// RequestsMetricsStub -
type RequestsMetricsStub struct {
	ProcessedResponseHandler func(alias string)
}

// ProcessedResponse -
func (stub *RequestsMetricsStub) ProcessedResponse(alias string) {
	if stub.ProcessedResponseHandler != nil {
		stub.ProcessedResponseHandler(alias)
	}
}

// IsInterfaceNil -
func (stub *RequestsMetricsStub) IsInterfaceNil() bool {
	return stub == nil
}
