package process

import (
	"errors"
	"fmt"
	"net/smtp"
	"os"
	"testing"

	logger "github.com/multiversx/mx-chain-logger-go"
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
		err := sender.SendEmail("to@email.com", "subject", "body")
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
		err := sender.SendEmail(
			"to@email.com",
			"Activate your account for the MultiversX Deep History Access",
			"Please click the link below to activate your account:\n   <b><a href=\"http://localhost:8080/api/activate?token=11223344556677889900\">Activate with token 11223344556677889900</a></b>")
		assert.Nil(t, err)
		assert.Equal(t, expectedBody, string(sentMsgBytes))
	})
}

func TestSmtpSender_FunctionalTest(t *testing.T) {
	smtpTo := os.Getenv("SMTP_TO")
	smtpFrom := os.Getenv("SMTP_FROM")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	if len(smtpTo) == 0 || len(smtpFrom) == 0 || len(smtpPassword) == 0 {
		t.Skip("this is a functional test, will need real credentials. Please define your environment variables SMTP_TO, SMTP_FROM and SMTP_PASSWORD so this test can work")
	}

	_ = logger.SetLogLevel("*:DEBUG")

	args := ArgsSmtpSender{
		SmtpPort: 587,
		SmtpHost: "smtp.gmail.com",
		From:     smtpFrom,
		Password: smtpPassword,
	}

	sender := NewSmtpSender(args)
	err := sender.SendEmail(
		smtpTo,
		"Activate your account for the MultiversX Deep History Access",
		"In order to activate your newly registered account for the MultiversX Deep History Access you need to click on the link below:<br><br><b><a href=\"http://localhost:8080/api/activate?token=EMAILTOKEN11223344556677889900\">Activate with token EMAILTOKEN11223344556677889900</a></b>",
	)
	assert.Nil(t, err)
}
