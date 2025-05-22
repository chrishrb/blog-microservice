package email

import (
	"bytes"
	_ "embed"
	"html/template"

	"github.com/chrishrb/blog-microservice/notification-service/channels"
	"golang.org/x/net/context"
	gomail "gopkg.in/mail.v2"
)

//go:embed templates/password-reset.tmpl
var passwordResetTemplate string

//go:embed templates/verify-account.tmpl
var verifyAccountTemplate string

type EmailChannel struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func NewEmailChannel(host string, port int, username, password, from string) (*EmailChannel, error) {
	return &EmailChannel{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}, nil
}

func (e *EmailChannel) SendPasswordReset(ctx context.Context, recipient string, variables channels.PasswordResetVariables) error {
	subject, body, err := e.parseEmailTemplate(passwordResetTemplate, variables)
	if err != nil {
		return err
	}
	return e.sendPlainTextEmail(recipient, subject, body)
}

func (e *EmailChannel) SendVerifyAccount(ctx context.Context, recipient string, variables channels.VerifyAccountVariables) error {
	subject, body, err := e.parseEmailTemplate(verifyAccountTemplate, variables)
	if err != nil {
		return err
	}
	return e.sendPlainTextEmail(recipient, subject, body)
}

func (e *EmailChannel) sendPlainTextEmail(recipient, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", e.from)
	m.SetHeader("To", recipient)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)
	dialer := gomail.NewDialer(e.host, e.port, e.username, e.password)
	return dialer.DialAndSend(m)
}

func (e *EmailChannel) parseEmailTemplate(templateStr string, variables any) (string, string, error) {
	t, err := template.New("email").Parse(templateStr)
	if err != nil {
		return "", "", err
	}

	// Parse subject
	var subjectTpl bytes.Buffer
	if err := t.ExecuteTemplate(&subjectTpl, "Subject", variables); err != nil {
		return "", "", err
	}
	subject := subjectTpl.String()

	// Parse body
	var bodyTpl bytes.Buffer
	if err := t.ExecuteTemplate(&bodyTpl, "Body", variables); err != nil {
		return "", "", err
	}
	body := bodyTpl.String()

	return subject, body, nil
}
