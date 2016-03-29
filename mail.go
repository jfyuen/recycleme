package recycleme

import (
	"errors"
	"fmt"
	"net/smtp"
	"strings"
)

type Mailer func(subject, body string) error

type emailConfig struct {
	host, recipient, username, password, sender string
	auth                                        smtp.Auth
}

func NewEmailConfig(host, recipient, username, password string) (emailConfig, error) {
	e := emailConfig{host: host, recipient: recipient, username: username, password: password}
	if e.host == "" || e.recipient == "" || e.username == "" || e.password == "" {
		return e, errors.New("invalid config parameters")
	}
	if !strings.Contains(e.host, ":") {
		return e, fmt.Errorf("no port specified for host %v", e.host)
	}
	e.auth = smtp.PlainAuth("", e.username, e.password, strings.Split(e.host, ":")[0])
	e.sender = "admin@howtorecycle.me"
	return e, nil
}

func (e emailConfig) createMessage(subject, body string) string {
	return fmt.Sprintf("From: %s\nTo: %s\nSubject: [RECYCLEME] %s\n\n%s", e.sender, e.recipient, subject, body)

}

func (e emailConfig) SendMail(subject, body string) error {
	msg := e.createMessage(subject, body)
	return smtp.SendMail(e.host, e.auth, e.sender, []string{e.recipient}, []byte(msg))
}
