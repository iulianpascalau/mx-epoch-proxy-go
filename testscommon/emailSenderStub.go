package testscommon

// EmailSenderStub is a stub for EmailSender
type EmailSenderStub struct {
	SendEmailHandler      func(to string, subject string, body string) error
	IsInterfaceNilHandler func() bool
}

// SendEmail -
func (stub *EmailSenderStub) SendEmail(to string, subject string, body string) error {
	if stub.SendEmailHandler != nil {
		return stub.SendEmailHandler(to, subject, body)
	}
	return nil
}

// IsInterfaceNil -
func (stub *EmailSenderStub) IsInterfaceNil() bool {
	if stub.IsInterfaceNilHandler != nil {
		return stub.IsInterfaceNilHandler()
	}
	return stub == nil
}
