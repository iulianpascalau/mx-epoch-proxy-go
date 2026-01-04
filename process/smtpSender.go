package process

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
)

const mimeHeaders = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

// BasicHTMLTemplate is a template for a basic email message. Can be used in testing
const BasicHTMLTemplate = `<!-- template.html -->
<!DOCTYPE html>
<html lang="en">
<body>
   {{.Body}}
</body>
</html>
`

type sendMailHandler func(host string, auth smtp.Auth, from string, to []string, msgBytes []byte) error

type smtpSender struct {
	smtpPort     int
	smtpHost     string
	from         string
	password     string
	sendMailFunc sendMailHandler
}

// ArgsSmtpSender represents the SMTP sender arguments used in the constructor function
type ArgsSmtpSender struct {
	SmtpPort int
	SmtpHost string
	From     string
	Password string
}

// NewSmtpSender creates a new SMTP email sender
func NewSmtpSender(args ArgsSmtpSender) *smtpSender {
	return &smtpSender{
		smtpPort:     args.SmtpPort,
		smtpHost:     args.SmtpHost,
		from:         args.From,
		password:     args.Password,
		sendMailFunc: sendMail,
	}
}

// SendEmail will try to send the email containing the subject and body
func (sender *smtpSender) SendEmail(to string, subject string, body any, htmlTemplate string) error {
	log.Debug("smtpSender.SendEmail sending messages", "to", to, "subject", subject, "body", body)

	auth := smtp.PlainAuth("", sender.from, sender.password, sender.smtpHost)
	msgBytes, err := createEmailBytes(body, subject, htmlTemplate)
	if err != nil {
		return err
	}

	err = sender.sendMailFunc(
		fmt.Sprintf("%s:%d", sender.smtpHost, sender.smtpPort),
		auth,
		sender.from,
		[]string{to},
		msgBytes,
	)
	if err != nil {
		return err
	}

	log.Debug("smtpSender.SendEmail: sent SMTP email")

	return nil
}

func sendMail(host string, auth smtp.Auth, from string, to []string, msgBytes []byte) error {
	return smtp.SendMail(host, auth, from, to, msgBytes)
}

func createEmailBytes(msg any, title string, htmlTemplate string) ([]byte, error) {
	var body bytes.Buffer

	mailTemplate := template.New("")
	_, err := mailTemplate.Parse(htmlTemplate)
	if err != nil {
		return nil, err
	}
	body.Write([]byte(fmt.Sprintf("Subject: %s \n%s\n\n", title, mimeHeaders)))

	err = mailTemplate.Execute(&body, msg)
	if err != nil {
		return nil, err
	}

	return body.Bytes(), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sender *smtpSender) IsInterfaceNil() bool {
	return sender == nil
}
