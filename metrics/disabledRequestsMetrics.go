package metrics

// DisabledRequestsMetrics is the disabled implementation of the requests metrics
type DisabledRequestsMetrics struct {
}

// ProcessedResponse does nothing
func (metrics *DisabledRequestsMetrics) ProcessedResponse(_ string) {
}

// GetAllKeyValues will return an empty slice
func (metrics *DisabledRequestsMetrics) GetAllKeyValues() []string {
	return make([]string, 0)
}

// IsInterfaceNil returns true if the value under the interface is nil
func (metrics *DisabledRequestsMetrics) IsInterfaceNil() bool {
	return metrics == nil
}
