package testscommon

// PerformanceMonitorStub -
type PerformanceMonitorStub struct {
	AddPerformanceMetricAsyncHandler func(label string)
}

// AddPerformanceMetric -
func (stub *PerformanceMonitorStub) AddPerformanceMetricAsync(label string) {
	if stub.AddPerformanceMetricAsyncHandler != nil {
		stub.AddPerformanceMetricAsyncHandler(label)
	}
}

// IsInterfaceNil -
func (stub *PerformanceMonitorStub) IsInterfaceNil() bool {
	return stub == nil
}
