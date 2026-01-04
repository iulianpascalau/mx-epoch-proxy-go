package process

import (
	"errors"
	"fmt"
	"html/template"
	"net/smtp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSmtpSender(t *testing.T) {
	t.Parallel()

	sender := NewSmtpSender(ArgsSmtpSender{})
	assert.NotNil(t, sender)
	assert.False(t, sender.IsInterfaceNil())
}

func TestSmtpSender_SendEmail(t *testing.T) {
	testArgs := ArgsSmtpSender{
		SmtpPort: 37,
		SmtpHost: "host.email.com",
		From:     "from@email.com",
		Password: "pass",
	}
	expectedErr := errors.New("expected error")

	t.Run("send mail function fails, should error", func(t *testing.T) {
		t.Parallel()

		sender := NewSmtpSender(testArgs)
		sender.sendMailFunc = func(host string, auth smtp.Auth, from string, to []string, msgBytes []byte) error {
			return expectedErr
		}
		err := sender.SendEmail(
			"to@email.com",
			"subject",
			struct {
				Body template.HTML
			}{
				Body: "body",
			},
			BasicHTMLTemplate)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})
	t.Run("sending message should work", func(t *testing.T) {
		t.Parallel()

		expectedBody := `Subject: Activate your account for the MultiversX Deep History Access 
MIME-version: 1.0;
Content-Type: text/html; charset="UTF-8";




<!DOCTYPE html>
<html lang="en">
<body>
   Please click the link below to activate your account:
   <b><a href="http://localhost:8080/api/activate?token=11223344556677889900">Activate with token 11223344556677889900</a></b>
</body>
</html>
`
		var sentMsgBytes []byte
		sender := NewSmtpSender(testArgs)
		sender.sendMailFunc = func(host string, auth smtp.Auth, from string, to []string, msgBytes []byte) error {
			assert.Equal(t, fmt.Sprintf("%s:%d", testArgs.SmtpHost, testArgs.SmtpPort), host)
			assert.Equal(t, testArgs.From, from)
			assert.Equal(t, []string{"to@email.com"}, to)
			sentMsgBytes = msgBytes

			return nil
		}

		message := "Please click the link below to activate your account:\n   <b><a href=\"http://localhost:8080/api/activate?token=11223344556677889900\">Activate with token 11223344556677889900</a></b>"
		err := sender.SendEmail(
			"to@email.com",
			"Activate your account for the MultiversX Deep History Access",
			struct {
				Body template.HTML
			}{
				Body: template.HTML(message),
			},
			BasicHTMLTemplate,
		)
		assert.Nil(t, err)
		assert.Equal(t, expectedBody, string(sentMsgBytes))
	})
}
