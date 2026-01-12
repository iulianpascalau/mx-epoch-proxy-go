package testscommon

// EmailSenderStub is a stub for EmailSender
type EmailSenderStub struct {
	SendEmailHandler      func(to string, subject string, body any, htmlTemplate string) error
	IsInterfaceNilHandler func() bool
}

// SendEmail -
func (stub *EmailSenderStub) SendEmail(to string, subject string, body any, htmlTemplate string) error {
	if stub.SendEmailHandler != nil {
		return stub.SendEmailHandler(to, subject, body, htmlTemplate)
	}
	return nil
}

// IsInterfaceNil -
func (stub *EmailSenderStub) IsInterfaceNil() bool {
	return stub == nil
}
