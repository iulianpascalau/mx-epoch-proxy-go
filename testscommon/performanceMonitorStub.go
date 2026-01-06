package testscommon

// PerformanceMonitorStub -
type PerformanceMonitorStub struct {
	AddPerformanceMetricCalled func(label string) error
}

// AddPerformanceMetric -
func (stub *PerformanceMonitorStub) AddPerformanceMetric(label string) error {
	if stub.AddPerformanceMetricCalled != nil {
		return stub.AddPerformanceMetricCalled(label)
	}
	return nil
}

// IsInterfaceNil -
func (stub *PerformanceMonitorStub) IsInterfaceNil() bool {
	return stub == nil
}
