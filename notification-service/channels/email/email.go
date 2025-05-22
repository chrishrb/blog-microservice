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
	t, err := template.New("email").Parse(passwordResetTemplate)
	if err != nil {
		return err
	}

	// Parse subject
	var subjectTpl bytes.Buffer
	if err := t.ExecuteTemplate(&subjectTpl, "Subject", variables); err != nil {
		return err
	}
	subject := subjectTpl.String()

	// Parse body
	var tpl bytes.Buffer
	if err := t.ExecuteTemplate(&tpl, "Body", variables); err != nil {
		return err
	}
	body := tpl.String()

	// Create a new message
	m := gomail.NewMessage()

	// Set email headers
	m.SetHeader("From", e.from)
	m.SetHeader("To", recipient)
	m.SetHeader("Subject", subject)

	// Set email body
	m.SetBody("text/plain", body)

	dialer := gomail.NewDialer(e.host, e.port, e.username, e.password)
	return dialer.DialAndSend(m)
}
