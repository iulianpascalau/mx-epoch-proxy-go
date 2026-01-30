package framework

// MockEmailSender -
type MockEmailSender struct {
	LastTo   string
	LastBody any
}

// SendEmail -
func (m *MockEmailSender) SendEmail(to string, _ string, body any, _ string) error {
	m.LastTo = to
	m.LastBody = body

	return nil
}

// IsInterfaceNil -
func (m *MockEmailSender) IsInterfaceNil() bool {
	return m == nil
}
