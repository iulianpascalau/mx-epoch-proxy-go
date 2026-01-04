package integrationTests

import (
	"html/template"
	"os"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/common"
	"github.com/iulianpascalau/mx-epoch-proxy-go/process"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendRealEmailWithTemplate(t *testing.T) {
	smtpTo := os.Getenv("SMTP_TO")
	smtpFrom := os.Getenv("SMTP_FROM")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	if len(smtpTo) == 0 || len(smtpFrom) == 0 || len(smtpPassword) == 0 {
		t.Skip("this is a functional test, will need real credentials. Please define your environment variables SMTP_TO, SMTP_FROM and SMTP_PASSWORD so this test can work")
	}

	_ = logger.SetLogLevel("*:DEBUG")

	args := process.ArgsSmtpSender{
		SmtpPort: 587,
		SmtpHost: "smtp.gmail.com",
		From:     smtpFrom,
		Password: smtpPassword,
	}

	token := "EMAILTOKEN" + common.GenerateKey()
	activationUrl := "http://localhost:8080/api/activate?token=" + token
	swaggerUrl := "http://localhost:8080/"

	bodyObject := struct {
		ActivationURL template.HTML
		SwaggerURL    template.HTML
	}{
		ActivationURL: template.HTML(activationUrl),
		SwaggerURL:    template.HTML(swaggerUrl),
	}

	emailTemplate, err := os.ReadFile("./../activation_email.html")
	require.Nil(t, err)

	sender := process.NewSmtpSender(args)
	err = sender.SendEmail(
		smtpTo,
		"Activate your account for the Deep History on MultiversX",
		bodyObject,
		string(emailTemplate),
	)
	assert.Nil(t, err)
}
